# Scenario

**Feature**: `assets ensure` downloads browser-trace extension into cache via BASE_URL httptest

```
XDG_CACHE_HOME=tmp (cold)
BROWSER_AGENT_ASSET_BASE_URL=httptest/releases/download
httptest serves browser-trace … extension.tar.gz by path
  -> HandleCLI(["assets", "ensure"])
  -> nil error
  -> CacheComplete(browser-trace, version, extension)
```

## Preconditions

- Fresh XDG temp (empty cache).
- `CLIServeExtensionTar` starts extension-only httptest and injects BASE_URL env.

## Steps

1. Set `CLIOp = CLIOpAssetsEnsure`.
2. CLIArgs `assets ensure`; XDG temp; CLIServeExtensionTar true.

## Context

- Ensure covers **extension only** for product browser-trace at current
  ClientVersion / v0.2.0 (no session-page).
- Progress may go to stderr; success is nil error + complete extension cache.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.CLIOp = CLIOpAssetsEnsure
	req.CLIArgs = []string{"assets", "ensure"}
	req.XDGCacheHome = t.TempDir()
	req.CLIEnv = map[string]string{
		EnvXDGCacheHome: req.XDGCacheHome,
	}
	req.CLIServeExtensionTar = true
	return nil
}
```
