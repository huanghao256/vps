package agent

import (
	"crypto/rand"
	"encoding/hex"
	"time"
)

// Run is a persisted snapshot of one inspection execution.
type Run struct {
	ID        string    `json:"id"`
	CheckIDs  []string  `json:"checkIds"`
	Status    string    `json:"status"`
	Score     int       `json:"score"`
	StartedAt time.Time `json:"startedAt"`
	EndedAt   time.Time `json:"endedAt,omitempty"`
	Results   []Result  `json:"results"`
}

// NewRun creates a running inspection record for the supplied check IDs.
func NewRun(checkIDs []string) Run {
	return Run{
		ID:        newRunID(),
		CheckIDs:  checkIDs,
		Status:    "running",
		StartedAt: time.Now().UTC(),
		Results:   []Result{},
	}
}

// Complete marks a run finished and computes its average score.
func (r *Run) Complete() {
	r.Status = "completed"
	r.EndedAt = time.Now().UTC()
	if len(r.Results) == 0 {
		r.Score = 0
		return
	}
	total := 0
	for _, result := range r.Results {
		total += result.Score
	}
	r.Score = total / len(r.Results)
}

func newRunID() string {
	var bytes [8]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		return time.Now().UTC().Format("20060102150405")
	}
	return hex.EncodeToString(bytes[:])
}
