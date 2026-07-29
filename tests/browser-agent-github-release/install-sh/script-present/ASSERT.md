## Expected

- Exit 0; install.sh readable.
- Body mentions `github.com/xhd2015/browser-agent`.
- Body builds asset name with `browser-agent-v` and target (`darwin-arm64` / `linux-amd64` etc.).
- Body mentions `INSTALL_TAG` and `GOPATH` or `/usr/local/bin`.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	if resp == nil {
		t.Fatal("resp is nil")
	}
	assertExitZero(t, resp)
	body := resp.InstallSHText
	if strings.TrimSpace(body) == "" {
		t.Fatalf("install.sh empty path=%s", resp.InstallSHPath)
	}
	if !strings.Contains(body, "github.com/xhd2015/browser-agent") {
		t.Fatalf("install.sh missing repo URL; path=%s", resp.InstallSHPath)
	}
	if !strings.Contains(body, "browser-agent-v") {
		t.Fatalf("install.sh missing browser-agent-v asset prefix; path=%s", resp.InstallSHPath)
	}
	if !strings.Contains(body, "darwin-arm64") && !strings.Contains(body, "linux-amd64") {
		t.Fatalf("install.sh missing platform target tokens; path=%s", resp.InstallSHPath)
	}
	if !strings.Contains(body, "INSTALL_TAG") {
		t.Fatalf("install.sh missing INSTALL_TAG; path=%s", resp.InstallSHPath)
	}
	if !strings.Contains(body, "GOPATH") && !strings.Contains(body, "/usr/local/bin") {
		t.Fatalf("install.sh missing install dest (GOPATH or /usr/local/bin); path=%s", resp.InstallSHPath)
	}
}
```
