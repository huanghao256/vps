package checks

import (
	"context"
	"fmt"
	"net"
	"sort"
	"time"

	"github.com/vps-inspector/vps-inspector/internal/agent"
)

// LatencyTarget describes one TCP connect latency probe.
type LatencyTarget struct {
	Name    string `json:"name"`
	Address string `json:"address"`
}

// LatencyCheck measures TCP connect latency to multiple public targets.
type LatencyCheck struct {
	targets []LatencyTarget
}

// NewLatencyCheck creates a latency check for the supplied targets.
func NewLatencyCheck(targets []LatencyTarget) LatencyCheck {
	return LatencyCheck{targets: targets}
}

// DefaultLatencyTargets returns globally reachable TCP latency probes.
func DefaultLatencyTargets() []LatencyTarget {
	return []LatencyTarget{
		{Name: "Cloudflare", Address: "1.1.1.1:443"},
		{Name: "Google DNS", Address: "8.8.8.8:443"},
		{Name: "Quad9", Address: "9.9.9.9:443"},
	}
}

// ID returns the stable API identifier for this check.
func (LatencyCheck) ID() string { return "network.tcp_latency" }

// Name returns the display name for this check.
func (LatencyCheck) Name() string { return "TCP Latency" }

// Description explains what the check measures.
func (LatencyCheck) Description() string {
	return "Measures TCP connect latency to known network targets."
}

// Category groups this check in API metadata.
func (LatencyCheck) Category() string { return "network" }

// Run measures TCP connect latency and returns median latency details.
func (c LatencyCheck) Run(ctx context.Context) agent.Result {
	started := time.Now().UTC()
	results := make([]map[string]any, 0, len(c.targets))
	successes := 0
	latencies := make([]float64, 0, len(c.targets))

	for _, target := range c.targets {
		latency, err := dialLatency(ctx, target.Address)
		item := map[string]any{
			"name":    target.Name,
			"address": target.Address,
		}
		if err != nil {
			item["status"] = "fail"
			item["error"] = err.Error()
		} else {
			successes++
			ms := float64(latency.Microseconds()) / 1000
			latencies = append(latencies, ms)
			item["status"] = "pass"
			item["latencyMs"] = ms
		}
		results = append(results, item)
	}

	if successes == 0 {
		return agent.TimedResult(c.ID(), agent.StatusFail, 0, "All latency probes failed.", map[string]any{"targets": results}, started, nil)
	}

	sort.Float64s(latencies)
	median := latencies[len(latencies)/2]
	score := latencyScore(median, successes, len(c.targets))
	status := statusFromScore(score)
	summary := fmt.Sprintf("Median TCP latency %.1f ms across %d/%d targets.", median, successes, len(c.targets))

	return agent.TimedResult(c.ID(), status, score, summary, map[string]any{
		"medianLatencyMs": median,
		"successes":       successes,
		"targets":         results,
	}, started, nil)
}

func dialLatency(ctx context.Context, address string) (time.Duration, error) {
	dialer := net.Dialer{Timeout: 5 * time.Second}
	started := time.Now()
	conn, err := dialer.DialContext(ctx, "tcp", address)
	if err != nil {
		return 0, err
	}
	_ = conn.Close()
	return time.Since(started), nil
}

func latencyScore(median float64, successes, total int) int {
	score := 100
	switch {
	case median > 280:
		score = 45
	case median > 180:
		score = 65
	case median > 90:
		score = 78
	case median > 40:
		score = 88
	default:
		score = 96
	}
	if successes < total {
		score -= (total - successes) * 12
	}
	if score < 0 {
		return 0
	}
	return score
}
