# Scenario

**Feature**: background references POST /v1/ext/hello for HTTP attach

```
POST http://127.0.0.1:{port}/v1/ext/hello
  { session_id, version?, features?, browser_product: "firefox", … }
  -> phase extension_connected (no WebSocket required)
```

## Preconditions

- BackgroundSourceTarget = hello-endpoint.

## Steps

1. Set `BackgroundSourceTarget = BgSrcHelloEndpoint`.

## Context

- Path token `/v1/ext/hello` is the GREEN marker (full URL host may be built dynamically).

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.BackgroundSourceTarget = BgSrcHelloEndpoint
	return nil
}
```
