package demo

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"

	"dlc/src/sema"
)

type RowSet map[string]struct{}
type OutputSets map[string]RowSet

// ReadOutputs implements the same whole-byte-row/LF contract as
// harness/answers.py. Missing and extra files fail even when empty.
func ReadOutputs(dir string, expected []string, limit int64) (OutputSets, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	want := map[string]bool{}
	for _, n := range expected {
		want[n] = true
	}
	actual := map[string]bool{}
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".csv") {
			if !e.Type().IsRegular() {
				return nil, fmt.Errorf("non-regular output %s", e.Name())
			}
			actual[e.Name()] = true
		}
	}
	if len(actual) != len(want) {
		return nil, fmt.Errorf("output file manifest differs: got %v, expected %v", sortedNames(actual), sortedNames(want))
	}
	for n := range want {
		if !actual[n] {
			return nil, fmt.Errorf("missing output file %s", n)
		}
	}
	out := OutputSets{}
	var total int64
	for _, name := range expected {
		p := filepath.Join(dir, name)
		info, err := os.Stat(p)
		if err != nil {
			return nil, err
		}
		total += info.Size()
		if total > limit {
			return nil, fmt.Errorf("output exceeds %d-byte display/verification limit", limit)
		}
		b, err := os.ReadFile(p)
		if err != nil {
			return nil, err
		}
		if len(b) > 0 && b[len(b)-1] != '\n' {
			return nil, fmt.Errorf("unterminated LF row in %s", name)
		}
		rows := RowSet{}
		if len(b) > 0 {
			for _, row := range bytes.Split(b[:len(b)-1], []byte{'\n'}) {
				rows[string(row)] = struct{}{}
			}
		}
		out[name] = rows
	}
	return out, nil
}
func sortedNames[T any](m map[string]T) []string {
	out := make([]string, 0, len(m))
	for n := range m {
		out = append(out, n)
	}
	sort.Strings(out)
	return out
}

type Row struct {
	Escaped   string   `json:"escaped"`
	RawBase64 string   `json:"raw_base64"`
	Columns   []string `json:"columns,omitempty"`
	Ambiguous bool     `json:"ambiguous"`
}

func displayRow(raw string, params []sema.RelationParam) Row {
	r := Row{Escaped: strconv.Quote(raw), RawBase64: base64.StdEncoding.EncodeToString([]byte(raw))}
	if len(params) == 0 {
		r.Ambiguous = raw != "()"
		return r
	}
	fields := strings.Split(raw, "\t")
	safe := len(fields) == len(params) && utf8.ValidString(raw)
	if safe {
		for i, p := range params {
			if p.Type == "number" {
				if _, err := strconv.ParseInt(fields[i], 10, 64); err != nil {
					safe = false
				}
			}
		}
	}
	if safe {
		r.Columns = fields
	} else {
		r.Ambiguous = true
	}
	return r
}

type OutputFile struct {
	Name   string               `json:"name"`
	Params []sema.RelationParam `json:"params"`
	Rows   []Row                `json:"rows"`
}

func displayOutputs(sets OutputSets, relations []sema.RelationInfo) []OutputFile {
	out := []OutputFile{}
	for _, name := range sortedNames(sets) {
		params := []sema.RelationParam{}
		for _, r := range relations {
			if r.Name+".csv" == name {
				params = r.Params
				break
			}
		}
		f := OutputFile{Name: name, Params: params, Rows: []Row{}}
		for _, row := range sortedNames(sets[name]) {
			f.Rows = append(f.Rows, displayRow(row, params))
		}
		out = append(out, f)
	}
	return out
}

type Difference struct {
	File        string   `json:"file"`
	MissingFile bool     `json:"missing_file"`
	ExtraFile   bool     `json:"extra_file"`
	Missing     []string `json:"missing"`
	Extra       []string `json:"extra"`
}
type Comparison struct {
	Available   bool         `json:"available"`
	Equal       bool         `json:"equal"`
	Reason      string       `json:"reason,omitempty"`
	Differences []Difference `json:"differences"`
}

func Compare(actual, expected OutputSets) Comparison {
	c := Comparison{Available: len(expected) > 0, Equal: true, Differences: []Difference{}}
	if !c.Available {
		c.Equal = false
		c.Reason = "no output declarations; no answer comparison"
		return c
	}
	names := map[string]bool{}
	for n := range actual {
		names[n] = true
	}
	for n := range expected {
		names[n] = true
	}
	for _, name := range sortedNames(names) {
		a, aok := actual[name]
		e, eok := expected[name]
		d := Difference{File: name, MissingFile: !aok, ExtraFile: !eok, Missing: []string{}, Extra: []string{}}
		for _, v := range sortedNames(e) {
			if _, ok := a[v]; !ok {
				d.Missing = append(d.Missing, strconv.Quote(v))
			}
		}
		for _, v := range sortedNames(a) {
			if _, ok := e[v]; !ok {
				d.Extra = append(d.Extra, strconv.Quote(v))
			}
		}
		if !aok || !eok || len(d.Missing) > 0 || len(d.Extra) > 0 {
			c.Equal = false
			c.Differences = append(c.Differences, d)
		}
	}
	return c
}
