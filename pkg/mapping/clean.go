package mapping

import (
	"fmt"
	"path/filepath"
	"strings"
)

// CleanOptions configures the dataset cleaning process.
type CleanOptions struct {
	MasterDir string
	Apply     bool
}

// CleanStats summarizes the actions taken by Clean.
type CleanStats struct {
	TotalEntries    int
	TitlesTrimmed   int
	SynonymsRemoved int
	FilesModified   int
}

// Clean normalizes titles, removes duplicate/empty synonyms, and removes synonyms identical to the title.
func Clean(opts CleanOptions) (CleanStats, error) {
	stats := CleanStats{}
	files := []string{"manga.yaml", "manhwa.yaml", "manhua.yaml", "novel.yaml", "overrides.yaml"}

	for _, fname := range files {
		fpath := filepath.Join(opts.MasterDir, fname)
		entries, err := LoadYAMLFile(fpath)
		if err != nil {
			return stats, fmt.Errorf("loading %s: %w", fname, err)
		}
		if len(entries) == 0 {
			continue
		}

		fileModified := false
		for i := range entries {
			stats.TotalEntries++

			// 1. Trim whitespace and normalize quotes on Title
			trimmedTitle := strings.TrimSpace(entries[i].Title)
			if (strings.HasPrefix(trimmedTitle, "\"") && strings.HasSuffix(trimmedTitle, "\"")) ||
				(strings.HasPrefix(trimmedTitle, "'") && strings.HasSuffix(trimmedTitle, "'")) {
				unquoted := strings.Trim(trimmedTitle, "\"'")
				if len(unquoted) > 0 {
					trimmedTitle = unquoted
				}
			}
			if strings.Contains(trimmedTitle, "''") {
				trimmedTitle = strings.ReplaceAll(trimmedTitle, "''", "'")
			}
			if strings.HasPrefix(trimmedTitle, "'") && strings.Contains(trimmedTitle, "\"") {
				trimmedTitle = strings.ReplaceAll(trimmedTitle, "\"", "")
			}
			if trimmedTitle != entries[i].Title {
				entries[i].Title = trimmedTitle
				stats.TitlesTrimmed++
				fileModified = true
			}

			// 2. Clean synonyms
			cleanMain := CleanTitle(entries[i].Title)
			seenSyn := make(map[string]bool)
			var validSyns []string

			for _, s := range entries[i].Synonyms {
				trimmed := strings.TrimSpace(s)
				if trimmed == "" {
					stats.SynonymsRemoved++
					fileModified = true
					continue
				}

				cSyn := CleanTitle(trimmed)
				// Discard if identical to main title
				if cSyn == "" || cSyn == cleanMain {
					stats.SynonymsRemoved++
					fileModified = true
					continue
				}

				// Deduplicate
				if seenSyn[cSyn] {
					stats.SynonymsRemoved++
					fileModified = true
					continue
				}

				seenSyn[cSyn] = true
				validSyns = append(validSyns, trimmed)
			}

			if len(validSyns) != len(entries[i].Synonyms) {
				fileModified = true
			}
			entries[i].Synonyms = validSyns
		}

		if fileModified {
			stats.FilesModified++
			if opts.Apply {
				if err := SaveYAMLFile(fpath, entries); err != nil {
					return stats, fmt.Errorf("saving %s: %w", fname, err)
				}
			}
		}
	}

	return stats, nil
}
