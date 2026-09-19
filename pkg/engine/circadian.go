package engine

import "time"

// SleepState represents current circadian consciousness.
type SleepState string

const (
	StateAwake       SleepState = "awake"
	StateWindingDown SleepState = "winding_down"
	StateAsleep      SleepState = "asleep"
	StateWaking      SleepState = "waking"
)

// Chronotype defines daily sleep-wake bias.
type Chronotype string

const (
	ChronotypeEarlyBird Chronotype = "early_bird" // Sleep 21:30 - 05:30
	ChronotypeBalanced  Chronotype = "balanced"   // Sleep 23:00 - 07:00
	ChronotypeNightOwl  Chronotype = "night_owl"  // Sleep 01:30 - 09:30
)

// CircadianProfile governs sleep schedule.
type CircadianProfile struct {
	Chronotype    Chronotype
	SleepHour     int // e.g. 23
	SleepMinute   int // e.g. 0
	WakeHour      int // e.g. 7
	WakeMinute    int // e.g. 0
	CurrentState  SleepState
}

// DefaultCircadianProfile returns a balanced sleep profile.
func DefaultCircadianProfile() CircadianProfile {
	return CircadianProfile{
		Chronotype:   ChronotypeBalanced,
		SleepHour:    23,
		SleepMinute:  0,
		WakeHour:     7,
		WakeMinute:   0,
		CurrentState: StateAwake,
	}
}

// Evaluate determines whether the agent should be asleep or awake at time t.
func (p *CircadianProfile) Evaluate(t time.Time) SleepState {
	hour := t.Hour()
	minute := t.Minute()
	currentMinutes := (hour * 60) + minute

	sleepMinutes := (p.SleepHour * 60) + p.SleepMinute
	wakeMinutes := (p.WakeHour * 60) + p.WakeMinute

	isNightTime := false
	if sleepMinutes > wakeMinutes {
		// e.g. 23:00 to 07:00
		isNightTime = currentMinutes >= sleepMinutes || currentMinutes < wakeMinutes
	} else {
		// sleep across standard day interval
		isNightTime = currentMinutes >= sleepMinutes && currentMinutes < wakeMinutes
	}

	if isNightTime {
		return StateAsleep
	}
	return StateAwake
}
