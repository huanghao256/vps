package checks

import "github.com/vps-inspector/vps-inspector/internal/agent"

func statusFromScore(score int) agent.Status {
	switch {
	case score >= 80:
		return agent.StatusPass
	case score >= 50:
		return agent.StatusWarn
	default:
		return agent.StatusFail
	}
}
