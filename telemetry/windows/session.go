//go:build windows
// +build windows

package windows

import (
	"fmt"
	"syscall"
	"unsafe"
)

// GetLogonSessions retrieves the list of logon sessions safely.
func GetLogonSessions() []string {
	var sessions []string

	secur32 := syscall.NewLazyDLL("secur32.dll")
	advapi32 := syscall.NewLazyDLL("advapi32.dll")

	procLsaEnumerateLogonSessions := secur32.NewProc("LsaEnumerateLogonSessions")
	procLsaGetLogonSessionData := secur32.NewProc("LsaGetLogonSessionData")
	procLsaFreeReturnBuffer := secur32.NewProc("LsaFreeReturnBuffer")
	procLsaNtStatusToWinError := advapi32.NewProc("LsaNtStatusToWinError")

	// LsaEnumerateLogonSessions(OUT PULONG LogonSessionCount, OUT PLUID *LogonSessionList)
	var count uint32
	var luidListPtr uintptr

	status, _, _ := procLsaEnumerateLogonSessions.Call(
		uintptr(unsafe.Pointer(&count)),
		uintptr(unsafe.Pointer(&luidListPtr)),
	)
	if status != 0 {
		// convert NTSTATUS to Win32 error if desired
		_, _, _ = procLsaNtStatusToWinError.Call(status)
		return sessions
	}
	if luidListPtr == 0 || count == 0 {
		return sessions
	}
	// ensure we free the returned LUID list when done
	defer func() {
		if luidListPtr != 0 {
			procLsaFreeReturnBuffer.Call(luidListPtr)
		}
	}()

	// Define LUID and SECURITY_LOGON_SESSION_DATA as used by these APIs
	type LUID struct {
		LowPart  uint32
		HighPart int32
	}
	type UNICODE_STRING struct {
		Length        uint16
		MaximumLength uint16
		Buffer        *uint16
	}
	type SECURITY_LOGON_SESSION_DATA struct {
		Size              uint32
		LogonId           LUID
		UserName          UNICODE_STRING
		LogonDomain       UNICODE_STRING
		AuthenticationPkg UNICODE_STRING
		LogonType         uint32
		Session           uint32
		Sid               uintptr
		LogonTime         int64
		LogonServer       UNICODE_STRING
		DnsDomainName     UNICODE_STRING
		Upn               UNICODE_STRING
	}

	// Safe upper bound for UTF-16 buffers (prevents accidentally huge slices)
	const maxUTF16Len = 1 << 20 // 1Mi 16-bit entries => ~2MiB

	luidSize := unsafe.Sizeof(LUID{})

	for i := uint32(0); i < count; i++ {
		// pointer to the i-th LUID in the returned array
		luidAddr := luidListPtr + uintptr(i)*luidSize
		if luidAddr == 0 {
			continue
		}
		pLuid := (*LUID)(unsafe.Pointer(luidAddr))

		var dataPtr uintptr
		status, _, _ = procLsaGetLogonSessionData.Call(
			uintptr(unsafe.Pointer(pLuid)),
			uintptr(unsafe.Pointer(&dataPtr)),
		)
		if status != 0 || dataPtr == 0 {
			// ignore this entry if it fails
			continue
		}

		// ensure we free this data block when done
		func() {
			defer procLsaFreeReturnBuffer.Call(dataPtr)

			dataStruct := (*SECURITY_LOGON_SESSION_DATA)(unsafe.Pointer(dataPtr))

			// helper to convert UNICODE_STRING to Go string safely
			utf16From := func(us UNICODE_STRING) string {
				if us.Buffer == nil || us.Length == 0 {
					return ""
				}
				// length is in bytes; convert to uint32 count of uint16
				count := int(us.Length / 2)
				if count <= 0 {
					return ""
				}
				if count > maxUTF16Len {
					count = maxUTF16Len
				}
				// build a slice header safely
				buf := (*[maxUTF16Len]uint16)(unsafe.Pointer(us.Buffer))[:count:count]
				return syscall.UTF16ToString(buf)
			}

			user := utf16From(dataStruct.UserName)
			domain := utf16From(dataStruct.LogonDomain)

			if user != "" {
				if domain == "" {
					sessions = append(sessions, user)
				} else {
					sessions = append(sessions, fmt.Sprintf("%s\\%s", domain, user))
				}
			}
		}()
	}

	return sessions
}
