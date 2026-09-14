package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// state remembers which slots were already reported, keyed by signature.
type state struct {
	// Seen maps a slot signature to its UTC start, so past entries can be pruned.
	Seen map[string]string `json:"seen"`
}

// loadState reads the state file. A missing file yields an empty state.
func loadState(path string) (*state, error) {
	raw, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return &state{Seen: map[string]string{}}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read state: %w", err)
	}
	var s state
	if err := json.Unmarshal(raw, &s); err != nil {
		return nil, fmt.Errorf("parse state JSON: %w", err)
	}
	if s.Seen == nil {
		s.Seen = map[string]string{}
	}
	return &s, nil
}

// save writes the state file, creating parent directories as needed.
func (s *state) save(path string) error {
	if dir := filepath.Dir(path); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("create state dir: %w", err)
		}
	}
	raw, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("encode state: %w", err)
	}
	if err := os.WriteFile(path, raw, 0o644); err != nil {
		return fmt.Errorf("write state: %w", err)
	}
	return nil
}

// newSlots returns the slots not seen before and records them as seen.
func (s *state) newSlots(slots []matchedSlot) []matchedSlot {
	var fresh []matchedSlot
	for _, slot := range slots {
		sig := slot.signature()
		if _, ok := s.Seen[sig]; ok {
			continue
		}
		s.Seen[sig] = slot.startUTC.UTC().Format(time.RFC3339)
		fresh = append(fresh, slot)
	}
	return fresh
}

// prune drops seen entries whose slot start is in the past, keeping the file
// from growing without bound.
func (s *state) prune(now time.Time) {
	for sig, startStr := range s.Seen {
		start, err := time.Parse(time.RFC3339, startStr)
		if err != nil || start.Before(now) {
			delete(s.Seen, sig)
		}
	}
}

// signatures returns the seen signatures sorted, for stable logging/tests.
func (s *state) signatures() []string {
	out := make([]string, 0, len(s.Seen))
	for sig := range s.Seen {
		out = append(out, sig)
	}
	sort.Strings(out)
	return out
}
