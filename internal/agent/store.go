package agent

import "sync"

type RunStore struct {
	mu   sync.RWMutex
	runs []Run
	max  int
}

func NewRunStore(max int) *RunStore {
	if max <= 0 {
		max = 20
	}
	return &RunStore{max: max}
}

func (s *RunStore) Add(run Run) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.runs = append([]Run{run}, s.runs...)
	if len(s.runs) > s.max {
		s.runs = s.runs[:s.max]
	}
}

func (s *RunStore) List() []Run {
	s.mu.RLock()
	defer s.mu.RUnlock()
	copied := make([]Run, len(s.runs))
	copy(copied, s.runs)
	return copied
}

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
