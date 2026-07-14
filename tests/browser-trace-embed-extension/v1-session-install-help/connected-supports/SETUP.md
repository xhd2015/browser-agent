# Scenario

**Feature**: connected + supports — install help not primary (requirement #7)

```
# Extension Agent announces capable hello
Test Client -> POST /v1/hello {version: 1.2.0, features: [browser-trace]}
Test Client -> GET /v1/session
Control Server -> connected=true, supports_browser_trace=true
# Hint is operational (connected/recording-ready), not primary Load-unpacked tutorial
```

## Preconditions

- Hello with version ≥ 1.2.0 and feature `browser-trace`.
- Install path may still be present in JSON (optional keep); must not be the only story.

## Steps

1. Set `DoHello = true`.
2. Set `HelloVersion = "1.2.0"`.
3. Set `HelloFeatures = ["browser-trace"]`.

## Context

- Soft requirement from design: only if easy. Assert demotes install language as primary hint.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.DoHello = true
	req.HelloVersion = "1.2.0"
	req.HelloFeatures = []string{"browser-trace"}
	return nil
}
```
