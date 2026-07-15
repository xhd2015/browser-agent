# Scenario

**Feature**: React SessionPageApp sets `document.title` from session id

```
Test Client -> read react/src/ui/SessionPageApp.tsx
  document.title = sid + " - Browser Agent" when sid known
  skip / leave static when sid empty
```

## Preconditions

- Mode `ModeReactSrc`.
- ModuleRoot resolved by root Setup.
- No npm/webpack — existence + content probes only.

## Steps

1. Set `Mode = ModeReactSrc`.
2. Leaf sets `ReactProbe`.

## Context

- Requirement surface: React client title (scenario 3 + empty-sid guard).
- Classic TDD: no title effect yet → RED until implementer.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModeReactSrc
	if req.ModuleRoot == "" {
		t.Fatal("ModuleRoot must be set by root Setup")
	}
	return nil
}
```
