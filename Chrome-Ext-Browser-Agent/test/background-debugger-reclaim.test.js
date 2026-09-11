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
  };
  const context = {
    URL,
    AbortController,
    WebSocket: { OPEN: 1, CONNECTING: 0 },
    console: { log() {}, warn() {}, error() {} },
    importScripts() {},
    // Run near-immediate timers; swallow long heal/attach-timeout timers so tests exit.
    setTimeout(fn, ms) {
      if (typeof fn !== "function") return 1;
      if (ms && ms >= 100) return 1;
      return setImmediate(fn);
    },
    clearTimeout() {},
    setInterval() { return 1; },
    clearInterval() {},
    fetch: async () => ({ ok: true, text: async () => "" }),
    chrome: chromeApi,
  };
  vm.createContext(context);
  vm.runInContext(
    source +
      ";globalThis.__backgroundTest = {" +
      "attachedTabs, attachDebugger, adoptDebuggerTargetsFromChrome, isAnotherDebuggerAttachedError, anotherDebuggerAttachedHint, detachDebugger" +
      "};",
    context,
  );
  return {
    api: context.__backgroundTest,
    chrome: chromeApi,
  };
}

test("isAnotherDebuggerAttachedError matches Chrome wording", () => {
  const { api } = loadBackground();
  assert.equal(
    api.isAnotherDebuggerAttachedError(
      new Error("Another debugger is already attached to the tab with id: 123."),
    ),
    true,
  );
  assert.equal(api.isAnotherDebuggerAttachedError(new Error("chrome.debugger.attach failed")), false);
  assert.match(api.anotherDebuggerAttachedHint(99), /Close DevTools/);
  assert.match(api.anotherDebuggerAttachedHint(99), /reload the Browser Agent extension/);
});

test("adoptDebuggerTargetsFromChrome restores orphaned attaches into attachedTabs", async () => {
  const { api } = loadBackground({
    debugger: {
      getTargets: (cb) =>
        cb([
          { tabId: 42, attached: true },
          { tabId: 43, attached: false },
          { attached: true },
        ]),
    },
  });
  assert.equal(api.attachedTabs.has(42), false);
  const adopted = await api.adoptDebuggerTargetsFromChrome("test");
  assert.equal(adopted, 1);
  assert.equal(api.attachedTabs.has(42), true);
  assert.equal(api.attachedTabs.has(43), false);
});

test("attachDebugger reclaims via getTargets after Another debugger error", async () => {
  let attachCalls = 0;
  let chromeRef;
  const loaded = loadBackground({
    debugger: {
      getTargets: (cb) => {
        // First pre-attach: empty (simulates map loss). After failed attach, reclaim sees orphan.
        if (attachCalls === 0) cb([]);
        else cb([{ tabId: 7, attached: true }]);
      },
      attach: (_dbg, _proto, cb) => {
        attachCalls += 1;
        // Simulate Chrome still holding the prior SW attach.
        chromeRef.runtime.lastError = {
          message: "Another debugger is already attached to the tab with id: 7.",
        };
        cb();
        chromeRef.runtime.lastError = null;
      },
      sendCommand: (_dbg, _method, _params, cb) => cb && cb(),
      detach: (_dbg, cb) => cb && cb(),
    },
  });
  chromeRef = loaded.chrome;

  await loaded.api.attachDebugger(7);
  assert.equal(loaded.api.attachedTabs.has(7), true);
  assert.equal(attachCalls, 1);
});

test("attachDebugger detach+retry succeeds when getTargets does not list the tab", async () => {
  let attachCalls = 0;
  let detachCalls = 0;
  let chromeRef;
  const loaded = loadBackground({
    debugger: {
      getTargets: (cb) => cb([]),
      attach: (_dbg, _proto, cb) => {
        attachCalls += 1;
        if (attachCalls === 1) {
          chromeRef.runtime.lastError = {
            message: "Another debugger is already attached to the tab with id: 8.",
          };
          cb();
          chromeRef.runtime.lastError = null;
          return;
        }
        chromeRef.runtime.lastError = null;
        cb();
      },
      detach: (_dbg, cb) => {
        detachCalls += 1;
        cb && cb();
      },
      sendCommand: (_dbg, _method, _params, cb) => cb && cb(),
    },
  });
  chromeRef = loaded.chrome;

  await loaded.api.attachDebugger(8);
  assert.equal(loaded.api.attachedTabs.has(8), true);
  assert.equal(attachCalls, 2);
  assert.ok(detachCalls >= 1);
});

test("attachDebugger surfaces actionable hint when reclaim cannot take the tab", async () => {
  let chromeRef;
  const loaded = loadBackground({
    debugger: {
      getTargets: (cb) => cb([]),
      attach: (_dbg, _proto, cb) => {
        chromeRef.runtime.lastError = {
          message: "Another debugger is already attached to the tab with id: 9.",
        };
        cb();
        chromeRef.runtime.lastError = null;
      },
      detach: (_dbg, cb) => cb && cb(),
      sendCommand: (_dbg, _method, _params, cb) => cb && cb(),
    },
  });
  chromeRef = loaded.chrome;

  await assert.rejects(() => loaded.api.attachDebugger(9), (err) => {
    assert.match(String(err && err.message), /Close DevTools/);
    assert.match(String(err && err.message), /reload the Browser Agent extension/);
    return true;
  });
});
