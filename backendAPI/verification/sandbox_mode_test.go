package verification

import (
	"strings"
	"testing"
)

func TestParseSandboxExecMode(t *testing.T) {
	for _, tc := range []struct {
		value string
		want  string
	}{
		{"", sandboxExecModeSealed},
		{"sealed", sandboxExecModeSealed},
		{"path", sandboxExecModePath},
		{"PATH", ""},
		{"Sealed", ""},
		{"auto", ""},
		{"disabled", ""},
		{" path ", ""},
		{" ", ""},
	} {
		t.Run(tc.value, func(t *testing.T) {
			got, err := parseSandboxExecMode(tc.value)
			if tc.want == "" {
				if err == nil || !strings.Contains(err.Error(), "VERIFIER_SANDBOX_EXEC_MODE must be sealed or path") {
					t.Fatalf("parseSandboxExecMode(%q) = %q, %v, want explicit mode error", tc.value, got, err)
				}
			} else if err != nil || got != tc.want {
				t.Fatalf("parseSandboxExecMode(%q) = %q, %v, want %q", tc.value, got, err, tc.want)
			}
		})
	}
}
