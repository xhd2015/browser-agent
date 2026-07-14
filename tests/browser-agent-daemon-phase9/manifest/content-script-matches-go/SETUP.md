# Scenario

**Feature**: manifest content_scripts matches go page (P5)

```
manifest.json
  content_scripts[].matches includes loopback host + /go path
  js includes contentScript.js
```

## Preconditions

- ManifestProbe = content-script-matches-go.

## Steps

1. Set ManifestProbe content-script-matches-go.

## Context

- Broad `127.0.0.1/*` without `/go` path is too wide; phase 9 scopes to session page.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.ManifestProbe = ManifestProbeContentScriptMatchesGo
	return nil
}
```