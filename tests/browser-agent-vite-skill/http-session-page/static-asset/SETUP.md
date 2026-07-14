# Scenario

**Feature**: committed fixture static asset served with HTTP 200 (B1)

```
GET /assets/session-page.js  (default AssetPath)
  -> 200, non-empty body
```

## Preconditions

- HTTPProbe = static-asset.
- Fixture includes `browseragent/embedded/session-page/assets/session-page.js`.

## Steps

1. Set HTTPProbe to static-asset.
2. Default AssetPath `/assets/session-page.js` unless overridden.

## Context

- No npm; fixture JS may be a one-line comment or stub export.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.HTTPProbe = HTTPProbeAsset
	if req.AssetPath == "" {
		req.AssetPath = "/assets/session-page.js"
	}
	return nil
}
```
