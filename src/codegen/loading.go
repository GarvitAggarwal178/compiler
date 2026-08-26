package codegen

// emitFactLoading emits load_facts(const char *factsDir), called once
// from main before evaluation. Mirrors eval/io.go's LoadFacts exactly:
// a missing <name>.facts file is zero rows, not an error (this
// project's own established fixture convention, not Soufflé's own
// hard-fail-on-missing-file behavior -- eval/io.go's own DESIGN.md note
// applies here unchanged).
func (g *generator) emitFactLoading() {
	g.w("static void load_facts(const char *facts_dir) {")
	g.w("\tchar path[4096];")
	g.w("\tchar cols[MAX_COLS][256];")
	for _, name := range g.inputNames {
		schema, ok := g.schemas[name]
		if !ok {
			continue
		}
		id := cIdent(name)
		arity := len(schema.Params)
		g.w("\t{")
		g.w("\t\tsnprintf(path, sizeof(path), \"%%s/%s.facts\", facts_dir);", name)
		g.w("\t\tFILE *f = fopen(path, \"r\");")
		g.w("\t\tif (f) {")
		if arity == 0 {
			g.w("\t\t\t%s_present = 1;", id)
		} else {
			g.w("\t\t\tint n;")
			g.w("\t\t\twhile ((n = read_facts_line(f, cols)) >= 0) {")
			g.w("\t\t\t\tif (n == 0) continue;")
			g.w("\t\t\t\tTuple_%s t;", id)
			for i, param := range schema.Params {
				if param.Type == "number" {
					g.w("\t\t\t\tt.c[%d] = atoll(cols[%d]);", i, i)
				} else {
					g.w("\t\t\t\tt.c[%d] = intern(cols[%d]);", i, i)
				}
			}
			g.w("\t\t\t\t%s_insert(t);", id)
			g.w("\t\t\t}")
		}
		g.w("\t\t\tfclose(f);")
		g.w("\t\t}")
		g.w("\t}")
	}
	g.w("}")
	g.w("")
}

// emitOutput emits write_output(const char *outDir), called once from
// main after evaluation. Tab-separated, no header -- Soufflé's own
// convention, matching eval/io.go's WriteOutput exactly (the
// differential harness sorts before comparing, CLAUDE.md section 6, so
// insertion order here is never significant).
func (g *generator) emitOutput() {
	g.w("static void write_output(const char *out_dir) {")
	g.w("\tchar path[4096];")
	g.w("\tFILE *f;")
	for _, name := range g.outputNames {
		schema, ok := g.schemas[name]
		if !ok {
			continue
		}
		id := cIdent(name)
		arity := len(schema.Params)
		g.w("\tsnprintf(path, sizeof(path), \"%%s/%s.csv\", out_dir);", name)
		g.w("\tf = fopen(path, \"w\");")
		g.w("\tif (f) {")
		if arity == 0 {
			g.w("\t\tif (%s_present) fprintf(f, \"\\n\");", id)
		} else {
			g.w("\t\tfor (size_t i = 0; i < %s_len; i++) {", id)
			for i, param := range schema.Params {
				sep := ""
				if i > 0 {
					sep = "\"\\t\""
					g.w("\t\t\tfputs(%s, f);", sep)
				}
				if param.Type == "number" {
					g.w("\t\t\tfprintf(f, \"%%lld\", (long long)%s_data[i].c[%d]);", id, i)
				} else {
					g.w("\t\t\tfputs(str_lookup(%s_data[i].c[%d]), f);", id, i)
				}
			}
			g.w("\t\t\tfputc('\\n', f);")
			g.w("\t\t}")
		}
		g.w("\t\tfclose(f);")
		g.w("\t}")
	}
	g.w("}")
	g.w("")
}

func (g *generator) emitMain() {
	g.w("int main(int argc, char **argv) {")
	g.w("\tif (argc != 3) { fprintf(stderr, \"usage: %%s <factsDir> <outDir>\\n\", argv[0]); return 2; }")
	g.w("\tload_facts(argv[1]);")
	g.w("\tevaluate();")
	g.w("\twrite_output(argv[2]);")
	g.w("\treturn 0;")
	g.w("}")
}
