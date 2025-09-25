package telemetry

import (
	"fmt"
	"net"
	"os"
	"runtime"
	"sort"
	"strings"
	"syscall"
	"time"
	"unsafe"
)

// ==============================
// Structs & Constants
// ==============================

type Telemetry struct {
	Username          string   `json:"username"`
	Hostname          string   `json:"hostname"`
	OS                string   `json:"os"`
	TimeOfDay         string   `json:"time_of_day"`
	DayOfWeek         string   `json:"day_of_week"`
	WorkingHours      bool     `json:"working_hours"`
	IdleSeconds       int      `json:"idle_seconds"`
	ActiveWindow      string   `json:"active_window"`
	ProcessList       []string `json:"process_list"`
	NetworkTraffic    []string `json:"network_traffic"`
	Drivers           []string `json:"drivers"`
	NetworkInfo       []string `json:"network_info"`
	USBDevices        []string `json:"usb_devices"`
	UptimeMinutes     int      `json:"uptime_minutes"`
	LogonSessions     []string `json:"logon_sessions"`
	BaselineProcesses []string `json:"baseline_processes"`
	DeviationCount    int      `json:"deviation_count"`
	SelfProcess       string   `json:"self_process"`
	ParentProcess     string   `json:"parent_process"`
}

const (
	TCP_TABLE_OWNER_PID_ALL = 5
	BASELINE_SNAPSHOTS      = 3
	BASELINE_WAIT           = 5 * time.Second
	TELEMETRY_SNAPSHOTS     = 2
	TELEMETRY_INTERVAL      = 8 * time.Second
)

type MIB_TCPROW_OWNER_PID struct {
	State      uint32
	LocalAddr  uint32
	LocalPort  uint32
	RemoteAddr uint32
	RemotePort uint32
	OwningPid  uint32
}

// ==============================
// Baseline + Telemetry Collection
// ==============================

func BuildBaseline(selfProc, parentProc string) []string {
	freq := make(map[string]int)
	for i := 0; i < BASELINE_SNAPSHOTS; i++ {
		procs := getProcesses()
		for _, p := range procs {
			if p == selfProc || p == parentProc {
				continue
			}
			freq[p]++
		}
		time.Sleep(BASELINE_WAIT)
	}
	var baseline []string
	for proc, count := range freq {
		if count >= BASELINE_SNAPSHOTS-1 {
			baseline = append(baseline, proc)
		}
	}
	sort.Strings(baseline)
	return baseline
}

func Collect(baseline []string, selfProc, parentProc string) Telemetry {
	freq := make(map[string]int)
	for i := 0; i < TELEMETRY_SNAPSHOTS; i++ {
		procs := getProcesses()
		for _, p := range procs {
			freq[p]++
		}
		time.Sleep(TELEMETRY_INTERVAL)
	}

	var current []string
	deviation := 0
	for proc, count := range freq {
		if proc == selfProc || proc == parentProc {
			continue
		}
		if count >= TELEMETRY_SNAPSHOTS-1 {
			current = append(current, proc)
			if !contains(baseline, proc) {
				deviation++
			}
		}
	}
	sort.Strings(current)

	now := time.Now()
	working := now.Hour() >= 9 && now.Hour() <= 17

	return Telemetry{
		Username:          os.Getenv("USERNAME"),
		Hostname:          getHostname(),
		OS:                runtime.GOOS + " " + runtime.GOARCH,
		TimeOfDay:         now.Format("15:04:05"),
		DayOfWeek:         now.Weekday().String(),
		WorkingHours:      working,
		IdleSeconds:       getIdleSeconds(),
		ActiveWindow:      getActiveWindowTitle(selfProc, parentProc),
		ProcessList:       current,
		Drivers:           getDrivers(),
		NetworkInfo:       getNetworkInfo(),
		NetworkTraffic:    getNetworkTraffic(),
		USBDevices:        getUSBDevices(),
		UptimeMinutes:     getUptimeMinutes(),
		LogonSessions:     getLogonSessions(),
		BaselineProcesses: baseline,
		DeviationCount:    deviation,
		SelfProcess:       selfProc,
		ParentProcess:     parentProc,
	}
}

func GetSelfAndParentNames() (string, string) {
	pid := os.Getpid()
	ppid := os.Getppid()

	self := ""
	parent := ""

	// Reuse snapshot logic
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

// ==============================
// Helpers 
// ==============================

func getProcesses() []string {
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

func getDrivers() []string {
	var drivers []string

	psapi := syscall.NewLazyDLL("psapi.dll")
	procEnumDeviceDrivers := psapi.NewProc("EnumDeviceDrivers")
	procGetDeviceDriverBaseName := psapi.NewProc("GetDeviceDriverBaseNameW")

	var cbNeeded uint32
	var arr [1024]uintptr

	ret, _, _ := procEnumDeviceDrivers.Call(
		uintptr(unsafe.Pointer(&arr[0])),
		uintptr(len(arr))*unsafe.Sizeof(arr[0]),
		uintptr(unsafe.Pointer(&cbNeeded)),
	)
	if ret == 0 {
		return drivers
	}

	count := cbNeeded / uint32(unsafe.Sizeof(arr[0]))
	for i := 0; i < int(count); i++ {
		var name [syscall.MAX_PATH]uint16
		n, _, _ := procGetDeviceDriverBaseName.Call(
			arr[i],
			uintptr(unsafe.Pointer(&name[0])),
			uintptr(len(name)),
		)
		if n == 0 {
			continue
		}
		drivers = append(drivers, strings.ToLower(syscall.UTF16ToString(name[:])))
	}

	return drivers
}

func getNetworkTraffic() []string {
	var results []string

	iphlpapi := syscall.NewLazyDLL("iphlpapi.dll")
	procGetExtendedTcpTable := iphlpapi.NewProc("GetExtendedTcpTable")

	var bufSize uint32
	procGetExtendedTcpTable.Call(
		0,
		uintptr(unsafe.Pointer(&bufSize)),
		0,
		uintptr(syscall.AF_INET),
		uintptr(TCP_TABLE_OWNER_PID_ALL),
		0,
	)

	buf := make([]byte, bufSize)
	ret, _, _ := procGetExtendedTcpTable.Call(
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(unsafe.Pointer(&bufSize)),
		0,
		uintptr(syscall.AF_INET),
		uintptr(TCP_TABLE_OWNER_PID_ALL),
		0,
	)
	if ret != 0 {
		return results
	}

	count := *(*uint32)(unsafe.Pointer(&buf[0]))
	offset := unsafe.Sizeof(count)

	for i := 0; i < int(count); i++ {
		row := (*MIB_TCPROW_OWNER_PID)(unsafe.Pointer(&buf[offset]))
		localPort := ntohs(uint16(row.LocalPort >> 16))
		remotePort := ntohs(uint16(row.RemotePort >> 16))

		localIP := net.IPv4(
			byte(row.LocalAddr),
			byte(row.LocalAddr>>8),
			byte(row.LocalAddr>>16),
			byte(row.LocalAddr>>24),
		).String()

		remoteIP := net.IPv4(
			byte(row.RemoteAddr),
			byte(row.RemoteAddr>>8),
			byte(row.RemoteAddr>>16),
			byte(row.RemoteAddr>>24),
		).String()

		results = append(results,
			fmt.Sprintf("%s:%d -> %s:%d (PID %d, State %d)",
				localIP, localPort,
				remoteIP, remotePort,
				row.OwningPid, row.State,
			),
		)

		offset += unsafe.Sizeof(*row)
	}

	return results
}

func ntohs(n uint16) uint16 {
	return (n<<8)&0xff00 | n>>8
}

func getIdleSeconds() int {
	type LASTINPUTINFO struct {
		cbSize uint32
		dwTime uint32
	}
	user32 := syscall.NewLazyDLL("user32.dll")
	procGetLastInputInfo := user32.NewProc("GetLastInputInfo")
	lii := LASTINPUTINFO{cbSize: 8}
	procGetLastInputInfo.Call(uintptr(unsafe.Pointer(&lii)))
	tickCount, _, _ := syscall.Syscall(syscall.NewLazyDLL("kernel32.dll").NewProc("GetTickCount").Addr(), 0, 0, 0, 0)
	return int((tickCount - uintptr(lii.dwTime)) / 1000)
}

func getActiveWindowTitle(selfProc, parentProc string) string {
	user32 := syscall.NewLazyDLL("user32.dll")
	getForegroundWindow := user32.NewProc("GetForegroundWindow")
	getWindowTextW := user32.NewProc("GetWindowTextW")
	hwnd, _, _ := getForegroundWindow.Call()
	buf := make([]uint16, 256)
	getWindowTextW.Call(hwnd, uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)))
	title := syscall.UTF16ToString(buf)
	titleLower := strings.ToLower(title)
	if strings.Contains(titleLower, "command prompt") ||
		strings.Contains(titleLower, "powershell") ||
		strings.Contains(titleLower, selfProc) ||
		strings.Contains(titleLower, parentProc) {
		return "none"
	}
	return title
}

func getUSBDevices() []string {
	var devices []string

	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	procGetLogicalDriveStrings := kernel32.NewProc("GetLogicalDriveStringsW")
	procGetDriveType := kernel32.NewProc("GetDriveTypeW")

	buf := make([]uint16, 254)
	n, _, _ := procGetLogicalDriveStrings.Call(
		uintptr(len(buf)),
		uintptr(unsafe.Pointer(&buf[0])),
	)
	if n == 0 {
		return devices
	}

	drives := syscall.UTF16ToString(buf[:n])
	for _, drive := range strings.Split(drives, "\x00") {
		if drive == "" {
			continue
		}

		dtype, _, _ := procGetDriveType.Call(uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(drive))))
		switch dtype {
		case 2:
			devices = append(devices, fmt.Sprintf("Removable: %s", drive))
		case 3:
			devices = append(devices, fmt.Sprintf("Fixed: %s", drive))
		case 4:
			devices = append(devices, fmt.Sprintf("Network: %s", drive))
		case 5:
			devices = append(devices, fmt.Sprintf("CD-ROM: %s", drive))
		case 6:
			devices = append(devices, fmt.Sprintf("RAM Disk: %s", drive))
		default:
			devices = append(devices, fmt.Sprintf("Unknown: %s", drive))
		}
	}

	return devices
}

func getNetworkInfo() []string {
	var info []string

	ifaces, err := net.Interfaces()
	if err != nil {
		return info
	}

	for _, iface := range ifaces {
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			info = append(info, fmt.Sprintf("%s: %s", iface.Name, addr.String()))
		}
	}

	return info
}

func getUptimeMinutes() int {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	procGetTickCount64 := kernel32.NewProc("GetTickCount64")

	uptimeMS, _, _ := procGetTickCount64.Call()
	return int(uptimeMS / 1000 / 60)
}

func getLogonSessions() []string {
	var sessions []string

	wtsapi32 := syscall.NewLazyDLL("wtsapi32.dll")
	procWTSEnumerateSessions := wtsapi32.NewProc("WTSEnumerateSessionsW")
	procWTSQuerySessionInformation := wtsapi32.NewProc("WTSQuerySessionInformationW")
	procWTSFreeMemory := wtsapi32.NewProc("WTSFreeMemory")

	var ppSessionInfo uintptr
	var count uint32

	ret, _, _ := procWTSEnumerateSessions.Call(
		0,
		0,
		1,
		uintptr(unsafe.Pointer(&ppSessionInfo)),
		uintptr(unsafe.Pointer(&count)),
	)
	if ret == 0 {
		return sessions
	}
	defer procWTSFreeMemory.Call(ppSessionInfo)

	type WTS_SESSION_INFO struct {
		SessionID       uint32
		pWinStationName *uint16
		State           uint32
	}

	const WTSUserName = 5

	entrySize := unsafe.Sizeof(WTS_SESSION_INFO{})
	for i := 0; i < int(count); i++ {
		entry := (*WTS_SESSION_INFO)(unsafe.Pointer(ppSessionInfo + uintptr(i)*entrySize))

		var buffer uintptr
		var bytesReturned uint32
		ret, _, _ = procWTSQuerySessionInformation.Call(
			0,
			uintptr(entry.SessionID),
			uintptr(WTSUserName),
			uintptr(unsafe.Pointer(&buffer)),
			uintptr(unsafe.Pointer(&bytesReturned)),
		)
		if ret != 0 && buffer != 0 {
			username := syscall.UTF16ToString((*[1 << 16]uint16)(unsafe.Pointer(buffer))[:bytesReturned/2])
			if username != "" {
				sessions = append(sessions, fmt.Sprintf("Session %d: %s", entry.SessionID, username))
			}
			procWTSFreeMemory.Call(buffer)
		}
	}

	return sessions
}

func getHostname() string {
	host, err := os.Hostname()
	if err != nil {
		return "unknown"
	}
	return host
}

func contains(list []string, target string) bool {
	for _, item := range list {
		if item == target {
			return true
		}
	}
	return false
}
