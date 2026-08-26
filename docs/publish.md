# Publish / release

Ship a verified feature worktree: land on main, bump `VERSION.txt`, tag, sign Firefox, publish the GitHub release, then remove the worktree.

## Steps

| # | cwd | Command |
|---|-----|---------|
| 1 | Feature worktree | `wrk --add-all --commit -m "<summary>" --merge-back` |
| 2 | Main | `cd "$(wrk --main --where)" && go run ./script/bump-version` |
| 3 | Main | `wrk --add-all --commit -m "bump version X -> Y" --push --tag-next --sync` |
| 4 | Main | `go run ./script/browser-agent/firefox/sign && go run ./script/github/release` |
| 5 | Feature worktree | `wrk --done` |

- **Step 2:** patch bump by default; `--minor` / `--major` / `--dry-run` available. Does not commit, tag, or sign.
- **Step 3:** replace `X` / `Y` with the versions printed by bump.
- **Step 4:** Firefox AMO sign may take **minutes to hours** — wait for it before relying on the release. Hydrate/archives detail: [assets-hydrate.md](./assets-hydrate.md).
- **Step 5:** only after land + release look good; do not `--done` with unmerged work still only on the worktree.

## Example

```bash
# feature worktree
wrk --add-all --commit -m "session page: clearer extension upgrade callout" --merge-back

cd "$(wrk --main --where)"
go run ./script/bump-version
# e.g. 1.0.15 → 1.0.16

wrk --add-all --commit -m "bump version 1.0.15 -> 1.0.16" --push --tag-next --sync
go run ./script/browser-agent/firefox/sign
go run ./script/github/release

# feature worktree
wrk --done
```
