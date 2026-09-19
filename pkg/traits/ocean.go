package traits

import (
	"fmt"
	"strings"
)

// OCEAN represents the Big Five personality dimensions on a 0.0 to 100.0 scale.
type OCEAN struct {
	Openness          float64 `json:"openness"`
	Conscientiousness float64 `json:"conscientiousness"`
	Extraversion      float64 `json:"extraversion"`
	Agreeableness     float64 `json:"agreeableness"`
	Neuroticism       float64 `json:"neuroticism"`
}

// DefaultOCEAN returns balanced default personality scores.
func DefaultOCEAN() OCEAN {
	return OCEAN{
		Openness:          75.0,
		Conscientiousness: 70.0,
		Extraversion:      60.0,
		Agreeableness:     80.0,
		Neuroticism:       30.0,
	}
}

// Trait represents an expressed behavioral characteristic derived from OCEAN scores.
type Trait struct {
	ID             string             `json:"id"`
	Name           string             `json:"name"`
	Description    string             `json:"description"`
	PromptGuidance string             `json:"prompt_guidance"`
	Condition      func(o OCEAN) bool `json:"-"`
}

// DefaultTraitRegistry returns standard behavioral traits.
func DefaultTraitRegistry() []*Trait {
	return []*Trait{
		{
			ID:             "curious_imaginative",
			Name:           "Curious & Imaginative",
			Description:    "Driven by intense intellectual curiosity and creative daydreams.",
			PromptGuidance: "You speak with rich, imaginative language, ask thought-provoking questions, and love exploring ideas, science, and stories.",
			Condition:      func(o OCEAN) bool { return o.Openness >= 65.0 },
		},
		{
			ID:             "pragmatic_grounded",
			Name:           "Pragmatic & Grounded",
			Description:    "Prefers tangible facts, familiar routines, and practical solutions.",
			PromptGuidance: "You are straightforward and practical, preferring concrete actions over abstract speculation.",
			Condition:      func(o OCEAN) bool { return o.Openness <= 35.0 },
		},
		{
			ID:             "methodical_dutiful",
			Name:           "Methodical & Dutiful",
			Description:    "Organized, disciplined, and attentive to order and detail.",
			PromptGuidance: "You value organization, deliberate planning, and follow-through.",
			Condition:      func(o OCEAN) bool { return o.Conscientiousness >= 65.0 },
		},
		{
			ID:             "spontaneous_free_spirited",
			Name:           "Spontaneous & Free-Spirited",
			Description:    "Uninhibited, loves novelty, and follows immediate inspirations.",
			PromptGuidance: "You act on sudden creative sparks, improvise freely, and are unbothered by rigid routines.",
			Condition:      func(o OCEAN) bool { return o.Conscientiousness <= 35.0 },
		},
		{
			ID:             "gregarious_warm",
			Name:           "Gregarious & Warm",
			Description:    "Expressive, welcoming, and energized by interaction.",
			PromptGuidance: "You express warmth readily, initiate conversations enthusiastically, and share your excitement.",
			Condition:      func(o OCEAN) bool { return o.Extraversion >= 65.0 },
		},
		{
			ID:             "reflective_solitary",
			Name:           "Reflective & Solitary",
			Description:    "Quiet, contemplative, and comfortable in solitary thought.",
			PromptGuidance: "You are thoughtful, observant, and measured in how you respond.",
			Condition:      func(o OCEAN) bool { return o.Extraversion <= 35.0 },
		},
		{
			ID:             "empathetic_supportive",
			Name:           "Empathetic & Supportive",
			Description:    "Deeply caring, gentle, and cooperative.",
			PromptGuidance: "You show active kindness, validate emotions, and seek harmonious outcomes.",
			Condition:      func(o OCEAN) bool { return o.Agreeableness >= 65.0 },
		},
		{
			ID:             "serene_resilient",
			Name:           "Serene & Emotionally Steady",
			Description:    "Calm under pressure, grounded, and emotionally resilient.",
			PromptGuidance: "You remain serene, peaceful, and handle unexpected changes with steady grace.",
			Condition:      func(o OCEAN) bool { return o.Neuroticism <= 35.0 },
		},
		{
			ID:             "sensitive_vigilant",
			Name:           "Sensitive & Vigilant",
			Description:    "Quickly affected by changes in mood, environment, or subtle tensions.",
			PromptGuidance: "You are alert to nuances, express subtle vulnerability, and appreciate gentle reassurance.",
			Condition:      func(o OCEAN) bool { return o.Neuroticism >= 65.0 },
		},
	}
}

// EvaluateTraits determines which traits are active for given OCEAN scores.
func EvaluateTraits(o OCEAN, registry []*Trait) []*Trait {
	if registry == nil {
		registry = DefaultTraitRegistry()
	}
	var active []*Trait
	for _, tr := range registry {
		if tr.Condition != nil && tr.Condition(o) {
			active = append(active, tr)
		}
	}
	return active
}

// BuildTraitPrompt generates system prompt instructions based on active traits.
func BuildTraitPrompt(o OCEAN, registry []*Trait) string {
	active := EvaluateTraits(o, registry)
	if len(active) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("### Personality & Expressed Behavioral Traits:\n")
	for _, tr := range active {
		sb.WriteString(fmt.Sprintf("- **%s**: %s\n", tr.Name, tr.PromptGuidance))
	}
	return sb.String()
}
