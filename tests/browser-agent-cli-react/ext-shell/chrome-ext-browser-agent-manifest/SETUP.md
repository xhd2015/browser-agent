# Scenario

**Feature**: Chrome-Ext-Browser-Agent manifest name + port 43761 (H1)

```
Chrome-Ext-Browser-Agent exists
  manifest name/description references Browser Agent
  hosts include 43761
```

## Preconditions

- Leaf under ModeExtShell (parent sets Mode).

## Steps

1. No extra fields (single leaf under grouping).

## Context

- Production shell may still be stub background; manifest is the contract.

```go
import (
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModeExtShell
	if req.ModuleRoot == "" {
		t.Fatal("ModuleRoot must be set by root Setup")
	}
	// Document expected on-disk root for Assert/Run (no create; implementer owns tree).
	_ = filepath.Join(req.ModuleRoot, "Chrome-Ext-Browser-Agent")
	return nil
}
```
