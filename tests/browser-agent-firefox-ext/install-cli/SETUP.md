# Scenario

**Feature**: install-firefox-extension CLI (markers, improved steps, color) and top-level help

```
Operator -> install-firefox-extension -> path + about:debugging + Load Temporary Add-on
Operator -> install-firefox-extension -> step 3 open folder; step 4 select manifest.json
Operator -> install-firefox-extension --color|--no-color -> ANSI / plain / conflict
Operator -> browser-agent --help -> lists install-firefox-extension
```

## Preconditions

- Mode = install-cli.
- TestHome isolates ensure/extract for install leaves.
- Help leaf uses HandleCLI("--help") only.
- Color leaves use HandleCLI with `--color` / `--no-color` under isolated HOME.
- Content leaves prefer InstallFirefoxExtensionWithHome (plain text on buffer).

## Steps

1. Set Mode = ModeInstallCLI.
2. Leaf sets InstallCLIOp.

## Context

- Exit 0 on success; stdout ends with trailing newline for install.
- Color conflict is exit 1 with "cannot be specified together".

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Mode = ModeInstallCLI
	return nil
}
```
