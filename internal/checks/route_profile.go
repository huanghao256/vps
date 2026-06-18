package checks

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/vps-inspector/vps-inspector/internal/agent"
)

// RouteTarget describes one carrier-oriented route probe.
type RouteTarget struct {
	Name    string `json:"name"`
	Carrier string `json:"carrier"`
	Address string `json:"address"`
}

// RouteProfileCheck estimates China carrier route quality from TCP probes.
type RouteProfileCheck struct {
	targets []RouteTarget
}

// NewRouteProfileCheck creates a route profile check for the supplied targets.
func NewRouteProfileCheck(targets []RouteTarget) RouteProfileCheck {
	return RouteProfileCheck{targets: targets}
}

// DefaultRouteTargets returns carrier probes used for line classification.
func DefaultRouteTargets() []RouteTarget {
	return []RouteTarget{
		{Name: "China Telecom DNS", Carrier: "Telecom", Address: "202.96.134.33:53"},
		{Name: "China Unicom DNS", Carrier: "Unicom", Address: "123.125.81.6:80"},
		{Name: "China Mobile DNS", Carrier: "Mobile", Address: "120.196.165.24:80"},
	}
}

// ID returns the stable API identifier for this check.
func (RouteProfileCheck) ID() string { return "network.route_profile" }

// Name returns the display name for this check.
func (RouteProfileCheck) Name() string { return "Route Profile" }

// Description explains what the check measures.
func (RouteProfileCheck) Description() string {
	return "Estimates carrier route quality using TCP probes to China carrier targets."
}

// Category groups this check in API metadata.
func (RouteProfileCheck) Category() string { return "line" }

// Run probes carrier targets and derives a Chinese line-quality profile.
func (c RouteProfileCheck) Run(ctx context.Context) agent.Result {
	started := time.Now().UTC()
	items := make([]map[string]any, 0, len(c.targets))
	successes := 0
	totalLatency := 0.0

	for _, target := range c.targets {
		latency, err := dialRoute(ctx, target.Address)
		item := map[string]any{"name": target.Name, "carrier": target.Carrier, "address": target.Address}
		if err != nil {
			item["status"] = "blocked"
			item["quality"] = "Unavailable"
			item["error"] = err.Error()
		} else {
			ms := float64(latency.Microseconds()) / 1000
			successes++
			totalLatency += ms
			item["status"] = "reachable"
			item["latencyMs"] = ms
			item["quality"] = routeQuality(ms)
		}
		items = append(items, item)
	}

	if successes == 0 {
		return agent.TimedResult(c.ID(), agent.StatusFail, 10, "三网线路探测均不可达。", map[string]any{
			"routes":     items,
			"lineType":   "线路不可达",
			"confidence": "低",
		}, started, nil)
	}

	avg := totalLatency / float64(successes)
	score := latencyScore(avg, successes, len(c.targets))
	lineType, confidence := classifyLineType(avg, successes, len(c.targets))
	return agent.TimedResult(c.ID(), statusFromScore(score), score, fmt.Sprintf("三网可达 %d/%d，平均延迟 %.0f ms，线路画像：%s。", successes, len(c.targets), avg, lineType), map[string]any{
		"routes":           items,
		"averageLatencyMs": avg,
		"reachable":        successes,
		"lineType":         lineType,
		"confidence":       confidence,
	}, started, nil)
}

func dialRoute(ctx context.Context, address string) (time.Duration, error) {
	dialer := net.Dialer{Timeout: 6 * time.Second}
	started := time.Now()
	conn, err := dialer.DialContext(ctx, "tcp", address)
	if err != nil {
		return 0, err
	}
	_ = conn.Close()
	return time.Since(started), nil
}

func routeQuality(ms float64) string {
	switch {
	case ms <= 150:
		return "优秀"
	case ms <= 220:
		return "良好"
	case ms <= 320:
		return "拥塞"
	default:
		return "较差"
	}
}

func classifyLineType(avg float64, successes, total int) (string, string) {
	if successes == total && avg <= 170 {
		return "CN2 GIA / 精品线路", "中"
	}
	if successes == total && avg <= 230 {
		return "AS9929 / 精品线路", "中"
	}
	if successes >= 2 && avg <= 260 {
		return "CMI / 优化线路", "中"
	}
	if successes >= 1 {
		return "普通 BGP / 国际线路", "低"
	}
	return "线路不可达", "低"
}
