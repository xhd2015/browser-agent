## Expected

- Exactly 2 seeds.
- Seed for logs.example.com/id: Env=live, Kind=app_logs, Title contains "ID live", Market=ID.
- Seed for grafana.example.com/stress: Env=test, Kind=grafana, Title contains "Stress", Market=MY.
- IDs non-empty.

## Side Effects

- None.

## Errors

- Empty Kind/Env/Title/Market when table columns exist fails this leaf.

## Exit Code

- 0.

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/browser-agent/script/debug/chaos-useful-links/seedload"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	assertResolveOK(t, resp)

	if len(resp.Seeds) != 2 {
		t.Fatalf("seed count=%d want 2; urls=%v", len(resp.Seeds), seedURLs(resp.Seeds))
	}
	assertWantURLs(t, resp.Seeds, req.WantURLs)

	bySub := map[string]seedload.Seed{}
	for _, s := range resp.Seeds {
		if strings.Contains(s.URL, "logs.example.com/id") {
			bySub["logs"] = s
		}
		if strings.Contains(s.URL, "grafana.example.com/stress") {
			bySub["grafana"] = s
		}
	}
	if len(bySub) != 2 {
		t.Fatalf("could not map both table URLs; seeds=%v", seedURLs(resp.Seeds))
	}

	logs := bySub["logs"]
	if logs.Env != "live" || logs.Kind != "app_logs" || logs.Market != "ID" {
		t.Fatalf("logs seed meta: env=%q kind=%q market=%q want live/app_logs/ID", logs.Env, logs.Kind, logs.Market)
	}
	if !strings.Contains(logs.Title, "ID live") && !strings.Contains(logs.Title, "app logs") {
		t.Fatalf("logs seed title %q want table title", logs.Title)
	}

	g := bySub["grafana"]
	if g.Env != "test" || g.Kind != "grafana" || g.Market != "MY" {
		t.Fatalf("grafana seed meta: env=%q kind=%q market=%q want test/grafana/MY", g.Env, g.Kind, g.Market)
	}
	if !strings.Contains(g.Title, "Stress") {
		t.Fatalf("grafana seed title %q want Stress board", g.Title)
	}

	for _, s := range resp.Seeds {
		if strings.TrimSpace(s.ID) == "" {
			t.Fatalf("empty seed ID: %+v", s)
		}
	}
}
```
