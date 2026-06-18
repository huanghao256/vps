package checks

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/vps-inspector/vps-inspector/internal/agent"
)

type StabilityCheck struct {
	targets []LatencyTarget
}

func NewStabilityCheck(targets []LatencyTarget) StabilityCheck {
	return StabilityCheck{targets: targets}
}

func (StabilityCheck) ID() string   { return "network.stability" }
func (StabilityCheck) Name() string { return "Stability" }
func (StabilityCheck) Description() string {
	return "Runs repeated TCP probes and estimates loss and jitter."
}
func (StabilityCheck) Category() string { return "stability" }

func (c StabilityCheck) Run(ctx context.Context) agent.Result {
	started := time.Now().UTC()
	var latencies []float64
	attempts := 0
	failures := 0

	for i := 0; i < 4; i++ {
		for _, target := range c.targets {
			attempts++
			latency, err := dialLatency(ctx, target.Address)
			if err != nil {
				failures++
				continue
			}
			latencies = append(latencies, float64(latency.Microseconds())/1000)
		}
	}

	loss := float64(failures) / float64(attempts) * 100
	jitter := jitterMs(latencies)
	score := stabilityScore(loss, jitter)
	summary := fmt.Sprintf("Loss %.1f%%, jitter %.1f ms across %d probes.", loss, jitter, attempts)
	return agent.TimedResult(c.ID(), statusFromScore(score), score, summary, map[string]any{
		"probeCount": attempts,
		"lossPct":    loss,
		"jitterMs":   jitter,
	}, started, nil)
}

func jitterMs(values []float64) float64 {
	if len(values) < 2 {
		return 0
	}
	total := 0.0
	for _, value := range values {
		total += value
	}
	mean := total / float64(len(values))
	variance := 0.0
	for _, value := range values {
		diff := value - mean
		variance += diff * diff
	}
	return math.Sqrt(variance / float64(len(values)))
}

func stabilityScore(loss, jitter float64) int {
	score := 100 - int(loss*3) - int(jitter/2)
	if score < 0 {
		return 0
	}
	if score > 100 {
		return 100
	}
	return score
}
