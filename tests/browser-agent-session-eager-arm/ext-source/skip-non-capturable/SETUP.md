# Scenario

**Feature**: eager auto-attach skips non-capturable URLs

```
eager attach candidate (register / create_tab / onUpdated)
  -> if !isCapturableTabURL(url): skip attach
  -> chrome://, chrome-extension://, devtools://, edge://, about: never auto-attached
```

## Preconditions

- ExtSourceTarget = skip-non-capturable.

## Steps

1. Set `ExtSourceTarget = ExtSrcSkipNonCapturable`.

## Context

- Job-path `isCapturableTabURL` in `pickTargetTabIdForSession` alone does **not**
  satisfy — guard must sit on **eager** attach paths.
- Expected **RED** until eager attach exists **and** guards with capturable check.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.ExtSourceTarget = ExtSrcSkipNonCapturable
	return nil
}
```
