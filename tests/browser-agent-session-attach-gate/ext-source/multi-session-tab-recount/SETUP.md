# Scenario

**Feature**: leave handler re-queries remaining session control tabs (multi-tab)

```
Two /go?session=S tabs in same window; one closed
  -> leave path re-queries open control tabs
  -> remaining >= 1 → stay armed (rebind entry.tabId); do not unregister/detach
Last control tab leaves
  -> remaining == 0 → teardown + detach
```

## Preconditions

- ExtSourceTarget = multi-session-tab-recount.

## Steps

1. Set `ExtSourceTarget = ExtSrcMultiSessionTabRecount`.

## Context

- Current master: `tabs.onRemoved` / navigate-away only match `entry.tabId` (last
  register wins) — closing the registered tab disarms even if another control tab
  remains. Telemetry `collectSessionPageTelemetry` is **not** leave recount.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.ExtSourceTarget = ExtSrcMultiSessionTabRecount
	return nil
}
```
