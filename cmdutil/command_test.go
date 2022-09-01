package cmdutil

import (
	"reflect"
	"testing"

	"github.com/go-yaml/yaml"
)

var commandTestCases = []struct {
	name     string
	src      string
	toString string
	toArgs   []string
	isEmpty  bool
}{
	{
		name:    "empty",
		src:     `command:`,
		isEmpty: true,
	},
	{
		name:     "one line string",
		src:      `command: echo hello`,
		toString: `echo hello`,
		toArgs:   []string{"/bin/sh", "-c", "echo hello"},
		isEmpty:  false,
	},
	{
		name:     "one line slice of string",
		src:      `command: [ "echo", "hello world" ]`,
		toString: `echo "hello world"`,
		toArgs:   []string{"echo", "hello world"},
		isEmpty:  false,
	},
	{
		name: "multi-line slice of string",
		src: `command:
  - echo
  - hello
  - world`,
		toString: `echo hello world`,
		toArgs:   []string{"echo", "hello", "world"},
		isEmpty:  false,
	},
}

func TestCommand(t *testing.T) {
	for _, tc := range commandTestCases {
		t.Run(tc.name, func(t *testing.T) {
			var conf struct {
				Command Command `yaml:"command"`
			}
			err := yaml.Unmarshal([]byte(tc.src), &conf)
			if err != nil {
				t.Fatalf("should not raise error: %v", err)
			}
			if got := conf.Command.IsEmpty(); got != tc.isEmpty {
				t.Errorf("IsEmpty(): expect %#v, got %#v", tc.isEmpty, got)
			}
			if !conf.Command.IsEmpty() {
				if s := conf.Command.String(); s != tc.toString {
					t.Errorf("String(): expect %#v, got %#v", tc.toString, s)
				}
				if got := conf.Command.ToArgs(); !reflect.DeepEqual(got, tc.toArgs) {
					t.Errorf("ToArgs(): expect %#v, got %#v", tc.toArgs, got)
				}
			}
		})
	}
}
