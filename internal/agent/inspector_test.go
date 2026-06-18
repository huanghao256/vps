package agent

import (
	"context"
	"testing"
	"time"
)

func TestInspectorRunsSelectedChecks(t *testing.T) {
	inspector := NewInspector([]Check{
		fakeCheck{id: "one", category: "a"},
		fakeCheck{id: "two", category: "b"},
	})

	run, err := inspector.Run(context.Background(), []string{"two"})
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}

	if len(run.Results) != 1 {
		t.Fatalf("expected one result, got %d", len(run.Results))
	}
	if run.Results[0].CheckID != "two" {
		t.Fatalf("unexpected check id: %s", run.Results[0].CheckID)
	}
	if run.Status != "completed" {
		t.Fatalf("unexpected status: %s", run.Status)
	}
}

func TestInspectorRejectsUnknownCheck(t *testing.T) {
	inspector := NewInspector([]Check{fakeCheck{id: "one", category: "a"}})

	if _, err := inspector.Run(context.Background(), []string{"missing"}); err == nil {
		t.Fatal("expected unknown check to fail")
	}
}

type fakeCheck struct {
	id       string
	category string
}

func (f fakeCheck) ID() string          { return f.id }
func (f fakeCheck) Name() string        { return f.id }
func (f fakeCheck) Description() string { return f.id }
func (f fakeCheck) Category() string    { return f.category }

func (f fakeCheck) Run(context.Context) Result {
	started := time.Now().UTC()
	return TimedResult(f.id, StatusPass, 90, "ok", nil, started, nil)
}
