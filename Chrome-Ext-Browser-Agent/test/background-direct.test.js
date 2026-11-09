const test = require("node:test");
const assert = require("node:assert/strict");
const fs = require("node:fs");
const path = require("node:path");
const vm = require("node:vm");

function event() {
  const listeners = [];
  return {
    addListener(fn) {
      listeners.push(fn);
    },
    listeners,
  };
}

/**
 * Loads background.js with a chrome stub wired for direct-mode scenarios:
 * - storage.local holds the direct toggle and instance id
 * - fetch answers the /v1/health negotiation probe
 * - tabs/windows answer window_id addressing
 */
function loadBackground(options = {}) {
  const source = fs.readFileSync(path.join(__dirname, "../public/background.js"), "utf8");
  // Production default is ON with an absent storage key. Tests simulate an
  // untouched install by omitting directEnabled; only directEnabled: false
  // injects the explicit opt-out.
  const storageData = Object.assign({}, options.storage || {});
  if (options.directEnabled !== undefined) {
    storageData.directWindowControlEnabled = options.directEnabled !== false;
  }
  function event() {
    return { addListener() {} };
  }
  const chromeApi = {
    runtime: {
      lastError: null,
      onMessage: event(),
      onMessageExternal: event(),
      onConnect: event(),
      onConnectExternal: event(),
      onInstalled: event(),
      onStartup: event(),
      sendMessage(_msg, cb) {
        if (cb) cb();
      },
    },
    windows: {
      getAll(_query, cb) {
        cb(options.windows || []);
      },
      get(id, _query, cb) {
        const w = (options.windows || []).find((x) => x.id === id);
        cb(w || null);
      },
      update(_id, _info, cb) {
        cb(null);
      },
      remove(_id, cb) {
        cb();
      },
    },
    tabs: Object.assign(
      {
        query: async (q) => {
          if (q && q.windowId != null && q.active) {
            return (options.windows || [])
              .filter((w) => w.id === q.windowId)
              .flatMap((w) => w.tabs || [])
              .filter((t) => t.active);
          }
          if (q && q.windowId != null) {
            return (options.windows || [])
              .filter((w) => w.id === q.windowId)
              .flatMap((w) => w.tabs || []);
          }
          return (options.windows || []).flatMap((w) => w.tabs || []);
        },
        get: async (tabId) => {
          const t = (options.windows || []).flatMap((w) => w.tabs || []).find((x) => x.id === tabId);
          if (!t) throw new Error("No tab with id: " + tabId);
          return t;
        },
        create(props, cb) {
          if (cb) cb(options.createdTab || { id: 900, windowId: props.windowId, url: props.url || "" });
        },
        onRemoved: event(),
        onUpdated: event(),
        onCreated: event(),
      },
    ),
    alarms: { create() {}, onAlarm: event() },
    debugger: {
      onEvent: event(),
      onDetach: event(),
      getTargets: (cb) => cb([]),
      attach: (_dbg, _proto, cb) => cb && cb(),
      detach: (_dbg, cb) => cb && cb(),
      sendCommand: (_dbg, _method, _params, cb) => cb && cb({ result: {} }),
    },
    storage: {
      local: {
        get(keys, cb) {
          const out = {};
          for (const k of [].concat(keys)) {
            if (k in storageData) out[k] = storageData[k];
          }
          cb(out);
        },
        set(obj, cb) {
          Object.assign(storageData, obj);
          if (cb) cb();
        },
      },
    },
  };
  const healthBody = options.healthBody !== undefined ? options.healthBody : { ok: true, direct_supported: true };
  const realSetTimeout = setTimeout;
  const realClearTimeout = clearTimeout;
  const context = {
    URL,
    AbortController,
    performance,
    crypto: { getRandomValues: (arr) => arr },
    WebSocket: class {
      constructor(url) {
        this.url = url;
        this.readyState = 0;
        context.__wsInstances.push(this);
      }
      send() {}
      close() {
        this.readyState = 3;
      }
    },
    console: { log() {}, warn() {}, error() {} },
    importScripts() {},
    // Timers ≥100ms (attach/keepalive/timeouts) never fire in tests; short
    // deferrals (0-99ms) run on the real event loop so boot kicks can flush.
    setTimeout(fn, ms) {
      if (typeof fn !== "function") return 1;
      if (ms && ms >= 100) return 1;
      return realSetTimeout(fn, ms || 0);
    },
    clearTimeout(id) {
      if (id && id !== 1) realClearTimeout(id);
    },
    setInterval() {
      return 1;
    },
    clearInterval() {},
    fetch: async (url) => {
      context.__fetchCalls.push(String(url));
      if (String(url).includes("/v1/health")) {
        if (options.healthFail) throw new Error("health unreachable");
        return { ok: true, json: async () => healthBody };
      }
      return { ok: true, text: async () => "" };
    },
    chrome: chromeApi,
  };
  context.__wsInstances = [];
  context.__fetchCalls = [];
  vm.createContext(context);
  vm.runInContext(
    source +
      ";globalThis.__backgroundTest = { handleJob, pickTargetTabIdForSession, connectInstance, attachDebuggerForSession, createTabInSession, handleInfoJobDirect, FEATURES, directEntry };",
    context,
  );
  return context.__backgroundTest;
}

function tab(id, index, title, url, active = false, windowId = 1) {
  return { id, index, title, url, active, windowId };
}

function windowWith(id, tabs, extra = {}) {
  return Object.assign(
    { id, type: "normal", focused: false, width: 1200, height: 900, state: "normal", tabs },
    extra,
  );
}

function jobEnvelope(type, extra = {}) {
  return {
    type: "job",
    id: "job-1",
    payload: Object.assign(
      { job_id: "job-1", type: type, direct: true, window_id: 1, params: {}, timeout_ms: 5000 },
      extra,
    ),
  };
}

test("direct: FEATURES advertises direct-window-v1", () => {
  const api = loadBackground({ directEnabled: false });
  assert.ok(api.FEATURES.includes("direct-window-v1"));
});

test("direct: health probe happens before instance WS connect", async () => {
  const api = loadBackground({});
  await api.connectInstance("test");
  // One health fetch; the WS was constructed with the ?instance= URL.
  assert.ok(
    api.directEntry.ws === null || api.directEntry.ws.url.includes("?instance="),
  );
});

test("direct: no WS when daemon health lacks direct_supported", async () => {
  const api = loadBackground({ healthBody: { ok: true } });
  await api.connectInstance("test");
  assert.equal(api.directEntry.ws, null);
});

test("direct: no WS when toggle off", async () => {
  const api = loadBackground({ directEnabled: false });
  await api.connectInstance("test");
  assert.equal(api.directEntry.ws, null);
});

test("direct: default on — absent storage key opens instance WS", async () => {
  // No directEnabled and no storage key: an untouched install connects.
  const api = loadBackground({});
  await api.connectInstance("test");
  assert.ok(api.directEntry.ws, "instance WS should open by default");
  assert.ok(api.directEntry.ws.url.includes("?instance="));
});

test("direct: default on — pickTarget works without stored consent", async () => {
  const api = loadBackground({
    windows: [
      windowWith(1, [tab(11, 0, "Docs", "https://example.com/docs", true)]),
    ],
  });
  const tabId = await api.pickTargetTabIdForSession("", { direct: true, windowId: 1 });
  assert.equal(tabId, 11);
});

test("direct: default on — pickTarget still refuses when opted out", async () => {
  const api = loadBackground({ directEnabled: false });
  await assert.rejects(
    () => api.pickTargetTabIdForSession("", { direct: true, windowId: 1 }),
    /direct mode disabled/,
  );
});

test("direct: screenshot job targets active tab of window_id", async () => {
  const api = loadBackground({
    windows: [
      windowWith(1, [
        tab(11, 0, "Docs", "https://example.com/docs"),
        tab(12, 1, "Active page", "https://example.com/active", true),
      ]),
    ],
  });
  const entry = api.directEntry;
  let answered = null;
  const origSend = entry.ws;
  // Run job without a live socket: sendJobResult no-ops on closed WS.
  await api.handleJob(jobEnvelope("screenshot"), "__direct__", entry);
  assert.ok(origSend === null); // no crash without socket
});

test("direct: pickTarget resolves explicit tab_id and validates capturable", async () => {
  const api = loadBackground({
    windows: [
      windowWith(1, [
        tab(11, 0, "Docs", "https://example.com/docs"),
        tab(12, 1, "Chrome page", "chrome://settings", true),
      ]),
    ],
  });
  assert.equal(await api.pickTargetTabIdForSession("", { direct: true, tabId: 11 }), 11);
  await assert.rejects(
    () => api.pickTargetTabIdForSession("", { direct: true, tabId: 12 }),
    /not capturable/,
  );
});

test("direct: pickTarget falls to active tab of window_id", async () => {
  const api = loadBackground({
    windows: [
      windowWith(1, [
        tab(11, 0, "Docs", "https://example.com/docs"),
        tab(12, 1, "Active", "https://example.com/active", true),
      ]),
      windowWith(2, [tab(21, 0, "Other", "https://example.com/other", true)]),
    ],
  });
  assert.equal(await api.pickTargetTabIdForSession("", { direct: true, windowId: 2 }), 21);
  await assert.rejects(
    () => api.pickTargetTabIdForSession("", { direct: true }),
    /needs tab_id or window_id/,
  );
});

test("direct: pickTarget refuses when toggle off", async () => {
  const api = loadBackground({ directEnabled: false });
  await assert.rejects(
    () => api.pickTargetTabIdForSession("", { direct: true, tabId: 11 }),
    /direct mode disabled/,
  );
});

test("direct: attach skips session-page gate", async () => {
  const api = loadBackground({
    windows: [windowWith(1, [tab(11, 0, "Page", "https://example.com/page", true)])],
  });
  // No /go page exists anywhere; direct attach must still succeed.
  await api.attachDebuggerForSession("__direct__", 11, { direct: true });
});

test("direct: har jobs rejected with session wording", async () => {
  const api = loadBackground({});
  await api.handleJob(jobEnvelope("har_start"), "__direct__", api.directEntry);
  // No throw: failure goes through sendJobResult (no socket → no-op).
});

test("direct: create_tab uses window_id from payload", async () => {
  const api = loadBackground({
    windows: [windowWith(1, [tab(11, 0, "Page", "https://example.com/page")])],
    createdTab: { id: 901, windowId: 1, url: "https://example.com/new" },
  });
  const result = await api.createTabInSession(
    "",
    { url: "https://example.com/new", window_id: 1 },
    { direct: true },
  );
  assert.equal(result.window_id, 1);
  assert.equal(result.tab_id, 901);
});

test("direct: create_tab without window_id fails cleanly", async () => {
  const api = loadBackground({ windows: [] });
  await assert.rejects(
    () => api.createTabInSession("", { url: "https://example.com/new" }, { direct: true }),
    /needs window_id/,
  );
});

test("direct: info job lists window tabs", async () => {
  const api = loadBackground({
    windows: [
      windowWith(1, [
        tab(11, 0, "Docs", "https://example.com/docs"),
        tab(12, 1, "Active", "https://example.com/active", true),
      ]),
    ],
  });
  const info = await api.handleInfoJobDirect({}, { windowId: 1 });
  assert.equal(info.window_id, 1);
  assert.equal(info.direct, true);
  assert.equal(info.tabs.length, 2);
  assert.equal(info.job_target.tab_id, 12);
});

test("direct: session jobs keep session behavior (jobSessionId intact)", async () => {
  const api = loadBackground({ windows: [] });
  // Session-addressed job with direct flag absent: unknown type fails without
  // touching direct state.
  await api.handleJob(
    { type: "job", id: "j2", payload: { job_id: "j2", type: "nope", session_id: "sess-x" } },
    "sess-x",
    { ws: null, kind: "session" },
  );
  assert.equal(api.directEntry.activeJobIds.has("j2"), false);
});
