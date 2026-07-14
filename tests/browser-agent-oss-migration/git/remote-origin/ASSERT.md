## Expected

- `git remote get-url origin` points at `xhd2015/browser-agent` on GitHub (SSH or HTTPS URL).

## Side Effects

- None.

## Errors

- Missing git repo or missing `origin` remote → leaf fails.

## Exit Code

- git command must succeed (exit 0).

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if resp.RunErr != "" {
		t.Fatalf("git remote get-url origin: %s\noutput: %s", resp.RunErr, resp.Stdout)
	}
	allowed := map[string]bool{
		"https://github.com/xhd2015/browser-agent":       true,
		"ssh://git@github.com/xhd2015/browser-agent":     true,
		"git@github.com:xhd2015/browser-agent.git":       true,
	}
	if !allowed[resp.RemoteURL] {
		t.Fatalf("origin URL = %q, want one of %v", resp.RemoteURL, []string{
			"https://github.com/xhd2015/browser-agent",
			"ssh://git@github.com/xhd2015/browser-agent",
			"git@github.com:xhd2015/browser-agent.git",
		})
	}
}
```