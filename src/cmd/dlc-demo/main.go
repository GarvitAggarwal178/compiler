// dlc-demo serves an offline, loopback compiler workbench.
package main

import (
	"context"
	"embed"
	"flag"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"dlc/src/demo"
)

//go:embed web/*
var web embed.FS

func main() {
	addr := flag.String("addr", "127.0.0.1:8080", "loopback address")
	repo := flag.String("repo", ".", "repository root")
	compiler := flag.String("compiler", "", "compiler executable (default: repository bin/dlc)")
	cc := flag.String("cc", env("CC", "cc"), "C compiler executable")
	oracle := flag.String("oracle", env("DLC_SOUFFLE", "souffle"), "optional Souffle executable")
	timeout := flag.Duration("timeout", 3*time.Second, "timeout per subprocess, maximum 30s")
	maxRuns := flag.Int("max-runs", 8, "retained temporary input snapshots")
	flag.Parse()
	if flag.NArg() != 0 {
		fmt.Fprintln(os.Stderr, "unexpected arguments")
		os.Exit(2)
	}
	if err := demo.CheckLoopback(*addr); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	assets, _ := fs.Sub(web, "web")
	workbench, err := demo.New(demo.Config{Repo: *repo, Compiler: *compiler, CC: *cc, Oracle: *oracle, Timeout: *timeout, MaxRuns: *maxRuns}, assets)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer workbench.Close()
	listener, err := net.Listen("tcp", *addr)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	server := &http.Server{Handler: workbench, ReadHeaderTimeout: 5 * time.Second}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	go func() {
		<-ctx.Done()
		shutdown, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = server.Shutdown(shutdown)
		_ = server.Close()
	}()
	for name, t := range workbench.ToolStatus() {
		if t.Available {
			fmt.Printf("%s: %s\n", name, t.Path)
		} else {
			fmt.Printf("%s unavailable: %s\n", name, t.Detail)
		}
	}
	fmt.Printf("dlc workbench: http://%s\n", listener.Addr())
	fmt.Printf("Per-command timeout: %s; snapshots are temporary and bounded. Ctrl-C stops all work.\n", *timeout)
	if err = server.Serve(listener); err != nil && err != http.ErrServerClosed {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
func env(name, fallback string) string {
	if s := os.Getenv(name); s != "" {
		return s
	}
	return fallback
}
