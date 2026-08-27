// conecheck is a small NIGHT-BATCH-04 B measurement aid: dumps
// guard.Decide's culprit/cone/declined sets as JSON for a given .dl
// file, so harness/night04_b_cone_gate.py can cross-check them against
// harness/cone_metric.py without re-deriving guard's own logic in
// Python. Not part of dlc's CLI surface (E's --explain is the real,
// user-facing version of this idea; this is a throwaway measurement
// tool, tools/ per CLAUDE.md's Lane B bucket).
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"

	"dlc/src/parser"
	"dlc/src/transform/guard"
)

type report struct {
	NoBindableQuery   bool     `json:"no_bindable_query"`
	IDBRelations      []string `json:"idb_relations"`
	CulpritPredicates []string `json:"culprit_predicates"`
	ConeRelations     []string `json:"cone_relations"`
	DeclinedRelations []string `json:"declined_relations"`
}

func sortedList(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: conecheck <file.dl>")
		os.Exit(2)
	}
	src, err := os.ReadFile(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, "read error:", err)
		os.Exit(1)
	}
	prog, errs := parser.Parse(src)
	if len(errs) != 0 {
		fmt.Fprintln(os.Stderr, "parse errors:", errs)
		os.Exit(1)
	}
	result, err := guard.Decide(prog)
	if err != nil {
		fmt.Fprintln(os.Stderr, "decide error:", err)
		os.Exit(1)
	}
	idb := map[string]bool{}
	for _, c := range prog.Clauses {
		idb[c.Head.Name] = true
	}
	out := report{
		NoBindableQuery:   result.NoBindableQuery,
		IDBRelations:      sortedList(idb),
		CulpritPredicates: sortedList(result.CulpritPredicates),
		ConeRelations:     sortedList(result.ConeRelations),
		DeclinedRelations: sortedList(result.DeclinedRelations),
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(out); err != nil {
		fmt.Fprintln(os.Stderr, "encode error:", err)
		os.Exit(1)
	}
}
