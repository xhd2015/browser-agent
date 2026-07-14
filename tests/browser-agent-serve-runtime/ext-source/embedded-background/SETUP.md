# Scenario

**Feature**: embedded mini extension background has WS agent protocol tokens (E3)

```
browseragent/embedded/extension/background.js (or ExtractEmbeddedExtension path)
  contains /v1/ws or ws:// + hello + job + result
```

## Preconditions

- ExtSourceTarget = embedded-background.
- BaseDir available for extract fallback.

## Steps

1. Set ExtSourceTarget embedded-background.

## Context

- CI embed must stay protocol-capable even if shell is larger.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.ExtSourceTarget = ExtSrcEmbeddedBackground
	return nil
}
```
