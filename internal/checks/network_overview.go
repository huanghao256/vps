package checks

import (
	"context"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/vps-inspector/vps-inspector/internal/agent"
)

// NetworkOverviewCheck collects lightweight traffic and socket counters.
type NetworkOverviewCheck struct{}

// NewNetworkOverviewCheck creates the network overview check.
func NewNetworkOverviewCheck() NetworkOverviewCheck {
	return NetworkOverviewCheck{}
}

// ID returns the stable API identifier for this check.
func (NetworkOverviewCheck) ID() string { return "network.overview" }

// Name returns the display name for this check.
func (NetworkOverviewCheck) Name() string { return "Network Overview" }

// Description explains what the check measures.
func (NetworkOverviewCheck) Description() string {
	return "Collects IP traffic counters and TCP/UDP connection counts."
}

// Category groups this check in API metadata.
func (NetworkOverviewCheck) Category() string { return "network" }

// Run collects lightweight network counters for the dashboard.
func (c NetworkOverviewCheck) Run(ctx context.Context) agent.Result {
	started := time.Now().UTC()
	if err := ctx.Err(); err != nil {
		return agent.TimedResult(c.ID(), agent.StatusFail, 0, "Network overview timed out.", nil, started, err)
	}

	rx, tx := readNetworkBytes()
	tcp := countProcNetRows("/proc/net/tcp") + countProcNetRows("/proc/net/tcp6")
	udp := countProcNetRows("/proc/net/udp") + countProcNetRows("/proc/net/udp6")

	return agent.TimedResult(c.ID(), agent.StatusPass, 90, "Network overview collected.", map[string]any{
		"receivedBytes":    rx,
		"transmittedBytes": tx,
		"tcpConnections":   tcp,
		"udpConnections":   udp,
	}, started, nil)
}

func readNetworkBytes() (uint64, uint64) {
	data, err := os.ReadFile("/proc/net/dev")
	if err != nil {
		return 0, 0
	}
	var rx uint64
	var tx uint64
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
		rxValue, _ := strconv.ParseUint(fields[0], 10, 64)
		txValue, _ := strconv.ParseUint(fields[8], 10, 64)
		rx += rxValue
		tx += txValue
	}
	return rx, tx
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
