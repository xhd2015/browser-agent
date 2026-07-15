# Scenario

**Feature**: operator docs for asset hydrate / cache / ensure (P7)

```
Test Client -> read docs/**/*.md, README.md, browseragent/SKILL.md
  -> combined text documents hydrate path for operators
```

## Preconditions

- Mode is docs.
- ModuleRoot resolved by root Setup.

## Steps

1. Set `Mode = ModeDocs`.

## Context

- Classic TDD — RED until implementer adds/updates operator docs.
- FS-only; no network.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModeDocs
	return nil
}
```
