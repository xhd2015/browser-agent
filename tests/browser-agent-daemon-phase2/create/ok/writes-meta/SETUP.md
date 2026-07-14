# Scenario

**Feature**: Create writes meta.json with discovery fields

```
Create(id) -> meta.json (session_id, addr, base_url, session_url, system_prompt_path, product, control_port)
```

## Preconditions

- CreateCase ok; SessionID `my-flow`.

## Steps

1. Inherit ok Setup (CreateCase + SessionID).

## Context

- Phase 2 omits extension_install_path in meta.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	return nil
}
```