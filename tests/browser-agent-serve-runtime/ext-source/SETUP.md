# Scenario

**Feature**: extension sources implement WS agent protocol tokens (filesystem)

```
# no Chrome
Test Client -> read Chrome-Ext-Browser-Agent public background/contentScript
Test Client -> read browseragent/embedded/extension background (or extract)
```

## Preconditions

- Mode is `ext-source`.
- ModuleRoot resolved by root Setup.
- No real browser; content asserts only.

## Steps

1. Set `Mode = ModeExtSource`.
2. Children set ExtSourceTarget.

## Context

- Requirement E1–E3. Protocol compliance for eval/info jobs when extension is real.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModeExtSource
	return nil
}
```
