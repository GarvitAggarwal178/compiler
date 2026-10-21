package demo

import (
	"context"
	"errors"
	"os/exec"
	"sync"
	"syscall"
)

type Command struct {
	Stage            string   `json:"stage"`
	Argv             []string `json:"argv"`
	Status           string   `json:"status"`
	ExitCode         int      `json:"exit_code"`
	Stdout           string   `json:"stdout"`
	Stderr           string   `json:"stderr"`
	PID              int      `json:"pid,omitempty"`
	GroupCleanup     bool     `json:"process_group_cleanup"`
	ProcessReaped    bool     `json:"process_reaped"`
	CaptureTruncated bool     `json:"capture_truncated"`
}
type cappedBuffer struct {
	mu       sync.Mutex
	data     []byte
	limit    int
	overflow bool
}

func (b *cappedBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	n := len(p)
	remaining := b.limit - len(b.data)
	if len(p) > remaining {
		b.overflow = true
		p = p[:remaining]
	}
	b.data = append(b.data, p...)
	return n, nil
}

// RunCommand starts a new process group and always waits after killing it.
// This bounds native children as well as the compiler itself on Linux/WSL.
func RunCommand(ctx context.Context, stage string, argv []string, captureLimit int) Command {
	result := Command{Stage: stage, Argv: argv, ExitCode: -1}
	if len(argv) == 0 {
		result.Status = "unavailable"
		return result
	}
	cmd := exec.Command(argv[0], argv[1:]...)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	out, errout := &cappedBuffer{limit: captureLimit}, &cappedBuffer{limit: captureLimit}
	cmd.Stdout = out
	cmd.Stderr = errout
	if err := cmd.Start(); err != nil {
		result.Status = "unavailable"
		result.Stderr = err.Error()
		return result
	}
	result.PID = cmd.Process.Pid
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	var err error
	select {
	case err = <-done:
		result.Status = "ok"
	case <-ctx.Done():
		_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
		err = <-done
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			result.Status = "timeout"
		} else {
			result.Status = "cancelled"
		}
	}
	// A child may outlive its group leader even after an ordinary exit.
	_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
	result.GroupCleanup = true
	result.ProcessReaped = true
	if cmd.ProcessState != nil {
		result.ExitCode = cmd.ProcessState.ExitCode()
	}
	if result.Status == "ok" && err != nil {
		result.Status = "error"
	}
	result.Stdout = string(out.data)
	result.Stderr = string(errout.data)
	result.CaptureTruncated = out.overflow || errout.overflow
	if result.Status == "ok" && result.CaptureTruncated {
		result.Status = "output_limit"
	}
	return result
}
