# Scenario

**Feature**: run posts job type run with file source (B2)

```
Write temp script.js with doctest-run-marker
Serve + fake WS
HandleCLI session run --session-id X --addr <base> <script.js>
  -> observed type=run
  -> params source (or expression) includes file content marker
  -> CLI nil error; stdout trailing \n
```

## Preconditions

- JobCLI = run.
- RunScriptBody has unique marker string.

## Steps

1. Set JobCLIRun.
2. Set RunScriptBody with marker `doctest-run-marker`.
3. Leave RunScriptPath empty so harness writes under BaseDir.

## Context

- Requirement B2. CLI reads file client-side; job carries source.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.JobCLI = JobCLIRun
	req.RunScriptBody = "// doctest-run-marker\nconsole.log('hello-from-run');\n"
	req.RunScriptPath = ""
	req.CLIArgs = nil
	return nil
}
```
