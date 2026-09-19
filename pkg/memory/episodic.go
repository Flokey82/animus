package memory

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

// EpisodicEvent represents an individual action, observation, or conversation occurrence.
type EpisodicEvent struct {
	Timestamp time.Time `json:"timestamp"`
	Category  string    `json:"category"` // "action", "chat", "ambient", "dream", "tool"
	Content   string    `json:"content"`
}

// EpisodicLogger records the chronological flow of an agent's day.
type EpisodicLogger struct {
	mu       sync.RWMutex
	events   []EpisodicEvent
	maxItems int
}

// NewEpisodicLogger creates a logger with the specified buffer capacity.
func NewEpisodicLogger(maxItems int) *EpisodicLogger {
	if maxItems <= 0 {
		maxItems = 100
	}
	return &EpisodicLogger{
		events:   make([]EpisodicEvent, 0, maxItems),
		maxItems: maxItems,
	}
}

// Log records a new event.
func (e *EpisodicLogger) Log(category, content string) {
	e.mu.Lock()
	defer e.mu.Unlock()

	ev := EpisodicEvent{
		Timestamp: time.Now(),
		Category:  category,
		Content:   content,
	}

	if len(e.events) >= e.maxItems {
		// Drop oldest event
		e.events = append(e.events[1:], ev)
	} else {
		e.events = append(e.events, ev)
	}
}

// FullLogText returns all recorded events formatted with timestamps.
func (e *EpisodicLogger) FullLogText() string {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if len(e.events) == 0 {
		return ""
	}

	var sb strings.Builder
	for _, ev := range e.events {
		sb.WriteString(fmt.Sprintf("[%s] (%s) %s\n", ev.Timestamp.Format("15:04"), ev.Category, ev.Content))
	}
	return strings.TrimSpace(sb.String())
}

// Recent returns the last n events.
func (e *EpisodicLogger) Recent(n int) []EpisodicEvent {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if n <= 0 || len(e.events) == 0 {
		return nil
	}
	if n > len(e.events) {
		n = len(e.events)
	}
	res := make([]EpisodicEvent, n)
	copy(res, e.events[len(e.events)-n:])
	return res
}

// Clear resets the episodic log (e.g. after morning consolidation).
func (e *EpisodicLogger) Clear() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.events = e.events[:0]
}
