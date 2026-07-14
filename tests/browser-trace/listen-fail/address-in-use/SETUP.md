# Scenario

**Feature**: address already in use → start fails with clear error

```
# Test occupies free loopback port, then browser-trace binds same Addr
Occupier.Listen(127.0.0.1:ephemeral)
browser-trace.Run(Addr=same) -> error mentions address / in use / listen
exit ≠ 0
```

## Preconditions

- `req.OccupyAddr = true`.
- `req.Addr` is a free port that `Run` will occupy first, then browser-trace fails on.

## Steps

1. Pick a free `127.0.0.1:port` and set `req.Addr`.
2. Leave extension mock off.
3. Run starts, occupies the port, then calls `browsertrace.Run` which must fail.

## Context

- Requirement scenario #1.
- Error text should mention listen failure and/or address in use (wording flexible).

```go
import (
	"net"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return err
	}
	addr := ln.Addr().String()
	_ = ln.Close()
	// Run will re-bind this exact address via OccupyAddr, then browser-trace fails.
	req.Addr = addr
	req.OccupyAddr = true
	req.ExtensionScript = ExtNone
	return nil
}
```
