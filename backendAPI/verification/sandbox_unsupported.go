//go:build !linux

package verification

import (
	"bytes"
	"context"
	"errors"
	"os"
)

type sandboxLauncher struct {
	SHA256       string
	PolicySHA256 string
}

func newSandboxLauncher(_, _, execMode string) (*sandboxLauncher, error) {
	if _, err := parseSandboxExecMode(execMode); err != nil {
		return nil, err
	}
	return nil, errors.New("NsJail compiler sandbox requires Linux")
}

func (s *sandboxLauncher) run(
	_ context.Context,
	_ *os.File,
	_ []string,
	_ *bytes.Reader,
	_ int,
	_ int,
) ([]byte, []byte, error) {
	return nil, nil, errors.New("NsJail compiler sandbox requires Linux")
}

func (s *sandboxLauncher) Close() error { return nil }
