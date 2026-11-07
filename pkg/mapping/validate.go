package mapping

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ValidateOptions contains paths needed for validation.
type ValidateOptions struct {
	MasterDir    string
	GeneratedDir string
}

// Validate checks schema rules, ID uniqueness, and generated synchronization.
func Validate(opts ValidateOptions) []string {
	var errs []string

	files := []string{"manga.yaml", "manhwa.yaml", "manhua.yaml", "novel.yaml", "overrides.yaml"}
	var allBase []Entry
	var overrides []Entry

	for _, fname := range files {
		fpath := filepath.Join(opts.MasterDir, fname)
		entries, err := LoadYAMLFile(fpath)
		if err != nil {
			errs = append(errs, fmt.Sprintf("[%s] %v", fname, err))
			continue
		}

		if fname == "overrides.yaml" {
			overrides = entries
		} else {
			allBase = append(allBase, entries...)
		}

		for idx, e := range entries {
			if strings.TrimSpace(e.Title) == "" {
				errs = append(errs, fmt.Sprintf("[%s] Entry #%d has an empty title", fname, idx+1))
			}
			if !AllowedTypes[strings.ToLower(e.Type)] {
				errs = append(errs, fmt.Sprintf("[%s] Entry '%s' has invalid type '%s' (allowed: manga, manhwa, manhua, novel, oneshot)", fname, e.Title, e.Type))
			}
			if len(e.Volumes) > 0 && !ValidateVolumeMap(e.Volumes) {
				errs = append(errs, fmt.Sprintf("[%s] Entry '%s' has an invalid volume map: volumes must start at 1, be sequential without gaps, and have strictly increasing chapter numbers", fname, e.Title))
			}
		}
	}

	// 2. ID uniqueness validation
	seenMAL := make(map[int]string)
	seenAL := make(map[int]string)
	seenKT := make(map[int]string)

	merged := MergeEntries(allBase, overrides)
	for _, e := range merged {
		for _, id := range e.MALID.AllNonZero() {
			if existing, found := seenMAL[id]; found {
				errs = append(errs, fmt.Sprintf("Duplicate MAL ID %d found in '%s' and '%s'", id, e.Title, existing))
			} else {
				seenMAL[id] = e.Title
			}
		}

		for _, id := range e.AniListID.AllNonZero() {
			if existing, found := seenAL[id]; found {
				errs = append(errs, fmt.Sprintf("Duplicate AniList ID %d found in '%s' and '%s'", id, e.Title, existing))
			} else {
				seenAL[id] = e.Title
			}
		}

		for _, id := range e.KitsuID.AllNonZero() {
			if existing, found := seenKT[id]; found {
				errs = append(errs, fmt.Sprintf("Duplicate Kitsu ID %d found in '%s' and '%s'", id, e.Title, existing))
			} else {
				seenKT[id] = e.Title
			}
		}
	}

	// 3. Verify generated sync (if generated files exist)
	mappingPath := filepath.Join(opts.GeneratedDir, "mapping.json")
	diskData, err := os.ReadFile(mappingPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Generated files not compiled yet (normal on fresh clone where generated/ is gitignored)
			return errs
		}
		errs = append(errs, fmt.Sprintf("Error reading %s: %s", mappingPath, err))
		return errs
	}

	var diskEntries []Entry
	if err := json.Unmarshal(diskData, &diskEntries); err != nil {
		errs = append(errs, fmt.Sprintf("Failed to parse %s: %s", mappingPath, err))
		return errs
	}

	if len(diskEntries) != len(merged) {
		errs = append(errs, fmt.Sprintf("generated/mapping.json count (%d) does not match master (%d). Run 'go run ./cmd/mapping build'", len(diskEntries), len(merged)))
	} else {
		for i := range merged {
			if merged[i].Title != diskEntries[i].Title {
				errs = append(errs, fmt.Sprintf("Mismatch at index %d: master '%s' vs generated '%s'", i, merged[i].Title, diskEntries[i].Title))
				break
			}
		}
	}

	return errs
}
