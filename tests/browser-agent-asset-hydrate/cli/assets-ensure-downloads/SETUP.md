# Scenario

**Feature**: `assets ensure` downloads into cache via BASE_URL httptest

```
XDG_CACHE_HOME=tmp (cold)
BROWSER_AGENT_ASSET_BASE_URL=httptest/releases/download
httptest serves session-page + extension tar.gz by path
  -> HandleCLI(["assets", "ensure"])
  -> nil error
  -> CacheComplete(session-page) && CacheComplete(extension)
```

## Preconditions

- Fresh XDG temp (empty cache).
- `CLIServeBothArchives` starts dual-archive httptest and injects BASE_URL env.

## Steps

1. Set `CLIOp = CLIOpAssetsEnsureDownloads`.
2. CLIArgs `assets ensure`; XDG temp; CLIServeBothArchives true.

## Context

- Ensure covers both kinds for current ClientVersion / v0.2.0.
- Progress may go to stderr; success is nil error + complete cache.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.CLIOp = CLIOpAssetsEnsureDownloads
	req.CLIArgs = []string{"assets", "ensure"}
	req.XDGCacheHome = t.TempDir()
	req.CLIEnv = map[string]string{
		EnvXDGCacheHome: req.XDGCacheHome,
	}
	req.CLIServeBothArchives = true
	return nil
}
```
