package main

import (
	"testing"
	"time"

	"github.com/xhd2015/browser-agent/script/debug/chaos-useful-links/seedload"
)

func TestDiceBootstrapAlwaysOpen(t *testing.T) {
	d := newDice(1)
	for i := 0; i < 20; i++ {
		op := d.pickOp(0, 5)
		if op != opOpenSeed {
			t.Fatalf("step %d: want open_seed with 0 tabs, got %s", i, op)
		}
	}
}

func TestDiceReproducible(t *testing.T) {
	seeds := []Seed{{ID: "a"}, {ID: "b"}, {ID: "c"}}
	seq1 := rollSequence(42, 30, seeds)
	seq2 := rollSequence(42, 30, seeds)
	if len(seq1) != len(seq2) {
		t.Fatalf("len mismatch")
	}
	for i := range seq1 {
		if seq1[i] != seq2[i] {
			t.Fatalf("step %d: %q != %q", i, seq1[i], seq2[i])
		}
	}
	// Different seed should diverge eventually.
	seq3 := rollSequence(43, 30, seeds)
	same := true
	for i := range seq1 {
		if seq1[i] != seq3[i] {
			same = false
			break
		}
	}
	if same {
		t.Fatalf("expected different seed to produce different sequence")
	}
}

func rollSequence(seed int64, n int, seeds []Seed) []string {
	d := newDice(seed)
	tabs := 0
	maxTabs := 5
	var out []string
	for i := 0; i < n; i++ {
		op := d.pickOp(tabs, maxTabs)
		s := d.pickSeed(seeds)
		out = append(out, string(op)+":"+s.ID)
		switch op {
		case opOpenSeed:
			if tabs < maxTabs {
				tabs++
			}
		}
	}
	return out
}

func TestDiceRespectsMaxTabs(t *testing.T) {
	d := newDice(99)
	for i := 0; i < 100; i++ {
		op := d.pickOp(5, 5)
		if op == opOpenSeed {
			t.Fatalf("open_seed at max tabs")
		}
	}
}

func TestClassifyRouting(t *testing.T) {
	c := classifyEval(
		"https://credit.log.some-org.io/log-search",
		"http://127.0.0.1:43761/go?session=sess-abc",
		time.Second,
		nil,
	)
	if c != classFailRouting {
		t.Fatalf("want FAIL_ROUTING, got %s", c)
	}
}

func TestClassifyAuthWall(t *testing.T) {
	c := classifyEval(
		"https://space.some-org.io/console/cmdb",
		"https://sso.some-org.io/login?next=...",
		time.Second,
		nil,
	)
	if c != classOKAuthWall {
		t.Fatalf("want OK_AUTH_WALL, got %s", c)
	}
}

func TestClassifyTimeout(t *testing.T) {
	c := classifyError(errString("timeout after 25s: context deadline exceeded"))
	if c != classFailTimeout {
		t.Fatalf("want FAIL_TIMEOUT, got %s", c)
	}
}

type errString string

func (e errString) Error() string { return string(e) }

func TestResolveRandomSeeds(t *testing.T) {
	resolved, err := seedload.ResolveSeedSource("", true, seedload.Options{})
	if err != nil {
		t.Fatal(err)
	}
	if len(resolved.Seeds) < 3 {
		t.Fatalf("expected >=3 random seeds, got %d", len(resolved.Seeds))
	}
	for _, s := range resolved.Seeds {
		if s.ID == "" || s.URL == "" {
			t.Fatalf("seed missing id/url: %+v", s)
		}
	}
}

func TestParseEvalIdentityNested(t *testing.T) {
	stdout := `{"ok":true,"data":{"href":"https://example.com/x","title":"Hi","ready":"complete"}}`
	href, title, ready := parseEvalIdentity(stdout)
	if href != "https://example.com/x" || title != "Hi" || ready != "complete" {
		t.Fatalf("got href=%q title=%q ready=%q", href, title, ready)
	}
}

func TestExtractTabID(t *testing.T) {
	cases := []string{
		`{"ok":true,"data":{"tab_id":12345}}`,
		`{"tab_id":99}`,
		`{"ok":true,"result":{"tab_id":7}}`,
	}
	want := []int64{12345, 99, 7}
	for i, c := range cases {
		got := extractTabID(c)
		if got != want[i] {
			t.Fatalf("case %d: got %d want %d", i, got, want[i])
		}
	}
}
