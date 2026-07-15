# Scenario

**Feature**: chaos-useful-links dynamic seed corpus from --links / --random-links

```
# links file
Link File (md/text) -> Extractor -> Seeds (deduped, historical-filtered)
  -> Resolved{Source.Type=links, Counts, Seeds}

# random catalog
Random Catalog -> RandomSeeds / ResolveSeedSource(random)
  -> ≥3 public https seeds (example.com, google.com, baidu.com)

# source mutex
Operator flags (--links | --random-links) -> Source Resolver
  -> error if neither or both
```

## Preconditions

- Package `github.com/xhd2015/browser-agent/script/debug/chaos-useful-links/seedload`
  is the implementer target (not yet present — classic TDD RED).
- Tree root is `tests/chaos-useful-links/`; **ModuleRoot** =
  `filepath.Clean(filepath.Join(DOCTEST_ROOT, "..", ".."))`.
- Fixtures live under `DOCTEST_ROOT/testdata/`; no host absolute org paths.
- **Classic TDD**: dynamic link load APIs and CLI flags are not implemented;
  `doctest test` must be RED (compile fail and/or assertion fail).
- No Chrome, no network fetch of remote link files.

## Steps

1. Resolve `ModuleRoot` from `DOCTEST_ROOT`.
2. Leave `Mode` empty at root (grouping Setup sets Mode).
3. Shared helpers in root `DOCTEST.md` are available to all leaves.

## Context

- Spec version **0.0.2**.
- Historical headings (case-insensitive match on heading text): Historical,
  Archived, Deprecated — skip until next same-or-higher level heading.
- MaxSeeds `0` means keep all after dedupe; positive N caps after dedupe.
- Seed.ID should be human-readable slug + short stable URL hash (asserted
  lightly: non-empty id + non-empty url when seeds present).

```go
import (
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.ModuleRoot = filepath.Clean(filepath.Join(DOCTEST_ROOT, "..", ".."))
	return nil
}
```
