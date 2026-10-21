package eval

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"dlc/src/ir"
	"dlc/src/sema"
)

// LoadFacts reads <factsDir>/<name>.facts (Soufflé's own tab-separated,
// no-header convention -- fixtures_lib.write_facts already produces
// exactly this) for every relation with an .input declaration, and
// Inserts each row -- never RecordSeedInsert/RecordIterationInsert
// (ir/DESIGN.md: EDB loads are never counted as rule-derived). A missing
// .facts file is treated as zero rows, not an error -- matching this
// project's own already-established fixture convention (an .input
// relation with no matching file is empty, not fatal) rather than
// mirroring Soufflé's own hard failure on a missing file (T2 found that
// exact Soufflé behavior, docs/reports/night02-T2-hostile.md, and it is
// not what a comparison harness wants here).
func (e *Evaluator) LoadFacts(factsDir string, schemas map[string]*sema.RelationSchema, inputNames map[string]bool) error {
	names := make([]string, 0, len(inputNames))
	for name := range inputNames {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		schema, ok := schemas[name]
		if !ok {
			continue
		}
		path := filepath.Join(factsDir, name+".facts")
		f, err := os.Open(path)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return fmt.Errorf("opening %s: %w", path, err)
		}
		rel := e.Relations[name]
		reader := bufio.NewReader(f)
		lineNo := 0
		for {
			line, readErr := reader.ReadString('\n')
			if readErr != nil && readErr != io.EOF {
				f.Close()
				return fmt.Errorf("reading %s: %w", path, readErr)
			}
			if len(line) == 0 && readErr == io.EOF {
				break
			}
			lineNo++
			line = strings.TrimSuffix(strings.TrimSuffix(line, "\n"), "\r")
			if strings.ContainsRune(line, 0) {
				f.Close()
				return fmt.Errorf("%s:%d: NUL byte in facts", path, lineNo)
			}
			cols := strings.Split(line, "\t")
			if len(schema.Params) == 0 && (line == "" || line == "()") {
				cols = nil
			}
			if len(cols) != len(schema.Params) {
				f.Close()
				return fmt.Errorf("%s:%d: expected %d columns, got %d", path, lineNo, len(schema.Params), len(cols))
			}
			tup := make(ir.Tuple, len(cols))
			for i, col := range cols {
				if schema.Params[i].Type == "number" {
					n, err := strconv.ParseInt(col, 10, 64)
					if err != nil {
						f.Close()
						return fmt.Errorf("%s:%d: column %d: %w", path, lineNo, i, err)
					}
					tup[i] = ir.NumberValue(n)
				} else {
					tup[i] = ir.StringValue(e.Strings.Intern(col))
				}
			}
			rel.Insert(tup)
		}
		f.Close()

	}
	return nil
}

// WriteOutput writes <outDir>/<name>.csv for every relation with an
// .output declaration, in the same tab-separated, no-header,
// insertion-order-doesn't-matter format Soufflé itself writes (the
// differential harness sorts before comparing, CLAUDE.md §6, so
// insertion order here is never significant).
func (e *Evaluator) WriteOutput(outDir string, schemas map[string]*sema.RelationSchema, outputNames map[string]bool) error {
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}
	for name := range outputNames {
		rel := e.Relations[name]
		path := filepath.Join(outDir, name+".csv")
		f, err := os.Create(path)
		if err != nil {
			return err
		}
		w := bufio.NewWriter(f)
		if rel != nil {
			for _, tup := range rel.All() {
				if len(tup) == 0 {
					fmt.Fprintln(w, "()")
					continue
				}
				cols := make([]string, len(tup))
				for i, v := range tup {
					cols[i] = e.valueString(v)
				}
				fmt.Fprintln(w, strings.Join(cols, "\t"))
			}
		}
		if err := w.Flush(); err != nil {
			f.Close()
			return fmt.Errorf("writing %s: %w", path, err)
		}
		if err := f.Close(); err != nil {
			return fmt.Errorf("closing %s: %w", path, err)
		}
	}
	return nil
}
