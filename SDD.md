# Software Design Document (SDD) - community-mapping

Status: Approved v2.2
Target Path: `community-mapping/SDD.md`

## 1. Objective

A community-maintained, public repository mapping identical works across major manga, manhwa, manhua, and novel metadata databases:
**MyAnimeList (MAL), AniList (AL), and Kitsu**, including volume ending chapter boundaries and alternative titles for:
- Manga (Japanese visual comics)
- Manhwa (Korean visual comics / webtoons)
- Manhua (Chinese and Taiwanese visual comics)
- Light Novels and Web Novels (Japanese, Korean, Chinese)
- One-shots

Architecture highlights:
- A **master** dataset authored and maintained in simple YAML (`master/*.yaml`) designed for easy community contributions.
- Automated **generated** datasets (`generated/*.json`, `generated/by-*.json`, `generated/*.lean.min.json`) consumed directly by sync services and media managers.
- High-performance **Go tooling** (`pkg/mapping/`, `cmd/mapping/`) providing instant builds and microsecond CI validation with zero runtime overhead.
- Value `0` represents "unmapped / does not exist on this provider".

## 2. Contributor-Centric Philosophy

Earlier drafts included bot-specific fields like `subtitles: []`, `sources: [...]`, `confidence: 0.99`, and `notes: [...]`.
These created unnecessary cognitive burden and discouraged human contributions.

In v2.0+:
- **No boilerplate fields**: If an entry has subtitles or alternate titles, they belong in `synonyms`. No mandatory empty lists (`subtitles: []`).
- **No script baggage in master**: Internal scoring, signals, and confidence weights exist only inside candidate selection logic, not in master YAML files.
- **Minimal human entry**: A contributor only specifies what is necessary: title, type, database IDs, and optionally volume ending chapters.

## 3. Data Format Strategy: YAML + JSON

| Layer | Format | Rationale |
|---|---|---|
| **Master** (human-edited & reviewed) | **YAML** | Human-readable, simple to edit on GitHub web interface, clean and reviewable diffs in Pull Requests. |
| **Generated** (application-consumed) | **JSON** (+ minified & lean) | Ultra-fast parsing in Go, Rust, Python, and JavaScript; eliminates YAML boolean and octal edge cases. |

Rule: Human contributors never edit `generated/*.json` directly. CI fails if generated files diverge from master source files.

## 4. Repository Layout

```
community-mapping/
|-- SDD.md                     # Software Design Document
|-- README.md                  # Contributor and usage guide
|-- LICENSE                    # MIT License
|-- schema/
|   |-- entry.schema.json      # JSON Schema validating every entry
|   `-- enums.yaml             # Allowed types and formats
|-- master/                    # Human-edited master YAML files (A-Z sorted, NO JSON HERE)
|   |-- manga.yaml             # Japanese manga & oneshots
|   |-- manhwa.yaml            # Korean manhwa & webtoons
|   |-- manhua.yaml            # Chinese & Taiwanese manhua
|   |-- novel.yaml             # Light novels & web novels
|   `-- overrides.yaml         # Manual corrections that take precedence
|-- generated/                 # Compiled distributions (DO NOT EDIT MANUALLY)
|   |-- mapping.db             # Complete unified SQLite database (all formats) (+ .sql)
|   |-- comics.db              # Visual comics SQLite database (manga + manhwa + manhua) (+ .sql)
|   |-- manga.db               # Manga SQLite database (+ .sql)
|   |-- manhwa.db              # Manhwa SQLite database (+ .sql)
|   |-- manhua.db              # Manhua SQLite database (+ .sql)
|   |-- novels.db              # Novels SQLite database (+ .sql)
|   |-- mapping.json           # Complete mapping array (+ min.json)
|   |-- mapping.lean.json      # Clean array without synonyms (+ min.json)
|   |-- comics.json            # Visual formats (manga + manhwa + manhua) (+ min.json)
|   |-- manga.json             # Manga only (+ min.json)
|   |-- manhwa.json            # Manhwa only (+ min.json)
|   |-- manhua.json            # Manhua only (+ min.json)
|   |-- novels.json            # Novels only (+ min.json)
|   |-- by-mal.json            # Direct index: MAL ID -> Entry (+ min.json)
|   |-- by-anilist.json        # Direct index: AniList ID -> Entry (+ min.json)
|   |-- by-kitsu.json          # Direct index: Kitsu ID -> Entry (+ min.json)
|   |-- by-title.json          # Direct index: Clean Title/Synonym -> Entry (+ min.json)
|   `-- unmapped.yaml          # Entries with any missing or zero ID
|-- pkg/mapping/               # Core Go package
|   |-- entry.go               # Entry model, IDVal, IDStrVal, and VolumeMap
|   |-- build.go               # Fast compiler from YAML to JSON, SQLite databases, and lookup indexes
|   |-- validate.go            # Microsecond validation suite (schema, IDs, volume order)
|   |-- clean.go               # Dataset normalizer and synonym deduplicator
|   |-- volumes.go             # Volume map validation and ongoing series pruning
|   `-- similarity.go          # Title cleaning and normalization utilities
|-- cmd/mapping/               # Standalone Go CLI
|   `-- main.go                # Subcommands: build, validate, clean
`-- .github/
    |-- workflows/
    |   |-- validate.yml       # Fast Go-based CI validation
    |   `-- release.yml        # Build and publish release distributions
    `-- PULL_REQUEST_TEMPLATE.md
```

## 5. Master Data Model

A master entry is concise and direct:

```yaml
- title: Solo Leveling
  type: manhwa # manga | manhwa | manhua | novel | oneshot
  mal_id: 121496
  anilist_id: 105398
  kitsu_id: 38486
  volumes:
    1: 47      # Volume 1 ends at chapter 47
    2: 68      # Volume 2 ends at chapter 68
```

### Rules
1. **Unindexed / Missing IDs**: Use `0` if an entry does not exist on that service (e.g. web novels not indexed on MAL).
2. **Split Entries (1-to-many)**: If a service splits a single publication into multiple catalog entries, a list of IDs is supported: e.g. `mal_id: [123, 456]` or `anilist_id: [109266, 130008]`.
3. **Volumes**: Optional map indicating ending chapter of each volume (e.g. `1: 7, 2: 16`). Keys are numerically sorted.

## 6. Go Tooling Performance

Using Go delivers significant advantages:
- **Build Speed**: Compiles `master/*.yaml` into all generated distributions, SQLite databases, and lookups in ~3 seconds for > 23,000 entries.
- **Validation Speed**: Checks schema, uniqueness of IDs across the entire repository, and sequential volume rules in < 400 milliseconds.
- **Zero Runtime Dependencies**: No `pip install`, virtualenvs, or external runtimes in GitHub Actions.

## 7. Public Tooling Commands

- **`build`**: Compiles master YAML files into all distributions, lean datasets, and SQLite databases in `generated/`.
- **`validate`**: Verifies data integrity, volume sequence rules, and ID uniqueness.
- **`clean`**: Normalizes whitespace and fixes formatting on master files.
