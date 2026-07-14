# Scenario

**Feature**: embed index HTML is non-empty and has React root mount (A1)

```
SessionPageFS -> index.html | session-page.html
  body non-empty
  id="root" OR data-browser-agent-root
```

## Preconditions

- ModeEmbedFS from parent.

## Steps

1. Run reads embed index via SessionPageFS.
2. Assert root mount marker.

## Context

- Fixture may be minimal static HTML; production Vite build replaces later.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	// Mode already ModeEmbedFS from parent grouping.
	return nil
}
```
