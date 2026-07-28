# Scenario

**Feature**: screenshot job uses tabs.captureVisibleTab (viewport)

```
handleJob type=screenshot
  -> browser.tabs.captureVisibleTab(windowId, { format })
  -> result { type: "screenshot", base64|dataUrl, … }
```

## Preconditions

- BackgroundSourceTarget = screenshot-capture-visible.

## Steps

1. Set `BackgroundSourceTarget = BgSrcScreenshotCaptureVisible`.

## Context

- Viewport capture only; full-page CDP is Phase 3.
- `Page.captureScreenshot` alone without captureVisibleTab is **not** Phase 2 path.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.BackgroundSourceTarget = BgSrcScreenshotCaptureVisible
	return nil
}
```
