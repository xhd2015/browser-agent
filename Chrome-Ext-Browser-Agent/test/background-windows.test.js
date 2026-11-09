const test = require("node:test");
const assert = require("node:assert/strict");
const fs = require("node:fs");
const path = require("node:path");
const vm = require("node:vm");

function event() {
  return { addListener() {} };
}

// Read the manifest so a version bump does not turn every window test red.
const MANIFEST_VERSION = JSON.parse(
  fs.readFileSync(path.join(__dirname, "../public/manifest.json"), "utf8"),
).version;

function loadBackground(windows, options = {}) {
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
    windows: {
      getAll(_query, cb) {
        cb(windows);
      },
      get(_id, _query, cb) {
        cb(windows.find((w) => w && w.id === _id) || null);
      },
      update(_id, _info, cb) {
        if (options.updates) options.updates.push(_info);
        if (options.updateError) {
          chromeApi.runtime.lastError = { message: options.updateError };
          cb();
          chromeApi.runtime.lastError = null;
          return;
        }
        const base = windows.find((w) => w && w.id === _id) || null;
        cb(base ? Object.assign({}, base, _info) : null);
      },
      remove(id, cb) {
        if (options.removeError) {
          chromeApi.runtime.lastError = { message: options.removeError };
          cb();
          chromeApi.runtime.lastError = null;
          return;
        }
        if (options.onRemove) options.onRemove(id);
        cb();
      },
    },
    tabs: Object.assign(
      {
        query: async () => [],
        get: async () => {
          throw new Error("tabs.get not stubbed");
        },
        create(_props, cb) {
          if (cb) cb(null);
        },
        onRemoved: event(),
        onUpdated: event(),
        onCreated: event(),
      },
    ),
    alarms: { create() {}, onAlarm: event() },
    system: options.displays
      ? {
          display: {
            getInfo(cb) {
              cb(options.displays);
            },
          },
        }
      : undefined,
    debugger: {
      onEvent: event(),
      onDetach: event(),
      getTargets: (cb) => cb([]),
      attach: (_dbg, _proto, cb) => cb && cb(),
      detach: (_dbg, cb) => cb && cb(),
      sendCommand: (_dbg, _method, _params, cb) => cb && cb(),
    },
    storage: {},
  };
  const context = {
    URL,
    AbortController,
    performance,
    WebSocket: { OPEN: 1, CONNECTING: 0 },
    console: { log() {}, warn() {}, error() {} },
    importScripts() {},
    setTimeout(fn) {
      if (typeof fn === "function") fn();
      return 1;
    },
    clearTimeout() {},
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
      ";globalThis.__backgroundTest = { handleWindowsJob, handleWindowJob, baClampWindowBounds, FEATURES };",
    context,
  );
  return context.__backgroundTest;
}

function tab(id, index, title, url, active = false) {
  return { id, index, title, url, active, windowId: 1 };
}

function windowWith(id, tabs, extra = {}) {
  return Object.assign(
    { id, type: "normal", focused: false, width: 1200, height: 900, state: "normal", tabs },
    extra,
  );
}

test("handleWindowsJob lists normal windows with tabs sorted by index", async () => {
  const api = loadBackground([
    windowWith(9, [
      tab(92, 1, "Second", "https://example.com/second", true),
      tab(91, 0, "  First\n page  ", "https://example.com/first"),
    ]),
  ]);
  const result = await api.handleWindowsJob();
  assert.equal(result.type, "windows");
  assert.equal(result.version, MANIFEST_VERSION);
  assert.ok(result.features.includes("all-windows"));
  assert.ok(result.features.includes("window-close"));
  assert.equal(result.windows.length, 1);
  const win = result.windows[0];
  assert.equal(win.window_id, 9);
  assert.equal(win.title, "First page");
  assert.equal(win.title_source, "tab");
  assert.equal(win.title_tab_id, 91);
  assert.deepEqual(
    Array.from(win.tabs, (t) => t.tab_id),
    [91, 92],
  );
  assert.equal(win.tabs[0].role, "user");
  assert.equal(win.tabs[0].capturable, true);
});

test("handleWindowsJob skips session pages for the title and records session ids", async () => {
  const api = loadBackground([
    windowWith(10, [
      tab(
        101,
        0,
        "Browser Agent",
        "http://127.0.0.1:43761/go?session=sess-7lm3dk",
      ),
      tab(102, 1, "Confluence — AI Critic", "https://confluence.example/page"),
    ]),
  ]);
  const result = await api.handleWindowsJob();
  const win = result.windows[0];
  assert.equal(win.title, "Confluence — AI Critic");
  assert.equal(win.title_tab_id, 102);
  assert.deepEqual(Array.from(win.session_ids), ["sess-7lm3dk"]);
  assert.equal(win.tabs[0].role, "session");
  assert.equal(win.tabs[1].role, "user");
});

test("handleWindowsJob follows the first page, not the active tab", async () => {
  const api = loadBackground([
    windowWith(11, [
      tab(111, 0, "Docs", "https://docs.example/"),
      tab(112, 1, "Mail", "https://mail.example/", true),
    ]),
  ]);
  const result = await api.handleWindowsJob();
  assert.equal(result.windows[0].title, "Docs");
  assert.equal(result.windows[0].title_tab_id, 111);
});

test("handleWindowsJob falls back from chrome pages to a session id to untitled", async () => {
  let api = loadBackground([
    windowWith(12, [
      tab(121, 0, "New Tab", "chrome://newtab/", true),
      tab(122, 1, "Settings", "chrome://settings/"),
    ]),
    windowWith(13, [tab(132, 0, "Settings", "chrome://settings/")]),
    windowWith(14, [
      tab(
        141,
        0,
        "Session",
        "http://localhost:43761/go?session=sess-only",
      ),
    ]),
    windowWith(15, []),
  ]);
  let result = await api.handleWindowsJob();
  assert.equal(result.windows[0].title, "New Tab");
  assert.equal(result.windows[0].title_tab_id, 121);
  assert.equal(result.windows[1].title, "Settings");
  assert.equal(result.windows[1].title_tab_id, 132);
  assert.equal(result.windows[2].title, "sess-only");
  assert.equal(result.windows[2].title_source, "session");
  assert.equal(result.windows[3].title, "Untitled window");
  assert.equal(result.windows[3].title_source, "none");

  api = loadBackground([
    windowWith(16, [
      tab(161, 0, "", "https://fallback.example/path"),
    ]),
  ]);
  result = await api.handleWindowsJob();
  assert.equal(result.windows[0].title, "fallback.example");
});

test("handleWindowsJob reports every window type and leaves filtering to callers", async () => {
  const api = loadBackground([
    windowWith(17, [tab(171, 0, "Popup", "https://popup.example/")], { type: "popup" }),
    windowWith(18, [tab(181, 0, "Main", "https://main.example/")]),
    windowWith(19, [tab(191, 0, "DevTools", "devtools://devtools/")], {
      type: "devtools",
    }),
  ]);
  const result = await api.handleWindowsJob();
  // A browser is half an operating system: /go and the rail want every window,
  // and each consumer filters by type instead of trusting the wire to hide some.
  assert.deepEqual(
    Array.from(result.windows, (w) => w.window_id),
    [17, 18, 19],
  );
  assert.deepEqual(
    Array.from(result.windows, (w) => w.type),
    ["popup", "normal", "devtools"],
  );
});

test("handleWindowsJob reports window metadata", async () => {
  const api = loadBackground([
    windowWith(19, [tab(191, 0, "Main", "https://main.example/")], {
      focused: true,
      width: 1440,
      height: 900,
      state: "normal",
    }),
  ]);
  const win = (await api.handleWindowsJob()).windows[0];
  assert.equal(win.focused, true);
  assert.equal(win.width, 1440);
  assert.equal(win.height, 900);
  assert.equal(win.state, "normal");
  assert.equal(win.type, "normal");
});

test("handleWindowJob close removes one normal window", async () => {
  const removed = [];
  const api = loadBackground(
    [windowWith(21, [tab(211, 0, "One", "https://one.example")]),
     windowWith(22, [tab(221, 0, "Two", "https://two.example")])],
    { onRemove: (id) => removed.push(id) },
  );
  const result = await api.handleWindowJob({ op: "close", window_id: 22 });
  assert.equal(result.type, "window");
  assert.equal(result.op, "close");
  assert.equal(result.window_id, 22);
  assert.equal(result.closed, true);
  assert.deepEqual(removed, [22]);
});

test("handleWindowJob close refuses the last normal window", async () => {
  const api = loadBackground([windowWith(21, [tab(211, 0, "One", "https://one.example")])]);
  await assert.rejects(
    () => api.handleWindowJob({ op: "close", window_id: 21 }),
    /last Chrome window/,
  );
});

test("handleWindowJob close reports a window that is already gone", async () => {
  const api = loadBackground([
    windowWith(21, []),
    windowWith(22, []),
  ]);
  await assert.rejects(
    () => api.handleWindowJob({ op: "close", window_id: 99 }),
    /already closed/,
  );
});

test("handleWindowJob close surfaces a chrome.windows.remove failure", async () => {
  const api = loadBackground(
    [windowWith(21, []), windowWith(22, [])],
    { removeError: "no permission" },
  );
  await assert.rejects(
    () => api.handleWindowJob({ op: "close", window_id: 22 }),
    /no permission/,
  );
});

// Chrome rejects chrome.windows.update with "Bounds must be at least 50%
// within visible screen space" when the resulting bounds keep less than half
// the window on screen — e.g. ai-critic's Mac preset restoring remembered
// bounds of a window the user later dragged mostly off-screen. The extension
// clamps the target into the hosting display's workarea first: move to fit,
// shrink only when the size itself cannot fit.

const WORKAREAS = [{ workArea: { left: 0, top: 24, width: 1728, height: 1093 } }];

test("clamp moves a window parked off the bottom/right back into the workarea", async () => {
  // Reproduces the live failure: window (488,610) 1426x1021 on a 1728x1117
  // display — 47% visible, so restoring the same bounds was rejected.
  const api = loadBackground(
    [windowWith(30, [tab(301, 0, "One", "https://one.example")], {
      left: 488,
      top: 610,
      width: 1426,
      height: 1021,
    })],
    { displays: WORKAREAS },
  );
  const clamped = await api.baClampWindowBounds(30, {
    left: 488,
    top: 610,
    width: 1426,
    height: 1021,
  });
  // Spread into a local object: values come from the vm realm and carry a
  // foreign Object.prototype that breaks deepStrictEqual reference checks.
  assert.deepEqual({ ...clamped }, { left: 302, top: 96, width: 1426, height: 1021 });
});

test("clamp fills position from the current window for size-only targets", async () => {
  // Resize path sends width/height only: position must come from the current
  // window, then every axis that hangs off-screen is pulled back in.
  const api = loadBackground(
    [windowWith(31, [tab(311, 0, "One", "https://one.example")], {
      left: 488,
      top: 610,
      width: 1426,
      height: 1021,
    })],
    { displays: WORKAREAS },
  );
  const clamped = await api.baClampWindowBounds(31, { width: 1426, height: 1021 });
  assert.deepEqual({ ...clamped }, { left: 302, top: 96, width: 1426, height: 1021 });
});

test("clamp shrinks a size that cannot fit the workarea", async () => {
  const api = loadBackground(
    [windowWith(32, [tab(321, 0, "One", "https://one.example")], { left: 0, top: 0 })],
    { displays: WORKAREAS },
  );
  const clamped = await api.baClampWindowBounds(32, { width: 3000, height: 1500 });
  assert.deepEqual({ ...clamped }, { left: 0, top: 24, width: 1728, height: 1093 });
});

test("clamp uses the display hosting the window", async () => {
  const api = loadBackground(
    [windowWith(33, [tab(331, 0, "One", "https://one.example")], {
      left: 1800,
      top: 100,
      width: 800,
      height: 600,
    })],
    {
      displays: [
        { workArea: { left: 0, top: 24, width: 1728, height: 1093 } },
        { workArea: { left: 1728, top: 0, width: 1728, height: 1200 } },
      ],
    },
  );
  const clamped = await api.baClampWindowBounds(33, {
    left: 3400,
    top: 200,
    width: 800,
    height: 1600,
  });
  assert.deepEqual({ ...clamped }, { left: 2656, top: 0, width: 800, height: 1200 });
});

test("clamp leaves non-geometry updates and no-display environments untouched", async () => {
  const api = loadBackground(
    [windowWith(34, [tab(341, 0, "One", "https://one.example")], { left: 488, top: 610 })],
  );
  assert.deepEqual({ ...(await api.baClampWindowBounds(34, { focused: true })) }, { focused: true });
  // Without chrome.system.display the legacy behavior holds (position filled
  // from the window, no clamping).
  const passthrough = await api.baClampWindowBounds(34, { width: 1426, height: 1021 });
  assert.deepEqual({ ...passthrough }, { left: 488, top: 610, width: 1426, height: 1021 });
});

test("handleWindowJob update clamps stale Mac-preset bounds instead of failing", async () => {
  const updates = [];
  const api = loadBackground(
    [windowWith(30, [tab(301, 0, "One", "https://one.example")], {
      left: 488,
      top: 610,
      width: 1426,
      height: 1021,
    })],
    { displays: WORKAREAS, updates },
  );
  const result = await api.handleWindowJob({
    op: "update",
    window_id: 30,
    left: 488,
    top: 610,
    width: 1426,
    height: 1021,
    state: "normal",
  });
  assert.equal(result.type, "window");
  assert.equal(result.op, "update");
  assert.deepEqual(updates.map((u) => ({ ...u })), [
    { left: 302, top: 96, width: 1426, height: 1021, state: "normal" },
  ]);
  assert.equal(result.window.top, 96);
});

test("handleWindowJob update augments a raw bounds rejection with guidance", async () => {
  const api = loadBackground(
    [windowWith(35, [tab(351, 0, "One", "https://one.example")], { left: 488, top: 610 })],
    {
      updateError:
        "Invalid value for bounds. Bounds must be at least 50% within visible screen space.",
    },
  );
  await assert.rejects(
    () => api.handleWindowJob({ op: "update", window_id: 35, width: 800, height: 600 }),
    /drag it back into view/,
  );
});
