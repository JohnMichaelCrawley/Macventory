package command

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

type Runner struct {
	Timeout time.Duration
}

type Result struct {
	Command   string
	Output    string
	Available bool
	Err       error
}

// Function: Run (Command)
func (runner Runner) Run(ctx context.Context, name string, args ...string) Result {
	label := strings.Join(append([]string{name}, args...), " ")
	path, err := exec.LookPath(name)
	if err != nil {
		return Result{Command: label}
	}
	commandCtx, cancel := context.WithTimeout(ctx, runner.Timeout)
	defer cancel()

	cmd := exec.CommandContext(commandCtx, path, args...)
	cmd.Env = append(os.Environ(), "LC_ALL=C", "LANG=C")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err = cmd.Run()

	output := strings.TrimSpace(stdout.String())
	stderrOutput := strings.TrimSpace(stderr.String())
	if err == nil && stderrOutput != "" {
		if output == "" {
			output = stderrOutput
		} else {
			output += "\n" + stderrOutput
		}
	}
	if err != nil {
		if errors.Is(commandCtx.Err(), context.DeadlineExceeded) {
			err = fmt.Errorf(
				"timed out after %s",
				runner.Timeout,
			)
		} else if stderrOutput != "" {
			err = fmt.Errorf(
				"%w: %s",
				err,
				firstLine(stderrOutput),
			)
		}
	}

	return Result{
		Command:   label,
		Output:    output,
		Available: true,
		Err:       err,
	}
}

// Function: First Line
func firstLine(value string) string {
	if i := strings.IndexByte(value, '\n'); i >= 0 {
		return value[:i]
	}
	return value
}
