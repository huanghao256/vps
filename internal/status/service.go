package status

import (
	"bufio"
	"context"
	"errors"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Service collects realtime Linux status snapshots for the dashboard.
type Service struct{}

// NewService returns a status collector with process-local uptime tracking.
func NewService() Service {
	return Service{}
}

// Snapshot is the API payload for one realtime status sample.
type Snapshot struct {
	CapturedAt  time.Time   `json:"capturedAt"`
	Health      Health      `json:"health"`
	CPU         CPU         `json:"cpu"`
	Memory      Memory      `json:"memory"`
	Disk        Disk        `json:"disk"`
	Uptime      Uptime      `json:"uptime"`
	Network     Network     `json:"network"`
	Connections Connections `json:"connections"`
	System      System      `json:"system"`
}

// Health summarizes overall load for the top-level status card.
type Health struct {
	Label   string  `json:"label"`
	LoadPct float64 `json:"loadPct"`
}

// CPU contains core count, usage, and Linux load averages.
type CPU struct {
	Cores    int     `json:"cores"`
	UsagePct float64 `json:"usagePct"`
	Load1    float64 `json:"load1"`
	Load5    float64 `json:"load5"`
	Load15   float64 `json:"load15"`
}

// Memory contains total, used, available, and percentage usage values.
type Memory struct {
	TotalBytes     uint64  `json:"totalBytes"`
	UsedBytes      uint64  `json:"usedBytes"`
	AvailableBytes uint64  `json:"availableBytes"`
	UsagePct       float64 `json:"usagePct"`
}

// Disk describes usage for the root filesystem.
type Disk struct {
	Mount      string  `json:"mount"`
	TotalBytes uint64  `json:"totalBytes"`
	UsedBytes  uint64  `json:"usedBytes"`
	FreeBytes  uint64  `json:"freeBytes"`
	UsagePct   float64 `json:"usagePct"`
}

// Uptime contains OS uptime and this process uptime in seconds.
type Uptime struct {
	SystemSeconds float64 `json:"systemSeconds"`
	AppSeconds    float64 `json:"appSeconds"`
}

// Network contains primary IP, total traffic counters, and sampled throughput.
type Network struct {
	IPv4             string  `json:"ipv4"`
	ReceivedBytes    uint64  `json:"receivedBytes"`
	TransmittedBytes uint64  `json:"transmittedBytes"`
	ReceiveMbps      float64 `json:"receiveMbps"`
	TransmitMbps     float64 `json:"transmitMbps"`
}

// Connections contains current TCP and UDP socket counts.
type Connections struct {
	TCP int `json:"tcp"`
	UDP int `json:"udp"`
}

// System contains host-level identity and process/thread counts.
type System struct {
	Hostname  string `json:"hostname"`
	OS        string `json:"os"`
	Threads   int    `json:"threads"`
	Processes int    `json:"processes"`
}

var appStartedAt = time.Now()

// Snapshot samples Linux status data. It returns an error when /proc is unavailable.
func (s Service) Snapshot(ctx context.Context) (Snapshot, error) {
	if _, err := os.Stat("/proc"); err != nil {
		return Snapshot{}, errors.New("system status requires Linux /proc")
	}

	firstCPU, err := readCPUStat()
	if err != nil {
		return Snapshot{}, err
	}
	firstNet := readNetworkCounters()

	timer := time.NewTimer(250 * time.Millisecond)
	select {
	case <-ctx.Done():
		timer.Stop()
		return Snapshot{}, ctx.Err()
	case <-timer.C:
	}

	secondCPU, err := readCPUStat()
	if err != nil {
		return Snapshot{}, err
	}
	secondNet := readNetworkCounters()

	hostname, _ := os.Hostname()
	mem := readMemory()
	disk := readRootDisk(ctx)
	uptime := readUptime()
	load1, load5, load15 := readLoad()
	cores := countCPUCores()
	cpuUsage := cpuUsagePct(firstCPU, secondCPU)

	return Snapshot{
		CapturedAt: time.Now().UTC(),
		Health: Health{
			Label:   healthLabel(cpuUsage, mem.UsagePct, disk.UsagePct),
			LoadPct: maxFloat(cpuUsage, maxFloat(mem.UsagePct, disk.UsagePct)),
		},
		CPU: CPU{
			Cores:    cores,
			UsagePct: cpuUsage,
			Load1:    load1,
			Load5:    load5,
			Load15:   load15,
		},
		Memory: mem,
		Disk:   disk,
		Uptime: Uptime{
			SystemSeconds: uptime,
			AppSeconds:    time.Since(appStartedAt).Seconds(),
		},
		Network: Network{
			IPv4:             primaryIPv4(),
			ReceivedBytes:    secondNet.rx,
			TransmittedBytes: secondNet.tx,
			ReceiveMbps:      rateMbps(firstNet.rx, secondNet.rx, 0.25),
			TransmitMbps:     rateMbps(firstNet.tx, secondNet.tx, 0.25),
		},
		Connections: Connections{
			TCP: countProcNetRows("/proc/net/tcp") + countProcNetRows("/proc/net/tcp6"),
			UDP: countProcNetRows("/proc/net/udp") + countProcNetRows("/proc/net/udp6"),
		},
		System: System{
			Hostname:  hostname,
			OS:        readOSPrettyName(),
			Threads:   countThreads(),
			Processes: countProcesses(),
		},
	}, nil
}

type cpuStat struct {
	total uint64
	idle  uint64
}

func readCPUStat() (cpuStat, error) {
	data, err := os.ReadFile("/proc/stat")
	if err != nil {
		return cpuStat{}, err
	}
	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 5 || fields[0] != "cpu" {
			continue
		}
		var total uint64
		var idle uint64
		for i, field := range fields[1:] {
			value, _ := strconv.ParseUint(field, 10, 64)
			total += value
			if i == 3 || i == 4 {
				idle += value
			}
		}
		return cpuStat{total: total, idle: idle}, nil
	}
	return cpuStat{}, errors.New("cpu row not found in /proc/stat")
}

func cpuUsagePct(a, b cpuStat) float64 {
	total := b.total - a.total
	idle := b.idle - a.idle
	if total == 0 {
		return 0
	}
	return clampPct((1 - float64(idle)/float64(total)) * 100)
}

func readMemory() Memory {
	values := readMeminfo()
	total := values["MemTotal"]
	available := values["MemAvailable"]
	if available == 0 {
		available = values["MemFree"]
	}
	used := uint64(0)
	if total > available {
		used = total - available
	}
	return Memory{
		TotalBytes:     total,
		UsedBytes:      used,
		AvailableBytes: available,
		UsagePct:       percent(used, total),
	}
}

func readMeminfo() map[string]uint64 {
	data, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return map[string]uint64{}
	}
	values := map[string]uint64{}
	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		value, _ := strconv.ParseUint(fields[1], 10, 64)
		values[strings.TrimSuffix(fields[0], ":")] = value * 1024
	}
	return values
}

func readRootDisk(ctx context.Context) Disk {
	cmd := exec.CommandContext(ctx, "df", "-B1", "/")
	output, err := cmd.Output()
	if err != nil {
		return Disk{Mount: "/"}
	}
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) < 2 {
		return Disk{Mount: "/"}
	}
	fields := strings.Fields(lines[1])
	if len(fields) < 6 {
		return Disk{Mount: "/"}
	}
	total, _ := strconv.ParseUint(fields[1], 10, 64)
	used, _ := strconv.ParseUint(fields[2], 10, 64)
	free, _ := strconv.ParseUint(fields[3], 10, 64)
	return Disk{
		Mount:      fields[5],
		TotalBytes: total,
		UsedBytes:  used,
		FreeBytes:  free,
		UsagePct:   percent(used, total),
	}
}

func readUptime() float64 {
	data, err := os.ReadFile("/proc/uptime")
	if err != nil {
		return 0
	}
	fields := strings.Fields(string(data))
	if len(fields) == 0 {
		return 0
	}
	value, _ := strconv.ParseFloat(fields[0], 64)
	return value
}

func readLoad() (float64, float64, float64) {
	data, err := os.ReadFile("/proc/loadavg")
	if err != nil {
		return 0, 0, 0
	}
	fields := strings.Fields(string(data))
	if len(fields) < 3 {
		return 0, 0, 0
	}
	load1, _ := strconv.ParseFloat(fields[0], 64)
	load5, _ := strconv.ParseFloat(fields[1], 64)
	load15, _ := strconv.ParseFloat(fields[2], 64)
	return load1, load5, load15
}

type networkCounters struct {
	rx uint64
	tx uint64
}

func readNetworkCounters() networkCounters {
	data, err := os.ReadFile("/proc/net/dev")
	if err != nil {
		return networkCounters{}
	}
	var counters networkCounters
	for _, line := range strings.Split(string(data), "\n") {
		if !strings.Contains(line, ":") {
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		if strings.TrimSpace(parts[0]) == "lo" {
			continue
		}
		fields := strings.Fields(parts[1])
		if len(fields) < 16 {
			continue
		}
		rx, _ := strconv.ParseUint(fields[0], 10, 64)
		tx, _ := strconv.ParseUint(fields[8], 10, 64)
		counters.rx += rx
		counters.tx += tx
	}
	return counters
}

func primaryIPv4() string {
	interfaces, err := net.Interfaces()
	if err != nil {
		return ""
	}
	for _, item := range interfaces {
		if item.Flags&net.FlagUp == 0 || item.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, err := item.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			var ip net.IP
			switch value := addr.(type) {
			case *net.IPNet:
				ip = value.IP
			case *net.IPAddr:
				ip = value.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			if ipv4 := ip.To4(); ipv4 != nil {
				return ipv4.String()
			}
		}
	}
	return ""
}

func countCPUCores() int {
	data, err := os.ReadFile("/proc/cpuinfo")
	if err != nil {
		return 0
	}
	count := 0
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "processor") {
			count++
		}
	}
	return count
}

func countProcesses() int {
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return 0
	}
	count := 0
	for _, entry := range entries {
		if entry.IsDir() && isDigits(entry.Name()) {
			count++
		}
	}
	return count
}

func countThreads() int {
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return 0
	}
	total := 0
	for _, entry := range entries {
		if !entry.IsDir() || !isDigits(entry.Name()) {
			continue
		}
		tasks, err := os.ReadDir(filepath.Join("/proc", entry.Name(), "task"))
		if err == nil {
			total += len(tasks)
		}
	}
	return total
}

func readOSPrettyName() string {
	data, err := os.ReadFile("/etc/os-release")
	if err != nil {
		return "Linux"
	}
	for _, line := range strings.Split(string(data), "\n") {
		key, value, ok := strings.Cut(line, "=")
		if ok && key == "PRETTY_NAME" {
			return strings.Trim(value, `"`)
		}
	}
	return "Linux"
}

func countProcNetRows(path string) int {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) <= 1 {
		return 0
	}
	return len(lines) - 1
}

func isDigits(value string) bool {
	for _, char := range value {
		if char < '0' || char > '9' {
			return false
		}
	}
	return value != ""
}

func healthLabel(cpu, mem, disk float64) string {
	load := maxFloat(cpu, maxFloat(mem, disk))
	switch {
	case load < 50:
		return "运行流畅"
	case load < 80:
		return "负载偏高"
	default:
		return "需要关注"
	}
}

func rateMbps(previous, current uint64, seconds float64) float64 {
	if current <= previous || seconds <= 0 {
		return 0
	}
	return float64(current-previous) * 8 / 1000 / 1000 / seconds
}

func percent(value, total uint64) float64 {
	if total == 0 {
		return 0
	}
	return clampPct(float64(value) / float64(total) * 100)
}

func clampPct(value float64) float64 {
	if value < 0 {
		return 0
	}
	if value > 100 {
		return 100
	}
	return value
}

func maxFloat(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
