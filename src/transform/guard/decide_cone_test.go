package guard

import (
	"reflect"
	"sort"
	"testing"

	"dlc/src/parser"
)

// TestConeClosureMatchesHarnessCulpritCycle cross-checks Go's
// ConeClosure against harness/cone_metric.py's already-validated
// NIGHT-BATCH-03 T9 result on culprit_cycle.dl with {p} declined: the
// cone must be EXACTLY {q, s} -- M2-M3-BUILD.md §7's own required gate
// ("dlc's own cone computation and the harness's must agree exactly").
func TestConeClosureMatchesHarnessCulpritCycle(t *testing.T) {
	prog, errs := parser.Parse([]byte(culpritCycleSrc))
	if len(errs) != 0 {
		t.Fatalf("parse errors: %v", errs)
	}
	cone := ConeClosure(prog, map[string]bool{"p": true})
	var got []string
	for k := range cone {
		got = append(got, k)
	}
	sort.Strings(got)
	want := []string{"q", "s"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ConeClosure({p}) = %v, want %v (harness/cone_metric.py's NIGHT-BATCH-03 T9 validated result)", got, want)
	}
}
