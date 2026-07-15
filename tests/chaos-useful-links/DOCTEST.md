# chaos-useful-links — dynamic link extraction (no corpus.json)

Classic TDD for the chaos harness seed source: seeds come from **`--links PATH`**
(markdown/text URL extraction) or **`--random-links`** (built-in public sites).
Hardcoded `corpus.json` and absolute host knowledge-workspace paths go away.

**No live Chrome.** **No network fetch of link files.** Library-level
`seedload` API under `script/debug/chaos-useful-links/seedload` (implementer
creates the package). Doctests call pure parse/load/resolve functions only.

| Surface | What is under test |
|---------|-------------------|
| Extract | Bare URL, markdown `[t](url)`, `` `url` ``, `<url>` → seeds |
| Normalize | Strip trailing `).,;]`; stable URL for dedupe |
| Dedupe | Same URL once after normalize |
| Historical | Skip under Historical/Archived/Deprecated headings unless include-archived |
| Random | ≥3 public https seeds (example.com, google.com, baidu.com) |
| Source resolve | Exactly one of links path / random-links; mutex + missing errors |
| MaxSeeds | 0 = all; N caps after dedupe |
| Metadata | GFM table Env/Kind/Title/Market/Link → Seed fields |

## Version

0.0.2

# DSN (Domain Specific Notion)

**Link File** is an arbitrary markdown or plain-text document on disk. It may
contain bare `http(s)` URLs, markdown links, backtick-wrapped URLs, angle-bracket
URLs, optional GFM tables with Env/Kind/Title/Market/Link columns, and section
headings that mark historical content.

**Extractor** walks the file (or raw text), finds candidate URLs, strips wrapping
and trailing punctuation, optionally rewrites bare `www.` hosts to `https://www.…`,
skips content under historical/archived/deprecated headings until the next
same-or-higher heading (unless include-archived), dedupes by normalized URL, and
builds **Seed** records (human-readable slug id + short stable URL hash, title
from table or host/path heuristic).

**Random Catalog** is a built-in list of public https sites used when the operator
asks for random-links instead of a file (includes example.com, google.com,
baidu.com and at least three seeds total).

**Source Resolver** accepts operator intent: either a links file path or
random-links. Exactly one must be set. Both or neither is an error. Resolved
output is a snapshot: source meta (type, path, optional sha256), counts
(raw/deduped/archived_skipped/after_filter), and the seed list.

**Test Client** calls the importable `seedload` package with fixtures under
`testdata/` (no host absolute paths) and asserts seeds, counts, and errors.

```text
# --links path
LoadSeedsFromFile(path, opts) -> Resolved{Source.Type=links, Seeds, Counts}

# text-only extract (same rules)
ExtractLinks(text, opts) -> Resolved{Seeds, Counts}

# --random-links
RandomSeeds() / ResolveSeedSource("", true, opts)
  -> hosts include example.com, google.com, baidu.com; len >= 3

# mutex / missing
ResolveSeedSource("", false, opts) -> error
ResolveSeedSource(path, true, opts) -> error
```

## Decision Tree

```
chaos-useful-links
├── extract/                              [LoadSeedsFromFile / ExtractLinks]
│   ├── mixed-forms/                        bare + md + backtick + angle → all seeds
│   ├── trailing-punctuation/               strip trailing ).,;]
│   └── dedupe/                             identical URLs → one seed
├── historical/                           [heading skip filter]
│   ├── skip-by-default/                    Historical/Archived/Deprecated omitted
│   └── include-archived/                   IncludeArchived keeps them
├── source/                               [ResolveSeedSource / RandomSeeds]
│   ├── random-links/                       public seeds ≥3 incl. example/google/baidu
│   ├── missing/                            neither links nor random → error
│   └── both-mutex/                         both set → error
├── max-seeds/                            [MaxSeeds option after dedupe]
│   ├── zero-keeps-all/                     MaxSeeds=0 → all unique seeds
│   └── caps-n/                             MaxSeeds=2 → exactly 2 seeds
└── metadata/                             [GFM table columns]
    └── gfm-table-columns/                  Env/Kind/Title/Market/Link map to Seed
```

### Parameter significance (high → low)

1. **API surface / Mode** — extract vs historical filter vs source resolve vs
   max-seeds vs metadata (different behaviors under test).
2. **Within extract** — form variants, punctuation, dedupe (parse outcomes).
3. **Within historical** — include-archived on vs off.
4. **Within source** — random happy path vs missing vs both (mutex).
5. **Within max-seeds** — 0 (all) vs N (cap).
6. **Metadata** — GFM column mapping (optional leaf).

## Test Index

| Leaf | Scenario |
|------|----------|
| `extract/mixed-forms` | Fixture with bare, markdown, backtick, angle URLs → ≥4 distinct seeds; URLs present |
| `extract/trailing-punctuation` | URLs followed by `).,;]` → clean URLs without those trailers |
| `extract/dedupe` | Same URL thrice (bare + md) + one other → 2 seeds after dedupe |
| `historical/skip-by-default` | Default opts skip historical/archived/deprecated section links |
| `historical/include-archived` | IncludeArchived=true keeps those section links |
| `source/random-links` | RandomSeeds / resolve random → ≥3 seeds; hosts example.com, google.com, baidu.com |
| `source/missing` | Resolve with neither links nor random → non-nil error |
| `source/both-mutex` | Resolve with both set → non-nil error |
| `max-seeds/zero-keeps-all` | 5 unique fixture URLs, MaxSeeds=0 → 5 seeds |
| `max-seeds/caps-n` | Same fixture, MaxSeeds=2 → 2 seeds |
| `metadata/gfm-table-columns` | Table Env/Kind/Title/Market/Link → Seed fields populated |

**Leaf count: 11**

## How to Run

```sh
doctest vet ./tests/chaos-useful-links
doctest test ./tests/chaos-useful-links   # RED until implementer lands seedload + CLI wiring
```

### Implementer contract (authoritative for GREEN)

Package path (importable; not `package main`):

```text
github.com/xhd2015/browser-agent/script/debug/chaos-useful-links/seedload
```

```text
type Seed struct {
    ID, URL, Kind, Env, Market, Title string
    Tags   []string
    Source string // e.g. "links:row" or "random-links"
}

type SourceMeta struct {
    Type   string // "links" | "random-links"
    Path   string // file path when Type=links; empty for random-links
    SHA256 string // of file contents when Type=links; empty for random-links
}

type Counts struct {
    Raw             int // candidates before dedupe (after historical filter when applied)
    Deduped         int
    ArchivedSkipped int
    AfterFilter     int // after MaxSeeds / kind/env filters
}

type Resolved struct {
    Source SourceMeta
    Counts Counts
    Seeds  []Seed
}

type Options struct {
    IncludeArchived bool
    MaxSeeds        int    // 0 = keep all after dedupe
    Kind            string // optional filter when metadata present
    Env             string // optional filter when metadata present
}

// ExtractLinks applies extract/normalize/dedupe/historical/max-seeds rules to text.
func ExtractLinks(text string, opts Options) (*Resolved, error)

// LoadSeedsFromFile reads path and runs the same pipeline; Source.Type=links,
// Path set, SHA256 of file bytes.
func LoadSeedsFromFile(path string, opts Options) (*Resolved, error)

// RandomSeeds returns the built-in public catalog (≥3 https URLs including
// hosts example.com, google.com, baidu.com). Source field "random-links".
func RandomSeeds() []Seed

// ResolveSeedSource enforces mutex: exactly one of (linksPath non-empty) or
// randomLinks. Neither or both → error. Dispatches to LoadSeedsFromFile or
// RandomSeeds (+ Options for max-seeds etc.).
func ResolveSeedSource(linksPath string, randomLinks bool, opts Options) (*Resolved, error)
```

CLI (later wiring by implementer; not required for these library leaves to GREEN):

```text
--links PATH | --random-links   (mutex, one required)
--include-archived
--max-seeds N   (0=all)
optional --kind / --env filters when metadata present
artifacts: out/.../corpus.resolved.json
remove corpus.json default load path
```

```go
import (
	"fmt"
	"net/url"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xhd2015/browser-agent/script/debug/chaos-useful-links/seedload"
)

// Mode — top-level surface under test.
const (
	ModeExtract    = "extract"
	ModeHistorical = "historical"
	ModeSource     = "source"
	ModeMaxSeeds   = "max-seeds"
	ModeMetadata   = "metadata"
)

// SourceOp — source/ resolve sub-operations.
const (
	SourceRandom  = "random-links"
	SourceMissing = "missing"
	SourceBoth    = "both-mutex"
)

// Request is narrowed root→leaf by Setup functions.
type Request struct {
	Mode string

	// ModuleRoot is workspace module directory.
	ModuleRoot string

	// Fixture is a filename under DOCTEST_ROOT/testdata (preferred).
	Fixture string

	// Text is inline document body when Fixture is empty (rare).
	Text string

	// Extract / load options.
	IncludeArchived bool
	MaxSeeds        int

	// Source resolve.
	SourceOp    string // random-links | missing | both-mutex
	LinksPath   string // for both-mutex: non-empty path
	RandomLinks bool

	// Expected helper hints (set by leaf Setup for asserts).
	WantURLs    []string // substrings or full URLs that must appear
	WantNotURLs []string // URLs that must not appear
	WantCount   int      // expected seed count; 0 means "do not assert exact count" unless WantCountSet
	WantCountSet bool
}

// Response holds seedload outcomes.
type Response struct {
	Resolved *seedload.Resolved
	Seeds    []seedload.Seed

	ResolveErr     error
	ResolveErrText string

	ExitCode int
}

func Run(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.Mode == "" {
		t.Fatal("Mode must be set by grouping/leaf Setup")
	}
	if req.ModuleRoot == "" {
		req.ModuleRoot = filepath.Clean(filepath.Join(DOCTEST_ROOT, "..", ".."))
	}

	switch req.Mode {
	case ModeExtract, ModeHistorical, ModeMaxSeeds, ModeMetadata:
		return runLoadOrExtract(t, req)
	case ModeSource:
		return runSource(t, req)
	default:
		return nil, fmt.Errorf("unknown Mode %q", req.Mode)
	}
}

func runLoadOrExtract(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	opts := seedload.Options{
		IncludeArchived: req.IncludeArchived,
		MaxSeeds:        req.MaxSeeds,
	}

	resp := &Response{ExitCode: 0}
	var (
		resolved *seedload.Resolved
		err      error
	)

	if req.Fixture != "" {
		path := fixturePath(req.Fixture)
		resolved, err = seedload.LoadSeedsFromFile(path, opts)
	} else if req.Text != "" {
		resolved, err = seedload.ExtractLinks(req.Text, opts)
	} else {
		return nil, fmt.Errorf("extract/historical/max-seeds/metadata requires Fixture or Text")
	}

	if err != nil {
		resp.ResolveErr = err
		resp.ResolveErrText = err.Error()
		resp.ExitCode = 1
		return resp, nil
	}
	resp.Resolved = resolved
	if resolved != nil {
		resp.Seeds = resolved.Seeds
	}
	return resp, nil
}

func runSource(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.SourceOp == "" {
		t.Fatal("SourceOp must be set by leaf Setup for ModeSource")
	}
	opts := seedload.Options{
		IncludeArchived: req.IncludeArchived,
		MaxSeeds:        req.MaxSeeds,
	}
	resp := &Response{ExitCode: 0}

	switch req.SourceOp {
	case SourceRandom:
		// Prefer ResolveSeedSource so source meta is filled; also exercises RandomSeeds path.
		resolved, err := seedload.ResolveSeedSource("", true, opts)
		if err != nil {
			resp.ResolveErr = err
			resp.ResolveErrText = err.Error()
			resp.ExitCode = 1
			return resp, nil
		}
		resp.Resolved = resolved
		if resolved != nil {
			resp.Seeds = resolved.Seeds
		}
		// RandomSeeds() must also be non-empty when resolve succeeds (catalog consistency).
		if direct := seedload.RandomSeeds(); len(direct) == 0 {
			return resp, fmt.Errorf("RandomSeeds() returned empty catalog")
		}
		return resp, nil

	case SourceMissing:
		resolved, err := seedload.ResolveSeedSource("", false, opts)
		resp.Resolved = resolved
		resp.ResolveErr = err
		if err != nil {
			resp.ResolveErrText = err.Error()
			resp.ExitCode = 1
		}
		return resp, nil

	case SourceBoth:
		path := req.LinksPath
		if path == "" {
			path = fixturePath("mixed.md")
		}
		resolved, err := seedload.ResolveSeedSource(path, true, opts)
		resp.Resolved = resolved
		resp.ResolveErr = err
		if err != nil {
			resp.ResolveErrText = err.Error()
			resp.ExitCode = 1
		}
		return resp, nil

	default:
		return nil, fmt.Errorf("unknown SourceOp %q", req.SourceOp)
	}
}

func fixturePath(name string) string {
	return filepath.Join(DOCTEST_ROOT, "testdata", name)
}

func seedURLs(seeds []seedload.Seed) []string {
	out := make([]string, 0, len(seeds))
	for _, s := range seeds {
		out = append(out, s.URL)
	}
	return out
}

func hasURLContaining(seeds []seedload.Seed, sub string) bool {
	for _, s := range seeds {
		if strings.Contains(s.URL, sub) {
			return true
		}
	}
	return false
}

func hasHost(seeds []seedload.Seed, host string) bool {
	for _, s := range seeds {
		u, err := url.Parse(s.URL)
		if err != nil {
			if strings.Contains(s.URL, host) {
				return true
			}
			continue
		}
		h := strings.ToLower(u.Hostname())
		if h == host || strings.HasSuffix(h, "."+host) {
			return true
		}
	}
	return false
}

func assertNoRunErr(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("Run transport/setup error: %v", err)
	}
}

func assertResolveOK(t *testing.T, resp *Response) {
	t.Helper()
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if resp.ResolveErr != nil {
		t.Fatalf("ResolveErr=%v want nil; text=%q", resp.ResolveErr, resp.ResolveErrText)
	}
	if resp.Resolved == nil {
		t.Fatal("Resolved is nil")
	}
}

func assertResolveErr(t *testing.T, resp *Response) {
	t.Helper()
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if resp.ResolveErr == nil {
		t.Fatalf("expected ResolveErr non-nil; got seeds=%v", seedURLs(resp.Seeds))
	}
	if resp.ExitCode == 0 {
		t.Fatalf("ExitCode=0 want non-zero on resolve error")
	}
}

func assertWantURLs(t *testing.T, seeds []seedload.Seed, want []string) {
	t.Helper()
	for _, w := range want {
		if !hasURLContaining(seeds, w) {
			t.Fatalf("missing seed URL containing %q; got %v", w, seedURLs(seeds))
		}
	}
}

func assertNotURLs(t *testing.T, seeds []seedload.Seed, not []string) {
	t.Helper()
	for _, n := range not {
		if hasURLContaining(seeds, n) {
			t.Fatalf("seed URL containing %q must be absent; got %v", n, seedURLs(seeds))
		}
	}
}

```
