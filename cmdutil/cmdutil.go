package cmdutil

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"time"

	"github.com/Songmu/timeout"
)

var (
	defaultTimeoutDuration = 30 * time.Second
	timeoutKillAfter       = 10 * time.Second
	errTimedOut            = errors.New("command timed out")
)

// RunCommand executes command with context
func RunCommand(ctx context.Context, command Command, user string, env []string, timeoutDuration time.Duration) (string, string, int, error) {
	args := command.ToArgs()
	if user != "" {
		args = append([]string{"sudo", "-Eu", user}, args...)
	}
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Env = append(os.Environ(), env...)
	outbuf, errbuf := new(bytes.Buffer), new(bytes.Buffer)
	cmd.Stdout, cmd.Stderr = outbuf, errbuf
	tio := &timeout.Timeout{
		Cmd:       cmd,
		Duration:  defaultTimeoutDuration,
		KillAfter: timeoutKillAfter,
	}
	if timeoutDuration > 0 {
		tio.Duration = timeoutDuration
	}
	exitStatus, err := tio.RunContext(ctx)
	exitCode := -1
	if err != nil {
		if terr, ok := err.(*timeout.Error); ok {
			exitCode = terr.ExitCode
		}
	} else {
		exitCode = exitStatus.GetChildExitCode()
		if exitStatus.IsTimedOut() && exitStatus.Signaled {
			err = errTimedOut
		}
	}
	return outbuf.String(), errbuf.String(), exitCode, err
}
