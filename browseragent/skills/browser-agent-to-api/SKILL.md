---
name: browser-agent-to-api
description: >-
  Capture a web application's browser network traffic as per-tab HAR files and
  derive a reusable API workflow from the observed requests. Use when the user
  asks to inspect website APIs, reverse-engineer browser requests, reproduce a
  browser action with direct HTTP calls, or runs /browser-agent-to-api.
---

# Browser Agent to API

Capture the user flow with Browser Agent, then use the HAR evidence to describe or implement the smallest reusable API sequence.

## Capture

1. Create or reuse one Chrome Browser Agent session. HAR capture is not supported by the Firefox extension.
2. Inspect the session before recording:

```bash
browser-agent session info --session-id <session-id> --json
```

3. Start capture. It enrolls every current user HTTP(S) tab in the session window and any eligible tab opened while recording:

```bash
browser-agent session har start <session-id>
```

4. Perform the browser operations that demonstrate the requested flow. Prefer explicit `--tab-id` for `eval`, `run`, `logs`, `screenshot`, and `cdp` operations.
5. End capture even when a browser operation fails, so useful partial traffic is exported:

```bash
browser-agent session har end <session-id>
```

The default output is `/tmp/browser-agent-<session-slug>/`. To choose another empty directory:

```bash
browser-agent session har end <session-id> --output-dir /tmp/my-api-capture
```

HAR files can contain authorization headers, cookies, request bodies, and sensitive responses. Do not print secrets or commit capture directories.

## Inspect the export

Use offline inspect commands (no live session required). Start with summary, then path inventory, then show the create/mutation calls:

```bash
DIR=/tmp/browser-agent-<session-slug>

browser-agent har inspect summary "$DIR"
browser-agent har inspect paths "$DIR" --host <app-host>
browser-agent har inspect entries "$DIR" --host <app-host> --method POST --limit 50
browser-agent har inspect show "$DIR" --match /api/example --json
```

`summary` reports `partial` and body coverage. Treat `partial: true`, tab errors, and missing response bodies as uncertainty. Do not claim an endpoint contract is complete when the capture is partial.

Inspect every tab listed by the manifest (default), not just the active tab. Prefer `--match` / `--path-substr` over reading raw HAR. Static assets and analytics are dropped by default (`--no-noise`). Keep:

- application API `POST`, `PUT`, `PATCH`, and `DELETE` requests;
- `GET` requests that load or verify task data;
- redirects, authentication exchanges, polling, pagination, and follow-up mutations required by the flow.

For each relevant request, extract:

- method, origin, path, and query parameters;
- request headers that affect behavior, naming sensitive values without reproducing them;
- cookie or token requirements;
- parsed request body and content type;
- response status, envelope, identifiers, pagination fields, and error shape;
- ordering and data dependencies between requests.

Prefer captured request and response bodies as evidence. A missing body is not evidence that the body is empty.

## Report what the capture implies

Present findings in whatever shape fits the question. Do **not** follow a fixed section template. The report must still make these information needs clear:

1. **Capture provenance** — export directory or `.har` path; that analysis used `browser-agent har inspect`; relevant tabs/entries; `partial` / body-coverage limits on confidence.
2. **API inventory with roles** — a high-level view of the APIs that matter for the goal (method + path + role: what the call is for). Prefer APIs needed for the task plus notable extras; say explicitly when an expected capability (e.g. update) was **not** observed.
3. **Lifecycle / composition** — how those APIs compose for the user goal: order, which response fields feed later requests, success checks, and optional verify or cleanup. This is the reusable workflow, not a dump of every HAR entry.
4. **Contracts for critical steps** — enough detail to reuse mutations and their dependencies (auth style without secrets, request/response shapes, success signals). Depth may vary by importance.
5. **Gaps and uncertainty** — missing bodies, failed requests, dynamic values, browser-only mechanisms, or flows the capture does not prove.

Include a code/client gap comparison or a reusable `curl`/project-language example only when it helps answer the request. Implement product changes only when the user asked for implementation.

## Verify a replacement

When direct API code is implemented:

1. Run it against a safe target.
2. Confirm response status and application-level success fields.
3. Re-read the affected browser state or repeat a short browser capture when visual state matters.
4. Verify the replacement does not depend on copied short-lived credentials or hard-coded captured identifiers.
