# Scenario

**Feature**: session page boot inject uses firefox when stamp indicates firefox

```
SessionNew(Browser=firefox) -> GET /go?session=<id>
  -> #browser-agent-boot / __BROWSER_AGENT.browser == "firefox"
```

## Preconditions

- Mode = boot-inject.
- Relies on snap path or browser field after SessionNew stamp
  (`sessionSnapIsFirefox` / FormatSessionBootJSONWithBrowser).

## Steps

1. Set Mode = ModeBootInject.

## Context

- injectSessionBoot already selects firefox when path contains
  browser-agent-firefox; this leaf is RED until Create/SessionNew stamps path.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Mode = ModeBootInject
	return nil
}
```
