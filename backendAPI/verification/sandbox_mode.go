package verification

import "fmt"

const (
	sandboxExecModeSealed = "sealed"
	sandboxExecModePath   = "path"
)

func parseSandboxExecMode(value string) (string, error) {
	switch value {
	case "", sandboxExecModeSealed:
		return sandboxExecModeSealed, nil
	case sandboxExecModePath:
		return sandboxExecModePath, nil
	default:
		return "", fmt.Errorf("VERIFIER_SANDBOX_EXEC_MODE must be sealed or path, got %q", value)
	}
}
