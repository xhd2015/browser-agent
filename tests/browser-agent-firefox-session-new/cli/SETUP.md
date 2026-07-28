# Scenario

**Feature**: HandleCLI session new --browser (help, firefox, unknown)

```
HandleCLI session new --browser firefox -> inject OpenFirefoxFn + path + about:debugging
HandleCLI session new --help            -> documents --browser
HandleCLI session new --browser safari  -> unknown browser error, nonzero
```

## Preconditions

- Mode = cli.
- inject.WithSessionNewHooks for open leaves; no real browser.

## Steps

1. Set Mode = ModeCLI.

## Context

- Help and unknown-browser leaves do not require a healthy daemon open path
  (unknown validates before spawn when possible; help is pure).

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Mode = ModeCLI
	return nil
}
```
