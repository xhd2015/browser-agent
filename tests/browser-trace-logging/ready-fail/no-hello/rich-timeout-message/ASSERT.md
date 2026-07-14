## Expected

- Exit code ≠ 0.
- Stderr (and/or returned error text) includes:
  - timeout / ready failure language
  - stage hint: `no_hello` **or** hello/connect language
  - session URL or `/go?session=`
  - install path / install / extension path hint when extract succeeds
    (if install path cannot be known in-test, at least “install” or
    host_permissions-style guidance is acceptable)
- Stdout must **not** be a success path line (no sole session-dir success print).

## Side Effects

- Session dir may exist with partial meta; no successful complete HAR required.

## Errors

- Ready-phase failure (extension never said hello).

## Exit Code

- Non-zero.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertExitNonZero(t, resp)

	text := combinedErrText(resp)
	low := strings.ToLower(text)

	hasTimeout := strings.Contains(low, "timeout") ||
		strings.Contains(low, "timed out") ||
		strings.Contains(low, "deadline") ||
		strings.Contains(low, "ready fail")
	if !hasTimeout {
		t.Fatalf("expected timeout/ready-failure language; got:\n%s", text)
	}

	hasStage := strings.Contains(low, "no_hello") ||
		strings.Contains(low, "hello") ||
		strings.Contains(low, "connect") ||
		strings.Contains(low, "did not connect")
	if !hasStage {
		t.Fatalf("expected stage hint (no_hello / hello / connect); got:\n%s", text)
	}

	hasSessionURL := strings.Contains(low, "/go?session=") ||
		strings.Contains(low, "go?session") ||
		strings.Contains(low, "session=") ||
		(strings.Contains(low, "http://") && strings.Contains(low, "session"))
	if !hasSessionURL {
		t.Fatalf("expected session URL or /go?session= in ready-timeout message; got:\n%s", text)
	}

	hasInstallHint := strings.Contains(low, "install") ||
		strings.Contains(low, "host_permission") ||
		strings.Contains(low, "extension") ||
		strings.Contains(low, "enable")
	if !hasInstallHint {
		t.Fatalf("expected install/extension enable hint; got:\n%s", text)
	}

	// Must not look like success-only stdout contract.
	if resp.ExitCode == 0 {
		t.Fatal("unexpected exit 0 on ready timeout")
	}
}
```
