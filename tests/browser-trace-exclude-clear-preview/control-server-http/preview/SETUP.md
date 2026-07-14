# Scenario

**Feature**: `GET /preview` HTML live viewer (req #5, #6)

```
Mock Extension ?-> POST /v1/entries {…}
Test Client -> GET /preview?session=<id>
Control Server -> text/html live viewer | 404
  (viewer polls /v1/entries or embeds snapshot)
```

## Preconditions

- Final probe is `ProbePreview`.
- Success responses are HTML; errors may be HTML or JSON with not-found.

## Steps

1. Set `Probe = ProbePreview` (`"preview"`).

## Context

- Preferred popup target: `http://127.0.0.1:43759/preview?session=…`
  (tests use free port base URL).
- Extension fallback `preview.html` is out of scope.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Probe = ProbePreview
	return nil
}
```
