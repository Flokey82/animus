package traits

import (
	"strings"
	"testing"
)

func TestEvaluateTraits(t *testing.T) {
	ocean := OCEAN{
		Openness:          80, // Curious
		Conscientiousness: 90, // Methodical
		Extraversion:      20, // Solitary
		Agreeableness:     85, // Empathetic
		Neuroticism:       10, // Serene
	}

	active := EvaluateTraits(ocean, nil)
	names := make([]string, 0, len(active))
	for _, tr := range active {
		names = append(names, tr.Name)
	}

	joined := strings.Join(names, ", ")
	expectedTraits := []string{
		"Curious & Imaginative",
		"Methodical & Dutiful",
		"Reflective & Solitary",
		"Empathetic & Supportive",
		"Serene & Emotionally Steady",
	}

	for _, exp := range expectedTraits {
		if !strings.Contains(joined, exp) {
			t.Errorf("expected trait %q in %q", exp, joined)
		}
	}

	prompt := BuildTraitPrompt(ocean, nil)
	if !strings.Contains(prompt, "### Personality & Expressed Behavioral Traits:") {
		t.Errorf("prompt missing header: %s", prompt)
	}
}
