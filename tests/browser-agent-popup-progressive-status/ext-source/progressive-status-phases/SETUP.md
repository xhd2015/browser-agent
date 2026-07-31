# Scenario

**Feature**: popup surfaces distinct progressive status phases

```
popup shell status area
  -> phase row/hook: daemon/control health
  -> phase row/hook: session/WS connected or waiting
  -> phase row/hook: debugger/armed (attach set)
  -> each phase updates independently (not single hung path)
```

## Preconditions

- ExtSourceTarget = progressive-status-phases.

## Steps

1. Set `ExtSourceTarget = ExtSrcProgressiveStatusPhases`.

## Context

- Current popup has only `#ctrl-status` + `#conn-hint` — **RED** until implementer
  adds session/ws and debugger/armed phase DOM hooks (ids, data-phase, or rows).
- Single control-health span does **not** satisfy this leaf.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.ExtSourceTarget = ExtSrcProgressiveStatusPhases
	return nil
}
```
