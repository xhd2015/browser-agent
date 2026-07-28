# Scenario

**Feature**: SessionPageApp passes resolveInstallBrowser into InstallGuideline (W2)

```
SessionPageApp
  installBrowser = resolveInstallBrowser(...)
  -> <InstallGuideline browser={installBrowser} ... />
```

## Preconditions

- Mode already `ModeReactSrc` from parent.

## Steps

1. Set `ReactProbe = ReactProbeWiringInstallBrowser`.

## Context

- Ensures helper result drives install UX, not a hard-coded chrome default only.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.ReactProbe = ReactProbeWiringInstallBrowser
	return nil
}
```
