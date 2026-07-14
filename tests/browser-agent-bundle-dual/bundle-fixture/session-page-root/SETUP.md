# Scenario

**Feature**: Bundle UseFixture stages session-page with root mount (A2)

```
Bundle(UseFixture) -> SessionPageDir/index.html
  root mount id="root" | data-browser-agent-root
```

## Preconditions

- ModeBundle parent Setup complete.
- BundlePasses = 1.

## Steps

1. Single Bundle pass; assert SessionPageDir content.

## Context

- Session-page index may be `index.html` or `session-page.html`.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	if req.Mode != ModeBundle {
		t.Fatalf("Mode = %q, want %q", req.Mode, ModeBundle)
	}
	req.BundlePasses = 1
	return nil
}
```
