package checks

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/vps-inspector/vps-inspector/internal/agent"
)

type RiskCheck struct {
	client *http.Client
}

func NewRiskCheck() RiskCheck {
	return RiskCheck{client: &http.Client{Timeout: 8 * time.Second}}
}

func (RiskCheck) ID() string   { return "risk.reputation" }
func (RiskCheck) Name() string { return "Risk Profile" }
func (RiskCheck) Description() string {
	return "Checks public IP visibility and common service reachability signals."
}
func (RiskCheck) Category() string { return "risk" }

func (c RiskCheck) Run(ctx context.Context) agent.Result {
	started := time.Now().UTC()
	trace, traceErr := c.cloudflareTrace(ctx)
	googleOK := c.headOK(ctx, "https://www.google.com/generate_204")
	githubOK := c.headOK(ctx, "https://github.com")

	score := 70
	signals := []string{}
	if traceErr == nil && trace["ip"] != "" {
		score += 10
		signals = append(signals, "Public IP detected")
	}
	if googleOK {
		score += 10
		signals = append(signals, "Google reachable")
	} else {
		score -= 12
		signals = append(signals, "Google probe failed")
	}
	if githubOK {
		score += 5
		signals = append(signals, "GitHub reachable")
	}
	if trace["warp"] != "" && trace["warp"] != "off" {
		score -= 8
		signals = append(signals, "Cloudflare WARP/proxy signal detected")
	}
	if score > 100 {
		score = 100
	}
	if score < 0 {
		score = 0
	}

	details := map[string]any{
		"ip":              trace["ip"],
		"country":         trace["loc"],
		"cloudflareColo":  trace["colo"],
		"googleReachable": googleOK,
		"githubReachable": githubOK,
		"signals":         signals,
		"note":            "This is a lightweight risk signal, not a full IP reputation database.",
	}
	if traceErr != nil {
		details["traceError"] = traceErr.Error()
	}

	return agent.TimedResult(c.ID(), statusFromScore(score), score, fmt.Sprintf("Risk score %d based on %d signals.", score, len(signals)), details, started, nil)
}

func (c RiskCheck) cloudflareTrace(ctx context.Context) (map[string]string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://cloudflare.com/cdn-cgi/trace", nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	values := map[string]string{}
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		key, value, ok := strings.Cut(scanner.Text(), "=")
		if ok {
			values[key] = value
		}
	}
	return values, scanner.Err()
}

func (c RiskCheck) headOK(ctx context.Context, url string) bool {
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, url, nil)
	if err != nil {
		return false
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode >= 200 && resp.StatusCode < 400
}
