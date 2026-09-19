package builtin

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/Flokey82/animus/pkg/tools"
	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
)

// ClockTool provides the agent with current wall clock time, date, and day of week.
type ClockTool struct {
	Location *time.Location
}

func NewClockTool(loc *time.Location) *ClockTool {
	if loc == nil {
		loc = time.Local
	}
	return &ClockTool{Location: loc}
}

func (c *ClockTool) Definition() openai.Tool {
	return openai.Tool{
		Type: openai.ToolTypeFunction,
		Function: &openai.FunctionDefinition{
			Name:        "check_current_time",
			Description: "Check the current wall clock time, date, day of week, and whether it is daytime or nighttime.",
			Parameters:  jsonschema.Definition{Type: jsonschema.Object},
		},
	}
}

func (c *ClockTool) Execute(ctx context.Context, args map[string]any) (string, error) {
	now := time.Now().In(c.Location)
	hour := now.Hour()
	isDaytime := hour >= 6 && hour < 20

	data := map[string]any{
		"time":        now.Format("15:04:05"),
		"date":        now.Format("2006-01-02"),
		"day_of_week": now.Weekday().String(),
		"timezone":    now.Location().String(),
		"is_daytime":  isDaytime,
	}
	return tools.MakeSuccessPayload(data), nil
}

func (c *ClockTool) Tags() []string {
	return nil // Universal tool
}

// CelestialMoonTool calculates astronomical moon phase and illumination.
type CelestialMoonTool struct{}

func NewCelestialMoonTool() *CelestialMoonTool {
	return &CelestialMoonTool{}
}

func (m *CelestialMoonTool) Definition() openai.Tool {
	return openai.Tool{
		Type: openai.ToolTypeFunction,
		Function: &openai.FunctionDefinition{
			Name:        "get_moon_phase",
			Description: "Calculate the current astronomical moon phase, percent illumination, and lunar age.",
			Parameters:  jsonschema.Definition{Type: jsonschema.Object},
		},
	}
}

func (m *CelestialMoonTool) Execute(ctx context.Context, args map[string]any) (string, error) {
	now := time.Now().UTC()
	// Reference new moon: Jan 6, 2000 18:14 UTC
	refNewMoon := time.Date(2000, 1, 6, 18, 14, 0, 0, time.UTC)
	synodicMonth := 29.53058867

	diffDays := now.Sub(refNewMoon).Hours() / 24.0
	cycles := diffDays / synodicMonth
	phaseFraction := cycles - math.Floor(cycles)
	ageDays := phaseFraction * synodicMonth
	illumination := (1.0 - math.Cos(phaseFraction*2.0*math.Pi)) / 2.0 * 100.0

	var phaseName string
	var desc string

	switch {
	case phaseFraction < 0.03 || phaseFraction >= 0.97:
		phaseName = "New Moon 🌑"
		desc = "The moon is cloaked in darkness between Earth and the Sun."
	case phaseFraction < 0.22:
		phaseName = "Waxing Crescent 🌒"
		desc = "A slender silver crescent glows in the western evening sky."
	case phaseFraction < 0.28:
		phaseName = "First Quarter 🌓"
		desc = "Half the lunar disk is illuminated, rising high at sunset."
	case phaseFraction < 0.47:
		phaseName = "Waxing Gibbous 🌔"
		desc = "More than half illuminated and waxing toward a full pearl."
	case phaseFraction < 0.53:
		phaseName = "Full Moon 🌕"
		desc = "A brilliant, fully illuminated orb shining all night long."
	case phaseFraction < 0.72:
		phaseName = "Waning Gibbous 🌖"
		desc = "A rich, waning moon rising late in the evening."
	case phaseFraction < 0.78:
		phaseName = "Last Quarter 🌗"
		desc = "Half illuminated, visible high in the early dawn sky."
	default:
		phaseName = "Waning Crescent 🌘"
		desc = "A delicate sliver before dawn, soon to join the sun."
	}

	res := map[string]any{
		"phase_name":           phaseName,
		"illumination_percent": fmt.Sprintf("%.1f%%", illumination),
		"age_days":             fmt.Sprintf("%.1f days", ageDays),
		"description":          desc,
	}
	return tools.MakeSuccessPayload(res), nil
}

func (m *CelestialMoonTool) Tags() []string {
	return []string{"nature", "night", "outdoor"}
}
