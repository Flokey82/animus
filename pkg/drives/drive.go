package drives

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"sync"
)

// Drive represents an individual homeostatic need or motive.
type Drive struct {
	Name             string  `json:"name"`               // e.g. "Energy", "Hunger", "Social", "Curiosity"
	Value            float64 `json:"value"`              // 0.0 (empty/depleted) to 100.0 (full/satisfied)
	Min              float64 `json:"min"`                // minimum bound (usually 0.0)
	Max              float64 `json:"max"`                // maximum bound (usually 100.0)
	DecayRatePerTick float64 `json:"decay_rate"`         // positive subtracts, negative adds
	CriticalLevel    float64 `json:"critical_level"`     // threshold below which drive causes acute distress
	WarningText      string  `json:"warning_text"`       // text prompt note when in critical state
}

// Urgency calculates how severely this drive demands attention (0.0 = content, 1.0 = acute crisis).
func (d *Drive) Urgency() float64 {
	if d.Value >= d.Max {
		return 0.0
	}
	if d.Value <= d.CriticalLevel {
		// Linear interpolation between 0 and critical level scaled 0.6 -> 1.0
		ratio := (d.CriticalLevel - d.Value) / math.Max(d.CriticalLevel, 1.0)
		return 0.6 + (ratio * 0.4)
	}
	// Linear interpolation between CriticalLevel and Max scaled 0.0 -> 0.6
	remaining := (d.Max - d.Value) / (d.Max - d.CriticalLevel)
	return remaining * 0.6
}

// Manager coordinates the full set of an agent's drives with thread-safety.
type Manager struct {
	mu     sync.RWMutex
	drives map[string]*Drive
}

// NewManager creates an empty drives manager.
func NewManager() *Manager {
	return &Manager{
		drives: make(map[string]*Drive),
	}
}

// RegisterDrive registers or replaces a drive.
func (m *Manager) RegisterDrive(d Drive) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if d.Max == 0 {
		d.Max = 100.0
	}
	if d.Value == 0 && d.Min == 0 {
		d.Value = d.Max
	}
	m.drives[d.Name] = &d
}

// Modify adjusts a drive value by a delta, clamping within [Min, Max].
func (m *Manager) Modify(name string, delta float64) float64 {
	m.mu.Lock()
	defer m.mu.Unlock()
	d, ok := m.drives[name]
	if !ok {
		return 0
	}
	d.Value += delta
	if d.Value > d.Max {
		d.Value = d.Max
	}
	if d.Value < d.Min {
		d.Value = d.Min
	}
	return d.Value
}

// Set explicitly sets a drive value.
func (m *Manager) Set(name string, val float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if d, ok := m.drives[name]; ok {
		if val > d.Max {
			val = d.Max
		}
		if val < d.Min {
			val = d.Min
		}
		d.Value = val
	}
}

// Get returns a copy of the specified drive.
func (m *Manager) Get(name string) (Drive, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	d, ok := m.drives[name]
	if !ok {
		return Drive{}, false
	}
	return *d, true
}

// All returns a slice of all drives.
func (m *Manager) All() []Drive {
	m.mu.RLock()
	defer m.mu.RUnlock()
	res := make([]Drive, 0, len(m.drives))
	for _, d := range m.drives {
		res = append(res, *d)
	}
	return res
}

// Tick applies the standard decay rates to all drives.
func (m *Manager) Tick() {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, d := range m.drives {
		d.Value -= d.DecayRatePerTick
		if d.Value > d.Max {
			d.Value = d.Max
		}
		if d.Value < d.Min {
			d.Value = d.Min
		}
	}
}

// MostUrgent returns the drive that is currently in greatest distress.
func (m *Manager) MostUrgent() *Drive {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if len(m.drives) == 0 {
		return nil
	}

	var best *Drive
	var highestUrgency float64 = -1

	for _, d := range m.drives {
		u := d.Urgency()
		if u > highestUrgency {
			highestUrgency = u
			cp := *d
			best = &cp
		}
	}
	return best
}

// Summary returns a human-readable prompt string summarizing current motives.
func (m *Manager) Summary() string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	items := make([]Drive, 0, len(m.drives))
	for _, d := range m.drives {
		items = append(items, *d)
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].Urgency() > items[j].Urgency()
	})

	var parts []string
	for _, it := range items {
		state := "Satisfied"
		if it.Value <= it.CriticalLevel {
			state = "CRITICAL"
		} else if it.Value < (it.Max * 0.5) {
			state = "Low"
		}
		parts = append(parts, fmt.Sprintf("%s: %.0f/%.0f (%s)", it.Name, it.Value, it.Max, state))
	}
	return strings.Join(parts, " | ")
}
