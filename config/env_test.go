package config

import (
	"reflect"
	"testing"
)

func TestMasked(t *testing.T) {
	testCases := []struct {
		name   string
		env    Env
		expect []string
	}{
		{
			name:   "nil",
			env:    nil,
			expect: nil,
		},
		{
			name:   "empty",
			env:    Env{},
			expect: nil,
		},
		{
			name:   "short and long values",
			env:    Env{"FOO=abc", "BAR=ABCDEFGH", "BAZ=12345=678"},
			expect: []string{"FOO=abc", "BAR=ABCD***", "BAZ=1234***"},
		},
		{
			name:   "empty value",
			env:    Env{"FOO="},
			expect: []string{"FOO="},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.env.Masked(); !reflect.DeepEqual(got, tc.expect) {
				t.Fatalf("expect %v, actual %v", tc.expect, got)
			}
		})
	}
}

func TestMaskEnvValue(t *testing.T) {
	testCases := []struct {
		envValue string
		expect   string
	}{
		{
			envValue: "AAA",
			expect:   "AAA",
		},
		{
			envValue: "BBBBBBBBBBB",
			expect:   "BBBB***",
		},
		{
			envValue: "CCC CC",
			expect:   "CCC ***",
		},
	}

	for _, tc := range testCases {
		if MaskEnvValue(tc.envValue) != tc.expect {
			t.Fatalf("expect %s, actual %s", tc.expect, MaskEnvValue(tc.envValue))
		}
	}
}
