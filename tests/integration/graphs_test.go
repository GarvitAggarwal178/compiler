package integration_test

import (
	"fmt"
	"math/rand"
	"reflect"
	"sort"
	"strings"
	"testing"

	"dlc/src/ast"
	"dlc/src/transform/guard"
)

// Independent reachability (BFS), rather than agreement between two engines
// sharing the same implementation. Seeds and graph size are fixed.
func TestGraphAnswersAgainstBFS(t *testing.T) {
	for seed := int64(0); seed < 12; seed++ {
		t.Run(fmt.Sprint(seed), func(t *testing.T) {
			rng := rand.New(rand.NewSource(seed))
			const n = 12
			adjacency := make([][]int, n+1)
			var source strings.Builder
			source.WriteString(".decl node(x:number)\n.decl edge(x:number,y:number)\n.decl reach(x:number,y:number)\n.decl missing(x:number,y:number)\n.decl out(y:number)\n.output out\nreach(x,y) :- edge(x,y).\nreach(x,y) :- edge(x,z),reach(z,y).\nmissing(x,y) :- node(x),node(y),!reach(x,y).\nout(y) :- missing(1,y).\n")
			for x := 1; x <= n; x++ {
				fmt.Fprintf(&source, "node(%d).\n", x)
				for y := 1; y <= n; y++ {
					if rng.Intn(10) == 0 {
						adjacency[x] = append(adjacency[x], y)
						fmt.Fprintf(&source, "edge(%d,%d).\n", x, y)
					}
				}
			}
			reached := map[int]bool{}
			queue := []int{1}
			for len(queue) > 0 {
				x := queue[0]
				queue = queue[1:]
				for _, y := range adjacency[x] {
					if !reached[y] {
						reached[y] = true
						queue = append(queue, y)
					}
				}
			}
			rows := []string{}
			for y := 1; y <= n; y++ {
				if !reached[y] {
					rows = append(rows, fmt.Sprint(y))
				}
			}
			sort.Strings(rows)
			want := map[string][]string{"out": rows}
			p, _, _ := checked(t, source.String())
			dir := t.TempDir()
			q := transformed(t, p, guard.Transformer{})
			for _, prog := range []*ast.Program{p, q} {
				for _, semi := range []bool{false, true} {
					if got := answers(t, prog, dir, semi); !reflect.DeepEqual(got, want) {
						t.Fatalf("semi=%t got %v; BFS=%v", semi, got, want)
					}
				}
			}
			if seed < 3 {
				if got := compiledAnswers(t, q, dir); !reflect.DeepEqual(got, want) {
					t.Fatalf("C=%v BFS=%v", got, want)
				}
			}
		})
	}
}
