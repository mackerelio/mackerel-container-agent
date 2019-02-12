package cmdutil

import (
	"context"
	"strings"
	"syscall"
	"testing"
	"time"
)

func TestRunCommand(t *testing.T) {
	testCases := []struct {
		name           string
		command        Command
		user           string
		env            []string
		timeout        time.Duration
		stdout, stderr string
		exitCode       int
		err            error
	}{
		{
			name:    "echo 1",
			command: CommandString("echo 1"),
			stdout:  "1\n",
		},
		{
			name:    "stdout stderr",
			command: CommandString("echo foobar && echo quxquux >&2"),
			stdout:  "foobar\n",
			stderr:  "quxquux\n",
		},
		{
			name:     "exit status",
			command:  CommandString("exit 42"),
			exitCode: 42,
			err:      nil,
		},
		{
			name:    "environment variables",
			command: CommandString("echo $FOO; echo $BAR >&2"),
			env:     []string{"FOO=foo bar", "BAR=qux quux"},
			stdout:  "foo bar\n",
			stderr:  "qux quux\n",
		},
		{
			name:     "timeout",
			command:  CommandString("sleep 3"),
			timeout:  100 * time.Millisecond,
			exitCode: 128 + int(syscall.SIGTERM),
			err:      errTimedOut,
		},
		{
			name:     "command not found",
			command:  CommandString("notfound"),
			exitCode: 127,
			stderr:   " not found\n",
		},
		{
			name:    "command args",
			command: CommandArgs([]string{"echo", "foo", "bar"}),
			stdout:  "foo bar\n",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			stdout, stderr, exitCode, err := RunCommand(ctx, tc.command, tc.user, tc.env, tc.timeout)
			if stdout != tc.stdout {
				t.Errorf("invalid stdout (out: %q, expect: %q)", stdout, tc.stdout)
			}
			if tc.stderr == "" && stderr != "" || !strings.Contains(stderr, tc.stderr) {
				t.Errorf("invalid stderr (out: %q, expect: %q)", stderr, tc.stderr)
			}
			if exitCode != tc.exitCode {
				t.Errorf("exitCode should be %d, but: %d", tc.exitCode, exitCode)
			}
			if err != tc.err {
				t.Errorf("err should be %v but: %v", tc.err, err)
			}
		})
	}
}
