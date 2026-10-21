package demo_test

import (
	"dlc/src/demo"
	"os"
	"path/filepath"
	"testing"
)

func TestByteRowsAndOutputManifest(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "out.csv")
	cases := [][]byte{nil, []byte("\n"), []byte("()\n"), []byte("a\u2028b\n"), []byte("a\nb\n"), []byte("a\r\n")}
	seen := []demo.OutputSets{}
	for _, b := range cases {
		_ = os.WriteFile(p, b, 0600)
		sets, e := demo.ReadOutputs(dir, []string{"out.csv"}, 1024)
		if e != nil {
			t.Fatal(e)
		}
		seen = append(seen, sets)
	}
	if demo.Compare(seen[0], seen[1]).Equal || demo.Compare(seen[3], seen[4]).Equal {
		t.Fatal("distinct bytes/rows conflated")
	}
	_ = os.WriteFile(p, []byte("a"), 0600)
	if _, e := demo.ReadOutputs(dir, []string{"out.csv"}, 1024); e == nil {
		t.Fatal("unterminated row accepted")
	}
	_ = os.WriteFile(p, nil, 0600)
	_ = os.WriteFile(filepath.Join(dir, "extra.csv"), nil, 0600)
	if _, e := demo.ReadOutputs(dir, []string{"out.csv"}, 1024); e == nil {
		t.Fatal("extra empty file ignored")
	}
	if demo.Compare(demo.OutputSets{}, demo.OutputSets{}).Available {
		t.Fatal("no outputs shown as verified")
	}
	if demo.Compare(demo.OutputSets{"out.csv": demo.RowSet{"a": {}}}, demo.OutputSets{"out.csv": demo.RowSet{"b": {}}}).Equal {
		t.Fatal("equal cardinality treated as equal")
	}
}
