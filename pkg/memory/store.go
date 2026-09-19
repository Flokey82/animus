package memory

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

// MemoryTier designates the durability and importance level of a memory.
type MemoryTier string

const (
	TierShortTerm MemoryTier = "short_term" // volatile impressions, topics
	TierLongTerm  MemoryTier = "long_term"  // significant facts, past events, relationships
	TierCore      MemoryTier = "core"       // fundamental identity beliefs and unbreakable truths
)

// Item represents a single stored memory unit.
type Item struct {
	ID        string     `json:"id"`
	Content   string     `json:"content"`
	Tier      MemoryTier `json:"tier"`
	Weight    float64    `json:"weight"` // 1.0 (trivial) to 10.0 (life-defining)
	CreatedAt time.Time  `json:"created_at"`
	Keywords  []string   `json:"keywords,omitempty"`
}

// Store coordinates multi-tiered memory storage, recall, and consolidation.
type Store struct {
	mu        sync.RWMutex
	shortTerm []Item
	longTerm  []Item
	core      []Item
}

// NewStore initializes a memory store.
func NewStore() *Store {
	return &Store{
		shortTerm: make([]Item, 0),
		longTerm:  make([]Item, 0),
		core:      make([]Item, 0),
	}
}

// Add inserts a new memory into the specified tier.
func (s *Store) Add(content string, tier MemoryTier, weight float64, keywords ...string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	item := Item{
		ID:        fmt.Sprintf("mem_%d", time.Now().UnixNano()),
		Content:   strings.TrimSpace(content),
		Tier:      tier,
		Weight:    weight,
		CreatedAt: time.Now(),
		Keywords:  keywords,
	}

	switch tier {
	case TierCore:
		s.core = append(s.core, item)
	case TierLongTerm:
		s.longTerm = append(s.longTerm, item)
	default:
		s.shortTerm = append(s.shortTerm, item)
	}
}

// Recall retrieves memories across all tiers matching any of the query terms.
func (s *Store) Recall(terms ...string) []Item {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var matches []Item
	all := append(append([]Item{}, s.core...), append(s.longTerm, s.shortTerm...)...)

	for _, item := range all {
		contentLower := strings.ToLower(item.Content)
		for _, term := range terms {
			tLower := strings.ToLower(strings.TrimSpace(term))
			if tLower != "" && strings.Contains(contentLower, tLower) {
				matches = append(matches, item)
				break
			}
		}
	}
	return matches
}

// ActiveMemoriesSummary compiles prominent memories for prompt composition.
func (s *Store) ActiveMemoriesSummary() string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var sb strings.Builder

	if len(s.core) > 0 {
		sb.WriteString("Core Beliefs & Identity:\n")
		for _, c := range s.core {
			sb.WriteString(fmt.Sprintf("- %s\n", c.Content))
		}
	}

	if len(s.longTerm) > 0 {
		limit := len(s.longTerm)
		if limit > 6 {
			limit = 6
		}
		sb.WriteString("Prominent Long-Term Memories:\n")
		for i := len(s.longTerm) - limit; i < len(s.longTerm); i++ {
			sb.WriteString(fmt.Sprintf("- %s\n", s.longTerm[i].Content))
		}
	}

	if len(s.shortTerm) > 0 {
		limit := len(s.shortTerm)
		if limit > 5 {
			limit = 5
		}
		sb.WriteString("Recent Thoughts & Impressions:\n")
		for i := len(s.shortTerm) - limit; i < len(s.shortTerm); i++ {
			sb.WriteString(fmt.Sprintf("- %s\n", s.shortTerm[i].Content))
		}
	}

	return strings.TrimSpace(sb.String())
}
