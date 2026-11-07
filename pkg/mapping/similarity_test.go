package mapping

import (
	"testing"
)

func TestCleanTitle(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Solo Leveling (manhwa)", "solo leveling"},
		{"Berserk (Manga)!", "berserk"},
		{"Omniscient Reader's Viewpoint [webtoon]", "omniscient reader s viewpoint"},
		{"Overlord (Novel)", "overlord"},
		{"   Multiple   Spaces   ", "multiple spaces"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := CleanTitle(tt.input)
			if got != tt.expected {
				t.Errorf("CleanTitle(%q) = %q; want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestTitleSimilarity(t *testing.T) {
	tests := []struct {
		titleA   string
		titleB   string
		minScore float64
		maxScore float64
	}{
		{"Solo Leveling", "Solo Leveling (manhwa)", 1.0, 1.0},
		{"Tower of God", "God of Highschool", 0.4, 0.6},
		{"Berserk", "Bleach", 0.0, 0.0},
		{"The Beginning After The End", "Beginning After The End", 0.7, 1.0},
		{"", "Test", 0.0, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.titleA+" vs "+tt.titleB, func(t *testing.T) {
			score := TitleSimilarity(tt.titleA, tt.titleB)
			if score < tt.minScore || score > tt.maxScore {
				t.Errorf("TitleSimilarity(%q, %q) = %f; want between %f and %f",
					tt.titleA, tt.titleB, score, tt.minScore, tt.maxScore)
			}
		})
	}
}

func TestExtractYear(t *testing.T) {
	tests := []struct {
		input string
		want  int
	}{
		{"2018-03-04", 2018},
		{"1999-12-31", 1999},
		{"2025", 2025},
		{"invalid", 0},
		{"", 0},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := ExtractYear(tt.input)
			if got != tt.want {
				t.Errorf("ExtractYear(%q) = %d; want %d", tt.input, got, tt.want)
			}
		})
	}
}
