package checks

import (
	"context"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/vps-inspector/vps-inspector/internal/agent"
)

// SystemCheck collects basic OS and process runtime information.
type SystemCheck struct{}

// NewSystemCheck creates the system information check.
func NewSystemCheck() SystemCheck {
	return SystemCheck{}
}

// ID returns the stable API identifier for this check.
func (SystemCheck) ID() string { return "system.info" }

// Name returns the display name for this check.
func (SystemCheck) Name() string { return "System Information" }

// Description explains what the check measures.
func (SystemCheck) Description() string { return "Collects OS, CPU, hostname, and memory basics." }

// Category groups this check in API metadata.
func (SystemCheck) Category() string { return "system" }

// Run collects basic runtime and host information.
func (c SystemCheck) Run(ctx context.Context) agent.Result {
	started := time.Now().UTC()
	select {
	case <-ctx.Done():
		return agent.TimedResult(c.ID(), agent.StatusFail, 0, "System check timed out.", nil, started, ctx.Err())
	default:
	}

	hostname, _ := os.Hostname()
	details := map[string]any{
		"hostname":      hostname,
		"os":            runtime.GOOS,
		"arch":          runtime.GOARCH,
		"cpuCores":      runtime.NumCPU(),
		"goroutines":    runtime.NumGoroutine(),
		"goVersion":     runtime.Version(),
		"uptimeSeconds": readLinuxUptime(),
		"memory":        readLinuxMemory(),
		"processUser":   os.Getenv("USER"),
		"processShell":  os.Getenv("SHELL"),
	}

	return agent.TimedResult(c.ID(), agent.StatusPass, 92, "System information collected.", details, started, nil)
}

func readLinuxMemory() map[string]uint64 {
	data, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return map[string]uint64{"totalBytes": 0, "availableBytes": 0, "usedBytes": 0}
	}
	values := map[string]uint64{}
	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		kb, err := strconv.ParseUint(fields[1], 10, 64)
		if err != nil {
			continue
		}
		values[strings.TrimSuffix(fields[0], ":")] = kb * 1024
	}
	total := values["MemTotal"]
	available := values["MemAvailable"]
	if available == 0 {
		available = values["MemFree"]
	}
	used := uint64(0)
	if total > available {
		used = total - available
	}
	return map[string]uint64{"totalBytes": total, "availableBytes": available, "usedBytes": used}
}

func readLinuxUptime() float64 {
	data, err := os.ReadFile("/proc/uptime")
	if err != nil {
		return 0
	}
	fields := strings.Fields(string(data))
	if len(fields) == 0 {
		return 0
	}
	value, err := strconv.ParseFloat(fields[0], 64)
	if err != nil {
		return 0
	}
	return value
}
