package browseragent

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// Noise rules aligned with the former summarise-har helper (static/analytics).
var (
	harNoiseHostSuffixes = []string{
		"dem.some-x.com",
		"google-analytics.com",
		"googletagmanager.com",
	}
	harNoisePathPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)/tags/web-(performance|custom)/`),
		regexp.MustCompile(`(?i)/dem/entrance/`),
		regexp.MustCompile(`(?i)\.(js|css|png|jpg|jpeg|gif|webp|svg|ico|woff2?|ttf|map)(\?|$)`),
	}
	harJWTPattern = regexp.MustCompile(`eyJ[A-Za-z0-9_\-]{10,}\.[A-Za-z0-9_\-]{10,}\.[A-Za-z0-9_\-]{10,}`)
)

type harFileDoc struct {
	Log struct {
		Creator struct {
			Name    string `json:"name"`
			Version string `json:"version"`
		} `json:"creator"`
		Entries []harLogEntry `json:"entries"`
	} `json:"log"`
}

type harLogEntry struct {
	StartedDateTime string `json:"startedDateTime"`
	Time            any    `json:"time"`
	Request         struct {
		Method   string          `json:"method"`
		URL      string          `json:"url"`
		Headers  []harNameValue  `json:"headers"`
		PostData *harPostData    `json:"postData"`
	} `json:"request"`
	Response struct {
		Status  int            `json:"status"`
		Headers []harNameValue `json:"headers"`
		Content struct {
			Text     string `json:"text"`
			Encoding string `json:"encoding"`
			Size     int    `json:"size"`
			MimeType string `json:"mimeType"`
		} `json:"content"`
	} `json:"response"`
}

type harNameValue struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type harPostData struct {
	MimeType string `json:"mimeType"`
	Text     string `json:"text"`
}

// harInspectSource is a loaded export directory or single .har file.
type harInspectSource struct {
	Path     string
	IsDir    bool
	Manifest *harManifest
	Tabs     []harInspectTab
}

type harInspectTab struct {
	TabID   int
	Title   string
	File    string // basename or absolute for single-file mode
	AbsPath string
	Entries []harLogEntry
}

type harInspectFilter struct {
	Hosts      []string
	Method     string
	PathSubstr string
	Match      string
	TabID      int
	TabIDSet   bool
	DropNoise  bool
}

type harInspectRow struct {
	TabID           int    `json:"tab_id,omitempty"`
	File            string `json:"file,omitempty"`
	Index           int    `json:"index"`
	Started         string `json:"started"`
	Method          string `json:"method"`
	URL             string `json:"url"`
	Path            string `json:"path"`
	Status          int    `json:"status"`
	RequestBodyLen  int    `json:"request_body_len"`
	ResponseBodyLen int    `json:"response_body_len"`
}

type harBodyCoverage struct {
	Entries          int `json:"entries"`
	WithRequestBody  int `json:"with_request_body"`
	WithResponseBody int `json:"with_response_body"`
}

type harPathCount struct {
	Method string `json:"method"`
	Path   string `json:"path"`
	Count  int    `json:"count"`
}

func isHARNoiseURL(raw string) bool {
	u, err := url.Parse(raw)
	if err != nil {
		return false
	}
	host := strings.ToLower(u.Hostname())
	for _, suf := range harNoiseHostSuffixes {
		if strings.HasSuffix(host, suf) {
			return true
		}
	}
	path := u.EscapedPath()
	if u.RawQuery != "" {
		path = path + "?" + u.RawQuery
	}
	for _, re := range harNoisePathPatterns {
		if re.MatchString(path) {
			return true
		}
	}
	return false
}

func harPathOnly(raw string) string {
	u, err := url.Parse(raw)
	if err != nil {
		return raw
	}
	if u.RawQuery != "" {
		return u.EscapedPath() + "?" + u.RawQuery
	}
	return u.EscapedPath()
}

func loadHARInspectSource(path string) (*harInspectSource, error) {
	path = filepath.Clean(strings.TrimSpace(path))
	if path == "" || path == "." {
		return nil, fmt.Errorf("path is required")
	}
	st, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	src := &harInspectSource{Path: path, IsDir: st.IsDir()}
	if !st.IsDir() {
		if !strings.EqualFold(filepath.Ext(path), ".har") {
			return nil, fmt.Errorf("path must be an export directory or a .har file")
		}
		doc, err := readHARFile(path)
		if err != nil {
			return nil, err
		}
		src.Tabs = []harInspectTab{{
			File:    filepath.Base(path),
			AbsPath: path,
			Entries: doc.Log.Entries,
		}}
		return src, nil
	}

	manifestPath := filepath.Join(path, "manifest.json")
	raw, err := os.ReadFile(manifestPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("export directory missing manifest.json: %s", manifestPath)
		}
		return nil, err
	}
	var manifest harManifest
	if err := json.Unmarshal(raw, &manifest); err != nil {
		return nil, fmt.Errorf("parse manifest.json: %w", err)
	}
	if manifest.SchemaVersion == 0 && manifest.SessionID == "" && len(manifest.Tabs) == 0 {
		return nil, fmt.Errorf("invalid manifest.json: missing schema_version/session tabs")
	}
	src.Manifest = &manifest
	if len(manifest.Tabs) == 0 {
		return nil, fmt.Errorf("manifest.json lists no tabs")
	}
	for _, tab := range manifest.Tabs {
		file := strings.TrimSpace(tab.File)
		if file == "" {
			return nil, fmt.Errorf("manifest tab %d missing file", tab.TabID)
		}
		abs := filepath.Join(path, filepath.Base(file))
		doc, err := readHARFile(abs)
		if err != nil {
			return nil, fmt.Errorf("tab %d (%s): %w", tab.TabID, file, err)
		}
		src.Tabs = append(src.Tabs, harInspectTab{
			TabID:   tab.TabID,
			Title:   tab.Title,
			File:    filepath.Base(file),
			AbsPath: abs,
			Entries: doc.Log.Entries,
		})
	}
	return src, nil
}

func readHARFile(path string) (*harFileDoc, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var doc harFileDoc
	if err := json.Unmarshal(raw, &doc); err != nil {
		return nil, fmt.Errorf("parse HAR %s: %w", path, err)
	}
	if doc.Log.Entries == nil {
		doc.Log.Entries = []harLogEntry{}
	}
	return &doc, nil
}

func (f harInspectFilter) accept(tab harInspectTab, entry harLogEntry) bool {
	if f.TabIDSet && tab.TabID != f.TabID {
		return false
	}
	rawURL := strings.TrimSpace(entry.Request.URL)
	if rawURL == "" {
		return false
	}
	if f.DropNoise && isHARNoiseURL(rawURL) {
		return false
	}
	if len(f.Hosts) > 0 {
		u, err := url.Parse(rawURL)
		if err != nil {
			return false
		}
		host := strings.ToLower(u.Hostname())
		ok := false
		for _, h := range f.Hosts {
			h = strings.ToLower(strings.TrimSpace(h))
			if h == "" {
				continue
			}
			if strings.Contains(host, h) || strings.HasSuffix(host, h) {
				ok = true
				break
			}
		}
		if !ok {
			return false
		}
	}
	if m := strings.TrimSpace(f.Method); m != "" {
		if !strings.EqualFold(entry.Request.Method, m) {
			return false
		}
	}
	path := harPathOnly(rawURL)
	if s := strings.TrimSpace(f.PathSubstr); s != "" {
		if !strings.Contains(path, s) && !strings.Contains(rawURL, s) {
			return false
		}
	}
	if s := strings.TrimSpace(f.Match); s != "" {
		if !strings.Contains(rawURL, s) && !strings.Contains(path, s) {
			return false
		}
	}
	return true
}

func (src *harInspectSource) filteredRows(filter harInspectFilter) []harInspectRow {
	var rows []harInspectRow
	for _, tab := range src.Tabs {
		for i, entry := range tab.Entries {
			if !filter.accept(tab, entry) {
				continue
			}
			reqBody := entryRequestBody(entry)
			respBody, _ := entryResponseBody(entry)
			rows = append(rows, harInspectRow{
				TabID:           tab.TabID,
				File:            tab.File,
				Index:           i,
				Started:         entry.StartedDateTime,
				Method:          entry.Request.Method,
				URL:             entry.Request.URL,
				Path:            harPathOnly(entry.Request.URL),
				Status:          entry.Response.Status,
				RequestBodyLen:  len(reqBody),
				ResponseBodyLen: len(respBody),
			})
		}
	}
	return rows
}

func entryRequestBody(entry harLogEntry) string {
	if entry.Request.PostData == nil {
		return ""
	}
	return entry.Request.PostData.Text
}

func entryResponseBody(entry harLogEntry) (text string, encoding string) {
	text = entry.Response.Content.Text
	encoding = entry.Response.Content.Encoding
	if strings.EqualFold(encoding, "base64") && text != "" {
		decoded, err := base64.StdEncoding.DecodeString(text)
		if err == nil {
			return string(decoded), encoding
		}
	}
	return text, encoding
}

func bodyCoverageFromRows(rows []harInspectRow) harBodyCoverage {
	c := harBodyCoverage{Entries: len(rows)}
	for _, r := range rows {
		if r.RequestBodyLen > 0 {
			c.WithRequestBody++
		}
		if r.ResponseBodyLen > 0 {
			c.WithResponseBody++
		}
	}
	return c
}

func pathInventory(rows []harInspectRow) []harPathCount {
	type key struct{ method, path string }
	counts := map[key]int{}
	for _, r := range rows {
		method := r.Method
		if method == "" {
			method = "?"
		}
		counts[key{method, r.Path}]++
	}
	out := make([]harPathCount, 0, len(counts))
	for k, n := range counts {
		out = append(out, harPathCount{Method: k.method, Path: k.path, Count: n})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Count != out[j].Count {
			return out[i].Count > out[j].Count
		}
		if out[i].Path != out[j].Path {
			return out[i].Path < out[j].Path
		}
		return out[i].Method < out[j].Method
	})
	return out
}

func findHAREntry(src *harInspectSource, filter harInspectFilter, index *int, indexSet bool) (harInspectTab, int, harLogEntry, error) {
	type hit struct {
		tab   harInspectTab
		index int
		entry harLogEntry
	}
	var hits []hit
	for _, tab := range src.Tabs {
		for i, entry := range tab.Entries {
			if indexSet {
				if filter.TabIDSet && tab.TabID != filter.TabID {
					continue
				}
				if !src.IsDir && *index == i {
					return tab, i, entry, nil
				}
				if src.IsDir {
					if !filter.TabIDSet {
						return harInspectTab{}, 0, harLogEntry{}, fmt.Errorf("directory show with --index requires --tab-id")
					}
					if tab.TabID == filter.TabID && i == *index {
						return tab, i, entry, nil
					}
				}
				continue
			}
			if !filter.accept(tab, entry) {
				continue
			}
			hits = append(hits, hit{tab: tab, index: i, entry: entry})
		}
	}
	if indexSet {
		return harInspectTab{}, 0, harLogEntry{}, fmt.Errorf("no entry at index %d", *index)
	}
	if len(hits) == 0 {
		match := strings.TrimSpace(filter.Match)
		if match == "" {
			match = strings.TrimSpace(filter.PathSubstr)
		}
		if match == "" {
			return harInspectTab{}, 0, harLogEntry{}, fmt.Errorf("no entries matched filters")
		}
		return harInspectTab{}, 0, harLogEntry{}, fmt.Errorf("no entries matched --match %s", match)
	}
	if len(hits) > 1 {
		var b strings.Builder
		fmt.Fprintf(&b, "%d entries matched; refine --match/--tab-id or use --index. candidates:", len(hits))
		limit := len(hits)
		if limit > 5 {
			limit = 5
		}
		for i := 0; i < limit; i++ {
			h := hits[i]
			fmt.Fprintf(&b, "\n  tab=%d index=%d %s %s", h.tab.TabID, h.index, h.entry.Request.Method, harPathOnly(h.entry.Request.URL))
		}
		if len(hits) > limit {
			fmt.Fprintf(&b, "\n  … %d more", len(hits)-limit)
		}
		return harInspectTab{}, 0, harLogEntry{}, fmt.Errorf("%s", b.String())
	}
	h := hits[0]
	return h.tab, h.index, h.entry, nil
}

func redactHARHeaderName(name string) bool {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "authorization", "cookie", "set-cookie", "x-token", "x-csrf-token", "proxy-authorization":
		return true
	default:
		return false
	}
}

func redactSecretsInString(s string) string {
	if s == "" {
		return s
	}
	s = regexp.MustCompile(`(?i)Bearer\s+[A-Za-z0-9._\-]+`).ReplaceAllString(s, "Bearer <REDACTED>")
	s = harJWTPattern.ReplaceAllString(s, "<REDACTED_JWT>")
	return s
}

func redactHeaders(headers []harNameValue, redact bool) map[string]string {
	out := map[string]string{}
	for _, h := range headers {
		name := h.Name
		val := h.Value
		if redact && redactHARHeaderName(name) {
			val = "<REDACTED>"
		} else if redact {
			val = redactSecretsInString(val)
		}
		// Prefer first occurrence; keep interesting casing of first.
		if _, ok := out[name]; !ok {
			out[name] = val
		}
	}
	return out
}

func parseJSONOrString(text string, redact bool) any {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil
	}
	var v any
	if err := json.Unmarshal([]byte(text), &v); err != nil {
		if redact {
			return redactSecretsInString(text)
		}
		return text
	}
	if redact {
		return redactJSONValue(v)
	}
	return v
}

func redactJSONValue(v any) any {
	switch t := v.(type) {
	case map[string]any:
		out := make(map[string]any, len(t))
		for k, val := range t {
			lk := strings.ToLower(k)
			if strings.Contains(lk, "token") || strings.Contains(lk, "authorization") || lk == "password" || lk == "cookie" {
				out[k] = "<REDACTED>"
				continue
			}
			out[k] = redactJSONValue(val)
		}
		return out
	case []any:
		out := make([]any, len(t))
		for i, val := range t {
			out[i] = redactJSONValue(val)
		}
		return out
	case string:
		return redactSecretsInString(t)
	default:
		return v
	}
}

func buildShowPayload(tab harInspectTab, index int, entry harLogEntry, redact bool) map[string]any {
	reqBody := entryRequestBody(entry)
	respBody, respEnc := entryResponseBody(entry)
	reqHeaders := redactHeaders(entry.Request.Headers, redact)
	// Keep a small relevant subset preference for readability but include all (redacted).
	payload := map[string]any{
		"tab_id": tab.TabID,
		"file":   tab.File,
		"index":  index,
		"started": entry.StartedDateTime,
		"method": entry.Request.Method,
		"url":    entry.Request.URL,
		"status": entry.Response.Status,
		"request": map[string]any{
			"headers": reqHeaders,
			"body":    parseJSONOrString(reqBody, redact),
		},
		"response": map[string]any{
			"headers":  redactHeaders(entry.Response.Headers, redact),
			"encoding": respEnc,
			"body":     parseJSONOrString(respBody, redact),
		},
	}
	return payload
}

func parseOptionalIndex(args []string) (value int, set bool, err error) {
	raw, ok := flagStringSet(args, "--index")
	if !ok || strings.TrimSpace(raw) == "" {
		return 0, false, nil
	}
	n, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil || n < 0 {
		return 0, false, fmt.Errorf("--index must be a non-negative integer")
	}
	return n, true, nil
}

func collectFlagValues(args []string, name string) []string {
	var out []string
	for i := 0; i < len(args); i++ {
		a := args[i]
		if a == name {
			if i+1 < len(args) {
				out = append(out, args[i+1])
				i++
			}
			continue
		}
		if strings.HasPrefix(a, name+"=") {
			out = append(out, strings.TrimPrefix(a, name+"="))
		}
	}
	return out
}

func hasFlag(args []string, name string) bool {
	for _, a := range args {
		if a == name {
			return true
		}
	}
	return false
}

func takeInspectPath(args []string) (string, []string, error) {
	// First non-flag positional is the path.
	for i := 0; i < len(args); i++ {
		a := args[i]
		if a == "" {
			continue
		}
		if strings.HasPrefix(a, "-") {
			// skip flag values for known flags that take an argument
			switch a {
			case "--host", "--method", "--path-substr", "--match", "--tab-id", "--index", "--limit":
				i++
			default:
				if strings.Contains(a, "=") {
					continue
				}
			}
			continue
		}
		rest := append([]string{}, args[:i]...)
		rest = append(rest, args[i+1:]...)
		return a, rest, nil
	}
	return "", args, fmt.Errorf("path is required (export directory or .har file)")
}
