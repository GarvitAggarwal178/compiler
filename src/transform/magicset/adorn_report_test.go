package magicset

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"dlc/src/parser"
)

// TestReportAdornedPredicates is a reporting aid for M2-M3-BUILD.md
// section 2's gate ("report the set of adorned predicates produced and
// the worklist iteration count"), run manually via `go test -run
// TestReportAdornedPredicates -v`, not part of the regular suite's
// correctness assertions.
func TestReportAdornedPredicates(t *testing.T) {
	repo := "/root/compiler"
	files := []string{
		"tests/corpus/BENCHMARK_FAMILY/same_generation_negation.dl",
		"tests/corpus/BENCHMARK_FAMILY/transitive_closure_bound.dl",
		"tests/corpus/BENCHMARK_FAMILY/ancestor_nonancestor.dl",
		"tests/corpus/BENCHMARK_FAMILY/reachability_complement.dl",
		"tests/corpus/BENCHMARK_FAMILY/culprit_cycle.dl",
		"tests/programs/p2.dl",
	}
	for _, f := range files {
		src, err := os.ReadFile(filepath.Join(repo, f))
		if err != nil {
			t.Fatal(err)
		}
		prog, errs := parser.Parse(src)
		if len(errs) != 0 {
			t.Fatalf("%s: parse errors: %v", f, errs)
		}
		q := FindQuery(prog)
		if q == nil {
			fmt.Printf("%s: no bindable query, no-op\n", f)
			continue
		}
		result, err := Adorn(prog, q)
		if err != nil {
			t.Fatalf("%s: adorn error: %v", f, err)
		}
		fmt.Printf("%s: iterations=%d adorned=%v\n", f, result.Iterations, result.Order)
	}
}
