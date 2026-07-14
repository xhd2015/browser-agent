# Scenario

**Feature**: HandleCLI open-chrome help dispatch

```
HandleCLI(open-chrome --help) -> managed profile documentation
```

## Preconditions

- ModeCLIDispatch.
- No Chrome launch.

## Steps

1. Set Mode = ModeCLIDispatch.

## Context

- Help must document managed profile semantics.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModeCLIDispatch
	return nil
}
```
