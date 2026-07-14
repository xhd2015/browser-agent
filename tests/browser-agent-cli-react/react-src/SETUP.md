# Scenario

**Feature**: Vite React source layout under module root (filesystem)

```
ModuleRoot/react/src/products/browser-agent.ts
ModuleRoot/react/src/apps/{session-page,popup}/main.tsx
ModuleRoot/react/src/ui/InstallGuideline.tsx
```

## Preconditions

- ModeReactSrc.
- ModuleRoot set by root Setup from DOCTEST_ROOT.
- No npm/webpack execution — existence + content probes only.

## Steps

1. Set Mode = ModeReactSrc.
2. ReactProbe set by leaf.

## Context

- Preferred path is `react/` per requirement; harness may note
  `project-api-capture-react/` but implementer should provide `react/`.

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
