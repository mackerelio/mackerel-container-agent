package config

import (
	"testing"
)

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
