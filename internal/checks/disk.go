package checks

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/vps-inspector/vps-inspector/internal/agent"
)

const diskSampleSize = 8 * 1024 * 1024

// DiskCheck measures small sequential disk read/write samples.
type DiskCheck struct{}

// NewDiskCheck creates the disk I/O check.
func NewDiskCheck() DiskCheck {
	return DiskCheck{}
}

// ID returns the stable API identifier for this check.
func (DiskCheck) ID() string { return "disk.io" }

// Name returns the display name for this check.
func (DiskCheck) Name() string { return "Disk I/O" }

// Description explains what the check measures.
func (DiskCheck) Description() string { return "Measures a small sequential write and read sample." }

// Category groups this check in API metadata.
func (DiskCheck) Category() string { return "storage" }

// Run executes the disk sample and returns normalized I/O metrics.
func (c DiskCheck) Run(ctx context.Context) agent.Result {
	started := time.Now().UTC()

	dir := os.TempDir()
	path := filepath.Join(dir, "vps-inspector-disk-sample.tmp")
	defer os.Remove(path)

	payload := make([]byte, diskSampleSize)
	if _, err := rand.Read(payload); err != nil {
		return agent.TimedResult(c.ID(), agent.StatusFail, 0, "Failed to prepare disk sample.", nil, started, err)
	}

	if err := ctx.Err(); err != nil {
		return agent.TimedResult(c.ID(), agent.StatusFail, 0, "Disk check timed out.", nil, started, err)
	}

	writeStarted := time.Now()
	if err := os.WriteFile(path, payload, 0o600); err != nil {
		return agent.TimedResult(c.ID(), agent.StatusFail, 0, "Disk write failed.", nil, started, err)
	}
	writeDuration := time.Since(writeStarted)

	readStarted := time.Now()
	readBack, err := os.ReadFile(path)
	if err != nil {
		return agent.TimedResult(c.ID(), agent.StatusFail, 0, "Disk read failed.", nil, started, err)
	}
	readDuration := time.Since(readStarted)
	if len(readBack) != len(payload) {
		return agent.TimedResult(c.ID(), agent.StatusFail, 0, "Disk read returned an unexpected size.", nil, started, errors.New("disk sample size mismatch"))
	}

	writeMBps := mbPerSecond(diskSampleSize, writeDuration)
	readMBps := mbPerSecond(diskSampleSize, readDuration)
	score := diskScore(writeMBps, readMBps)
	status := statusFromScore(score)

	details := map[string]any{
		"sampleBytes":      diskSampleSize,
		"tempDir":          dir,
		"writeMegabytesPS": writeMBps,
		"readMegabytesPS":  readMBps,
	}
	summary := fmt.Sprintf("Write %.1f MB/s, read %.1f MB/s.", writeMBps, readMBps)
	return agent.TimedResult(c.ID(), status, score, summary, details, started, nil)
}

func mbPerSecond(bytes int, duration time.Duration) float64 {
	if duration <= 0 {
		return 0
	}
	return (float64(bytes) / 1024 / 1024) / duration.Seconds()
}

func diskScore(writeMBps, readMBps float64) int {
	avg := (writeMBps + readMBps) / 2
	switch {
	case avg >= 250:
		return 95
	case avg >= 120:
		return 82
	case avg >= 60:
		return 68
	case avg >= 25:
		return 50
	default:
		return 30
	}
}
