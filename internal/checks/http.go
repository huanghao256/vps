package checks

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/vps-inspector/vps-inspector/internal/agent"
)

type HTTPTarget struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type HTTPReachabilityCheck struct {
	client  *http.Client
	targets []HTTPTarget
}

func NewHTTPReachabilityCheck(targets []HTTPTarget) HTTPReachabilityCheck {
	return HTTPReachabilityCheck{
		client:  &http.Client{Timeout: 8 * time.Second},
		targets: targets,
	}
}

func DefaultHTTPTargets() []HTTPTarget {
	return []HTTPTarget{
		{Name: "Cloudflare Trace", URL: "https://cloudflare.com/cdn-cgi/trace"},
		{Name: "Google Generate 204", URL: "https://www.google.com/generate_204"},
		{Name: "GitHub", URL: "https://github.com"},
	}
}

func (HTTPReachabilityCheck) ID() string   { return "network.http_reachability" }
func (HTTPReachabilityCheck) Name() string { return "HTTP Reachability" }
func (HTTPReachabilityCheck) Description() string {
	return "Checks outbound HTTPS reachability to common services."
}
func (HTTPReachabilityCheck) Category() string { return "network" }

func (c HTTPReachabilityCheck) Run(ctx context.Context) agent.Result {
	started := time.Now().UTC()
	results := make([]map[string]any, 0, len(c.targets))
	successes := 0

	for _, target := range c.targets {
		probeStarted := time.Now()
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, target.URL, nil)
		if err != nil {
			results = append(results, map[string]any{"name": target.Name, "url": target.URL, "status": "fail", "error": err.Error()})
			continue
		}
		req.Header.Set("User-Agent", "vps-inspector/0.1")

		resp, err := c.client.Do(req)
		item := map[string]any{"name": target.Name, "url": target.URL}
		if err != nil {
			item["status"] = "fail"
			item["error"] = err.Error()
		} else {
			_ = resp.Body.Close()
			elapsed := float64(time.Since(probeStarted).Microseconds()) / 1000
			item["statusCode"] = resp.StatusCode
			item["latencyMs"] = elapsed
			if resp.StatusCode >= 200 && resp.StatusCode < 400 {
				successes++
				item["status"] = "pass"
			} else {
				item["status"] = "warn"
			}
		}
		results = append(results, item)
	}

	score := int(float64(successes) / float64(len(c.targets)) * 100)
	status := statusFromScore(score)
	summary := fmt.Sprintf("Reached %d/%d HTTPS targets.", successes, len(c.targets))

	return agent.TimedResult(c.ID(), status, score, summary, map[string]any{
		"successes": successes,
		"targets":   results,
	}, started, nil)
}
