## Expected

- Load succeeds; 5 clean seeds.
- No seed URL ends with `)`, `.`, `,`, `;`, or `]`.
- Want path segments present (trail, comma, semi, paren, brack).

## Side Effects

- None.

## Errors

- Keeping trailing punctuation on URLs fails this leaf.

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

	if req.WantCountSet && len(resp.Seeds) != req.WantCount {
		t.Fatalf("seed count=%d want %d; urls=%v", len(resp.Seeds), req.WantCount, seedURLs(resp.Seeds))
	}
	assertWantURLs(t, resp.Seeds, req.WantURLs)

	trailers := []string{")", ".", ",", ";", "]"}
	for _, s := range resp.Seeds {
		u := s.URL
		for _, tr := range trailers {
			if strings.HasSuffix(u, tr) {
				t.Fatalf("URL %q still has trailing %q", u, tr)
			}
		}
		// Exact clean forms preferred.
		if strings.Contains(u, "trail") && u != "https://example.com/trail" {
			// allow only clean trail URL
			if !strings.HasSuffix(u, "/trail") || strings.ContainsAny(u[len(u)-1:], ").,;]") {
				t.Fatalf("trail URL not cleaned: %q", u)
			}
		}
	}
}
```
