package codegen

func (g *generator) emitPrelude() {
	g.sb.WriteString(preludeTemplate)
}

// emitRelationStorage emits, per declared relation: a fixed-arity tuple
// struct, a growable array of them, and a chained hash index on column 0
// (§4 item 1's own wording: "nested loops over relations with hash
// indices") for insert-dedup and for LookupByFirstColumn-equivalent
// joins -- the same one-index-on-column-0 choice ir.Relation makes
// (ir/DESIGN.md), for the same reason (naive is explicitly permitted;
// a real index-selection pass would look at actual adornments).
// Zero-arity relations get a plain presence flag instead of an array --
// an empty struct is not standard C.
func (g *generator) emitRelationStorage() {
	for _, name := range g.relNames() {
		schema := g.schemas[name]
		arity := len(schema.Params)
		id := cIdent(name)
		if arity == 0 {
			g.w("static int %s_present = 0;", id)
			continue
		}
		g.w("typedef struct { int64_t c[%d]; } Tuple_%s;", arity, id)
		g.w("static Tuple_%s *%s_data = NULL;", id, id)
		g.w("static size_t %s_len = 0, %s_cap = 0;", id, id)
		g.w("#define %s_HASH_BUCKETS 65536", id)
		g.w("typedef struct HNode_%s { size_t row; struct HNode_%s *next; } HNode_%s;", id, id, id)
		g.w("static HNode_%s *%s_buckets[%s_HASH_BUCKETS];", id, id, id)
		g.w("")
		g.w("static int %s_contains(Tuple_%s t) {", id, id)
		g.w("\tunsigned h = (unsigned)((uint64_t)t.c[0] %% %s_HASH_BUCKETS);", id)
		g.w("\tfor (HNode_%s *n = %s_buckets[h]; n; n = n->next) {", id, id)
		g.w("\t\tif (memcmp(&%s_data[n->row], &t, sizeof(Tuple_%s)) == 0) return 1;", id, id)
		g.w("\t}")
		g.w("\treturn 0;")
		g.w("}")
		g.w("")
		g.w("static int %s_insert(Tuple_%s t) {", id, id) // returns 1 if newly inserted
		g.w("\tif (%s_contains(t)) return 0;", id)
		g.w("\tif (%s_len == %s_cap) { %s_cap = %s_cap ? %s_cap * 2 : 16; %s_data = realloc(%s_data, %s_cap * sizeof(Tuple_%s)); }",
			id, id, id, id, id, id, id, id, id)
		g.w("\t%s_data[%s_len] = t;", id, id)
		g.w("\tunsigned h = (unsigned)((uint64_t)t.c[0] %% %s_HASH_BUCKETS);", id)
		g.w("\tHNode_%s *n = malloc(sizeof(HNode_%s)); n->row = %s_len; n->next = %s_buckets[h];", id, id, id, id)
		g.w("\t%s_buckets[h] = n;", id)
		g.w("\t%s_len++;", id)
		g.w("\treturn 1;")
		g.w("}")
		g.w("")
		// No separate "lookup by first column" function returning a
		// materialized list: the join codegen (evaluation.go) walks the
		// hash bucket's linked list directly inline instead, so there is
		// no fixed-size scratch buffer anywhere to overflow regardless
		// of how many rows share a first-column value (a real risk that
		// was caught and removed before this was ever generated into a
		// program run against the family's larger scale points -- see
		// DESIGN.md).
	}
}
