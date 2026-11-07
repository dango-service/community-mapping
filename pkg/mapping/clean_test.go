package mapping

import (
	"os"
	"path/filepath"
	"testing"
)

func TestClean(t *testing.T) {
	tmpDir := t.TempDir()

	sampleManga := `
- title: "  Solo Leveling  "
  type: manhwa
  mal_id: 121496
  anilist_id: 105398
  kitsu_id: 38486
  synonyms:
    - "Solo Leveling"
    - "   "
    - "Only I Level Up"
    - "only i level up"
    - "나 혼자만 레벨업"
`
	if err := os.WriteFile(filepath.Join(tmpDir, "manhwa.yaml"), []byte(sampleManga), 0644); err != nil {
		t.Fatalf("writing sample file: %v", err)
	}

	stats, err := Clean(CleanOptions{
		MasterDir: tmpDir,
		Apply:     true,
	})
	if err != nil {
		t.Fatalf("Clean failed: %v", err)
	}

	if stats.TitlesTrimmed != 1 {
		t.Errorf("expected 1 title trimmed, got %d", stats.TitlesTrimmed)
	}
	if stats.SynonymsRemoved != 3 { // "Solo Leveling" (matches title), "   " (empty), "only i level up" (duplicate)
		t.Errorf("expected 3 synonyms removed, got %d", stats.SynonymsRemoved)
	}

	// Read back and check
	entries, err := LoadYAMLFile(filepath.Join(tmpDir, "manhwa.yaml"))
	if err != nil {
		t.Fatalf("loading cleaned file: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Title != "Solo Leveling" {
		t.Errorf("expected trimmed title 'Solo Leveling', got %q", entries[0].Title)
	}
	if len(entries[0].Synonyms) != 2 {
		t.Errorf("expected 2 synonyms left ('Only I Level Up', '나 혼자만 레벨업'), got %d", len(entries[0].Synonyms))
	}
}
