# community-mapping

Community-maintained cross-database identity and volume mapping dataset for manga, manhwa, manhua, and light novels across **MyAnimeList (MAL)**, **AniList (AL)**, and **Kitsu**.

This repository maps external catalog IDs and volume-to-chapter boundaries so media managers, scrobblers, and reader applications (such as Kavita, Komga, Mihon, Suwayomi, and AniSync) can match works and accurately convert chapter progress to volume completions without ambiguity.

---

## Available Distributions

Pre-compiled distributions are automatically published on every release:

| Distribution | Format | Description |
|---|---|---|
| `mapping.db` | SQLite | Complete database with indexed tables for entries, volumes, and search titles |
| `comics.db` | SQLite | Visual formats only (manga, manhwa, manhua) |
| `novels.db` | SQLite | Light novels and web novels |
| `mapping.lean.json` | JSON | Minimal JSON dataset (IDs and volume boundaries, no title index overhead) |
| `mapping.json` | JSON | Complete JSON dataset including all alternative titles |
| `distributions-bundle.zip` | Archive | Single zip bundle containing all SQLite databases, JSONs, and SQL dumps |

Distributions can be downloaded directly from the [latest Release](https://github.com/dango-service/community-mapping/releases/latest).

---

## Repository Structure

```text
community-mapping/
|-- master/                    Master YAML dataset (alphabetically ordered)
|   |-- manga.yaml             Japanese manga
|   |-- manhwa.yaml            Korean manhwa and webtoons
|   |-- manhua.yaml            Chinese and Taiwanese manhua
|   |-- novel.yaml             Light novels and web novels
|   `-- overrides.yaml         Manual overrides
|-- data/
|   `-- titles.json.gz         Snapshot of verified alternative titles
|-- pkg/mapping/               Dataset compilation, validation, and normalization
|-- cmd/mapping/               Go CLI entrypoint (build, validate, clean)
`-- schema/                    JSON Schema and content type definitions
```

---

## Data Model

Entries in `master/*.yaml` maintain a minimal, human-auditable structure:

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

### Fields
- `title`: Canonical romanized or English title of the work.
- `type`: Content format (`manga`, `manhwa`, `manhua`, `novel`, `oneshot`).
- `mal_id`: Numeric MyAnimeList ID (or list of IDs if catalog split; `0` if unindexed).
- `anilist_id`: Numeric AniList ID (or list of IDs; `0` if unindexed).
- `kitsu_id`: Numeric Kitsu ID (or list of IDs; `0` if unindexed).
- `volumes`: Optional key-value map specifying the ending chapter of each physical volume.

> Official alternative names (Romaji, English, and Native titles) are automatically injected from the verified metadata snapshot during distribution builds, keeping the master YAML files concise and free of redundant text.

---

## Local Tooling

The dataset compiler is written in Go:

### Build Distributions
```bash
go run ./cmd/mapping build
```
Compiles `master/*.yaml` into SQLite databases, partitioned subsets, and lookup indexes in `generated/`.

### Validate Integrity
```bash
go run ./cmd/mapping validate
```
Checks for ID uniqueness, required fields, strictly sequential volume ranges, and data consistency.

### Format and Lint
```bash
go run ./cmd/mapping clean
go run ./cmd/mapping clean --apply
```

---

## Contributing

Pull requests adding missing mappings or correcting volume boundaries are welcome.

1. Fork the repository.
2. Edit the relevant file in `master/` (`manga.yaml`, `manhwa.yaml`, `manhua.yaml`, or `novel.yaml`).
3. Ensure entries remain sorted alphabetically by `title`.
4. Run validation before committing:
   ```bash
   go run ./cmd/mapping build
   go run ./cmd/mapping validate
   ```
5. Submit your Pull Request. CI will verify the dataset automatically.

---

## License

This project is licensed under the [MIT License](LICENSE).
