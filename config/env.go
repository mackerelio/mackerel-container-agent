package config

import (
	"fmt"
	"sort"
	"strings"
)

// Env represents environment variables
type Env []string

// UnmarshalYAML defines unmarshaler from YAML
func (env *Env) UnmarshalYAML(unmarshal func(v any) error) (err error) {
	var envMap map[string]string
	if err = unmarshal(&envMap); err != nil {
		return err
	}
	if *env, err = buildEnv(envMap); err != nil {
		return err
	}
	return nil
}

func buildEnv(envMap map[string]string) ([]string, error) {
	if len(envMap) == 0 {
		return nil, nil
	}
	env := make([]string, 0, len(envMap))
	for k, v := range envMap {
		if strings.Contains(k, "=") {
			return nil, fmt.Errorf("key of env should not contain \"=\", but got %q", k)
		}
		k = strings.Trim(k, " ")
		if k == "" {
			continue
		}
		env = append(env, k+"="+v)
	}
	sort.Strings(env)
	return env, nil
}

// MaskEnvValue return masked env value ex) FOOBARBAZ -> FOOB***
func MaskEnvValue(s string) string {
	if len(s) < 4 {
		return s
	} else {
		return s[:4] + "***"
	}
}
