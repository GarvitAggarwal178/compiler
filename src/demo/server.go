package demo

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"dlc/src/inspect"
	"dlc/src/ir"
	"dlc/src/sema"
)

type Config struct {
	Repo, Compiler, CC, Oracle, TempDir string
	Timeout                             time.Duration
	MaxRuns                             int
	MaxOutputBytes                      int64
}
type Tool struct {
	Available bool   `json:"available"`
	Path      string `json:"path,omitempty"`
	Detail    string `json:"detail,omitempty"`
}
type Example struct {
	ID          string            `json:"id"`
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Source      string            `json:"source"`
	Facts       map[string]string `json:"facts"`
}
type manifestExample struct {
	ID            string            `json:"id"`
	Title         string            `json:"title"`
	Description   string            `json:"description"`
	SourcePath    string            `json:"source_path"`
	FactPaths     map[string]string `json:"fact_paths"`
	FactDirectory string            `json:"fact_directory"`
}
type Request struct {
	Source      string            `json:"source"`
	Facts       map[string]string `json:"facts"`
	Mode        string            `json:"mode"`
	InspectOnly bool              `json:"inspect_only"`
}
type InputFile struct {
	Name   string `json:"name"`
	SHA256 string `json:"sha256"`
	Bytes  int    `json:"bytes"`
}
type WorkCounter struct {
	Relation        string         `json:"relation"`
	Derived         int            `json:"derived"`
	Seed            int            `json:"seed"`
	SCCLocalBuckets map[string]int `json:"scc_local_buckets,omitempty"`
}
type Evaluation struct {
	Mode         string        `json:"mode"`
	Engine       string        `json:"engine"`
	Materialized int           `json:"materialized"`
	Attempts     int64         `json:"attempts"`
	Relations    []WorkCounter `json:"relations"`
	Outputs      []OutputFile  `json:"outputs"`
	Comparison   Comparison    `json:"comparison"`
}
type Result struct {
	RunID        string          `json:"run_id"`
	Status       string          `json:"status"`
	Message      string          `json:"message,omitempty"`
	SourceSHA256 string          `json:"source_sha256"`
	InputSHA256  string          `json:"input_sha256"`
	Mode         string          `json:"mode"`
	Engine       string          `json:"engine"`
	Inputs       []InputFile     `json:"inputs"`
	Inspection   *inspect.Report `json:"inspection,omitempty"`
	Commands     []Command       `json:"commands"`
	Evaluations  []Evaluation    `json:"evaluations"`
	Comparison   *Comparison     `json:"comparison,omitempty"`
}
type NativeResult struct {
	RunID        string       `json:"run_id"`
	Status       string       `json:"status"`
	Message      string       `json:"message,omitempty"`
	SourceSHA256 string       `json:"source_sha256"`
	InputSHA256  string       `json:"input_sha256"`
	Mode         string       `json:"mode"`
	GeneratedC   string       `json:"generated_c,omitempty"`
	Commands     []Command    `json:"commands"`
	Outputs      []OutputFile `json:"outputs"`
	Comparison   *Comparison  `json:"comparison,omitempty"`
}
type storedRun struct {
	id, dir, mode, sourceHash, inputHash string
	relations                            []sema.RelationInfo
	expected                             []string
	baseline                             OutputSets
	ready                                bool
	active                               int
	op                                   sync.Mutex
}
type Server struct {
	cfg      Config
	tools    map[string]Tool
	examples []Example
	root     string
	mux      *http.ServeMux
	ctx      context.Context
	cancel   context.CancelFunc
	mu       sync.Mutex
	runs     map[string]*storedRun
	order    []string
	serial   uint64
	slots    chan struct{}
	wg       sync.WaitGroup
}

func discover(name string) Tool {
	p, e := exec.LookPath(name)
	if e != nil {
		return Tool{Detail: e.Error()}
	}
	p, _ = filepath.Abs(p)
	return Tool{Available: true, Path: p}
}
func New(cfg Config, assets fs.FS) (*Server, error) {
	if cfg.Timeout == 0 {
		cfg.Timeout = 3 * time.Second
	}
	if cfg.Timeout < time.Millisecond || cfg.Timeout > 30*time.Second {
		return nil, fmt.Errorf("timeout must be between 1ms and 30s")
	}
	if cfg.MaxRuns == 0 {
		cfg.MaxRuns = 8
	}
	if cfg.MaxRuns < 1 {
		return nil, fmt.Errorf("max-runs must be positive")
	}
	if cfg.MaxOutputBytes == 0 {
		cfg.MaxOutputBytes = 2 << 20
	}
	cfg.Repo, _ = filepath.Abs(cfg.Repo)
	if cfg.Compiler == "" {
		cfg.Compiler = filepath.Join(cfg.Repo, "bin/dlc")
	}
	if cfg.CC == "" {
		cfg.CC = "cc"
	}
	if cfg.Oracle == "" {
		cfg.Oracle = "souffle"
	}
	root, err := os.MkdirTemp(cfg.TempDir, "dlc-demo-")
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(context.Background())
	s := &Server{cfg: cfg, root: root, ctx: ctx, cancel: cancel, runs: map[string]*storedRun{}, slots: make(chan struct{}, 2), mux: http.NewServeMux()}
	s.tools = map[string]Tool{"compiler": discover(cfg.Compiler), "go": discover("go"), "cc": discover(cfg.CC), "oracle": discover(cfg.Oracle)}
	if err = s.loadExamples(); err != nil {
		s.Close()
		return nil, err
	}
	s.mux.HandleFunc("GET /api/preflight", s.preflight)
	s.mux.HandleFunc("GET /api/examples", func(w http.ResponseWriter, r *http.Request) { writeJSON(w, s.examples) })
	s.mux.HandleFunc("POST /api/run", s.run)
	s.mux.HandleFunc("POST /api/native/{id}", s.native)
	s.mux.HandleFunc("POST /api/oracle/{id}", s.oracle)
	if assets != nil {
		s.mux.Handle("GET /", http.FileServer(http.FS(assets)))
	}
	return s, nil
}
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Origin") != "" && r.Header.Get("Origin") != "http://"+r.Host {
		http.Error(w, "cross-origin request rejected", http.StatusForbidden)
		return
	}
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self'; img-src 'self' data:; connect-src 'self'; object-src 'none'; frame-ancestors 'none'")
	s.mux.ServeHTTP(w, r)
}
func (s *Server) Close() { s.cancel(); s.wg.Wait(); _ = os.RemoveAll(s.root) }
func (s *Server) preflight(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]any{"tools": s.tools, "timeout_ms": s.cfg.Timeout.Milliseconds(), "max_runs": s.cfg.MaxRuns, "source_limit": 65536, "facts_limit": 2 << 20, "output_limit": s.cfg.MaxOutputBytes, "modes": []string{"guarded", "guarded-compact"}})
}
func (s *Server) loadExamples() error {
	b, e := os.ReadFile(filepath.Join(s.cfg.Repo, "fixtures/demo/manifest.json"))
	if e != nil {
		return e
	}
	var manifest []manifestExample
	if e = json.Unmarshal(b, &manifest); e != nil {
		return e
	}
	for _, m := range manifest {
		source, e := s.repoFile(m.SourcePath)
		if e != nil {
			return e
		}
		x := Example{ID: m.ID, Title: m.Title, Description: m.Description, Source: string(source), Facts: map[string]string{}}
		paths := m.FactPaths
		if paths == nil {
			paths = map[string]string{}
		}
		if m.FactDirectory != "" {
			dir := filepath.Join(s.cfg.Repo, m.FactDirectory)
			entries, e := os.ReadDir(dir)
			if e != nil {
				return e
			}
			for _, f := range entries {
				if strings.HasSuffix(f.Name(), ".facts") {
					paths[strings.TrimSuffix(f.Name(), ".facts")] = filepath.Join(m.FactDirectory, f.Name())
				}
			}
		}
		for name, p := range paths {
			b, e := s.repoFile(p)
			if e != nil {
				return e
			}
			x.Facts[name] = string(b)
		}
		s.examples = append(s.examples, x)
	}
	return nil
}
func (s *Server) repoFile(p string) ([]byte, error) {
	if filepath.IsAbs(p) || p == ".." || strings.HasPrefix(filepath.Clean(p), ".."+string(filepath.Separator)) {
		return nil, fmt.Errorf("example path outside repository")
	}
	return os.ReadFile(filepath.Join(s.cfg.Repo, p))
}
func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(v)
}
func decode(w http.ResponseWriter, r *http.Request, v any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 3<<20)
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	if err := d.Decode(v); err != nil {
		http.Error(w, err.Error(), 400)
		return false
	}
	var extra any
	if d.Decode(&extra) != io.EOF {
		http.Error(w, "expected one JSON request", 400)
		return false
	}
	return true
}
func (s *Server) begin(w http.ResponseWriter, r *http.Request) (context.Context, func(), bool) {
	select {
	case s.slots <- struct{}{}:
	default:
		http.Error(w, "two runs already active; retry after cancellation", 429)
		return nil, nil, false
	}
	s.wg.Add(1)
	ctx, cancel := context.WithCancel(r.Context())
	stop := context.AfterFunc(s.ctx, cancel)
	return ctx, func() { stop(); cancel(); <-s.slots; s.wg.Done() }, true
}
func (s *Server) command(ctx context.Context, stage string, argv []string) Command {
	c, cancel := context.WithTimeout(ctx, s.cfg.Timeout)
	defer cancel()
	return RunCommand(c, stage, argv, 8<<20)
}
func (s *Server) newRun(mode string) (*storedRun, error) {
	dir, err := os.MkdirTemp(s.root, "run-")
	if err != nil {
		return nil, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.serial++
	id := strconv.FormatUint(s.serial, 10) + "-" + filepath.Base(dir)
	run := &storedRun{id: id, dir: dir, mode: mode, active: 1}
	s.runs[id] = run
	s.order = append(s.order, id)
	s.prune()
	return run, nil
}
func (s *Server) prune() {
	for len(s.order) > s.cfg.MaxRuns {
		removed := false
		for i, id := range s.order {
			r := s.runs[id]
			if r.active == 0 {
				_ = os.RemoveAll(r.dir)
				delete(s.runs, id)
				s.order = append(s.order[:i], s.order[i+1:]...)
				removed = true
				break
			}
		}
		if !removed {
			return
		}
	}
}
func (s *Server) release(run *storedRun) { s.mu.Lock(); defer s.mu.Unlock(); run.active--; s.prune() }
func hash(b []byte) string               { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }
func snapshotHash(source string, facts map[string]string) string {
	h := sha256.New()
	part := func(b []byte) { _ = binary.Write(h, binary.BigEndian, uint64(len(b))); _, _ = h.Write(b) }
	part([]byte(source))
	for _, n := range sortedNames(facts) {
		part([]byte(n))
		part([]byte(facts[n]))
	}
	return hex.EncodeToString(h.Sum(nil))
}
func (s *Server) run(w http.ResponseWriter, r *http.Request) {
	var req Request
	if !decode(w, r, &req) {
		return
	}
	if req.Mode == "" {
		req.Mode = "guarded"
	}
	if req.Mode != "guarded" && req.Mode != "guarded-compact" {
		http.Error(w, "unsupported demo mode", 400)
		return
	}
	if len(req.Source) > 65536 {
		http.Error(w, "source exceeds 65536 bytes", 413)
		return
	}
	total := 0
	for _, f := range req.Facts {
		total += len(f)
		if len(f) > 512<<10 {
			http.Error(w, "one facts file exceeds 512 KiB", 413)
			return
		}
	}
	if total > 2<<20 {
		http.Error(w, "facts exceed 2 MiB", 413)
		return
	}
	ctx, end, ok := s.begin(w, r)
	if !ok {
		return
	}
	defer end()
	run, err := s.newRun(req.Mode)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer s.release(run)
	result := Result{RunID: run.id, Status: "error", Mode: req.Mode, Engine: "Go semi-naive", SourceSHA256: hash([]byte(req.Source)), InputSHA256: snapshotHash(req.Source, req.Facts), Inputs: []InputFile{}, Commands: []Command{}, Evaluations: []Evaluation{}}
	finish := func() { writeJSON(w, result) }
	source := filepath.Join(run.dir, "source.dl")
	if err = os.WriteFile(source, []byte(req.Source), 0600); err != nil {
		result.Message = err.Error()
		finish()
		return
	}
	cmd := s.command(ctx, "inspect", []string{s.cfg.Compiler, "inspect", source, "--transformer=" + req.Mode})
	result.Commands = append(result.Commands, cmd)
	var report inspect.Report
	if err = json.Unmarshal([]byte(cmd.Stdout), &report); err == nil {
		result.Inspection = &report
	}
	if cmd.Status != "ok" {
		result.Status = cmd.Status
		if result.Inspection != nil {
			result.Status = result.Inspection.Status
		}
		if result.Status == "" {
			result.Status = cmd.Status
		}
		result.Message = cmd.Stderr
		finish()
		return
	}
	if result.Inspection == nil || report.Status != "ok" {
		result.Message = "compiler inspection returned no valid report"
		finish()
		return
	}
	run.relations = report.Source.Relations
	facts := map[string]string{}
	for _, rel := range run.relations {
		if rel.Input {
			facts[rel.Name] = ""
		}
		if rel.Output {
			run.expected = append(run.expected, rel.Name+".csv")
		}
	}
	for name, data := range req.Facts {
		if _, ok := facts[name]; !ok {
			result.Status = "invalid_input"
			result.Message = "facts submitted for undeclared input " + name
			finish()
			return
		}
		facts[name] = data
	}
	factdir := filepath.Join(run.dir, "facts")
	if err = os.Mkdir(factdir, 0700); err != nil {
		result.Message = err.Error()
		finish()
		return
	}
	for _, name := range sortedNames(facts) {
		data := []byte(facts[name])
		if err = os.WriteFile(filepath.Join(factdir, name+".facts"), data, 0600); err != nil {
			result.Message = err.Error()
			finish()
			return
		}
		result.Inputs = append(result.Inputs, InputFile{Name: name + ".facts", SHA256: hash(data), Bytes: len(data)})
	}
	result.InputSHA256 = snapshotHash(req.Source, facts)
	run.sourceHash = result.SourceSHA256
	run.inputHash = result.InputSHA256
	if req.InspectOnly {
		result.Status = "inspected"
		finish()
		return
	}
	var baseline OutputSets
	for _, variant := range []struct{ mode, engine, sub string }{{"passthrough", "semi-naive", "run-seminaive"}, {req.Mode, "semi-naive", "run-seminaive"}, {"passthrough", "naive", "run"}, {req.Mode, "naive", "run"}} {
		out := filepath.Join(run.dir, variant.mode+"-"+variant.engine)
		cmd := s.command(ctx, variant.mode+"/"+variant.engine, []string{s.cfg.Compiler, variant.sub, source, factdir, out, "--transformer=" + variant.mode})
		result.Commands = append(result.Commands, cmd)
		if cmd.Status != "ok" {
			result.Status = cmd.Status
			result.Message = "execution failed at " + cmd.Stage
			finish()
			return
		}
		actual, err := ReadOutputs(out, run.expected, s.cfg.MaxOutputBytes)
		if err != nil {
			result.Status = "output_error"
			result.Message = err.Error()
			finish()
			return
		}
		if baseline == nil {
			baseline = actual
			run.baseline = baseline
		}
		metrics, err := readMetrics(out, cmd.Stdout)
		if err != nil {
			result.Message = err.Error()
			finish()
			return
		}
		comparison := Compare(actual, baseline)
		result.Evaluations = append(result.Evaluations, Evaluation{Mode: variant.mode, Engine: variant.engine, Materialized: metrics.total, Attempts: metrics.attempts, Relations: metrics.relations, Outputs: displayOutputs(actual, run.relations), Comparison: comparison})
		if comparison.Available && !comparison.Equal {
			result.Status = "mismatch"
			result.Comparison = &comparison
			finish()
			return
		}
	}
	c := Compare(baseline, baseline)
	result.Comparison = &c
	result.Status = "ok"
	s.mu.Lock()
	run.ready = true
	s.mu.Unlock()
	finish()
}

type metrics struct {
	total     int
	attempts  int64
	relations []WorkCounter
}

func readMetrics(dir, stdout string) (metrics, error) {
	var doc ir.ProfileDoc
	var m metrics
	raw, err := os.ReadFile(filepath.Join(dir, "profile.json"))
	if err != nil {
		return m, err
	}
	if err = json.Unmarshal(raw, &doc); err != nil {
		return m, err
	}
	var response struct {
		Status   string `json:"status"`
		Attempts int64  `json:"derivation_attempts"`
	}
	if err = json.Unmarshal([]byte(stdout), &response); err != nil || response.Status != "ok" {
		return m, fmt.Errorf("missing execution counters: %v", err)
	}
	m.attempts = response.Attempts
	m.relations = []WorkCounter{}
	for _, name := range sortedNames(doc.Root.Program.Relation) {
		r := doc.Root.Program.Relation[name]
		c := WorkCounter{Relation: name, Seed: r.NumTuples, Derived: r.NumTuples, SCCLocalBuckets: map[string]int{}}
		for bucket, values := range r.Iteration {
			c.Derived += values.NumTuples
			c.SCCLocalBuckets[bucket] = values.NumTuples
		}
		m.total += c.Derived
		m.relations = append(m.relations, c)
	}
	return m, nil
}
func (s *Server) acquireRun(w http.ResponseWriter, id string) (*storedRun, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	run, ok := s.runs[id]
	if !ok {
		http.Error(w, "run expired; run the current inputs again", 410)
		return nil, false
	}
	if !run.ready {
		http.Error(w, "successful Go run required", 409)
		return nil, false
	}
	run.active++
	return run, true
}
func (s *Server) native(w http.ResponseWriter, r *http.Request) { s.external(w, r, false) }
func (s *Server) oracle(w http.ResponseWriter, r *http.Request) { s.external(w, r, true) }
func (s *Server) external(w http.ResponseWriter, r *http.Request, oracle bool) {
	ctx, end, ok := s.begin(w, r)
	if !ok {
		return
	}
	defer end()
	run, ok := s.acquireRun(w, r.PathValue("id"))
	if !ok {
		return
	}
	defer s.release(run)
	run.op.Lock()
	defer run.op.Unlock()
	result := NativeResult{RunID: run.id, Status: "error", Mode: run.mode, SourceSHA256: run.sourceHash, InputSHA256: run.inputHash, Commands: []Command{}, Outputs: []OutputFile{}}
	finish := func() { writeJSON(w, result) }
	dir, err := os.MkdirTemp(run.dir, "external-")
	if err != nil {
		result.Message = err.Error()
		finish()
		return
	}
	source, facts := filepath.Join(run.dir, "source.dl"), filepath.Join(run.dir, "facts")
	out := filepath.Join(dir, "output")
	tool := "cc"
	if oracle {
		tool = "oracle"
	}
	if !s.tools[tool].Available {
		result.Status = "unavailable"
		result.Message = tool + " unavailable: " + s.tools[tool].Detail
		finish()
		return
	}
	commands := [][]string{}
	stages := []string{}
	if oracle {
		_ = os.Mkdir(out, 0700)
		commands = append(commands, []string{s.tools["oracle"].Path, "-F", facts, "-D", out, source})
		stages = append(stages, "Souffle answer verification")
	} else {
		cfile, native := filepath.Join(dir, "query.c"), filepath.Join(dir, "query")
		commands = [][]string{{s.cfg.Compiler, "codegen", source, cfile, "--transformer=" + run.mode}, {s.tools["cc"].Path, "-std=c11", "-O2", cfile, "-o", native}, {native, facts, out}}
		stages = []string{"generate C", "compile C", "execute native C"}
	}
	for i, argv := range commands {
		cmd := s.command(ctx, stages[i], argv)
		result.Commands = append(result.Commands, cmd)
		if !oracle && i == 0 && cmd.Status == "ok" {
			b, err := os.ReadFile(argv[3])
			if err != nil {
				result.Message = err.Error()
				finish()
				return
			}
			if int64(len(b)) > s.cfg.MaxOutputBytes {
				result.Status = "output_limit"
				result.Message = "generated C exceeds display limit"
				finish()
				return
			}
			result.GeneratedC = string(b)
		}
		if cmd.Status != "ok" {
			result.Status = cmd.Status
			result.Message = "failed at " + stages[i]
			finish()
			return
		}
	}
	actual, err := ReadOutputs(out, run.expected, s.cfg.MaxOutputBytes)
	if err != nil {
		result.Status = "output_error"
		result.Message = err.Error()
		finish()
		return
	}
	result.Outputs = displayOutputs(actual, run.relations)
	comparison := Compare(actual, run.baseline)
	result.Comparison = &comparison
	result.Status = "ok"
	if comparison.Available && !comparison.Equal {
		result.Status = "mismatch"
	}
	finish()
}

// CheckLoopback prevents accidental exposure of native execution over a LAN.
func CheckLoopback(addr string) error {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return err
	}
	if host == "localhost" {
		return nil
	}
	ip := net.ParseIP(host)
	if ip == nil || !ip.IsLoopback() {
		return fmt.Errorf("demo address must use loopback")
	}
	return nil
}

func (s *Server) ToolStatus() map[string]Tool {
	out := map[string]Tool{}
	for k, v := range s.tools {
		out[k] = v
	}
	return out
}
