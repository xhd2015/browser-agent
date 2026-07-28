# Scenario

**Feature**: install-firefox-extension splits open-folder (step 3) from select-manifest.json (step 4)

```
InstallFirefoxExtensionWithHome(stdout, TestHome)
  -> step 3 open folder (path)
  -> step 4 select manifest.json
  -> about:addons / .xpi caution
  -> warning: restart note
```

## Preconditions

- InstallCLIOp = steps-folder-then-manifest.
- TestHome set by root Setup (HOME isolation via package WithHome).
- Auto color off on bytes.Buffer — plain-text content asserts only.

## Steps

1. Set InstallCLIOp = InstallCLIOpStepsFolderThenManifest.

## Context

- Classic TDD: RED while step 3 still conflates folder + manifest.json and restart lacks `warning:`.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.InstallCLIOp = InstallCLIOpStepsFolderThenManifest
	return nil
}
```
