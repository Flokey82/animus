package drives

import (
	"testing"
)

func TestDrivesManager(t *testing.T) {
	mgr := NewManager()

	mgr.RegisterDrive(Drive{
		Name:             "Energy",
		Value:            80,
		Max:              100,
		CriticalLevel:    20,
		DecayRatePerTick: 5,
	})

	mgr.RegisterDrive(Drive{
		Name:             "Hunger",
		Value:            15, // Already critical
		Max:              100,
		CriticalLevel:    25,
		DecayRatePerTick: 2,
	})

	urgent := mgr.MostUrgent()
	if urgent == nil || urgent.Name != "Hunger" {
		t.Fatalf("expected Hunger to be most urgent, got %+v", urgent)
	}

	// Apply tick
	mgr.Tick()

	e, _ := mgr.Get("Energy")
	if e.Value != 75 {
		t.Errorf("expected Energy 75, got %f", e.Value)
	}

	h, _ := mgr.Get("Hunger")
	if h.Value != 13 {
		t.Errorf("expected Hunger 13, got %f", h.Value)
	}

	// Satisfy hunger
	mgr.Modify("Hunger", 70)
	hAfter, _ := mgr.Get("Hunger")
	if hAfter.Value != 83 {
		t.Errorf("expected Hunger 83, got %f", hAfter.Value)
	}

	// Summary should contain both
	summary := mgr.Summary()
	if summary == "" {
		t.Errorf("expected non-empty summary")
	}
}
