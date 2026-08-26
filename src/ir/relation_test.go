package ir

import "testing"

func TestInsertDedup(t *testing.T) {
	r := NewRelation("p", 2)
	if !r.Insert(Tuple{NumberValue(1), NumberValue(2)}) {
		t.Fatalf("expected first insert to report newly-added")
	}
	if r.Insert(Tuple{NumberValue(1), NumberValue(2)}) {
		t.Fatalf("expected duplicate insert to report NOT newly-added")
	}
	if r.Len() != 1 {
		t.Fatalf("expected exactly 1 stored tuple, got %d", r.Len())
	}
}

func TestStringVsNumberNeverCollide(t *testing.T) {
	// A symbol id 5 and a number 5 must not be treated as the same Value.
	r := NewRelation("p", 1)
	r.Insert(Tuple{NumberValue(5)})
	if !r.Insert(Tuple{StringValue(5)}) {
		t.Fatalf("expected StringValue(5) to be a distinct tuple from NumberValue(5)")
	}
	if r.Len() != 2 {
		t.Fatalf("expected 2 distinct tuples, got %d", r.Len())
	}
}

func TestLookupByFirstColumn(t *testing.T) {
	r := NewRelation("edge", 2)
	r.Insert(Tuple{NumberValue(1), NumberValue(2)})
	r.Insert(Tuple{NumberValue(1), NumberValue(3)})
	r.Insert(Tuple{NumberValue(2), NumberValue(3)})

	got := r.LookupByFirstColumn(NumberValue(1))
	if len(got) != 2 {
		t.Fatalf("expected 2 tuples with first column 1, got %d: %v", len(got), got)
	}
	got2 := r.LookupByFirstColumn(NumberValue(99))
	if len(got2) != 0 {
		t.Fatalf("expected 0 tuples for an absent key, got %d", len(got2))
	}
}

func TestZeroArityRelationDoesNotPanic(t *testing.T) {
	r := NewRelation("p", 0)
	if !r.Insert(Tuple{}) {
		t.Fatalf("expected the (only possible) zero-arity tuple to insert once")
	}
	if r.Insert(Tuple{}) {
		t.Fatalf("expected a second zero-arity insert to be a duplicate")
	}
	_ = r.LookupByFirstColumn(NumberValue(0)) // must not panic
}

func TestInsertNeverTouchesStats(t *testing.T) {
	r := NewRelation("p", 1)
	r.Insert(Tuple{NumberValue(1)})
	r.Insert(Tuple{NumberValue(2)})
	if r.Stats.Total() != 0 {
		t.Fatalf("expected Insert alone to leave Stats untouched (EDB loads must not be counted), got Total()=%d", r.Stats.Total())
	}
}

func TestRecordSeedAndIterationInserts(t *testing.T) {
	r := NewRelation("p", 1)
	r.RecordSeedInsert()
	r.RecordSeedInsert()
	r.RecordIterationInsert(0)
	r.RecordIterationInsert(0)
	r.RecordIterationInsert(2) // sparse -- iteration 1 never fires
	if r.Stats.Total() != 5 {
		t.Fatalf("expected Total()=5 (2 seed + 2+0+1 iterations), got %d", r.Stats.Total())
	}
	if len(r.Stats.IterationInserts) != 3 {
		t.Fatalf("expected IterationInserts to have grown to length 3 (indices 0,1,2), got %v", r.Stats.IterationInserts)
	}
}

func TestStringTableInternRoundTrips(t *testing.T) {
	st := NewStringTable()
	id1 := st.Intern("foo")
	id2 := st.Intern("bar")
	id1Again := st.Intern("foo")
	if id1 != id1Again {
		t.Fatalf("expected interning the same string twice to return the same id")
	}
	if id1 == id2 {
		t.Fatalf("expected different strings to get different ids")
	}
	if st.Lookup(id1) != "foo" || st.Lookup(id2) != "bar" {
		t.Fatalf("Lookup did not round-trip: %q %q", st.Lookup(id1), st.Lookup(id2))
	}
}

func TestEmitProfileMatchesSouffleShape(t *testing.T) {
	relations := map[string]*Relation{
		"edge": NewRelation("edge", 2), // pure EDB, no rule-derived tuples ever recorded
		"reach": func() *Relation {
			r := NewRelation("reach", 2)
			r.RecordSeedInsert()
			r.RecordIterationInsert(0)
			r.RecordIterationInsert(0)
			r.RecordIterationInsert(1)
			return r
		}(),
	}
	doc := EmitProfile(relations)
	rel := doc.Root.Program.Relation
	if rel["edge"].NumTuples != 0 || rel["edge"].Iteration != nil {
		t.Fatalf("expected edge (EDB, never recorded) to show num-tuples=0 and no iteration block, got %+v", rel["edge"])
	}
	reach := rel["reach"]
	if reach.NumTuples != 1 {
		t.Fatalf("expected reach's seed num-tuples=1, got %d", reach.NumTuples)
	}
	if reach.Iteration["0"].NumTuples != 2 || reach.Iteration["1"].NumTuples != 1 {
		t.Fatalf("expected iteration 0=2, 1=1, got %+v", reach.Iteration)
	}
}
