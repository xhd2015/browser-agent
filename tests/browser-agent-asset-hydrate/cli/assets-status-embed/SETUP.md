# Scenario

**Feature**: `assets status` reports session-page and extension

```
HandleCLI(["assets", "status"], env{XDG_CACHE_HOME=tmp})
  -> nil error
  -> stdout mentions session-page and extension
  -> mentions complete/true/embed or cache (status of embed and/or cache)
  -> trailing \n
  -> no network required
```

## Preconditions

- XDG temp set (status may print cache paths under it).
- Live package embed is typically complete in this repo — status should still
  name both kinds.

## Steps

1. Set `CLIOp = CLIOpAssetsStatusEmbed`.
2. CLIArgs `assets status`; XDGCacheHome temp.

## Context

- Status must not require BASE_URL / HTTP.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.CLIOp = CLIOpAssetsStatusEmbed
	req.CLIArgs = []string{"assets", "status"}
	req.XDGCacheHome = t.TempDir()
	req.CLIEnv = map[string]string{
		EnvXDGCacheHome: req.XDGCacheHome,
	}
	return nil
}
```
