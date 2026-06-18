package agent

import (
	"context"
	"errors"
	"sort"
	"sync"
	"time"
)

// Inspector coordinates registered checks and aggregates their results into runs.
type Inspector struct {
	checks map[string]Check
}

// NewInspector indexes checks by ID for later execution.
func NewInspector(checks []Check) *Inspector {
	registered := make(map[string]Check, len(checks))
	for _, check := range checks {
		registered[check.ID()] = check
	}
	return &Inspector{checks: registered}
}

// ListChecks returns deterministic check metadata for the API and UI.
func (i *Inspector) ListChecks() []CheckInfo {
	items := make([]CheckInfo, 0, len(i.checks))
	for _, check := range i.checks {
		items = append(items, CheckInfo{
			ID:          check.ID(),
			Name:        check.Name(),
			Description: check.Description(),
			Category:    check.Category(),
		})
	}
	sort.Slice(items, func(a, b int) bool {
		if items[a].Category == items[b].Category {
			return items[a].ID < items[b].ID
		}
		return items[a].Category < items[b].Category
	})
	return items
}

// Run executes the selected checks concurrently and returns one completed run.
func (i *Inspector) Run(ctx context.Context, ids []string) (Run, error) {
	selected, err := i.selectChecks(ids)
	if err != nil {
		return Run{}, err
	}

	run := NewRun(checkIDs(selected))
	var wg sync.WaitGroup
	results := make(chan Result, len(selected))

	for _, check := range selected {
		wg.Add(1)
		go func(check Check) {
			defer wg.Done()
			checkCtx, cancel := context.WithTimeout(ctx, 25*time.Second)
			defer cancel()
			results <- check.Run(checkCtx)
		}(check)
	}

	wg.Wait()
	close(results)

	for result := range results {
		run.Results = append(run.Results, result)
	}
	sort.Slice(run.Results, func(a, b int) bool {
		return run.Results[a].CheckID < run.Results[b].CheckID
	})
	run.Complete()

	return run, nil
}

func (i *Inspector) selectChecks(ids []string) ([]Check, error) {
	if len(ids) == 0 {
		selected := make([]Check, 0, len(i.checks))
		for _, check := range i.checks {
			selected = append(selected, check)
		}
		return selected, nil
	}

	selected := make([]Check, 0, len(ids))
	for _, id := range ids {
		check, ok := i.checks[id]
		if !ok {
			return nil, errors.New("unknown check: " + id)
		}
		selected = append(selected, check)
	}
	return selected, nil
}

func checkIDs(checks []Check) []string {
	ids := make([]string, 0, len(checks))
	for _, check := range checks {
		ids = append(ids, check.ID())
	}
	sort.Strings(ids)
	return ids
}
