package agent

import (
	"context"
	"time"
)

type Check interface {
	ID() string
	Name() string
	Description() string
	Category() string
	Run(context.Context) Result
}

type CheckInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Category    string `json:"category"`
}

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

type Status string

const (
	StatusPass Status = "pass"
	StatusWarn Status = "warn"
	StatusFail Status = "fail"
	StatusSkip Status = "skip"
)

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
