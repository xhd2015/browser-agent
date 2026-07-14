---
name: analyse-har
description: >
  Analyse HAR files captured by Chrome-Ext-Capture-API to reverse-engineer the
  correct API calls for a user-described task.
---

# Analyse HAR

Reverse-engineer the **correct API contract** from a live browser recording, then
compare it to the project's existing client code and propose a fix.

## Inputs (required)

1. **HAR file path** — exported from `Chrome-Ext-Capture-API` (creator name:
   `API Capture - HAR Recorder`, version **≥ 1.1.0** recommended).
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

**First check body coverage** with `summarize_har.py`. If request/response bodies
are present, use them as primary evidence. Only fall back to heuristics when
bodies are missing (old extension build or failed capture).

## Workflow

### Step 1 — Summarize the HAR

Run the helper script (adjust `--host` to the app under study):

```bash
python3 skills/analyse-har/scripts/summarize_har.py /path/to/recording.har \
  --host app.example.com
```

For JSON output (includes parsed request/response summaries):

```bash
python3 skills/analyse-har/scripts/summarize_har.py /path/to/recording.har \
  --host app.example.com --json
```

Check `body_coverage` in JSON output:

- `with_request_body` / `entries` should match API POST count
- `with_response_body` / `entries` should match for v1.1.0+ captures

Also open the raw HAR when you need referer headers, auth headers, or fields
the summary truncates.

### Step 2 — Filter noise

**Exclude** by default:

- Static assets (`.js`, `.css`, images, fonts)
- Analytics / DEM telemetry (`dem.some-x.com`, `web-performance`, `web-custom`)
- Preflight `HEAD`/`OPTIONS` unless debugging CORS

**Keep**:

- `POST`/`PUT`/`PATCH`/`DELETE` to the application API host
- `GET` only when clearly part of the task flow

### Step 3 — Align narrative to timeline

Map the user's steps to entry indices (chronological `startedDateTime`):

1. Identify the **trigger** entry (user action → first API call).
2. Group subsequent calls into a **transaction** until the UI would settle.
3. Note **repeated motifs** (e.g. update → checkpoint → detail).

### Step 4 — Extract API contract

For each relevant call, document:

| Field | Source |
|-------|--------|
| Method + path | `request.url` (path only, strip origin) |
| Request body | `request.postData.text` (parse JSON if present) |
| Response envelope | `response.content.text` → `{code, msg, result}` |
| Key result fields | e.g. `result.id`, `result.jiraKey`, `result.data[]` |
| Auth | `Authorization` header (note Bearer, do not copy token) |
| Success signal | `response.status` 2xx + `code: 0` in JSON body |
| Ordering | index in filtered timeline |
| Idempotency / hops | count of repeated endpoints |

Produce a **sequence diagram** (mermaid or bullet list) for multi-step flows.

### Step 5 — Compare to project code

Search the repo for existing client wrappers and build a **gap table**:

| HAR (correct) | Code (current) | Gap |
|---------------|----------------|-----|
| `v3/update_field` | `update_field` | wrong endpoint version |
| `retry_requirement_checkpoint` | missing | post-update step absent |
| `v3/detail` | `detail-v2` | wrong detail endpoint |

Compare **both** request shapes and response handling.

### Step 6 — Propose fix (output template)

Deliver this structure to the user:

```markdown
## Task
<one sentence>

## HAR evidence
- File: <path>
- Extension: API Capture - HAR Recorder v1.1.0+
- Body coverage: request N/N, response N/N
- Relevant entries: <indices>
- Sequence: <ordered list of method path>

## Correct API contract
### Step 1: <name>
- `POST /api/...`
- Request: `{ ... }`
- Response: `{ code: 0, result: { ... } }`
- Notes: ...

## Root cause
<why current code fails>

## Proposed code changes
- `<client>.ts`: <functions/endpoints to change>
- `<panel>.tsx`: <UI/orchestration changes if any>
- Backend proxy: <none | reason>

## Open risks
- <only if bodies missing or base64-encoded>
```

Do **not** implement unless the user asks. This skill is analysis-first.

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
2. Re-run `summarize_har.py --json` and confirm paths, request bodies, and
   `response_body_summary.code === 0` for success paths.
3. If doctests exist for the feature, run `doctest test` on the relevant tree.