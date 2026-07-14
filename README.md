# browser-agent

Browser automation and tracing toolkit: Chrome extension, control server, session UI, and embedded assets for driving and recording browser workflows.

## Components

| Path | Purpose |
|------|---------|
| `browseragent/` | Core Go library (WebSocket control, CDP jobs, embedded extension) |
| `cmd/browser-agent/` | CLI to install, bundle, and serve the browser agent |
| `cmd/browser-trace/` | CLI for HAR capture / browser trace workflows |
| `Chrome-Ext-Browser-Agent/` | Browser agent Chrome extension sources |
| `Chrome-Ext-Capture-API/` | Network capture extension sources |
| `react/` | Session page and popup React apps |
| `har-viewer/` | **Legacy** HAR file viewer (moved from repo root) |

## Quick start

Install and run the browser agent from the module root:

```sh
go run ./cmd/browser-agent install
go run ./cmd/browser-agent serve
```

Run unit tests:

```sh
go test ./browseragent/...
```

## Legacy HAR viewer (`har-viewer/`)

The original project-api-capture HAR viewer lives under `har-viewer/`. Casement integration was removed for the OSS release.

Place `.har` files in a directory, then:

```sh
cd har-viewer
go run ./ --dir /path/to/hars
```

Dev mode (Vite hot-reload):

```sh
go run ./har-viewer/script/dev
```

Build embedded frontend:

```sh
go run ./har-viewer/script/build
```

## Development

- Module path: `github.com/xhd2015/browser-agent`
- Doc-style tests: `doctest test ./tests/browser-agent` and `./tests/browser-trace`
- Remote: `https://github.com/xhd2015/browser-agent`

## License

See [LICENSE](LICENSE).