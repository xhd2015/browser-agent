# Scenario

**Feature**: RunDaemon writes daemon discovery server.json

```
RunDaemon -> {BaseDir}/server.json with pid, addr, base_url, base_dir
```

## Preconditions

- RunDaemonOp writes-server-json.

## Steps

1. Set `RunDaemonOp = RunDaemonOpWritesServerJSON`.

## Context

- PID must be `os.Getpid()` of the test process running RunDaemon.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.RunDaemonOp = RunDaemonOpWritesServerJSON
	return nil
}
```