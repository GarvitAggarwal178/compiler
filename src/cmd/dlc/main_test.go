package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestCLIRejectsErrorsAndCompilesGuarded(t *testing.T) {
	dir := t.TempDir()
	bin := filepath.Join(dir, "dlc")
	goexe := filepath.Join(os.Getenv("GOROOT"), "bin", "go")
	if _, err := os.Stat(goexe); err != nil {
		goexe = "go"
	}
	if b, err := exec.Command(goexe, "build", "-buildvcs=false", "-o", bin, ".").CombinedOutput(); err != nil {
		t.Fatalf("build: %v\n%s", err, b)
	}
	badParse := filepath.Join(dir, "parse.dl")
	badSema := filepath.Join(dir, "sema.dl")
	good := filepath.Join(dir, "good.dl")
	for name, text := range map[string]string{badParse: ".decl p(\n", badSema: ".decl p(x:number)\np(x).\n", good: ".decl p(x:number,y:number)\np(1,2).\n.decl out(y:number)\n.output out\nout(y) :- p(1,y).\n"} {
		if err := os.WriteFile(name, []byte(text), 0600); err != nil {
			t.Fatal(err)
		}
	}
	for _, args := range [][]string{
		{"parse", badParse}, {"check", badSema}, {"explain", badSema}, {"roundtrip", badParse},
		{"run", badParse, dir, dir}, {"run-seminaive", badSema, dir, dir},
		{"codegen", badSema, filepath.Join(dir, "bad.c")}, {"emit", badSema},
		{"emit", good, "--typo=guarded"}, {"emit", good, "--transformer=missing"},
	} {
		t.Run(strings.Join(args[:2], "_"), func(t *testing.T) {
			if b, err := exec.Command(bin, args...).CombinedOutput(); err == nil {
				t.Fatalf("expected failure: %v\n%s", args, b)
			}
		})
	}
	for _, flag := range []string{"--transformer=guarded", "--transformer=passthrough"} {
		cpath := filepath.Join(dir, "out.c")
		if b, err := exec.Command(bin, "codegen", good, cpath, flag).CombinedOutput(); err != nil {
			t.Fatalf("codegen: %v\n%s", err, b)
		}
		b, err := os.ReadFile(cpath)
		if err != nil {
			t.Fatal(err)
		}
		if flag == "--transformer=guarded" && !strings.Contains(string(b), "magic_p_bf") {
			t.Fatal("optimized C must contain the demand seed")
		}
	}
}
