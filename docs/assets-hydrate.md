# Asset hydrate

How **browser-agent** (and **browser-trace**) obtain session-page and extension
assets when the binary was built without a full fat embed.

## Why hydrate exists

**Git** tracks only `placeholder.txt` under each embed root:

```text
browseragent/embedded/extension/placeholder.txt
browseragent/embedded/session-page/placeholder.txt
browsertrace/embedded/extension/placeholder.txt
```

Generated SPA/extension trees are **gitignored** (regenerate with bundle/install).
**Fat release** binaries and local builds after `script/*/bundle` ship a complete
`//go:embed` tree. Operators who **`go install`** from the module (or build with
placeholders only) get an **incomplete embed**. Hydrate fills the gap from
GitHub release archives without committing built assets.

### Staging a fat embed locally

```bash
# Auto-bundles when embeds are incomplete (placeholders only):
go run ./script/browser-agent/install
go run ./script/browser-trace/install

# Or stage only (no install):
go run ./script/browser-agent/bundle
go run ./script/browser-agent/bundle --fixture   # mini, no vite
go run ./script/browser-trace/bundle
```

Pass `--force-bundle` on install to refresh even when the on-disk embed looks complete.

## Cache layout

Assets materialize under a local cache (never the real user cache in tests;
operators use the real home cache):

```text
# Prefer XDG when set:
$XDG_CACHE_HOME/browser-agent/asset-cache/{product}/{version}/{kind}/

# Default when XDG_CACHE_HOME is unset:
~/.cache/browser-agent/asset-cache/{product}/{version}/{kind}/
```

Examples:

- `~/.cache/browser-agent/asset-cache/browser-agent/v0.2.0/session-page/`
- `~/.cache/browser-agent/asset-cache/browser-agent/v0.2.0/extension/`
- `~/.cache/browser-agent/asset-cache/browser-trace/v0.2.0/extension/`

## Runtime resolution

1. If the live embed is **complete** → use embed (source `embed`); no download.
2. Else if the local **asset-cache** is complete → use cache (source `cache`).
3. Else **download** a version-pinned archive and write the cache:
   `{BaseURL}/v{version}/{product}_v{version}_{kind}.tar.gz`  
   (never the `latest` tag).

Optional env:

- `BROWSER_AGENT_ASSET_BASE_URL` — download base (e.g. GitHub release download root)
- `HTTPS_PROXY` / standard proxy env — respected by the default HTTP client

## Operator CLI

```bash
# Report embed + cache completeness (no network)
browser-agent assets status

# Ensure both session-page and extension for this binary version
browser-agent assets ensure

browser-agent assets --help
```

**browser-trace** (extension only):

```bash
browser-trace assets status
browser-trace assets ensure
browser-trace assets --help
```

`assets ensure` uses `BROWSER_AGENT_ASSET_BASE_URL` when set, writes under
`~/.cache/browser-agent/asset-cache` (or `$XDG_CACHE_HOME/browser-agent/asset-cache`),
and is a no-op network-wise when the cache is already complete.

## Release archive names

Publish archives whose basenames match `browseragent.AssetReleaseNames(version)`,
for example for `v0.2.0`:

- `browser-agent_v0.2.0_session-page.tar.gz`
- `browser-agent_v0.2.0_extension.tar.gz`
- `browser-trace_v0.2.0_extension.tar.gz`

## Publishing release assets

From the browser-agent module root, pack the three hydrate archives from on-disk
embeds with **`script/github/release-assets`**:

```bash
# Pack only (writes under --out; no network / no gh)
go run ./script/github/release-assets --out ./dist --version v0.2.0
```

Sources:

- `browseragent/embedded/session-page`
- `browseragent/embedded/extension`
- `browsertrace/embedded/extension`

Optional **`--upload`** wraps **`gh`**: creates the GitHub release for the version
tag if it is missing, then uploads the packed archives with
`gh release upload --clobber`:

```bash
go run ./script/github/release-assets --out ./dist --version v0.2.0 --upload
```

Default is pack-only; use `--upload` only when you intend to publish to GitHub
(requires `gh` on `PATH` and an authenticated repo).

## Summary

| Situation | Behavior |
|-----------|----------|
| Fat release (complete embed) | Offline; serve from embed |
| `go install` / incomplete embed | Download into `~/.cache/browser-agent` asset-cache on ensure or first use |
| Explicit operator control | `browser-agent assets ensure` / `assets status` |
| Publish hydrate archives | `go run ./script/github/release-assets` (pack); add `--upload` for `gh` create/clobber |
