package agent

import (
	"context"
	"time"
)

// Check is the contract implemented by every VPS inspection module.
type Check interface {
	ID() string
	Name() string
	Description() string
	Category() string
	Run(context.Context) Result
}

// CheckInfo is the public metadata shown to API clients before running checks.
type CheckInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Category    string `json:"category"`
}

// Result captures the normalized outcome of a single inspection check.
type Result struct {
	CheckID   string         `json:"checkId"`
	Status    Status         `json:"status"`
	Score     int            `json:"score"`
	Summary   string         `json:"summary"`
	Details   map[string]any `json:"details,omitempty"`
	StartedAt time.Time      `json:"startedAt"`
	EndedAt   time.Time      `json:"endedAt"`
	Error     string         `json:"error,omitempty"`
}

// Status is the coarse health state returned by a check.
type Status string

const (
	// StatusPass means the check completed successfully and scored well.
	StatusPass Status = "pass"
	// StatusWarn means the check completed but found degraded behavior.
	StatusWarn Status = "warn"
	// StatusFail means the check failed or produced an unacceptable score.
	StatusFail Status = "fail"
	// StatusSkip means the check was intentionally not executed.
	StatusSkip Status = "skip"
)

// TimedResult builds a Result with consistent timing and error formatting.
func TimedResult(checkID string, status Status, score int, summary string, details map[string]any, started time.Time, err error) Result {
	result := Result{
		CheckID:   checkID,
		Status:    status,
		Score:     score,
		Summary:   summary,
		Details:   details,
		StartedAt: started,
		EndedAt:   time.Now().UTC(),
	}
	if err != nil {
		result.Error = err.Error()
	}
	return result
}
