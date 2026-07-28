# Scenario

**Feature**: cdp must not be only the Phase-2 generic unknown job-type path

```
# Phase 2 (insufficient):
handleJob default:
  // Unknown types (e.g. cdp) remain unimplemented
  "not implemented: firefox job type=" + jobType

# Phase 3 (required):
handleJob case "cdp":
  method matrix Page.navigate | Runtime.evaluate | unsupported error
```

## Preconditions

- BackgroundSourceTarget = cdp-not-generic-unknown-only.

## Steps

1. Set `BackgroundSourceTarget = BgSrcCdpNotGenericUnknownOnly`.

## Context

- Combines dedicated case + method matrix anti-pattern check via
  `isCdpOnlyGenericUnknown`.
- This is the explicit requirement: **not only generic unknown job type for cdp**.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.BackgroundSourceTarget = BgSrcCdpNotGenericUnknownOnly
	return nil
}
```
