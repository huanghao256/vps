package agent

import "testing"

func TestRunStoreKeepsNewestWithinLimit(t *testing.T) {
	store := NewRunStore(2)
	first := NewRun([]string{"a"})
	second := NewRun([]string{"b"})
	third := NewRun([]string{"c"})

	store.Add(first)
	store.Add(second)
	store.Add(third)

	runs := store.List()
	if len(runs) != 2 {
		t.Fatalf("expected two runs, got %d", len(runs))
	}
	if runs[0].ID != third.ID || runs[1].ID != second.ID {
		t.Fatal("expected newest runs first within limit")
	}
	if _, ok := store.Get(first.ID); ok {
		t.Fatal("expected oldest run to be evicted")
	}
}
