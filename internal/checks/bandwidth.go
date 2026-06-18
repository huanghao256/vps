package checks

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/vps-inspector/vps-inspector/internal/agent"
)

// BandwidthCheck estimates network throughput with small HTTP upload/download samples.
type BandwidthCheck struct {
	client *http.Client
}

// NewBandwidthCheck creates a bandwidth check with bounded HTTP timeouts.
func NewBandwidthCheck() BandwidthCheck {
	return BandwidthCheck{client: &http.Client{Timeout: 18 * time.Second}}
}

// ID returns the stable API identifier for this check.
func (BandwidthCheck) ID() string { return "network.bandwidth" }

// Name returns the display name for this check.
func (BandwidthCheck) Name() string { return "Bandwidth" }

// Description explains what the check measures.
func (BandwidthCheck) Description() string {
	return "Runs a small Cloudflare download and upload sample."
}

// Category groups this check in API metadata.
func (BandwidthCheck) Category() string { return "bandwidth" }

// Run executes the bandwidth samples and returns normalized throughput metrics.
func (c BandwidthCheck) Run(ctx context.Context) agent.Result {
	started := time.Now().UTC()
	downMbps, downErr := c.download(ctx, 8_000_000)
	upMbps, upErr := c.upload(ctx, 2_000_000)

	details := map[string]any{
		"downloadMbps": downMbps,
		"uploadMbps":   upMbps,
		"sampleNote":   "Small samples are used to avoid wasting VPS traffic.",
	}
	score := bandwidthScore(downMbps, upMbps)
	status := statusFromScore(score)

	if downErr != nil {
		details["downloadError"] = downErr.Error()
	}
	if upErr != nil {
		details["uploadError"] = upErr.Error()
	}
	if downErr != nil && upErr != nil {
		return agent.TimedResult(c.ID(), agent.StatusFail, 0, "Bandwidth samples failed.", details, started, downErr)
	}

	summary := fmt.Sprintf("Download %.1f Mbps, upload %.1f Mbps.", downMbps, upMbps)
	return agent.TimedResult(c.ID(), status, score, summary, details, started, nil)
}

func (c BandwidthCheck) download(ctx context.Context, bytesToRead int64) (float64, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("https://speed.cloudflare.com/__down?bytes=%d", bytesToRead), nil)
	if err != nil {
		return 0, err
	}
	started := time.Now()
	resp, err := c.client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	written, err := io.Copy(io.Discard, io.LimitReader(resp.Body, bytesToRead))
	if err != nil {
		return 0, err
	}
	return megabitsPerSecond(written, time.Since(started)), nil
}

func (c BandwidthCheck) upload(ctx context.Context, size int64) (float64, error) {
	body := bytes.NewReader(make([]byte, size))
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://speed.cloudflare.com/__up", body)
	if err != nil {
		return 0, err
	}
	req.Header.Set("Content-Type", "application/octet-stream")
	started := time.Now()
	resp, err := c.client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)
	return megabitsPerSecond(size, time.Since(started)), nil
}

func megabitsPerSecond(bytes int64, duration time.Duration) float64 {
	if duration <= 0 {
		return 0
	}
	return (float64(bytes) * 8 / 1000 / 1000) / duration.Seconds()
}

func bandwidthScore(download, upload float64) int {
	weighted := download*0.75 + upload*0.25
	switch {
	case weighted >= 500:
		return 96
	case weighted >= 200:
		return 88
	case weighted >= 100:
		return 78
	case weighted >= 50:
		return 65
	case weighted >= 20:
		return 50
	default:
		return 30
	}
}
