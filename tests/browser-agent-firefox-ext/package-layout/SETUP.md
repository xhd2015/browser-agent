# Scenario

**Feature**: Firefox-Ext-Browser-Agent/public source package layout on ModuleRoot

```
ModuleRoot/Firefox-Ext-Browser-Agent/public
  -> manifest.json (MV3 + gecko.id, no debugger)
  -> background.js, contentScript.js, popup.html, popup.js
```

## Preconditions

- Mode = package-layout.
- ModuleRoot from root Setup.
- Read-only FS probes; no build/npm.

## Steps

1. Set Mode = ModePackageLayout.
2. Leaf sets PackageLayoutOp.

## Context

- Classic TDD: package missing → FileExists false until implementer adds sources.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Mode = ModePackageLayout
	return nil
}
```
