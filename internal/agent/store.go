package agent

import "sync"

// RunStore keeps a bounded in-memory history of recent inspection runs.
type RunStore struct {
	mu   sync.RWMutex
	runs []Run
	max  int
}

// NewRunStore creates a run store with a positive retention limit.
func NewRunStore(max int) *RunStore {
	if max <= 0 {
		max = 20
	}
	return &RunStore{max: max}
}

// Add stores a run as the newest entry and evicts old history when needed.
func (s *RunStore) Add(run Run) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.runs = append([]Run{run}, s.runs...)
	if len(s.runs) > s.max {
		s.runs = s.runs[:s.max]
	}
}

// List returns a copy of recent runs in newest-first order.
func (s *RunStore) List() []Run {
	s.mu.RLock()
	defer s.mu.RUnlock()
	copied := make([]Run, len(s.runs))
	copy(copied, s.runs)
	return copied
}

// Get returns a run by ID when it is still retained.
func (s *RunStore) Get(id string) (Run, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, run := range s.runs {
		if run.ID == id {
			return run, true
		}
	}
	return Run{}, false
}
