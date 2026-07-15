# Scenario

**Feature**: GET `/go?session=<id>` — injected session page HTML title

```
Test Client -> GET /go?session=<id>
Registry Control Server -> SPA HTML with <title>{id} - Browser Agent</title>
```

## Preconditions

- Mode `ModeGoHTML`.
- Registry control handler serves embedded SPA + `injectSessionBoot` when embed present.

## Steps

1. Set `Mode = ModeGoHTML`.
2. Leaf pre-creates session and probes title.

## Context

- Primary behavior leaf for inject path (prefer HTTP over source).
- Classic TDD: currently static `Browser Agent Session` → RED until implementer.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModeGoHTML
	return nil
}
```
