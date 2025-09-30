//go:build linux
// +build linux

package linux

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// GetSelfAndParentNames returns names of current and parent processes.
func GetSelfAndParentNames() (string, string) {
	self, _ := os.Readlink("/proc/self/exe")

	ppid := os.Getppid()
	parentPath, _ := os.Readlink(filepath.Join("/proc", strconv.Itoa(ppid), "exe"))

	return filepath.Base(self), filepath.Base(parentPath)
}

// GetProcesses enumerates processes from /proc and returns names (exported).
func GetProcesses() []string {
	var procs []string
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return procs
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		pid := e.Name()
		// skip non-numeric directories
		if _, err := strconv.Atoi(pid); err != nil {
			continue
		}
		// preferred: /proc/<pid>/comm
		commPath := filepath.Join("/proc", pid, "comm")
		if b, err := os.ReadFile(commPath); err == nil {
			name := strings.TrimSpace(string(b))
			if name != "" {
				procs = append(procs, name)
				continue
			}
		}
		// fallback: read exe symlink
		exePath := filepath.Join("/proc", pid, "exe")
		if exe, err := os.Readlink(exePath); err == nil {
			procs = append(procs, filepath.Base(exe))
		}
	}
	return procs
}
