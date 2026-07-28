# Scenario

**Feature**: React InstallGuideline + SessionPageApp Firefox install UX

```
Test Client -> read react/src/ui/InstallGuideline.tsx
  firefox: about:debugging + temporary add-on + browser-agent-firefox
  chrome: chrome://extensions preserved

Test Client -> read react/src/ui/SessionPageApp.tsx
  wires firefox browser into install panel
```

## Preconditions

- Mode `ModeReactSrc`.
- ModuleRoot resolved by root Setup.
- No npm/webpack — existence + content probes only.

## Steps

1. Set `Mode = ModeReactSrc`.
2. Leaf sets `ReactProbe`.

## Context

- Requirement surfaces R1–R3.
- Classic TDD: Firefox markers absent today → RED on R1/R2; R3 GREEN.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Mode = ModeReactSrc
	if req.ModuleRoot == "" {
		t.Fatal("ModuleRoot must be set by root Setup")
	}
	return nil
}
```
