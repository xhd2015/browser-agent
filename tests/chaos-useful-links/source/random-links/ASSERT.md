## Expected

- Resolve succeeds with ≥3 seeds.
- Hosts include example.com, google.com, baidu.com.
- Source.Type is `random-links`.
- Each seed URL is https.

## Side Effects

- None (no network).

## Errors

- Missing required hosts or empty catalog fails this leaf.

## Exit Code

- 0.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	assertResolveOK(t, resp)

	if len(resp.Seeds) < 3 {
		t.Fatalf("random seed count=%d want ≥3; urls=%v", len(resp.Seeds), seedURLs(resp.Seeds))
	}
	if resp.Resolved.Source.Type != "random-links" {
		t.Fatalf("Source.Type=%q want random-links", resp.Resolved.Source.Type)
	}

	for _, host := range []string{"example.com", "google.com", "baidu.com"} {
		if !hasHost(resp.Seeds, host) {
			t.Fatalf("missing host %q in seeds %v", host, seedURLs(resp.Seeds))
		}
	}
	for _, s := range resp.Seeds {
		if !strings.HasPrefix(s.URL, "https://") {
			t.Fatalf("seed URL must be https: %q", s.URL)
		}
	}
}
```
