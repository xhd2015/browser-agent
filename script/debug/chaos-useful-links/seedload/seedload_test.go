package seedload

import (
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func testdata(name string) string {
	_, file, _, _ := runtime.Caller(0)
	// seedload -> chaos-useful-links -> debug -> script -> module root
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..", "..", ".."))
	return filepath.Join(root, "tests", "chaos-useful-links", "testdata", name)
}

func TestMixedForms(t *testing.T) {
	r, err := LoadSeedsFromFile(testdata("mixed.md"), Options{})
	if err != nil {
		t.Fatal(err)
	}
	if r.Source.Type != "links" || r.Source.SHA256 == "" {
		t.Fatalf("source meta: %+v", r.Source)
	}
	if len(r.Seeds) != 4 {
		t.Fatalf("count=%d urls=%v", len(r.Seeds), urls(r.Seeds))
	}
	for _, want := range []string{"/bare", "/md-link", "/backtick", "/angle"} {
		if !hasSub(r.Seeds, want) {
			t.Fatalf("missing %s in %v", want, urls(r.Seeds))
		}
	}
}

func TestTrailingPunct(t *testing.T) {
	r, err := LoadSeedsFromFile(testdata("trailing-punct.txt"), Options{})
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Seeds) != 5 {
		t.Fatalf("count=%d urls=%v", len(r.Seeds), urls(r.Seeds))
	}
	for _, s := range r.Seeds {
		for _, tr := range []string{")", ".", ",", ";", "]"} {
			if strings.HasSuffix(s.URL, tr) {
				t.Fatalf("trailing %q on %q", tr, s.URL)
			}
		}
	}
}

func TestDedupe(t *testing.T) {
	r, err := LoadSeedsFromFile(testdata("dedupe.md"), Options{})
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Seeds) != 2 || r.Counts.Deduped != 2 || r.Counts.Raw < 3 {
		t.Fatalf("seeds=%d raw=%d deduped=%d urls=%v", len(r.Seeds), r.Counts.Raw, r.Counts.Deduped, urls(r.Seeds))
	}
}

func TestHistoricalSkip(t *testing.T) {
	r, err := LoadSeedsFromFile(testdata("historical.md"), Options{IncludeArchived: false})
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Seeds) != 3 {
		t.Fatalf("count=%d urls=%v", len(r.Seeds), urls(r.Seeds))
	}
	if r.Counts.ArchivedSkipped < 1 {
		t.Fatalf("ArchivedSkipped=%d", r.Counts.ArchivedSkipped)
	}
	if hasSub(r.Seeds, "old.example.com") {
		t.Fatalf("old should be skipped: %v", urls(r.Seeds))
	}
}

func TestHistoricalInclude(t *testing.T) {
	r, err := LoadSeedsFromFile(testdata("historical.md"), Options{IncludeArchived: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Seeds) != 7 {
		t.Fatalf("count=%d urls=%v", len(r.Seeds), urls(r.Seeds))
	}
	if r.Counts.ArchivedSkipped != 0 {
		t.Fatalf("ArchivedSkipped=%d", r.Counts.ArchivedSkipped)
	}
}

func TestMaxSeeds(t *testing.T) {
	all, err := LoadSeedsFromFile(testdata("max-seeds.md"), Options{MaxSeeds: 0})
	if err != nil {
		t.Fatal(err)
	}
	if len(all.Seeds) != 5 || all.Counts.AfterFilter != 5 || all.Counts.Deduped != 5 {
		t.Fatalf("all: %+v seeds=%v", all.Counts, urls(all.Seeds))
	}
	cap2, err := LoadSeedsFromFile(testdata("max-seeds.md"), Options{MaxSeeds: 2})
	if err != nil {
		t.Fatal(err)
	}
	if len(cap2.Seeds) != 2 || cap2.Counts.Deduped != 5 || cap2.Counts.AfterFilter != 2 {
		t.Fatalf("cap: %+v seeds=%v", cap2.Counts, urls(cap2.Seeds))
	}
}

func TestGFMTable(t *testing.T) {
	r, err := LoadSeedsFromFile(testdata("useful-links-table.md"), Options{})
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Seeds) != 2 {
		t.Fatalf("count=%d urls=%v", len(r.Seeds), urls(r.Seeds))
	}
	by := map[string]Seed{}
	for _, s := range r.Seeds {
		if strings.Contains(s.URL, "logs.example.com") {
			by["logs"] = s
		}
		if strings.Contains(s.URL, "grafana.example.com") {
			by["grafana"] = s
		}
	}
	logs := by["logs"]
	if logs.Env != "live" || logs.Kind != "app_logs" || logs.Market != "ID" || !strings.Contains(logs.Title, "ID live") {
		t.Fatalf("logs meta: %+v", logs)
	}
	g := by["grafana"]
	if g.Env != "test" || g.Kind != "grafana" || g.Market != "MY" || !strings.Contains(g.Title, "Stress") {
		t.Fatalf("grafana meta: %+v", g)
	}
}

func TestSourceMutex(t *testing.T) {
	if _, err := ResolveSeedSource("", false, Options{}); err == nil {
		t.Fatal("expected missing error")
	}
	if _, err := ResolveSeedSource("x.md", true, Options{}); err == nil {
		t.Fatal("expected both error")
	}
	r, err := ResolveSeedSource("", true, Options{})
	if err != nil {
		t.Fatal(err)
	}
	if r.Source.Type != "random-links" || len(r.Seeds) < 3 {
		t.Fatalf("%+v", r)
	}
}

func urls(seeds []Seed) []string {
	out := make([]string, len(seeds))
	for i, s := range seeds {
		out[i] = s.URL
	}
	return out
}

func hasSub(seeds []Seed, sub string) bool {
	for _, s := range seeds {
		if strings.Contains(s.URL, sub) {
			return true
		}
	}
	return false
}
