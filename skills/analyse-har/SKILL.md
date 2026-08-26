---
name: analyse-har
description: >
  Analyse HAR files or Browser Agent HAR export directories with
  `browser-agent har inspect` to reverse-engineer the correct API calls for a
  user-described task.
---

# Analyse HAR

Reverse-engineer the **correct API contract** from a live browser recording
(Chrome-Ext-Capture-API `.har` or Browser Agent export dir), then optionally
compare it to the project's existing client code.

## Inputs (required)

1. **HAR path** — a `.har` file or a Browser Agent export directory
   (`manifest.json` + `tab-*.har`). Chrome-Ext captures should be from
   `API Capture - HAR Recorder` **≥ 1.1.0** when bodies matter.
2. **User narrative** — what they did on the page in order.
3. **Task goal** — what the app should accomplish.

If any input is missing, ask once before analysing.

## Chrome-Ext-Capture-API capture model

Build the extension before recording:

```bash
go run ./script/chrome-ext-capture-api/build
```

Then load unpacked from `Chrome-Ext-Capture-API/build/` in `chrome://extensions`.

Read `Chrome-Ext-Capture-API/src/background.js` when unsure. As of **v1.1.0**:

| Captured | Notes |
|----------|-------|
| Request method, URL, headers | via `Network.requestWillBeSent` |
| Request `postData.text` | via `maxPostDataSize` + `Network.getRequestPostData` fallback |
| Response status, size, mimeType, timing | via `Network.responseReceived` / `loadingFinished` |
| Response JSON bodies | via `Network.getResponseBody` → `response.content.text` |

| Not captured / partial | Notes |
|----------------------|-------|
| Full cookie jar | cookies arrays stay empty |
| Binary response bodies | may appear as `content.encoding: "base64"` |
| v1.0.0 recordings | request/response bodies often empty — rebuild extension |

**First check body coverage** with `browser-agent har inspect`. If request/response
bodies are present, use them as primary evidence. Only fall back to heuristics when
bodies are missing (old extension build or failed capture).

## Workflow

### Step 1 — Summarize the HAR

Use the offline inspect CLI (works on a single `.har` or a Browser Agent export dir):

```bash
browser-agent har inspect summary /path/to/recording.har
browser-agent har inspect paths /path/to/recording.har --host app.example.com
browser-agent har inspect entries /path/to/recording.har --host app.example.com --method POST
browser-agent har inspect show /path/to/recording.har --match /api/example --json
```

Check `body_coverage` in `summary --json`:

- `with_request_body` / `entries` should match API POST count
- `with_response_body` / `entries` should match for v1.1.0+ captures

Secrets in headers/bodies are redacted by default. Use `--no-redact` only for local debugging.

### Step 2 — Filter noise

Noise (static assets, analytics / DEM telemetry) is **excluded by default**
(`--no-noise`). Pass `--noise` only when debugging those requests.

**Keep**:

- `POST`/`PUT`/`PATCH`/`DELETE` to the application API host
- `GET` only when clearly part of the task flow

### Step 3 — Align narrative to timeline

Map the user's steps to entry indices (chronological `startedDateTime`):

1. Identify the **trigger** entry (user action → first API call).
2. Group subsequent calls into a **transaction** until the UI would settle.
3. Note **repeated motifs** (e.g. update → checkpoint → detail).

### Step 4 — Extract contracts for critical calls

For lifecycle-critical calls, pull method/path, request body, response envelope,
key result fields, auth style (do not copy tokens), success signal, and ordering
dependencies from `har inspect show` (prefer `--json`). Prefer captured bodies as
evidence; a missing body is not evidence that the body is empty.

### Step 5 — Compare to project code (when relevant)

If the repo has an API client for the same flow, compare HAR endpoints, request
shapes, ordering, and response handling to the code and call out gaps. Skip this
when the user only asked what the browser APIs are.

### Step 6 — Report findings (information requirements)

Present findings in whatever shape fits the question. Do **not** follow a fixed
section template. The report must still make these information needs clear
(same bar as `browser-agent-to-api`):

1. **Capture provenance** — HAR path or Browser Agent export dir; that analysis
   used `browser-agent har inspect`; relevant entries; body-coverage / partial limits.
2. **API inventory with roles** — high-level method + path + role for APIs that
   matter for the goal (plus notable extras). Say when an expected capability
   (e.g. update) was **not** observed.
3. **Lifecycle / composition** — how those APIs compose for the task: order,
   which response fields feed later requests, success checks, optional verify/cleanup.
   Not a dump of every HAR entry.
4. **Contracts for critical steps** — enough detail to reuse mutations and their
   dependencies (auth without secrets, request/response shapes, success signals).
5. **Gaps and uncertainty** — missing bodies, failed requests, dynamic values,
   or flows the capture does not prove.

Include root-cause / proposed code changes or a reusable example only when they
help answer the request. Do **not** implement unless the user asks. This skill is
analysis-first.

## Heuristics

### When bodies are missing (legacy v1.0.0 HAR)

Ask user to rebuild and re-record:

```bash
go run ./script/chrome-ext-capture-api/build
```

Then rely on:

1. **Call order** and **count** (strong signal)
2. **Referer** URL (query params)
3. **Response size** stability across hops
4. **Parallel working UI** on the same host

### v3 vs legacy naming

If HAR shows `/foo/v3/bar` and code uses `/foo/bar` or `/foo/bar-v2`, treat as
endpoint version mismatch — primary fix candidate.

## Verification (optional)

After a fix is implemented:

1. Re-record with Chrome-Ext-Capture-API (v1.1.0+) while repeating the same user flow.
2. Re-run `browser-agent har inspect show … --match … --json` and confirm paths,
   request bodies, and response `code === 0` for success paths.
3. If doctests exist for the feature, run `doctest test` on the relevant tree.