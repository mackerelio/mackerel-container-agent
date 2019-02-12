package cmdutil

import (
	"fmt"
	"strings"
	"unicode"
)

// Command represents a plugin/probe command to allow string and []string.
type Command struct {
	command interface{}
}

// UnmarshalYAML defines unmarshaler from YAML.
func (c *Command) UnmarshalYAML(unmarshal func(v interface{}) error) (err error) {
	var s string
	if err := unmarshal(&s); err == nil {
		c.command = s
		return nil
	}
	var ss []string
	if err := unmarshal(&ss); err != nil {
		return err
	}
	c.command = ss
	return nil
}

// String defines the string representation of Command.
func (c Command) String() string {
	switch cmd := c.command.(type) {
	case string:
		return cmd
	case []string:
		args := make([]string, len(cmd))
		for i, arg := range cmd {
			if strings.IndexFunc(arg, func(c rune) bool { return unicode.IsSpace(c) }) >= 0 {
				args[i] = fmt.Sprintf("%q", arg)
				continue
			}
			args[i] = arg
		}
		return strings.Join(args, " ")
	default:
		panic("unexpected command type")
	}
}

// ToArgs returns the command arguments.
func (c Command) ToArgs() []string {
	switch cmd := c.command.(type) {
	case string:
		return []string{"/bin/sh", "-c", cmd}
	case []string:
		return cmd
	default:
		panic("unexpected command type")
	}
}

// IsEmpty returns the command is empty.
func (c Command) IsEmpty() bool {
	switch cmd := c.command.(type) {
	case string:
		return cmd == ""
	case []string:
		return len(cmd) == 0
	default:
		return true
	}
}

// CommandString returns a Command of string.
func CommandString(s string) Command {
	return Command{command: s}
}

// CommandArgs returns a Command of []string.
func CommandArgs(ss []string) Command {
	return Command{command: ss}
}
