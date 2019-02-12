package metric

import "testing"

func Test_SanitizeMetricKey(t *testing.T) {
	testCases := []struct {
		src, expected string
	}{
		{"", ""},
		{"foo.bar", "foo_bar"},
		{"foo-^bar.qux#&quux", "foo-_bar_qux__quux"},
	}
	for _, testCase := range testCases {
		got := SanitizeMetricKey(testCase.src)
		if got != testCase.expected {
			t.Errorf("SanitizeMetricKey(%q) should be %q but got %q", testCase.src, testCase.expected, got)
		}
	}
}
