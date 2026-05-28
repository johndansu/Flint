package ipc

import (
	"crypto/md5"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// SocketPath returns the daemon socket path for a given workspace directory.
// Format: ~/.flint/daemon-<workspace-hash>.sock
func SocketPath(flintDir, workspaceRoot string) string {
	hash := fmt.Sprintf("%x", md5.Sum([]byte(workspaceRoot)))[:8]
	name := fmt.Sprintf("daemon-%s.sock", hash)
	return filepath.Join(flintDir, name)
}

// removeSocket removes an existing socket file (ignores not-exist errors).
func removeSocket(path string) error {
	err := os.Remove(path)
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

// IsSupported reports whether Unix domain sockets are available on this platform.
// On Windows they require Windows 10 1803+ but work in WSL2 always.
func IsSupported() bool {
	return runtime.GOOS != "windows"
}
