# Scenario

**Feature**: embedded mini background mentions CDP evaluate + job type tokens (D3)

```
Read browseragent/embedded/extension/background.js (or extract)
  at least Runtime.evaluate
  job type tokens present (not pure stub-only without CDP names)
```

## Preconditions

- ExtSourceTarget = embedded-cdp-tokens.

## Steps

1. Set ExtSrcEmbeddedCDPTokens.

## Context

- Requirement D3. Mini may be thinner than shell; still not stub-only without CDP names.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.ExtSourceTarget = ExtSrcEmbeddedCDPTokens
	return nil
}
```
