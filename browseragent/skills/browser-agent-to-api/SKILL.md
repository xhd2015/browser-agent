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

Read `manifest.json` first. It records the capture status, warnings, aggregate counts, and one HAR filename per tab.

```bash
python3 - <<'PY' /tmp/browser-agent-<session-slug>/manifest.json
import json, sys
manifest = json.load(open(sys.argv[1]))
print(json.dumps(manifest, indent=2))
PY
```

Treat `partial: true`, tab errors, and missing response bodies as uncertainty. Do not claim an endpoint contract is complete when the manifest says capture was partial.

Analyze every HAR named by the manifest, not just the active tab. Combine relevant entries by `startedDateTime` across files. Filter static assets, fonts, images, analytics, telemetry, and routine preflight traffic unless they are part of the question. Keep:

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

## Produce the API workflow

Report:

1. **Capture evidence** — output directory, relevant tab files, manifest partial status, and the selected timeline.
2. **API sequence** — ordered method/path list with request shape, response shape, and values passed to later calls.
3. **Authentication and session state** — required headers/cookies and how callers should obtain them, with secrets redacted.
4. **Behavioral details** — pagination, retries, polling, idempotency, redirects, and success conditions.
5. **Uncertainty** — missing bodies, failed requests, dynamic values, or browser-only mechanisms that were not proven.
6. **Reusable example** — `curl`, Python, or project-language code using placeholders for secrets and environment-specific identifiers.

When project code already contains an API client, compare its endpoint, request shape, ordering, and response handling to the HAR. Implement changes only when the user requested implementation; otherwise provide the evidence and proposed changes.

## Verify a replacement

When direct API code is implemented:

1. Run it against a safe target.
2. Confirm response status and application-level success fields.
3. Re-read the affected browser state or repeat a short browser capture when visual state matters.
4. Verify the replacement does not depend on copied short-lived credentials or hard-coded captured identifiers.
