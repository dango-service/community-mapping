package mapping

import (
	"bufio"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// BuildOptions holds paths for compiling distributions.
type BuildOptions struct {
	MasterDir    string
	GeneratedDir string
}

// LoadYAMLFile loads entries from a YAML file.
func LoadYAMLFile(path string) ([]Entry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}

	if len(strings.TrimSpace(string(data))) == 0 {
		return nil, nil
	}

	var entries []Entry
	if err := yaml.Unmarshal(data, &entries); err != nil {
		return nil, fmt.Errorf("parsing YAML from %s: %w", path, err)
	}

	return entries, nil
}

// SaveYAMLFile writes entries to a YAML file in clean sorted order.
func SaveYAMLFile(path string, entries []Entry) error {
	sort.Slice(entries, func(i, j int) bool {
		return strings.ToLower(entries[i].Title) < strings.ToLower(entries[j].Title)
	})
	data, err := yaml.Marshal(entries)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// MergeEntries combines datasets, allowing overrides to supersede base entries.
func MergeEntries(baseEntries []Entry, overrides []Entry) []Entry {
	overrideIndex := make(map[string]Entry)
	for _, ov := range overrides {
		for _, id := range ov.MALID.AllNonZero() {
			overrideIndex[fmt.Sprintf("mal:%d", id)] = ov
		}
		for _, id := range ov.AniListID.AllNonZero() {
			overrideIndex[fmt.Sprintf("al:%d", id)] = ov
		}
		for _, id := range ov.KitsuID.AllNonZero() {
			overrideIndex[fmt.Sprintf("kt:%d", id)] = ov
		}
	}

	seen := make(map[string]bool)
	var result []Entry

	for _, item := range baseEntries {
		finalItem := item

		// Check for override
		for _, id := range item.MALID.AllNonZero() {
			if ov, ok := overrideIndex[fmt.Sprintf("mal:%d", id)]; ok {
				finalItem = ov
				break
			}
		}
		for _, id := range item.AniListID.AllNonZero() {
			if ov, ok := overrideIndex[fmt.Sprintf("al:%d", id)]; ok {
				finalItem = ov
				break
			}
		}
		for _, id := range item.KitsuID.AllNonZero() {
			if ov, ok := overrideIndex[fmt.Sprintf("kt:%d", id)]; ok {
				finalItem = ov
				break
			}
		}

		key := fmt.Sprintf("%s:%s", finalItem.Type, strings.ToLower(strings.TrimSpace(finalItem.Title)))
		if !seen[key] {
			seen[key] = true
			result = append(result, finalItem)
		}
	}

	sort.Slice(result, func(i, j int) bool {
		return strings.ToLower(result[i].Title) < strings.ToLower(result[j].Title)
	})

	return result
}

// Build compiles master YAML files to generated JSON distribution packages.
func Build(opts BuildOptions) error {
	mangaPath := filepath.Join(opts.MasterDir, "manga.yaml")
	manhwaPath := filepath.Join(opts.MasterDir, "manhwa.yaml")
	manhuaPath := filepath.Join(opts.MasterDir, "manhua.yaml")
	novelPath := filepath.Join(opts.MasterDir, "novel.yaml")
	overridesPath := filepath.Join(opts.MasterDir, "overrides.yaml")

	manga, _ := LoadYAMLFile(mangaPath)
	manhwa, _ := LoadYAMLFile(manhwaPath)
	manhua, _ := LoadYAMLFile(manhuaPath)
	novels, _ := LoadYAMLFile(novelPath)
	overrides, _ := LoadYAMLFile(overridesPath)

	allBase := append(append(append(manga, manhwa...), manhua...), novels...)
	allEntries := MergeEntries(allBase, overrides)

	// Load official titles/synonyms snapshot
	snapshotPath := filepath.Join(opts.MasterDir, "..", "data", "titles.json.gz")
	titleSnapshot, _ := loadTitleSnapshot(snapshotPath)

	// Enrich allEntries with official synonyms from snapshot
	for i := range allEntries {
		e := &allEntries[i]
		seen := make(map[string]bool)
		for _, s := range e.Synonyms {
			seen[strings.ToLower(strings.TrimSpace(s))] = true
		}

		// Find snapshot synonyms by MAL, AL, KT, or title
		var snapSyn []string
		if malID := e.MALID.First(); malID > 0 {
			snapSyn = titleSnapshot[fmt.Sprintf("mal:%d", malID)]
		}
		if len(snapSyn) == 0 {
			if alID := e.AniListID.First(); alID > 0 {
				snapSyn = titleSnapshot[fmt.Sprintf("al:%d", alID)]
			}
		}
		if len(snapSyn) == 0 {
			if ktID := e.KitsuID.First(); ktID > 0 {
				snapSyn = titleSnapshot[fmt.Sprintf("kt:%d", ktID)]
			}
		}
		if len(snapSyn) == 0 {
			snapSyn = titleSnapshot[fmt.Sprintf("title:%s", strings.ToLower(strings.TrimSpace(e.Title)))]
		}

		cleanMain := strings.ToLower(strings.TrimSpace(e.Title))
		for _, s := range snapSyn {
			ts := strings.TrimSpace(s)
			low := strings.ToLower(ts)
			if ts != "" && !seen[low] && low != cleanMain {
				seen[low] = true
				e.Synonyms = append(e.Synonyms, ts)
			}
		}
	}

	_ = os.RemoveAll(opts.GeneratedDir)
	if err := os.MkdirAll(opts.GeneratedDir, 0755); err != nil {
		return fmt.Errorf("creating generated dir: %w", err)
	}

	// 1. Separate into groups
	var comics []Entry
	var novelList []Entry
	var mangaList []Entry
	var manhwaList []Entry
	var manhuaList []Entry
	var leanEntries []LeanEntry

	for _, e := range allEntries {
		leanEntries = append(leanEntries, e.ToLean())

		t := strings.ToLower(e.Type)
		if t == "novel" {
			novelList = append(novelList, e)
		} else {
			comics = append(comics, e)
		}

		switch t {
		case "manga", "oneshot":
			mangaList = append(mangaList, e)
		case "manhwa":
			manhwaList = append(manhwaList, e)
		case "manhua":
			manhuaList = append(manhuaList, e)
		}
	}

	// 2. mapping.json & mapping.min.json (Full dataset: comics + novels)
	if err := writeJSON(filepath.Join(opts.GeneratedDir, "mapping.json"), allEntries, false); err != nil {
		return err
	}
	if err := writeJSON(filepath.Join(opts.GeneratedDir, "mapping.min.json"), allEntries, true); err != nil {
		return err
	}

	// 3. mapping.lean.json & mapping.lean.min.json (Lean dataset: IDs + volumes without heavy synonyms)
	if err := writeJSON(filepath.Join(opts.GeneratedDir, "mapping.lean.json"), leanEntries, false); err != nil {
		return err
	}
	if err := writeJSON(filepath.Join(opts.GeneratedDir, "mapping.lean.min.json"), leanEntries, true); err != nil {
		return err
	}

	// 4. comics.json & comics.min.json (The visual formats: manga + manhwa + manhua)
	if err := writeJSON(filepath.Join(opts.GeneratedDir, "comics.json"), comics, false); err != nil {
		return err
	}
	if err := writeJSON(filepath.Join(opts.GeneratedDir, "comics.min.json"), comics, true); err != nil {
		return err
	}

	// 5. novels.json & novels.min.json (Novels only)
	if err := writeJSON(filepath.Join(opts.GeneratedDir, "novels.json"), novelList, false); err != nil {
		return err
	}
	if err := writeJSON(filepath.Join(opts.GeneratedDir, "novels.min.json"), novelList, true); err != nil {
		return err
	}

	// 6. Granular format distributions
	if err := writeJSON(filepath.Join(opts.GeneratedDir, "manga.json"), mangaList, false); err != nil {
		return err
	}
	if err := writeJSON(filepath.Join(opts.GeneratedDir, "manga.min.json"), mangaList, true); err != nil {
		return err
	}
	if err := writeJSON(filepath.Join(opts.GeneratedDir, "manhwa.json"), manhwaList, false); err != nil {
		return err
	}
	if err := writeJSON(filepath.Join(opts.GeneratedDir, "manhwa.min.json"), manhwaList, true); err != nil {
		return err
	}
	if err := writeJSON(filepath.Join(opts.GeneratedDir, "manhua.json"), manhuaList, false); err != nil {
		return err
	}
	if err := writeJSON(filepath.Join(opts.GeneratedDir, "manhua.min.json"), manhuaList, true); err != nil {
		return err
	}

	// 7. Lookups by service ID
	byMAL := make(map[string]Entry)
	byAniList := make(map[string]Entry)
	byKitsu := make(map[string]Entry)
	byTitle := make(map[string]Entry)
	var unmapped []Entry

	for _, e := range allEntries {
		malIDs := e.MALID.AllNonZero()
		alIDs := e.AniListID.AllNonZero()
		ktIDs := e.KitsuID.AllNonZero()

		for _, id := range malIDs {
			byMAL[strconv.Itoa(id)] = e
		}
		for _, id := range alIDs {
			byAniList[strconv.Itoa(id)] = e
		}
		for _, id := range ktIDs {
			byKitsu[strconv.Itoa(id)] = e
		}

		// Index by clean canonical title
		if ct := CleanTitle(e.Title); ct != "" {
			if _, exists := byTitle[ct]; !exists {
				byTitle[ct] = e
			}
		}
		// Index by clean synonyms
		for _, syn := range e.Synonyms {
			if cs := CleanTitle(syn); cs != "" {
				if _, exists := byTitle[cs]; !exists {
					byTitle[cs] = e
				}
			}
		}

		if len(malIDs) == 0 || len(alIDs) == 0 || len(ktIDs) == 0 {
			unmapped = append(unmapped, e)
		}
	}

	if err := writeJSON(filepath.Join(opts.GeneratedDir, "by-mal.json"), byMAL, false); err != nil {
		return err
	}
	if err := writeJSON(filepath.Join(opts.GeneratedDir, "by-mal.min.json"), byMAL, true); err != nil {
		return err
	}

	if err := writeJSON(filepath.Join(opts.GeneratedDir, "by-anilist.json"), byAniList, false); err != nil {
		return err
	}
	if err := writeJSON(filepath.Join(opts.GeneratedDir, "by-anilist.min.json"), byAniList, true); err != nil {
		return err
	}

	if err := writeJSON(filepath.Join(opts.GeneratedDir, "by-kitsu.json"), byKitsu, false); err != nil {
		return err
	}
	if err := writeJSON(filepath.Join(opts.GeneratedDir, "by-kitsu.min.json"), byKitsu, true); err != nil {
		return err
	}

	if err := writeJSON(filepath.Join(opts.GeneratedDir, "by-title.json"), byTitle, false); err != nil {
		return err
	}
	if err := writeJSON(filepath.Join(opts.GeneratedDir, "by-title.min.json"), byTitle, true); err != nil {
		return err
	}

	// 8. unmapped.yaml
	unmappedYAML, err := yaml.Marshal(unmapped)
	if err != nil {
		return fmt.Errorf("marshaling unmapped.yaml: %w", err)
	}
	if err := os.WriteFile(filepath.Join(opts.GeneratedDir, "unmapped.yaml"), unmappedYAML, 0644); err != nil {
		return err
	}

	// 9. SQLite databases & SQL dumps
	// 9.1 Complete unified database (all formats)
	if err := buildSingleSQLite(opts.GeneratedDir, "mapping", allEntries); err != nil {
		return fmt.Errorf("building mapping sqlite: %w", err)
	}

	// 9.2 Comics database (manga + manhwa + manhua)
	if err := buildSingleSQLite(opts.GeneratedDir, "comics", comics); err != nil {
		return fmt.Errorf("building comics sqlite: %w", err)
	}

	// 9.3 Category databases
	if err := buildSingleSQLite(opts.GeneratedDir, "manga", mangaList); err != nil {
		return fmt.Errorf("building manga sqlite: %w", err)
	}
	if err := buildSingleSQLite(opts.GeneratedDir, "manhwa", manhwaList); err != nil {
		return fmt.Errorf("building manhwa sqlite: %w", err)
	}
	if err := buildSingleSQLite(opts.GeneratedDir, "manhua", manhuaList); err != nil {
		return fmt.Errorf("building manhua sqlite: %w", err)
	}
	if err := buildSingleSQLite(opts.GeneratedDir, "novels", novelList); err != nil {
		return fmt.Errorf("building novels sqlite: %w", err)
	}

	return nil
}

func buildSingleSQLite(generatedDir string, baseName string, entries []Entry) error {
	sqlPath := filepath.Join(generatedDir, baseName+".sql")
	dbPath := filepath.Join(generatedDir, baseName+".db")

	f, err := os.Create(sqlPath)
	if err != nil {
		return fmt.Errorf("creating %s.sql: %w", baseName, err)
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	w.WriteString("PRAGMA synchronous = OFF;\n")
	w.WriteString("PRAGMA journal_mode = MEMORY;\n")
	w.WriteString("BEGIN TRANSACTION;\n\n")

	w.WriteString("DROP TABLE IF EXISTS entries;\n")
	w.WriteString("CREATE TABLE entries (\n")
	w.WriteString("    id INTEGER PRIMARY KEY,\n")
	w.WriteString("    title TEXT NOT NULL,\n")
	w.WriteString("    type TEXT NOT NULL,\n")
	w.WriteString("    mal_id INTEGER,\n")
	w.WriteString("    anilist_id INTEGER,\n")
	w.WriteString("    kitsu_id INTEGER\n")
	w.WriteString(");\n\n")

	w.WriteString("DROP TABLE IF EXISTS synonyms;\n")
	w.WriteString("CREATE TABLE synonyms (\n")
	w.WriteString("    entry_id INTEGER NOT NULL,\n")
	w.WriteString("    title TEXT NOT NULL\n")
	w.WriteString(");\n\n")

	w.WriteString("DROP TABLE IF EXISTS volumes;\n")
	w.WriteString("CREATE TABLE volumes (\n")
	w.WriteString("    entry_id INTEGER NOT NULL,\n")
	w.WriteString("    volume INTEGER NOT NULL,\n")
	w.WriteString("    end_chapter INTEGER NOT NULL,\n")
	w.WriteString("    PRIMARY KEY (entry_id, volume)\n")
	w.WriteString(");\n\n")

	escapeSQL := func(s string) string {
		return strings.ReplaceAll(s, "'", "''")
	}

	for i, e := range entries {
		entryID := i + 1
		w.WriteString(fmt.Sprintf("INSERT INTO entries VALUES (%d, '%s', '%s', %d, %d, %d);\n",
			entryID,
			escapeSQL(e.Title),
			escapeSQL(e.Type),
			e.MALID.First(),
			e.AniListID.First(),
			e.KitsuID.First(),
		))

		for _, syn := range e.Synonyms {
			cleanSyn := strings.TrimSpace(syn)
			if cleanSyn != "" {
				w.WriteString(fmt.Sprintf("INSERT INTO synonyms VALUES (%d, '%s');\n", entryID, escapeSQL(cleanSyn)))
			}
		}

		for volStr, ch := range e.Volumes {
			vNum, err := strconv.Atoi(volStr)
			if err == nil {
				w.WriteString(fmt.Sprintf("INSERT INTO volumes VALUES (%d, %d, %d);\n", entryID, vNum, ch))
			}
		}
	}

	w.WriteString("\nCREATE INDEX idx_entries_title ON entries(title);\n")
	w.WriteString("CREATE INDEX idx_entries_mal ON entries(mal_id);\n")
	w.WriteString("CREATE INDEX idx_entries_anilist ON entries(anilist_id);\n")
	w.WriteString("CREATE INDEX idx_entries_kitsu ON entries(kitsu_id);\n")
	w.WriteString("CREATE INDEX idx_entries_type ON entries(type);\n")
	w.WriteString("CREATE INDEX idx_synonyms_title ON synonyms(title);\n")
	w.WriteString("CREATE INDEX idx_synonyms_entry ON synonyms(entry_id);\n")
	w.WriteString("CREATE INDEX idx_volumes_entry ON volumes(entry_id);\n")
	w.WriteString("COMMIT;\n")

	if err := w.Flush(); err != nil {
		return fmt.Errorf("flushing %s.sql: %w", baseName, err)
	}

	// If sqlite3 binary is installed, build the binary sqlite file
	if sqliteBin, err := exec.LookPath("sqlite3"); err == nil && sqliteBin != "" {
		_ = os.Remove(dbPath)
		cmd := exec.Command(sqliteBin, dbPath)
		cmd.Stdin = strings.NewReader(fmt.Sprintf(".read \"%s\"\n", filepath.ToSlash(sqlPath)))
		if out, err := cmd.CombinedOutput(); err != nil {
			fmt.Printf("Warning: sqlite3 execution error for %s: %v (%s)\n", baseName, err, string(out))
		} else {
			fmt.Printf("Generated SQLite database: %s\n", filepath.ToSlash(filepath.Join("generated", baseName+".db")))
		}
	}

	return nil
}

func writeJSON(path string, data interface{}, minified bool) error {
	var raw []byte
	var err error
	if minified {
		raw, err = json.Marshal(data)
	} else {
		raw, err = json.MarshalIndent(data, "", "  ")
	}
	if err != nil {
		return fmt.Errorf("marshaling %s: %w", path, err)
	}
	if !minified {
		raw = append(raw, '\n')
	}
	return os.WriteFile(path, raw, 0644)
}

type titleSnapshotRecord struct {
	Title    string   `json:"title"`
	MAL      int      `json:"mal,omitempty"`
	AL       int      `json:"al,omitempty"`
	KT       int      `json:"kt,omitempty"`
	Synonyms []string `json:"synonyms"`
}

func loadTitleSnapshot(path string) (map[string][]string, error) {
	result := make(map[string][]string)
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return result, nil
		}
		return nil, err
	}
	defer f.Close()

	var reader io.Reader = f
	if strings.HasSuffix(path, ".gz") {
		gz, err := gzip.NewReader(f)
		if err != nil {
			return nil, err
		}
		defer gz.Close()
		reader = gz
	}

	var records []titleSnapshotRecord
	if err := json.NewDecoder(reader).Decode(&records); err != nil {
		return nil, err
	}

	for _, r := range records {
		if len(r.Synonyms) == 0 {
			continue
		}
		if r.MAL > 0 {
			result[fmt.Sprintf("mal:%d", r.MAL)] = r.Synonyms
		}
		if r.AL > 0 {
			result[fmt.Sprintf("al:%d", r.AL)] = r.Synonyms
		}
		if r.KT > 0 {
			result[fmt.Sprintf("kt:%d", r.KT)] = r.Synonyms
		}
		if r.Title != "" {
			result[fmt.Sprintf("title:%s", strings.ToLower(strings.TrimSpace(r.Title)))] = r.Synonyms
		}
	}

	return result, nil
}
