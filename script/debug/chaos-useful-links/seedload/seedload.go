// Package seedload extracts chaos harness seeds from markdown/text link files
// or a built-in random public catalog. No network fetches.
package seedload

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/url"
	"os"
	"regexp"
	"sort"
	"strings"
	"unicode"
)

// Seed is one chaos corpus URL with optional metadata.
type Seed struct {
	ID     string   `json:"id"`
	URL    string   `json:"url"`
	Kind   string   `json:"kind,omitempty"`
	Env    string   `json:"env,omitempty"`
	Market string   `json:"market,omitempty"`
	Title  string   `json:"title,omitempty"`
	Tags   []string `json:"tags,omitempty"`
	Source string   `json:"source,omitempty"` // e.g. "links:row" or "random-links"
}

// SourceMeta describes where seeds came from.
type SourceMeta struct {
	Type   string `json:"type"`             // "links" | "random-links"
	Path   string `json:"path,omitempty"`   // file path when Type=links
	SHA256 string `json:"sha256,omitempty"` // of file contents when Type=links
}

// Counts tracks pipeline stages.
type Counts struct {
	Raw             int `json:"raw"`              // candidates before dedupe (after historical filter)
	Deduped         int `json:"deduped"`          // after URL dedupe
	ArchivedSkipped int `json:"archived_skipped"` // historical/archived/deprecated section drops
	AfterFilter     int `json:"after_filter"`     // after MaxSeeds / kind/env filters
}

// Resolved is a seed snapshot plus source meta and counts.
type Resolved struct {
	Source SourceMeta `json:"source"`
	Counts Counts     `json:"counts"`
	Seeds  []Seed     `json:"seeds"`
}

// Options control extract/load/resolve filters.
type Options struct {
	IncludeArchived bool
	MaxSeeds        int    // 0 = keep all after dedupe
	Kind            string // optional filter when metadata present
	Env             string // optional filter when metadata present
}

var (
	// headingRe matches ATX markdown headings (# … ######).
	headingRe = regexp.MustCompile(`^(#{1,6})\s+(.+?)\s*$`)
	// mdLinkRe captures markdown link destinations: [text](url)
	mdLinkRe = regexp.MustCompile(`\[[^\]]*\]\(([^)\s]+)\)`)
	// angleURLRe captures <http(s)://...>
	angleURLRe = regexp.MustCompile(`<(https?://[^>\s]+)>`)
	// backtickURLRe captures `http(s)://...`
	// Pattern: one-or-more backticks, capture https?:// until backtick/whitespace, closing backticks.
	backtickURLRe = regexp.MustCompile("`+(https?://[^`\\s]+)`+")
	// bareURLRe captures bare http(s) URLs (also used as a fallback inside lines).
	bareURLRe = regexp.MustCompile(`https?://[^\s<>"'` + "`" + `]+`)
	// tableSepRe detects GFM separator rows (|---|---|)
	tableSepRe = regexp.MustCompile(`^\|?\s*:?-+:?\s*(\|\s*:?-+:?\s*)+\|?\s*$`)
	// archivedHeading keywords (case-insensitive substring match on heading text).
	archivedKeywordRe = regexp.MustCompile(`(?i)\b(historical|archived|deprecated)\b`)
)

// ExtractLinks applies extract/normalize/dedupe/historical/max-seeds rules to text.
func ExtractLinks(text string, opts Options) (*Resolved, error) {
	return extract(text, opts, SourceMeta{})
}

// LoadSeedsFromFile reads path and runs the same pipeline; Source.Type=links,
// Path set, SHA256 of file bytes.
func LoadSeedsFromFile(path string, opts Options) (*Resolved, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	sum := sha256.Sum256(raw)
	meta := SourceMeta{
		Type:   "links",
		Path:   path,
		SHA256: hex.EncodeToString(sum[:]),
	}
	return extract(string(raw), opts, meta)
}

// RandomSeeds returns the built-in public catalog (≥3 https URLs including
// hosts example.com, google.com, baidu.com). Source field "random-links".
func RandomSeeds() []Seed {
	catalog := []struct {
		url   string
		title string
	}{
		{"https://example.com/", "Example Domain"},
		{"https://www.google.com/", "Google"},
		{"https://www.baidu.com/", "Baidu"},
	}
	out := make([]Seed, 0, len(catalog))
	for _, c := range catalog {
		out = append(out, Seed{
			ID:     makeSeedID(c.url),
			URL:    c.url,
			Title:  c.title,
			Source: "random-links",
		})
	}
	return out
}

// ResolveSeedSource enforces mutex: exactly one of (linksPath non-empty) or
// randomLinks. Neither or both → error. Dispatches to LoadSeedsFromFile or
// RandomSeeds (+ Options for max-seeds etc.).
func ResolveSeedSource(linksPath string, randomLinks bool, opts Options) (*Resolved, error) {
	hasLinks := strings.TrimSpace(linksPath) != ""
	switch {
	case !hasLinks && !randomLinks:
		return nil, fmt.Errorf("seed source required: pass --links PATH or --random-links")
	case hasLinks && randomLinks:
		return nil, fmt.Errorf("seed source mutex: cannot set both --links and --random-links")
	case randomLinks:
		return resolveRandom(opts)
	default:
		return LoadSeedsFromFile(linksPath, opts)
	}
}

func resolveRandom(opts Options) (*Resolved, error) {
	seeds := RandomSeeds()
	// Apply MaxSeeds / kind / env on the catalog as well.
	counts := Counts{
		Raw:     len(seeds),
		Deduped: len(seeds),
	}
	seeds = applyFilters(seeds, opts, &counts)
	return &Resolved{
		Source: SourceMeta{Type: "random-links"},
		Counts: counts,
		Seeds:  seeds,
	}, nil
}

type candidate struct {
	url    string
	kind   string
	env    string
	market string
	title  string
	tags   []string
	source string
}

func extract(text string, opts Options, meta SourceMeta) (*Resolved, error) {
	// First pass: detect which line ranges are under archived headings.
	lines := splitLines(text)
	lineArchived := markArchivedLines(lines)

	// Collect table-row metadata keyed by normalized URL (first wins for meta).
	tableMeta := parseGFMTables(lines)

	var (
		rawCandidates   []candidate
		archivedSkipped int
	)

	// Walk document extracting URL occurrences in order.
	// Prefer structured extractors (md/angle/backtick/table) then bare.
	for i, line := range lines {
		inArchived := lineArchived[i]
		// Skip pure heading lines for bare extraction noise (URLs rarely in headings).
		if headingRe.MatchString(strings.TrimSpace(line)) {
			continue
		}

		// GFM table data rows: if this line is a table data row with a Link column, use that.
		if tm, ok := tableMeta.rowAt[i]; ok {
			u := normalizeURL(tm.url)
			if u == "" {
				continue
			}
			if inArchived && !opts.IncludeArchived {
				archivedSkipped++
				continue
			}
			rawCandidates = append(rawCandidates, candidate{
				url:    u,
				kind:   tm.kind,
				env:    tm.env,
				market: tm.market,
				title:  tm.title,
				source: "links:row",
			})
			continue
		}

		found := extractURLsFromLine(line)
		for _, rawU := range found {
			u := normalizeURL(rawU)
			if u == "" {
				continue
			}
			if inArchived && !opts.IncludeArchived {
				archivedSkipped++
				continue
			}
			c := candidate{url: u, source: "links"}
			// Attach table meta if this URL also appears in a table (dedupe later keeps first).
			if m, ok := tableMeta.byURL[u]; ok {
				c.kind = m.kind
				c.env = m.env
				c.market = m.market
				c.title = m.title
				c.source = "links:row"
			}
			rawCandidates = append(rawCandidates, c)
		}
	}

	counts := Counts{
		Raw:             len(rawCandidates),
		ArchivedSkipped: archivedSkipped,
	}

	// Dedupe by normalized URL; keep first occurrence (stable order).
	seen := make(map[string]struct{}, len(rawCandidates))
	deduped := make([]candidate, 0, len(rawCandidates))
	for _, c := range rawCandidates {
		key := c.url
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		deduped = append(deduped, c)
	}
	counts.Deduped = len(deduped)

	seeds := make([]Seed, 0, len(deduped))
	for _, c := range deduped {
		title := c.title
		if title == "" {
			title = heuristicTitle(c.url)
		}
		seeds = append(seeds, Seed{
			ID:     makeSeedID(c.url),
			URL:    c.url,
			Kind:   c.kind,
			Env:    c.env,
			Market: c.market,
			Title:  title,
			Tags:   c.tags,
			Source: c.source,
		})
	}

	seeds = applyFilters(seeds, opts, &counts)

	return &Resolved{
		Source: meta,
		Counts: counts,
		Seeds:  seeds,
	}, nil
}

func applyFilters(seeds []Seed, opts Options, counts *Counts) []Seed {
	out := seeds
	if k := strings.TrimSpace(opts.Kind); k != "" {
		anyKind := false
		for _, s := range seeds {
			if s.Kind != "" {
				anyKind = true
				break
			}
		}
		if anyKind {
			filtered := make([]Seed, 0, len(out))
			for _, s := range out {
				if s.Kind == k {
					filtered = append(filtered, s)
				}
			}
			out = filtered
		}
	}
	if e := strings.TrimSpace(opts.Env); e != "" {
		anyEnv := false
		for _, s := range seeds {
			if s.Env != "" {
				anyEnv = true
				break
			}
		}
		if anyEnv {
			filtered := make([]Seed, 0, len(out))
			for _, s := range out {
				if s.Env == e {
					filtered = append(filtered, s)
				}
			}
			out = filtered
		}
	}
	if opts.MaxSeeds > 0 && len(out) > opts.MaxSeeds {
		out = append([]Seed(nil), out[:opts.MaxSeeds]...)
	}
	counts.AfterFilter = len(out)
	return out
}

// markArchivedLines returns per-line whether content is under an archived heading.
// Skip under headings matching historical|archived|deprecated until the next
// same-or-higher level heading (unless caller ignores via IncludeArchived).
func markArchivedLines(lines []string) []bool {
	out := make([]bool, len(lines))
	// archivedLevel > 0 means currently inside archived section at that heading depth.
	archivedLevel := 0
	for i, line := range lines {
		trim := strings.TrimSpace(line)
		if m := headingRe.FindStringSubmatch(trim); m != nil {
			level := len(m[1])
			title := m[2]
			if archivedLevel > 0 && level <= archivedLevel {
				// Exit archived section.
				archivedLevel = 0
			}
			if archivedKeywordRe.MatchString(title) {
				archivedLevel = level
				out[i] = true // heading itself counts as archived context
				continue
			}
			// Non-archived heading at this level (or deeper when not in archive):
			if archivedLevel > 0 && level > archivedLevel {
				out[i] = true // nested under archive
				continue
			}
			out[i] = false
			continue
		}
		out[i] = archivedLevel > 0
	}
	return out
}

type tableRowMeta struct {
	url, kind, env, market, title string
}

type tableIndex struct {
	// rowAt maps line index → meta for that data row (Link column present).
	rowAt map[int]tableRowMeta
	// byURL maps normalized URL → first table meta.
	byURL map[string]tableRowMeta
}

func parseGFMTables(lines []string) tableIndex {
	idx := tableIndex{
		rowAt: make(map[int]tableRowMeta),
		byURL: make(map[string]tableRowMeta),
	}

	for i := 0; i < len(lines); i++ {
		headerCells := splitTableRow(lines[i])
		if len(headerCells) < 2 {
			continue
		}
		if i+1 >= len(lines) || !tableSepRe.MatchString(strings.TrimSpace(lines[i+1])) {
			continue
		}

		// Map column names (case-insensitive).
		col := map[string]int{}
		for ci, h := range headerCells {
			key := strings.ToLower(strings.TrimSpace(h))
			col[key] = ci
		}
		linkCol, hasLink := col["link"]
		if !hasLink {
			// Not a useful-links style table; skip.
			continue
		}
		envCol, hasEnv := col["env"]
		kindCol, hasKind := col["kind"]
		titleCol, hasTitle := col["title"]
		marketCol, hasMarket := col["market"]

		// Consume separator + data rows.
		j := i + 2
		for j < len(lines) {
			row := strings.TrimSpace(lines[j])
			if row == "" || headingRe.MatchString(row) {
				break
			}
			cells := splitTableRow(lines[j])
			if len(cells) == 0 {
				break
			}
			// If next line looks like a new table header+sep, stop.
			if j+1 < len(lines) && tableSepRe.MatchString(strings.TrimSpace(lines[j+1])) {
				break
			}
			// Skip stray separator-looking rows.
			if tableSepRe.MatchString(row) {
				j++
				continue
			}

			linkCell := cellAt(cells, linkCol)
			u := firstURLIn(linkCell)
			if u == "" {
				// Maybe the whole cell is a bare URL after normalize.
				u = normalizeURL(strings.TrimSpace(linkCell))
			}
			if u == "" {
				j++
				continue
			}
			u = normalizeURL(u)
			meta := tableRowMeta{url: u}
			if hasEnv {
				meta.env = strings.TrimSpace(cellAt(cells, envCol))
			}
			if hasKind {
				meta.kind = strings.TrimSpace(cellAt(cells, kindCol))
			}
			if hasTitle {
				meta.title = strings.TrimSpace(cellAt(cells, titleCol))
			}
			if hasMarket {
				meta.market = strings.TrimSpace(cellAt(cells, marketCol))
			}
			idx.rowAt[j] = meta
			if _, ok := idx.byURL[u]; !ok {
				idx.byURL[u] = meta
			}
			j++
		}
		i = j - 1
	}
	return idx
}

func splitTableRow(line string) []string {
	trim := strings.TrimSpace(line)
	if !strings.Contains(trim, "|") {
		return nil
	}
	// Trim outer pipes.
	trim = strings.TrimPrefix(trim, "|")
	trim = strings.TrimSuffix(trim, "|")
	parts := strings.Split(trim, "|")
	cells := make([]string, len(parts))
	for i, p := range parts {
		cells[i] = strings.TrimSpace(p)
	}
	return cells
}

func cellAt(cells []string, i int) string {
	if i < 0 || i >= len(cells) {
		return ""
	}
	return cells[i]
}

func firstURLIn(s string) string {
	if m := mdLinkRe.FindStringSubmatch(s); len(m) > 1 {
		return m[1]
	}
	if m := angleURLRe.FindStringSubmatch(s); len(m) > 1 {
		return m[1]
	}
	if m := backtickURLRe.FindStringSubmatch(s); len(m) > 1 {
		return m[1]
	}
	if m := bareURLRe.FindString(s); m != "" {
		return m
	}
	return ""
}

// extractURLsFromLine finds URLs in a non-table line via md/angle/backtick/bare forms.
// Dedupes within the line by span to avoid double-counting the same occurrence.
func extractURLsFromLine(line string) []string {
	type span struct {
		start, end int
		url        string
	}
	var spans []span

	addMatches := func(re *regexp.Regexp, group int) {
		for _, loc := range re.FindAllStringSubmatchIndex(line, -1) {
			// loc: full start/end, then groups
			var u string
			var s, e int
			if group == 0 {
				s, e = loc[0], loc[1]
				u = line[s:e]
			} else {
				si := 2 * group
				if si+1 >= len(loc) || loc[si] < 0 {
					continue
				}
				s, e = loc[si], loc[si+1]
				u = line[s:e]
			}
			spans = append(spans, span{start: s, end: e, url: u})
		}
	}

	// Prefer structured forms first for cleaner capture; bare fills gaps.
	addMatches(mdLinkRe, 1)
	addMatches(angleURLRe, 1)
	addMatches(backtickURLRe, 1)

	// Bare: only if not overlapping an already-captured span.
	for _, loc := range bareURLRe.FindAllStringIndex(line, -1) {
		s, e := loc[0], loc[1]
		overlap := false
		for _, sp := range spans {
			if s < sp.end && e > sp.start {
				overlap = true
				break
			}
		}
		if !overlap {
			spans = append(spans, span{start: s, end: e, url: line[s:e]})
		}
	}

	// Sort by start position for stable document order.
	sort.SliceStable(spans, func(i, j int) bool {
		return spans[i].start < spans[j].start
	})

	out := make([]string, 0, len(spans))
	for _, sp := range spans {
		out = append(out, sp.url)
	}
	return out
}

func normalizeURL(raw string) string {
	u := strings.TrimSpace(raw)
	if u == "" {
		return ""
	}
	// Strip common wrapping punctuation.
	u = strings.Trim(u, "<>\"'")
	// Optional: bare www. → https://www.
	if strings.HasPrefix(strings.ToLower(u), "www.") {
		u = "https://" + u
	}
	// Strip trailing punctuation from set ).,;]
	for {
		if u == "" {
			return ""
		}
		last := u[len(u)-1]
		if strings.ContainsRune(").,;]", rune(last)) {
			u = u[:len(u)-1]
			continue
		}
		break
	}
	// Must be http(s).
	lower := strings.ToLower(u)
	if !strings.HasPrefix(lower, "http://") && !strings.HasPrefix(lower, "https://") {
		return ""
	}
	// Basic parse check.
	parsed, err := url.Parse(u)
	if err != nil || parsed.Host == "" {
		return ""
	}
	return u
}

func makeSeedID(rawURL string) string {
	slug := urlSlug(rawURL)
	sum := sha256.Sum256([]byte(rawURL))
	short := hex.EncodeToString(sum[:3]) // 6 hex chars
	if slug == "" {
		return short
	}
	// Keep IDs compact.
	if len(slug) > 48 {
		slug = slug[:48]
		slug = strings.Trim(slug, "-")
	}
	return slug + "-" + short
}

func urlSlug(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return slugify(rawURL)
	}
	host := strings.TrimPrefix(strings.ToLower(u.Hostname()), "www.")
	path := strings.Trim(u.Path, "/")
	base := host
	if path != "" {
		// Use last path segment for brevity when deep.
		parts := strings.Split(path, "/")
		if len(parts) > 2 {
			path = strings.Join(parts[len(parts)-2:], "-")
		}
		base = host + "-" + path
	}
	return slugify(base)
}

func slugify(s string) string {
	var b strings.Builder
	prevDash := false
	for _, r := range strings.ToLower(s) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
			prevDash = false
			continue
		}
		if !prevDash {
			b.WriteByte('-')
			prevDash = true
		}
	}
	return strings.Trim(b.String(), "-")
}

func heuristicTitle(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	host := u.Hostname()
	path := strings.Trim(u.Path, "/")
	if path == "" {
		return host
	}
	return host + "/" + path
}

func splitLines(text string) []string {
	// Preserve empty lines; normalize CRLF.
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	return strings.Split(text, "\n")
}
