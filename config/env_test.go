package config

import (
	"testing"
)

func TestFilterMackerelEnv(t *testing.T) {
	env := []string{"FOO", "MACKEREL_FOO", "FOO_MACKEREL_BAR"}
	mackerel_env := FilterMackerelEnv(env)
	if len(mackerel_env) != 1 {
		t.Error("length of mackerel_env should be 1")
	}
	if mackerel_env[0] != "MACKEREL_FOO" {
		t.Error("MACKEREL_FOO should be mackerel_env")
	}
}
