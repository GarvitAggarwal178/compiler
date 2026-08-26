// Package ir holds dlc's runtime representation: interned values, tuple
// storage, and per-relation indices. §3.7. Lane B.
package ir

import "strconv"

// Value is one column's runtime value: either a number (int64) or a
// symbol, stored as an interned id into a StringTable rather than the
// string itself -- comparing two symbols is then an int32 comparison,
// and a Value stays a small, comparable (usable as a Go map key) struct
// regardless of which case it holds.
type Value struct {
	IsString bool
	Num      int64
	StrID    int32
}

func NumberValue(n int64) Value  { return Value{Num: n} }
func StringValue(id int32) Value { return Value{IsString: true, StrID: id} }

// Tuple is one relation row. Represented as a slice (arity varies per
// relation, so a fixed-size array type per arity isn't practical), so a
// Tuple itself is not comparable/usable as a map key directly --
// tupleKey below builds a comparable string key instead.
type Tuple []Value

// StringTable interns symbol strings to small integer ids: repeated
// symbols compare and hash as ints instead of repeatedly comparing full
// strings, and Value stays a fixed-size struct.
type StringTable struct {
	ids  map[string]int32
	strs []string
}

func NewStringTable() *StringTable {
	return &StringTable{ids: map[string]int32{}}
}

func (t *StringTable) Intern(s string) int32 {
	if id, ok := t.ids[s]; ok {
		return id
	}
	id := int32(len(t.strs))
	t.strs = append(t.strs, s)
	t.ids[s] = id
	return id
}

func (t *StringTable) Lookup(id int32) string {
	return t.strs[id]
}

// RelationStats is the instrumentation §3.7 requires: tuples inserted by
// rule evaluation, per phase. EDB loads are never counted here -- see
// DESIGN.md; Insert alone never touches these fields, only
// RecordSeedInsert/RecordIterationInsert do, and those are called by the
// evaluator (§3.8/§3.9), never by fact loading.
type RelationStats struct {
	SeedInserts      int   // tuples derived by this relation's non-recursive ("seed") rules
	IterationInserts []int // per semi-naive iteration, 0-indexed; naive evaluation (§3.8) records everything into iteration 0
}

// Total is the same derived-tuple metric this project has used
// throughout (docs/MEASUREMENTS.md): seed plus every iteration's delta,
// i.e. every tuple a rule ever derived, counted once.
func (s RelationStats) Total() int {
	total := s.SeedInserts
	for _, c := range s.IterationInserts {
		total += c
	}
	return total
}

// Relation is one relation's tuple storage plus exactly one index: a
// naive choice (§3.7 explicitly allows this), keyed on column 0. See
// DESIGN.md for what a real index-selection pass would do instead.
type Relation struct {
	Name  string
	Arity int
	Stats RelationStats

	tuples []Tuple
	seen   map[string]bool
	idx0   map[Value][]int // column-0 value -> row indices into tuples; empty/unused if Arity == 0
}

func NewRelation(name string, arity int) *Relation {
	return &Relation{
		Name: name, Arity: arity,
		seen: map[string]bool{}, idx0: map[Value][]int{},
	}
}

// Insert adds t if it is not already present (set semantics -- Datalog
// relations never hold duplicate tuples) and returns whether it was
// newly added. Does not itself touch Stats; EDB fact loading calls this
// and nothing else, so loaded facts are never counted as rule-derived
// (§3.7: "exclude EDB loads").
func (r *Relation) Insert(t Tuple) bool {
	key := tupleKey(t)
	if r.seen[key] {
		return false
	}
	r.seen[key] = true
	idx := len(r.tuples)
	r.tuples = append(r.tuples, t)
	if r.Arity > 0 {
		r.idx0[t[0]] = append(r.idx0[t[0]], idx)
	}
	return true
}

// RecordSeedInsert and RecordIterationInsert are what the evaluator
// calls, immediately after a successful (newly-added) Insert coming
// from a non-recursive or recursive rule respectively -- never for an
// EDB load.
func (r *Relation) RecordSeedInsert() { r.Stats.SeedInserts++ }

func (r *Relation) RecordIterationInsert(iteration int) {
	for len(r.Stats.IterationInserts) <= iteration {
		r.Stats.IterationInserts = append(r.Stats.IterationInserts, 0)
	}
	r.Stats.IterationInserts[iteration]++
}

// All returns every tuple currently stored, in insertion order.
func (r *Relation) All() []Tuple { return r.tuples }

func (r *Relation) Len() int { return len(r.tuples) }

// LookupByFirstColumn returns every currently-stored tuple whose column
// 0 equals v, via the idx0 index -- O(matches), not a scan of every
// tuple in the relation. Panics-free even for a zero-arity relation
// (idx0 is simply always empty then; LookupByFirstColumn should not be
// called on one, but returns an empty slice rather than panicking if it
// is).
func (r *Relation) LookupByFirstColumn(v Value) []Tuple {
	rows := r.idx0[v]
	out := make([]Tuple, len(rows))
	for i, idx := range rows {
		out[i] = r.tuples[idx]
	}
	return out
}

// tupleKey builds a comparable string key for set-membership dedup.
// Values are separated by a byte that cannot appear inside the encoded
// form of any single value (0x1f, ASCII unit separator, never produced
// by strconv.FormatInt or a StrID's decimal digits) so no two distinct
// tuples can collide onto the same key.
func tupleKey(t Tuple) string {
	b := make([]byte, 0, len(t)*8)
	for i, v := range t {
		if i > 0 {
			b = append(b, 0x1f)
		}
		if v.IsString {
			b = append(b, 's')
			b = strconv.AppendInt(b, int64(v.StrID), 10)
		} else {
			b = append(b, 'n')
			b = strconv.AppendInt(b, v.Num, 10)
		}
	}
	return string(b)
}
