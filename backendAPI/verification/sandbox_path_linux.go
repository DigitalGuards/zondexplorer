//go:build linux

package verification

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"syscall"
)

// Per-launcher filesystem operations let tests exercise root ownership checks
// without changing process credentials or sharing mutable package state.
type sandboxPathOps struct {
	lstat func(string) (os.FileInfo, error)
	open  func(string) (*os.File, error)
	stat  func(*os.File) (os.FileInfo, error)
}

func validateSandboxPathInfo(info os.FileInfo, path string, directory bool) error {
	if info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("NsJail sandbox launcher path must not contain symlinks: %q", path)
	}
	if directory {
		if !info.IsDir() {
			return fmt.Errorf("NsJail sandbox launcher parent must be a directory: %q", path)
		}
	} else if !info.Mode().IsRegular() {
		return fmt.Errorf("NsJail sandbox launcher must be a regular file: %q", path)
	}
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok || stat == nil {
		return fmt.Errorf("NsJail sandbox launcher path ownership is unavailable: %q", path)
	}
	if stat.Uid != 0 {
		return fmt.Errorf("NsJail sandbox launcher path must be owned by uid 0: %q", path)
	}
	if info.Mode().Perm()&0o022 != 0 {
		return fmt.Errorf("NsJail sandbox launcher path must not be group or world writable: %q", path)
	}
	if !directory {
		if info.Mode().Perm()&0o111 == 0 {
			return fmt.Errorf("NsJail sandbox launcher is not executable: %q", path)
		}
		if info.Size() < 1 || info.Size() > maxNativeExecutableBytes {
			return fmt.Errorf("NsJail sandbox launcher size %d is outside the allowed range", info.Size())
		}
	}
	return nil
}

func (ops sandboxPathOps) checkPath(path string) (os.FileInfo, error) {
	if !filepath.IsAbs(path) {
		return nil, fmt.Errorf("VERIFIER_SANDBOX_BIN must be an absolute path")
	}
	// Walk from / without cleaning away components such as symlink/../.
	// Each prefix must be protected before resolving the next component.
	paths := []string{"/"}
	for i := 1; i < len(path); i++ {
		if path[i] == '/' {
			paths = append(paths, path[:i])
		}
	}
	paths = append(paths, path)
	var info os.FileInfo
	for i, component := range paths {
		var err error
		info, err = ops.lstat(component)
		if err != nil {
			return nil, fmt.Errorf("stat NsJail sandbox launcher path %q: %w", component, err)
		}
		if err := validateSandboxPathInfo(info, component, i < len(paths)-1); err != nil {
			return nil, err
		}
	}
	return info, nil
}

// These infos have already passed validateSandboxPathInfo's Linux stat check.
func sameSandboxPathFile(a, b os.FileInfo) bool {
	left := a.Sys().(*syscall.Stat_t)
	right := b.Sys().(*syscall.Stat_t)
	return left.Dev == right.Dev && left.Ino == right.Ino &&
		a.Size() == b.Size() && left.Mtim == right.Mtim && left.Ctim == right.Ctim
}

// verify keeps the checked descriptor open through launch. Execution uses the
// pathname so AppArmor can attach its profile. Root administrators are trusted
// to preserve the checked path between verification and exec.
func (ops sandboxPathOps) verify(path, expectedSHA256 string) (*os.File, error) {
	pathInfo, err := ops.checkPath(path)
	if err != nil {
		return nil, err
	}
	source, err := ops.open(path)
	if err != nil {
		return nil, fmt.Errorf("open NsJail sandbox launcher without symlink following: %w", err)
	}
	keepOpen := false
	defer func() {
		if !keepOpen {
			_ = source.Close()
		}
	}()
	info, err := ops.stat(source)
	if err != nil {
		return nil, fmt.Errorf("stat open NsJail sandbox launcher: %w", err)
	}
	if err := validateSandboxPathInfo(info, path, false); err != nil {
		return nil, err
	}
	if !sameSandboxPathFile(pathInfo, info) {
		return nil, fmt.Errorf("NsJail sandbox launcher changed while opening")
	}
	if err := validateStaticLinuxAMD64ELF(source, "NsJail sandbox launcher"); err != nil {
		return nil, err
	}
	h := sha256.New()
	written, err := io.Copy(h, io.LimitReader(source, maxNativeExecutableBytes+1))
	if err != nil {
		return nil, fmt.Errorf("hash NsJail sandbox launcher: %w", err)
	}
	afterInfo, err := ops.stat(source)
	if err != nil {
		return nil, fmt.Errorf("stat verified NsJail sandbox launcher: %w", err)
	}
	if err := validateSandboxPathInfo(afterInfo, path, false); err != nil {
		return nil, err
	}
	currentInfo, err := ops.checkPath(path)
	if err != nil {
		return nil, err
	}
	if written != info.Size() || !sameSandboxPathFile(info, afterInfo) || !sameSandboxPathFile(info, currentInfo) {
		return nil, fmt.Errorf("NsJail sandbox launcher changed while verifying")
	}
	actualSHA256 := hex.EncodeToString(h.Sum(nil))
	if actualSHA256 != expectedSHA256 {
		return nil, fmt.Errorf("NsJail sandbox launcher SHA-256 mismatch: want %s, got %s", expectedSHA256, actualSHA256)
	}
	keepOpen = true
	return source, nil
}
