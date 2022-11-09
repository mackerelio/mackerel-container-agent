package config

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

// Env represents environment variables
type Env []string

// UnmarshalYAML defines unmarshaler from YAML
func (env *Env) UnmarshalYAML(unmarshal func(v interface{}) error) (err error) {
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

// Return env starting with "MACKEREL"
func FilterMackerelEnv(env Env) Env {
	var mackerel_env []string
	r := regexp.MustCompile(`^MACKEREL`)
	for _, v := range env {
		if r.MatchString(v) {
			mackerel_env = append(mackerel_env, v)
		}
	}
	return mackerel_env
}
