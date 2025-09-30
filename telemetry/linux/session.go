//go:build linux
// +build linux

package linux

import (
	"encoding/binary"
	"os"
	"unsafe"
)

// structure based on /usr/include/utmp.h
type utmp struct {
	Type    int16
	Pad_c   [2]byte
	Pid     int32
	Line    [32]byte
	ID      [4]byte
	User    [32]byte
	Host    [256]byte
	Exit    [4]byte
	Session int32
	Tv      [16]byte
	AddrV6  [16]byte
	Unused  [20]byte
}

// GetLogonSessions parses /var/run/utmp and returns usernames.
func GetLogonSessions() []string {
	file, err := os.Open("/var/run/utmp")
	if err != nil {
		return []string{}
	}
	defer file.Close()

	var sessions []string
	entrySize := int(unsafe.Sizeof(utmp{}))
	buf := make([]byte, entrySize)

	for {
		_, err := file.Read(buf)
		if err != nil {
			break
		}
		var u utmp
		binary.Read(file, binary.LittleEndian, &u)
		user := string(u.User[:])
		if user != "" {
			sessions = append(sessions, user)
		}
	}
	return sessions
}
