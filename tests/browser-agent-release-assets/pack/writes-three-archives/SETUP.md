# Scenario

**Feature**: pack-only writes three non-empty tar.gz archives with release basenames

```
go run ./script/github/release-assets --out DIR --version v0.2.0
  -> DIR/browser-agent_v0.2.0_session-page.tar.gz  (size > 0)
  -> DIR/browser-agent_v0.2.0_extension.tar.gz     (size > 0)
  -> DIR/browser-trace_v0.2.0_extension.tar.gz     (size > 0)
  exit 0
```

## Preconditions

- Parent pack Setup sets Mode, OutDir, Version, Args.
- Embed sources exist under ModuleRoot.

## Steps

1. Rely on parent pack Setup for Mode/OutDir/Version/Args.
2. No extra flags (pack-only; no `--upload`).

## Context

- Asserts filesystem under `--out`, not stdout shape (stdout may log paths).

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	// Parent already set ModePack + OutDir + Version + Args.
	if req.Mode != ModePack {
		t.Fatalf("Mode=%q want %q", req.Mode, ModePack)
	}
	if req.OutDir == "" {
		t.Fatal("OutDir empty after pack Setup")
	}
	return nil
}
```
