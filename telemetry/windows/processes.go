//go:build windows
// +build windows

package windows

import (
	"os"
	"strings"
	"syscall"
	"unsafe"
)

// getProcesses enumerates all processes running on Windows.
func GetProcesses() []string {
	var processes []string

	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	procCreateToolhelp32Snapshot := kernel32.NewProc("CreateToolhelp32Snapshot")
	procProcess32First := kernel32.NewProc("Process32FirstW")
	procProcess32Next := kernel32.NewProc("Process32NextW")
	procCloseHandle := kernel32.NewProc("CloseHandle")

	const TH32CS_SNAPPROCESS = 0x00000002

	type ProcessEntry32 struct {
		Size              uint32
		CntUsage          uint32
		ProcessID         uint32
		DefaultHeapID     uintptr
		ModuleID          uint32
		Threads           uint32
		ParentProcessID   uint32
		PriorityClassBase int32
		Flags             uint32
		ExeFile           [syscall.MAX_PATH]uint16
	}

	snap, _, _ := procCreateToolhelp32Snapshot.Call(uintptr(TH32CS_SNAPPROCESS), 0)
	if snap < 0 {
		return processes
	}
	defer procCloseHandle.Call(snap)

	var entry ProcessEntry32
	entry.Size = uint32(unsafe.Sizeof(entry))

	ret, _, _ := procProcess32First.Call(snap, uintptr(unsafe.Pointer(&entry)))
	if ret == 0 {
		return processes
	}

	for {
		exe := syscall.UTF16ToString(entry.ExeFile[:])
		processes = append(processes, strings.ToLower(exe))

		ret, _, _ = procProcess32Next.Call(snap, uintptr(unsafe.Pointer(&entry)))
		if ret == 0 {
			break
		}
	}

	return processes
}

// GetSelfAndParentNames returns the executable names of the current and parent process.
func GetSelfAndParentNames() (string, string) {
	pid := os.Getpid()
	ppid := os.Getppid()

	self := ""
	parent := ""

	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	procCreateToolhelp32Snapshot := kernel32.NewProc("CreateToolhelp32Snapshot")
	procProcess32First := kernel32.NewProc("Process32FirstW")
	procProcess32Next := kernel32.NewProc("Process32NextW")
	procCloseHandle := kernel32.NewProc("CloseHandle")

	const TH32CS_SNAPPROCESS = 0x00000002

	type ProcessEntry32 struct {
		Size              uint32
		CntUsage          uint32
		ProcessID         uint32
		DefaultHeapID     uintptr
		ModuleID          uint32
		Threads           uint32
		ParentProcessID   uint32
		PriorityClassBase int32
		Flags             uint32
		ExeFile           [syscall.MAX_PATH]uint16
	}

	snap, _, _ := procCreateToolhelp32Snapshot.Call(uintptr(TH32CS_SNAPPROCESS), 0)
	if snap < 0 {
		return self, parent
	}
	defer procCloseHandle.Call(snap)

	var entry ProcessEntry32
	entry.Size = uint32(unsafe.Sizeof(entry))

	ret, _, _ := procProcess32First.Call(snap, uintptr(unsafe.Pointer(&entry)))
	if ret == 0 {
		return self, parent
	}

	for {
		exe := strings.ToLower(syscall.UTF16ToString(entry.ExeFile[:]))
		if int(entry.ProcessID) == pid {
			self = exe
		}
		if int(entry.ProcessID) == ppid {
			parent = exe
		}
		ret, _, _ = procProcess32Next.Call(snap, uintptr(unsafe.Pointer(&entry)))
		if ret == 0 {
			break
		}
	}

	return self, parent
}
