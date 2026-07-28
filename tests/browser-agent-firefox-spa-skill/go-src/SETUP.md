# Scenario

**Feature**: Go inject / fallback session HTML install markers for Firefox

```
Test Client -> read browseragent/server.go
  injectSessionBoot / writeFallbackSessionHTML / install helpers
  firefox: about:debugging + temporary add-on + browser-agent-firefox
  chrome: chrome://extensions preserved
```

## Preconditions

- Mode `ModeGoSrc`.
- ModuleRoot resolved by root Setup.
- Read-only FS under `browseragent/`.

## Steps

1. Set `Mode = ModeGoSrc`.
2. Leaf sets `GoSrcProbe`.

## Context

- Requirement surfaces G1–G2.
- extension_firefox.go may contribute about:debugging CLI copy; session page
  inject must also surface Firefox markers for SPA/fallback HTML.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Mode = ModeGoSrc
	if req.ModuleRoot == "" {
		t.Fatal("ModuleRoot must be set by root Setup")
	}
	return nil
}
```
