## Pull Request Description

Please briefly describe the entry or change you are contributing.

### Type of Contribution
- [ ] New work mapping (`manga`, `manhwa`, `manhua`)
- [ ] New work mapping (`novel`)
- [ ] Correction or override to an existing mapping

### Links for Verification
Provide at least one catalog link for verification:
- **MyAnimeList**: https://myanimelist.net/manga/...
- **AniList**: https://anilist.co/manga/...
- **Kitsu**: https://kitsu.app/manga/...

### Quality Checklist
- [ ] Verified that IDs correspond to the same work.
- [ ] If an entry does not exist on a service, used `0`.
- [ ] Ran `go run ./cmd/mapping build` and committed updated `generated/` files.
- [ ] Ran `go run ./cmd/mapping validate` (exited with zero errors).
- [ ] No emojis in commit messages or file content.
