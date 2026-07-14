---
name: summarize-the-workflow
description: >
  After a successful browser co-pilot session (Chrome + browser-agent + agent
  tries/corrections), distill a reusable workflow playbook. Create a new
  workflow markdown file or update an existing/outdated one with a brief dated
  changelog. Prefer steps and key pages over scripts. Triggers: summarize the
  workflow, /summarize-the-workflow, write playbook from this session, update
  workflow from what we just did.
---

# Summarize the workflow

Turn a **successful** Chrome + browser-agent co-pilot session into a durable
playbook so the same class of request can be repeated later with less thrash.

## When to use

- User opened a site in Chrome, connected a **browser-agent session**, and had
  the agent act on their behalf.
- After tries/corrections, the goal **succeeded** (or user explicitly wants a
  partial/WIP playbook).
- User asks to **summarize**, **create**, or **update** a workflow doc from
  that session.

## When not to use

- Session still failed and user did not ask for a WIP playbook.
- User only wants a one-off dump of chat (not a reusable workflow).
- Unrelated pure-code design tasks with no browser co-pilot path.

## Inputs

Ask once if missing:

1. **Create vs update**
   - **Create:** new file path, or user says “whatever” / don’t care.
   - **Update:** path to an existing workflow (may be outdated, legacy, or incomplete).
2. **Name / slug** (optional) — see [Naming](#naming).
3. **Scope** (optional) — default: steps + key pages/components; no long scripts
   unless user asks for an automation appendix.
4. **Success bar** — confirm what “done” meant (e.g. edit page open vs fully published).

### Output location

| User says | Agent does |
|-----------|------------|
| Explicit path | Use that path (create or update) |
| “whatever” / default / don’t care | `/tmp/BROWSER_AGENT_WORKFLOW_<slug>.md` |
| Nothing about location | **Always ask** before writing |

Do not invent a project path without asking.

### Naming

Mix **user-supplied** and **agent-derived**:

| Situation | Behavior |
|-----------|----------|
| User gives a path, filename, or slug | Prefer that |
| User gives a loose title only | Normalize to kebab-case slug; refine with agent-derived clarity if needed |
| No name + default location | Agent derives short `<slug>` from the goal (kebab-case) |

Examples: goal “apply config change for region/env” → `apply-config-change-region-env`;
user says “use name config-apply-vn” → `config-apply-vn`.

## Create vs update

### Create

Write a new markdown playbook using the [Output template](#output-template).

### Update

1. Read the existing file fully.
2. **Merge** section by section: keep still-valid content; fix outdated steps;
   add missing steps/pages; remove or rewrite wrong guidance.
3. Do **not** wipe good sections for a full rewrite unless the doc is mostly wrong.
4. Prepend or maintain a **Changelog** with brief **dated bullets** (newest first).

Changelog shape:

```markdown
## Changelog

- **YYYY-MM-DD** — Short what changed (added / fixed / removed).
- **YYYY-MM-DD** — …
```

Rules: one short bullet per update session when possible; no essay; no full transcript.

## Distill procedure

1. **Restate the goal** in one sentence (what someone can re-request later).
2. **Separate happy path from dead ends.** Only the happy path is the main steps.
   Failures and false paths go under **Watch-outs** (brief).
3. **Generalize** one-off IDs when useful (ticket numbers → “duty ticket”; keep
   durable patterns like `…_live_{region}`).
4. **Keep concrete** durable UI landmarks: product URL patterns, tab/env names,
   button labels, namespace naming conventions.
5. **Extract a domain model** only if the UI encodes a hierarchy or entities
   that operators must understand (project → module → env → …).
6. **Preconditions / inputs** only if the session actually depended on them
   (login, permissions, connected browser-agent session, required facts before
   acting). Do **not** add a special “CLI prep” chapter or favor any product CLI.
   If a terminal command was load-bearing, it appears as a normal step or
   precondition—not a first-class tooling section.
7. **Success criteria** and **non-goals** (e.g. stop before submit on live).
8. Write the file; stop. Do not re-run the browser ops unless the user asks.

## Output template

Use this outline (omit empty body sections; rename slightly if clearer for the domain).
**Always include the YAML front matter** on create; on update, refresh `updated` and
keep other fields accurate.

```markdown
---
name: <slug>
title: <short title>
description: >
  <one-liner: what request this answers>
created: YYYY-MM-DD
updated: YYYY-MM-DD
source: browser-agent-copilot
status: active
---

# Workflow: <short title>

<one-liner: what request this answers>

## Changelog

- **YYYY-MM-DD** — Initial playbook from co-pilot session.
  (or update bullets when revising)

## Preconditions

- …

## Inputs required

- …

## Domain model (if useful)

| Layer | What it is | Example / pattern |
|-------|------------|-------------------|
| … | … | … |

Mental model: `… → … → …`

## Steps

### A. …
1. …

### B. …
1. …

## Key pages / components

| UI surface | Why it matters |
|------------|----------------|
| … | … |

## Watch-outs

- …

## Success criteria

- …

## Non-goals

- …
```

Front matter fields:

| Field | Required | Notes |
|-------|----------|--------|
| `name` | yes | kebab-case slug (user-supplied and/or agent-derived) |
| `title` | yes | Human-readable short title |
| `description` | yes | One-liner goal; folded `>` ok |
| `created` | yes | `YYYY-MM-DD`; set once on create |
| `updated` | yes | `YYYY-MM-DD`; bump on every update |
| `source` | yes | Default `browser-agent-copilot` |
| `status` | optional | e.g. `active`, `draft`, `deprecated` |


### Style rules

- Prefer **steps + key pages/components** over brittle automation scripts.
- Prefer tables and short checklists.
- Prefer human gates where risk is high (e.g. live publish: open edit/PR page,
  do not submit unless asked).
- Redact secrets, tokens, cookies, and PII.
- Do not dump the raw transcript as the workflow.
- Do not document every failed attempt as if it were the happy path.

## Anti-patterns

| Avoid | Prefer |
|-------|--------|
| Full chat paste | Distilled steps |
| One-off CDP/`eval` blobs as the playbook | Landmarks + ordered steps |
| Silent default path into a product repo | Always ask; `/tmp/…` only on “whatever” |
| Assuming a favorite CLI ecosystem | Generic preconditions only if real |
| Rewriting a good existing doc from scratch | Merge + brief dated changelog |

## Example invocation

User (after a successful co-pilot session):

> `/summarize-the-workflow` — write this up; update
> `path/to/WORKFLOW_….md` if it exists, else create. whatever is fine for name.

Agent:

1. Confirm success bar and create vs update path (or default
   `/tmp/BROWSER_AGENT_WORKFLOW_<slug>.md` if user said whatever for location).
2. Distill happy path + watch-outs from the session.
3. Create new file, or merge into existing + dated changelog bullet.
4. Report the path written and a one-line summary of what the workflow covers.
