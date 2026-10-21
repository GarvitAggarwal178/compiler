package integration_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"dlc/src/ast"
	"dlc/src/transform/compact"
	"dlc/src/transform/guard"
)

// Explicitly optional locally; make oracle requires a Souffle executable so
// release verification cannot silently turn an absent oracle into a pass.
func TestSouffleDifferential(t *testing.T) {
	oracle := os.Getenv("DLC_SOUFFLE")
	if oracle == "" {
		oracle = "souffle"
	}
	path, err := exec.LookPath(oracle)
	if err != nil {
		t.Skip("Souffle unavailable; run make oracle with DLC_SOUFFLE set")
	}
	for _, tc := range preservationCases() {
		t.Run(tc.name, func(t *testing.T) {
			p, _, _ := checked(t, tc.source)
			facts := factsDir(t, tc.facts)
			out := t.TempDir()
			source := filepath.Join(t.TempDir(), "input.dl")
			if err := os.WriteFile(source, []byte(tc.source), 0600); err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()
			if b, err := exec.CommandContext(ctx, path, "-F", facts, "-D", out, source).CombinedOutput(); err != nil {
				t.Fatalf("Souffle: %v\n%s", err, b)
			}
			outputs := map[string]bool{}
			for _, d := range p.Decls {
				if d.Kind == ast.DeclOutput {
					outputs[d.Name] = true
				}
			}
			want := readAnswers(t, out, outputs)
			for _, prog := range []*ast.Program{p, transformed(t, p, guard.Transformer{}), transformed(t, p, compact.Transformer{})} {
				for _, semi := range []bool{false, true} {
					got := answers(t, prog, facts, semi)
					if !reflect.DeepEqual(got, want) {
						t.Fatalf("semi=%t: dlc=%v Souffle=%v", semi, got, want)
					}
				}
			}
			for _, q := range []*ast.Program{transformed(t, p, guard.Transformer{}), transformed(t, p, compact.Transformer{})} {
				if got := compiledAnswers(t, q, facts); !reflect.DeepEqual(got, want) {
					t.Fatalf("C=%v Souffle=%v", got, want)
				}
			}
		})
	}
}
