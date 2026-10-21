package demo_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"syscall"
	"testing"
	"time"

	"dlc/src/demo"
)

var repo, compiler string

func TestMain(m *testing.M) {
	repo, _ = filepath.Abs("../..")
	dir, e := os.MkdirTemp("", "dlc-demo-tests-")
	if e != nil {
		panic(e)
	}
	compiler = filepath.Join(dir, "dlc")
	cmd := exec.Command("go", "build", "-buildvcs=false", "-o", compiler, "./src/cmd/dlc")
	cmd.Dir = repo
	if b, e := cmd.CombinedOutput(); e != nil {
		fmt.Fprintln(os.Stderr, string(b), e)
		os.RemoveAll(dir)
		os.Exit(1)
	}
	code := m.Run()
	_ = os.RemoveAll(dir)
	os.Exit(code)
}
func server(t *testing.T, c demo.Config) (*httptest.Server, *demo.Server) {
	t.Helper()
	c.Repo = repo
	c.Compiler = compiler
	if c.Timeout == 0 {
		c.Timeout = time.Second
	}
	s, e := demo.New(c, nil)
	if e != nil {
		t.Fatal(e)
	}
	h := httptest.NewServer(s)
	t.Cleanup(func() { h.Close(); s.Close() })
	return h, s
}
func post(t *testing.T, h *httptest.Server, path string, v any, target any) int {
	t.Helper()
	b, _ := json.Marshal(v)
	r, e := http.Post(h.URL+path, "application/json", bytes.NewReader(b))
	if e != nil {
		t.Fatal(e)
	}
	defer r.Body.Close()
	body, _ := io.ReadAll(r.Body)
	if target != nil && r.StatusCode == 200 {
		reflect.ValueOf(target).Elem().SetZero()
		if e = json.Unmarshal(body, target); e != nil {
			t.Fatalf("decode: %v %s", e, body)
		}
	}
	return r.StatusCode
}
func graphSource(t *testing.T) (string, map[string]string) {
	t.Helper()
	b, e := os.ReadFile(filepath.Join(repo, "fixtures/demo/reachability.dl"))
	if e != nil {
		t.Fatal(e)
	}
	facts := map[string]string{}
	for _, n := range []string{"edge", "node"} {
		b, e := os.ReadFile(filepath.Join(repo, "fixtures/demo/facts", n+".facts"))
		if e != nil {
			t.Fatal(e)
		}
		facts[n] = string(b)
	}
	return string(b), facts
}
func TestActualSnapshotsNativeAndIndependentOracle(t *testing.T) {
	h, _ := server(t, demo.Config{Timeout: 3 * time.Second})
	source, facts := graphSource(t)
	var r demo.Result
	if post(t, h, "/api/run", demo.Request{Source: source, Facts: facts, Mode: "guarded"}, &r) != 200 || r.Status != "ok" {
		t.Fatalf("run: %+v", r)
	}
	if !r.Comparison.Equal || len(r.Evaluations) != 4 {
		t.Fatal(r)
	}
	if r.Evaluations[0].Materialized != 69 || r.Evaluations[1].Materialized != 38 || r.Evaluations[0].Attempts != 69 || r.Evaluations[1].Attempts != 46 {
		t.Fatal("unexpected measured graph counters", r.Evaluations)
	}
	originalID, originalHash := r.RunID, r.InputSHA256
	facts["edge"] += "4\t5\n"
	var edited demo.Result
	post(t, h, "/api/run", demo.Request{Source: source, Facts: facts, Mode: "guarded"}, &edited)
	if edited.Status != "ok" || edited.InputSHA256 == originalHash {
		t.Fatal(edited)
	}
	if len(edited.Evaluations[1].Outputs[0].Rows) != 1 {
		t.Fatal("edited facts not executed")
	}
	var native demo.NativeResult
	post(t, h, "/api/native/"+originalID, struct{}{}, &native)
	if native.Status != "ok" || !native.Comparison.Equal || native.InputSHA256 != originalHash || !strings.Contains(native.GeneratedC, "source SHA-256") || !strings.Contains(native.GeneratedC, "source rule_") {
		t.Fatalf("native: %+v", native)
	}
	if _, e := exec.LookPath("souffle"); e == nil {
		var oracle demo.NativeResult
		post(t, h, "/api/oracle/"+originalID, struct{}{}, &oracle)
		if oracle.Status != "ok" || !oracle.Comparison.Equal {
			t.Fatal(oracle)
		}
	} else {
		t.Log("Souffle unavailable; separate make oracle is required")
	}
}
func TestMissingToolsErrorsAndNoOutputComparison(t *testing.T) {
	h, _ := server(t, demo.Config{CC: "/missing/dlc-cc", Oracle: "/missing/dlc-oracle"})
	var r demo.Result
	post(t, h, "/api/run", demo.Request{Source: ".decl out(x:number) .output out out(7).", Mode: "guarded"}, &r)
	for _, path := range []string{"/api/native/", "/api/oracle/"} {
		var result demo.NativeResult
		post(t, h, path+r.RunID, struct{}{}, &result)
		if result.Status != "unavailable" || result.Comparison != nil {
			t.Fatal(result)
		}
	}
	post(t, h, "/api/run", demo.Request{Source: ".decl out(x:number) .output out out(x).", Mode: "guarded"}, &r)
	if r.Status != "rejected" || r.Comparison != nil || len(r.Evaluations) > 0 {
		t.Fatal("stale/false success", r)
	}
	post(t, h, "/api/run", demo.Request{Source: ".decl out(x:number) out(1).", Mode: "guarded"}, &r)
	if r.Status != "ok" || r.Comparison.Available {
		t.Fatal("vacuous output verification", r)
	}
	if post(t, h, "/api/run", demo.Request{Source: ".decl out() out().", Facts: map[string]string{"../../outside": "bad"}}, &r) != 200 || r.Status != "invalid_input" {
		t.Fatal("undeclared filesystem input accepted", r)
	}
	if post(t, h, "/api/run", demo.Request{Source: strings.Repeat("x", 65537)}, nil) != 413 {
		t.Fatal("source limit missing")
	}
	if post(t, h, "/api/run", map[string]string{"path": "/etc/passwd"}, nil) != 400 {
		t.Fatal("unknown request field accepted")
	}
}
func TestTimeoutAndSuccessfulNextRun(t *testing.T) {
	h, _ := server(t, demo.Config{Timeout: 150 * time.Millisecond})
	var result demo.Result
	post(t, h, "/api/run", demo.Request{Source: ".decl p(x:number) .output p p(0). p(x+1) :- p(x).", Mode: "guarded"}, &result)
	if result.Status != "timeout" || result.Comparison != nil {
		t.Fatal(result)
	}
	for _, c := range result.Commands {
		if c.PID > 0 && c.Status == "timeout" {
			if syscall.Kill(c.PID, 0) != syscall.ESRCH {
				t.Fatal("timed-out process remains")
			}
		}
	}
	post(t, h, "/api/run", demo.Request{Source: ".decl out(x:number) .output out out(7).", Mode: "guarded"}, &result)
	if result.Status != "ok" || !result.Comparison.Equal {
		t.Fatal(result)
	}
}
func TestProcessGroupCleanup(t *testing.T) {
	dir := t.TempDir()
	script, pidfile := filepath.Join(dir, "child"), filepath.Join(dir, "pid")
	if e := os.WriteFile(script, []byte("#!/bin/sh\nsleep 60 &\necho $! > \"$1\"\nwait\n"), 0700); e != nil {
		t.Fatal(e)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	c := demo.RunCommand(ctx, "child cleanup", []string{script, pidfile}, 1024)
	if c.Status != "timeout" || !c.GroupCleanup {
		t.Fatal(c)
	}
	b, e := os.ReadFile(pidfile)
	if e != nil {
		t.Fatal(e)
	}
	pid := strings.TrimSpace(string(b))
	state, e := os.ReadFile("/proc/" + pid + "/stat")
	if e == nil && !strings.Contains(string(state), ") Z ") {
		t.Fatalf("child still live: %s", state)
	}
}
func TestBoundedRunRetentionAndLoopback(t *testing.T) {
	temp := t.TempDir()
	h, _ := server(t, demo.Config{TempDir: temp, MaxRuns: 2})
	first := ""
	for i := 0; i < 4; i++ {
		var result demo.Result
		post(t, h, "/api/run", demo.Request{Source: ".decl a() a().", Mode: "guarded", InspectOnly: true}, &result)
		if i == 0 {
			first = result.RunID
		}
	}
	if post(t, h, "/api/native/"+first, struct{}{}, nil) != 410 {
		t.Fatal("unbounded snapshot retention")
	}
	paths, _ := filepath.Glob(filepath.Join(temp, "dlc-demo-*", "run-*"))
	if len(paths) > 2 {
		t.Fatal(paths)
	}
	if demo.CheckLoopback("0.0.0.0:8080") == nil || demo.CheckLoopback("127.0.0.1:8080") != nil {
		t.Fatal("loopback check")
	}
}

func TestCancellationCaptureLimitsAndCompileFailure(t *testing.T) {
	dir := t.TempDir()
	script := filepath.Join(dir, "wait")
	if err := os.WriteFile(script, []byte("#!/bin/sh\nsleep 60\n"), 0700); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	time.AfterFunc(50*time.Millisecond, cancel)
	command := demo.RunCommand(ctx, "cancel", []string{script}, 1024)
	if command.Status != "cancelled" || !command.ProcessReaped || syscall.Kill(command.PID, 0) != syscall.ESRCH {
		t.Fatal(command)
	}
	command = demo.RunCommand(context.Background(), "capture", []string{"/usr/bin/printf", "abcdef"}, 3)
	if command.Status != "output_limit" || !command.CaptureTruncated || command.Stdout != "abc" {
		t.Fatal(command)
	}
	h, _ := server(t, demo.Config{CC: "/usr/bin/false"})
	var r demo.Result
	post(t, h, "/api/run", demo.Request{Source: ".decl out() .output out out().", Mode: "guarded"}, &r)
	var native demo.NativeResult
	post(t, h, "/api/native/"+r.RunID, struct{}{}, &native)
	if native.Status != "error" || native.Comparison != nil || len(native.Commands) != 2 || native.Commands[1].ExitCode == 0 {
		t.Fatal(native)
	}
	post(t, h, "/api/run", demo.Request{Source: ".decl input(x:number) .input input .decl out(x:number) .output out out(x):-input(x).", Facts: map[string]string{"input": "not-a-number\n"}, Mode: "guarded"}, &r)
	if r.Status != "error" || r.Comparison != nil {
		t.Fatal("malformed input became verified", r)
	}
}

func TestCCompilerAndActualNativeTimeouts(t *testing.T) {
	dir := t.TempDir()
	slowCC := filepath.Join(dir, "slow-cc")
	if err := os.WriteFile(slowCC, []byte("#!/bin/sh\nsleep 60 &\nwait\n"), 0700); err != nil {
		t.Fatal(err)
	}
	h, _ := server(t, demo.Config{CC: slowCC, Timeout: 200 * time.Millisecond})
	var result demo.Result
	post(t, h, "/api/run", demo.Request{Source: ".decl out() .output out out().", Mode: "guarded"}, &result)
	if result.Status != "ok" {
		t.Fatal(result)
	}
	var native demo.NativeResult
	post(t, h, "/api/native/"+result.RunID, struct{}{}, &native)
	if native.Status != "timeout" || native.Comparison != nil || len(native.Commands) != 2 {
		t.Fatal(native)
	}
	stopped := native.Commands[1]
	if !stopped.ProcessReaped || !stopped.GroupCleanup || syscall.Kill(stopped.PID, 0) != syscall.ESRCH {
		t.Fatal("C compiler was not cleaned up", stopped)
	}
	source, binary := filepath.Join(dir, "forever.c"), filepath.Join(dir, "forever")
	if err := os.WriteFile(source, []byte("int main(void) { for (;;) {} }\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if b, err := exec.Command("cc", "-std=c11", source, "-o", binary).CombinedOutput(); err != nil {
		t.Fatalf("native timeout fixture compile: %v %s", err, b)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	command := demo.RunCommand(ctx, "actual native timeout", []string{binary}, 1024)
	if command.Status != "timeout" || !command.ProcessReaped || syscall.Kill(command.PID, 0) != syscall.ESRCH {
		t.Fatal("native timeout was not reaped", command)
	}
}
