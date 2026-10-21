package codegen

// The C backend shares the Go evaluator's exact-arity TSV contract.
func (g *generator) emitFactLoading() {
	g.w("static void load_facts(const char *facts_dir) {")
	for _, name := range g.inputNames {
		schema := g.schemas[name]
		id := cIdent(name)
		arity := len(schema.Params)
		g.w("\t{")
		g.w("\t\tchar *path=file_path(facts_dir,\"%s\",\".facts\");", name)
		g.w("\t\tFILE *f=fopen(path,\"r\");")
		g.w("\t\tif(!f && errno!=ENOENT){perror(path);exit(1);}")
		g.w("\t\tif(f) {")
		g.w("\t\t\tchar *line=NULL;size_t capacity=0,line_no=0;ssize_t length;")
		g.w("\t\t\twhile((length=getline(&line,&capacity,f))>=0) {")
		g.w("\t\t\t\tline_no++;if(length && line[length-1]=='\\n')line[--length]=0;")
		g.w("\t\t\t\tif(length && line[length-1]=='\\r')line[--length]=0;")
		g.w("\t\t\t\tif(memchr(line,0,(size_t)length))facts_error(path,line_no,\"NUL byte in facts\");")
		if arity == 0 {
			g.w("\t\t\t\tif(length && strcmp(line,\"()\"))facts_error(path,line_no,\"expected zero columns\");")
			g.w("\t\t\t\t%s_present=1;", id)
		} else {
			g.w("\t\t\t\tchar *cols[%d];", arity)
			g.w("\t\t\t\tif(split_fields(line,cols,%d)!=%d)facts_error(path,line_no,\"expected %d columns\");", arity, arity, arity)
			g.w("\t\t\t\tTuple_%s t;", id)
			for i, param := range schema.Params {
				if param.Type == "number" {
					g.w("\t\t\t\tt.c[%d]=parse_number(cols[%d],path,line_no);", i, i)
				} else {
					g.w("\t\t\t\tt.c[%d]=intern(cols[%d]);", i, i)
				}
			}
			g.w("\t\t\t\t%s_insert(t);", id)
		}
		g.w("\t\t\t}")
		g.w("\t\t\tif(ferror(f)){perror(path);exit(1);}free(line);if(fclose(f)){perror(path);exit(1);}")
		g.w("\t\t}free(path);")
		g.w("\t}")
	}
	g.w("}\n")
}

func (g *generator) emitOutput() {
	g.w("static void write_output(const char *out_dir) {")
	for _, name := range g.outputNames {
		schema := g.schemas[name]
		id := cIdent(name)
		g.w("\t{")
		g.w("\t\tchar *path=file_path(out_dir,\"%s\",\".csv\");", name)
		g.w("\t\tFILE *f=fopen(path,\"w\");if(!f){perror(path);exit(1);}")
		if len(schema.Params) == 0 {
			g.w("\t\tif(%s_present)fputs(\"()\\n\",f);", id)
		} else {
			g.w("\t\tfor(size_t i=0;i<%s_len;i++) {", id)
			for i, param := range schema.Params {
				if i > 0 {
					g.w("\t\t\tfputc('\\t',f);")
				}
				if param.Type == "number" {
					g.w("\t\t\tfprintf(f,\"%%lld\",(long long)%s_data[i].c[%d]);", id, i)
				} else {
					g.w("\t\t\tfputs(str_lookup(%s_data[i].c[%d]),f);", id, i)
				}
			}
			g.w("\t\t\tfputc('\\n',f);\n\t\t}")
		}
		g.w("\t\tint failed=ferror(f);if(fclose(f)||failed){perror(path);exit(1);}free(path);")
		g.w("\t}")
	}
	g.w("}\n")
}

func (g *generator) emitMain() {
	g.w("int main(int argc,char **argv) {")
	g.w("\tif(argc!=3){fprintf(stderr,\"usage: %%s <factsDir> <outDir>\\n\",argv[0]);return 2;}")
	g.w("\tload_facts(argv[1]);evaluate();ensure_output_dir(argv[2]);write_output(argv[2]);cleanup();return 0;\n}")
}

func (g *generator) emitCleanup() {
	g.w("static void cleanup(void) {")
	for _, name := range g.relNames() {
		if len(g.schemas[name].Params) == 0 {
			continue
		}
		id := cIdent(name)
		g.w("\tfor(size_t i=0;i<%s_HASH_BUCKETS;i++){HNode_%s *p=%s_buckets[i];while(p){HNode_%s *next=p->next;free(p);p=next;}}", id, id, id, id)
		g.w("\tfree(%s_data);", id)
	}
	g.w("\tfor(size_t i=0;i<str_len;i++)free(str_table[i]);free(str_table);\n}\n")
}
