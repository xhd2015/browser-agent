const test = require("node:test");
const assert = require("node:assert/strict");
const fs = require("node:fs");
const path = require("node:path");
const vm = require("node:vm");

function event() {
  return { addListener() {}, removeListener() {} };
}

function loadBackground(overrides = {}) {
  const source = fs.readFileSync(path.join(__dirname, "../public/background.js"), "utf8");
  const realSetTimeout = setTimeout;
  const realClearTimeout = clearTimeout;
  const chromeApi = {
    runtime: {
      lastError: null,
      onMessage: event(),
      onMessageExternal: event(),
      onConnect: event(),
      onConnectExternal: event(),
      onInstalled: event(),
      onStartup: event(),
    },
    tabs: Object.assign(
      {
        query: async () => [],
        get: async () => {
          throw new Error("tabs.get not stubbed");
        },
        onRemoved: event(),
        onUpdated: event(),
        onCreated: event(),
      },
      overrides.tabs || {},
    ),
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
  };
  const context = {
    URL,
    AbortController,
    performance,
    WebSocket: { OPEN: 1, CONNECTING: 0 },
    console: { log() {}, warn() {}, error() {} },
    importScripts() {},
    setTimeout(fn, ms) {
      if (typeof fn !== "function") return 1;
      if (ms && ms >= 100) return 1;
      return realSetTimeout(fn, ms);
    },
    clearTimeout(id) {
      if (id && id !== 1) realClearTimeout(id);
    },
    setInterval() {
      return 1;
    },
    clearInterval() {},
    fetch: async () => ({ ok: true, text: async () => "" }),
    chrome: chromeApi,
  };
  vm.createContext(context);
  vm.runInContext(
    source +
      ";globalThis.__backgroundTest = {" +
      "attachDebugger, attachedTabs, isDebuggerNotAttachedError" +
      "};",
    context,
  );
  return { api: context.__backgroundTest, chrome: chromeApi };
}

test("attachDebugger reattaches when getTargets claims a stale target", async () => {
  let realAttached = false;
  let attaches = 0;
  let detaches = 0;
  let enables = 0;
  const { api, chrome } = loadBackground({
    debugger: {
      getTargets: (cb) => cb([{ tabId: 11, attached: true }]),
      attach: (_dbg, _proto, cb) => {
        attaches += 1;
        realAttached = true;
        chrome.runtime.lastError = null;
        cb();
      },
      detach: (_dbg, cb) => {
        detaches += 1;
        realAttached = false;
        chrome.runtime.lastError = null;
        cb();
      },
      sendCommand: (_dbg, method, _params, cb) => {
        if (method === "Runtime.enable") enables += 1;
        chrome.runtime.lastError = realAttached
          ? null
          : { message: "Debugger is not attached to the tab with id: 11." };
        cb();
      },
    },
  });

  assert.equal(await api.attachDebugger(11), true);
  assert.equal(attaches, 1, "stale adoption must fall through to a real attach");
  assert.equal(detaches, 1, "stale debugger bookkeeping must be released first");
  assert.ok(enables >= 2, "attach path verifies Runtime.enable");
  assert.equal(api.attachedTabs.get(11), "verified");
});

test("attachDebugger reuses a verified target without touching Chrome", async () => {
  let getTargets = 0;
  let attaches = 0;
  const { api } = loadBackground({
    debugger: {
      getTargets: (cb) => {
        getTargets += 1;
        cb([]);
      },
      attach: (_dbg, _proto, cb) => {
        attaches += 1;
        cb();
      },
    },
  });
  api.attachedTabs.set(11, "verified");

  assert.equal(await api.attachDebugger(11), true);
  assert.equal(getTargets, 0);
  assert.equal(attaches, 0);
});

test("attachDebugger keeps a fresh attach when Runtime.enable has a non-fatal error", async () => {
  let attaches = 0;
  const { api, chrome } = loadBackground({
    debugger: {
      attach: (_dbg, _proto, cb) => {
        attaches += 1;
        chrome.runtime.lastError = null;
        cb();
      },
      sendCommand: (_dbg, _method, _params, cb) => {
        chrome.runtime.lastError = { message: "Some transient enable glitch" };
        cb();
      },
    },
  });

  assert.equal(await api.attachDebugger(11), true);
  assert.equal(attaches, 1);
  assert.equal(api.attachedTabs.get(11), "verified");
});

test("isDebuggerNotAttachedError only matches the Chrome detach error", () => {
  const { api } = loadBackground();
  assert.equal(
    api.isDebuggerNotAttachedError(
      new Error("Debugger is not attached to the tab with id: 598526369."),
    ),
    true,
  );
  assert.equal(api.isDebuggerNotAttachedError(new Error("Another debugger is already attached")), false);
  assert.equal(api.isDebuggerNotAttachedError(null), false);
});
