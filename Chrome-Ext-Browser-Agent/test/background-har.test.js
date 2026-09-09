const test = require("node:test");
const assert = require("node:assert/strict");
const fs = require("node:fs");
const path = require("node:path");
const vm = require("node:vm");

function event() {
  return { addListener() {} };
}

function loadBackground(overrides = {}) {
  const source = fs.readFileSync(path.join(__dirname, "../public/background.js"), "utf8");
  const context = {
    URL,
    AbortController,
    WebSocket: { OPEN: 1, CONNECTING: 0 },
    console: { log() {}, warn() {}, error() {} },
    importScripts() {},
    setTimeout() { return 1; },
    clearTimeout() {},
    setInterval() { return 1; },
    clearInterval() {},
    fetch: overrides.fetch || (async () => ({ ok: true, text: async () => "" })),
    chrome: {
      runtime: {
        lastError: null,
        onMessage: event(),
        onMessageExternal: event(),
        onInstalled: event(),
        onStartup: event(),
      },
      tabs: {
        query: async () => [],
        onRemoved: event(),
        onUpdated: event(),
        onCreated: event(),
      },
      alarms: { create() {}, onAlarm: event() },
      debugger: Object.assign(
        {
          onEvent: event(),
          onDetach: event(),
          getTargets: (cb) => cb([]),
          attach: (_dbg, _proto, cb) => cb && cb(),
          detach: (_dbg, cb) => cb && cb(),
          sendCommand: (_dbg, _method, _params, cb) => cb && cb(),
        },
        overrides.debugger || {},
      ),
      storage: {},
    },
  };
  vm.createContext(context);
  vm.runInContext(
    source +
      ";globalThis.__backgroundTest = {" +
      "isSessionGoPageURL, isSessionGoPageURLAtPort, parseGoSessionFromURL, validateHARUploadURL, handleRegisterMessage, sessions, snapshotHARTab, harTabResult, uploadHARTab, sendJobResult, flushPendingJobResults, handleResultAck, pendingHAREndCaptureIDs, claimJobExecution, finishJobExecution, isHAROwnedTab, idleDebuggerLeftoverTabIDs, harCaptures, attachedTabs, attachDebugger, adoptDebuggerTargetsFromChrome, isAnotherDebuggerAttachedError, anotherDebuggerAttachedHint, detachDebugger" +
      "};",
    context,
  );
  return context.__backgroundTest;
}

test("recognizes only exact loopback session control pages", () => {
  const { isSessionGoPageURL, isSessionGoPageURLAtPort, parseGoSessionFromURL } = loadBackground();

  assert.equal(isSessionGoPageURL("http://127.0.0.1:43761/go?session=s", "s"), true);
  assert.equal(isSessionGoPageURL("https://localhost/go?other=1&session=s", "s"), true);
  assert.equal(isSessionGoPageURL("https://example.test/go?session=s", "s"), false);
  assert.equal(isSessionGoPageURL("http://127.0.0.1:43761/application/go?session=s", "s"), false);
  assert.equal(isSessionGoPageURL("http://127.0.0.1:43761/go/child?session=s", "s"), false);
  assert.equal(isSessionGoPageURLAtPort("http://127.0.0.1:43761/go?session=s", 43761), true);
  assert.equal(isSessionGoPageURLAtPort("http://localhost:3000/go?session=s", 43761), false);
  assert.equal(parseGoSessionFromURL("https://example.test/go?session=s"), null);
});

test("rejects untrusted registration and HAR upload destinations", () => {
  const { handleRegisterMessage, validateHARUploadURL, sessions } = loadBackground();
  assert.equal(
    handleRegisterMessage(
      { session_id: "s", control_port: 43761 },
      { tab: { id: 1, windowId: 2, url: "http://127.0.0.1:43761/not-go?session=s" } },
    ),
    false,
  );
  assert.equal(
    handleRegisterMessage(
      { session_id: "other", control_port: 43761 },
      { tab: { id: 1, windowId: 2, url: "http://127.0.0.1:43761/go?session=s" } },
    ),
    false,
  );
  assert.equal(
    handleRegisterMessage(
      { session_id: "s", tabId: 999, windowId: 999, control_port: 3000 },
      { tab: { id: 7, windowId: 8, url: "http://127.0.0.1:43761/go?session=s" } },
    ),
    true,
  );
  assert.equal(sessions.get("s").tabId, 7);
  assert.equal(sessions.get("s").windowId, 8);
  assert.equal(sessions.get("s").controlPort, 43761);
  assert.match(validateHARUploadURL("http://127.0.0.1:43761/v1/har/artifact", 43761), /^http:/);
  assert.throws(
    () => validateHARUploadURL("https://attacker.test/v1/har/artifact", 43761),
    /session control server/,
  );
  assert.throws(
    () => validateHARUploadURL("http://127.0.0.1:3000/v1/har/artifact", 43761),
    /session control server/,
  );
});

test("queues unsent job results and flushes them after reconnect", () => {
  const { sendJobResult, flushPendingJobResults, handleResultAck, pendingHAREndCaptureIDs, isHAROwnedTab, harCaptures } = loadBackground();
  const entry = { ws: null, pendingResults: new Map() };
  sendJobResult(entry, "job-1", true, { type: "har_end", capture_id: "capture-1" }, "");
  sendJobResult(entry, "job-2", false, { type: "har_end", capture_id: "capture-failed" }, "failed");
  sendJobResult(entry, "job-3", true, { type: "eval", capture_id: "capture-eval" }, "");
  assert.equal(entry.pendingResults.size, 3);
  assert.deepEqual(Array.from(pendingHAREndCaptureIDs(entry)), ["capture-1"]);

  const sent = [];
  entry.ws = { readyState: 1, send: (raw) => sent.push(JSON.parse(raw)) };
  flushPendingJobResults(entry);
  assert.equal(entry.pendingResults.size, 3);
  assert.equal(sent.length, 3);
  assert.equal(sent[0].payload.job_id, "job-1");
  assert.equal(sent[0].payload.data.type, "har_end");

  handleResultAck({ type: "result_ack", payload: { job_id: "job-1" } }, entry);
  assert.equal(entry.pendingResults.size, 2);
  handleResultAck({ type: "result_ack", id: "job-2", payload: {} }, entry);
  handleResultAck({ type: "result_ack", payload: { id: "job-3" } }, entry);
  assert.equal(entry.pendingResults.size, 0);
});

test("treats in-progress enrollment as HAR-owned", () => {
  const { isHAROwnedTab, harCaptures } = loadBackground();
  harCaptures.set("s", { enrolling: new Map([[42, Promise.resolve()]]) });
  assert.equal(isHAROwnedTab(42), true);
  assert.equal(isHAROwnedTab(43), false);
});

test("deduplicates active and completed job deliveries", () => {
  const { claimJobExecution, finishJobExecution } = loadBackground();
  const entry = {
    ws: null,
    pendingResults: new Map(),
    activeJobIds: new Set(),
    completedJobIds: new Set(),
  };
  assert.equal(claimJobExecution(entry, "job-1"), true);
  assert.equal(claimJobExecution(entry, "job-1"), false);
  finishJobExecution(entry, "job-1");
  assert.equal(claimJobExecution(entry, "job-1"), false);
});

test("idle cleanup excludes tabs being enrolled for HAR", () => {
  const { idleDebuggerLeftoverTabIDs, harCaptures, attachedTabs } = loadBackground();
  harCaptures.set("s", { enrolling: new Map([[42, Promise.resolve()]]) });
  attachedTabs.set(42, true);
  attachedTabs.set(43, true);
  assert.deepEqual(Array.from(idleDebuggerLeftoverTabIDs()), [43]);
});

test("marks daemon quota rejection as a permanent upload failure", async () => {
  const { uploadHARTab } = loadBackground({
    fetch: async () => ({ ok: false, status: 413, text: async () => "too large" }),
  });
  const state = {
    tabId: 9,
    snapshot: { log: { version: "1.2", entries: [] } },
    snapshotJSON: '{"log":{"version":"1.2","entries":[]}}',
    uploaded: false,
    uploadPromise: null,
  };
  const capture = {
    uploadURL: "http://127.0.0.1:43761/v1/har/artifact",
    sessionId: "s",
    captureId: "c",
    artifactToken: "t",
  };
  await assert.rejects(
    uploadHARTab(capture, state),
    (error) => error.harPermanentUploadFailure === true && /HTTP 413/.test(error.message),
  );
});

test("uploads retry from the same immutable HAR snapshot used by metadata", async () => {
  const bodies = [];
  let attempts = 0;
  const { snapshotHARTab, harTabResult, uploadHARTab } = loadBackground({
    fetch: async (_url, options) => {
      bodies.push(options.body);
      attempts += 1;
      return { ok: attempts > 1, status: 503, text: async () => "retry" };
    },
  });
  const liveHAR = {
    log: {
      version: "1.2",
      entries: [{ request: { url: "https://example.test/one" } }],
    },
  };
  let builds = 0;
  const state = {
    tabId: 4,
    title: "Example",
    url: "https://example.test/one",
    state: "open",
    recorder: {
      errors: [],
      buildHAR() {
        builds += 1;
        return JSON.parse(JSON.stringify(liveHAR));
      },
    },
    errors: [],
    uploaded: false,
    uploadPromise: null,
    snapshot: null,
    snapshotJSON: "",
    snapshotResult: null,
  };
  const capture = {
    uploadURL: "http://127.0.0.1:43761/v1/har/artifact",
    sessionId: "s",
    captureId: "c",
    artifactToken: "t",
  };

  snapshotHARTab(state);
  liveHAR.log.entries.push({ request: { url: "https://example.test/two" } });
  assert.equal(Object.isFrozen(state.snapshot), true);
  assert.equal(harTabResult(state).request_count, 1);
  assert.equal(JSON.parse(state.snapshotJSON).log.entries.length, 1);

  await assert.rejects(uploadHARTab(capture, state), /HTTP 503 retry/);
  await uploadHARTab(capture, state);

  assert.equal(builds, 1);
  assert.equal(bodies.length, 2);
  assert.equal(bodies[0], bodies[1]);
  assert.equal(state.uploaded, true);
});
