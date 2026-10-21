package sema

import (
	"dlc/src/parser"
	"encoding/json"
	"testing"
)

func TestCanonicalDiagnosticsAndSCCExport(t *testing.T) {
	p, errs := parser.Parse([]byte(`.decl z() .decl a() .decl input() .input input z() :- !z(). a() :- !a().`))
	if len(errs) > 0 {
		t.Fatal(errs)
	}
	ds, _ := CheckStratification(p)
	first, _ := json.Marshal(ds)
	for i := 0; i < 100; i++ {
		ds, _ := CheckStratification(p)
		b, _ := json.Marshal(ds)
		if string(b) != string(first) {
			t.Fatal("unstable diagnostic")
		}
	}
	g := InspectGraph(p)
	if len(g.InvalidSCCs) != 2 || g.InvalidSCCs[0][0] != "a" || g.InvalidSCCs[1][0] != "z" {
		t.Fatal(g.InvalidSCCs)
	}
	for _, r := range g.Relations {
		if r.Name == "input" && r.SCC != nil {
			t.Fatal("input-only SCC")
		}
	}
}
