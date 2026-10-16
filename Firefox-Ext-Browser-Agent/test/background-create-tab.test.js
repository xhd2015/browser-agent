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
  const realSetTimeout = setTimeout;
  const realClearTimeout = clearTimeout;
  const browserApi = {
    runtime: {
      onMessage: event(),
      onInstalled: event(),
      onStartup: event(),
      getURL: (p) => p,
    },
    tabs: Object.assign(
      {
        query: async () => [],
        get: async () => {
          throw new Error("tabs.get not stubbed");
        },
        create: async () => null,
        onRemoved: event(),
        onUpdated: event(),
      },
      overrides.tabs || {},
    ),
    alarms: { create() {}, onAlarm: event() },
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
    browser: browserApi,
  };
  vm.createContext(context);
  vm.runInContext(
    source +
      ";globalThis.__backgroundTest = {" +
      "createTabInSession, waitForCapturableTabURL, isCapturableTabURL, sessions, pickTargetTabIdForSession" +
      "};",
    context,
  );
  return context.__backgroundTest;
}

function arm(api) {
  api.sessions.set("sess-1", { windowId: 5 });
}

test("createTabInSession waits until the committed URL is capturable", async () => {
  let gets = 0;
  const api = loadBackground({
    tabs: {
      async create(props) {
        assert.equal(props.url, "https://casement.scredit.io");
        assert.equal(props.windowId, 5);
        return {
          id: 216808382,
          windowId: 5,
          url: "",
          pendingUrl: "https://casement.scredit.io",
          status: "loading",
        };
      },
      async get(tabId) {
        gets += 1;
        if (gets === 1) {
          return {
            id: tabId,
            windowId: 5,
            url: "",
            pendingUrl: "https://casement.scredit.io",
          };
        }
        return { id: tabId, windowId: 5, url: "https://casement.scredit.io/" };
      },
    },
  });
  arm(api);
  const started = Date.now();
  const result = await api.createTabInSession("sess-1", {
    url: "https://casement.scredit.io",
  });
  const elapsed = Date.now() - started;
  assert.equal(result.type, "create_tab");
  assert.equal(result.tab_id, 216808382);
  assert.equal(result.url, "https://casement.scredit.io/");
  assert.equal(result.window_id, 5);
  assert.equal(gets, 2);
  assert.ok(elapsed >= 40, "expected one capturable poll");
  assert.ok(elapsed < 2000, "did not wait the full capturable timeout");
});

test("waitForCapturableTabURL times out on about:blank without sleeping", async () => {
  const api = loadBackground({
    tabs: {
      async get() {
        return { id: 7, url: "about:blank" };
      },
    },
  });
  const started = Date.now();
  await assert.rejects(() => api.waitForCapturableTabURL(7, 0), (err) => {
    assert.match(err.message, /create-tab timed out waiting for a capturable URL/);
    assert.match(err.message, /tab 7, last url about:blank/);
    return true;
  });
  assert.ok(Date.now() - started < 200);
});

test("waitForCapturableTabURL times out on an empty URL", async () => {
  const api = loadBackground({
    tabs: {
      async get() {
        return { id: 8, url: "" };
      },
    },
  });
  await assert.rejects(
    () => api.waitForCapturableTabURL(8, 0),
    /create-tab timed out waiting for a capturable URL \(tab 8, last url \(empty\)\)/,
  );
});

test("createTabInSession rejects a stable non-capturable URL immediately", async () => {
  let gets = 0;
  const api = loadBackground({
    tabs: {
      async create() {
        return { id: 9, windowId: 5, url: "moz-extension://abc/page.html" };
      },
      async get() {
        gets += 1;
        return { id: 9, windowId: 5, url: "moz-extension://abc/page.html" };
      },
    },
  });
  arm(api);
  const started = Date.now();
  await assert.rejects(
    () => api.createTabInSession("sess-1", { url: "https://casement.scredit.io" }),
    (err) => {
      assert.match(err.message, /^create-tab URL is not capturable/);
      assert.match(err.message, /tab 9/);
      assert.match(err.message, /last url moz-extension:\/\/abc\/page\.html/);
      return true;
    },
  );
  assert.equal(gets, 1);
  assert.ok(Date.now() - started < 200);
});

test("createTabInSession returns a capturable URL from tabs.create without waiting", async () => {
  let gets = 0;
  const api = loadBackground({
    tabs: {
      async create() {
        return { id: 10, windowId: 5, url: "https://casement.scredit.io/" };
      },
      async get() {
        gets += 1;
        return { id: 10, windowId: 5, url: "" };
      },
    },
  });
  arm(api);
  const started = Date.now();
  const result = await api.createTabInSession("sess-1", {
    url: "https://casement.scredit.io",
  });
  assert.equal(result.tab_id, 10);
  assert.equal(result.url, "https://casement.scredit.io/");
  assert.equal(gets, 0);
  assert.ok(Date.now() - started < 200);
});

test("createTabInSession does not wait when no URL is requested", async () => {
  let gets = 0;
  const api = loadBackground({
    tabs: {
      async create(props) {
        assert.equal(props.url, undefined);
        return { id: 11, windowId: 5, url: "" };
      },
      get() {
        gets += 1;
        return new Promise(() => {});
      },
    },
  });
  arm(api);
  const started = Date.now();
  const result = await api.createTabInSession("sess-1", {});
  assert.equal(result.tab_id, 11);
  assert.equal(result.url, "");
  assert.equal(gets, 0);
  assert.ok(Date.now() - started < 200);
});

test("pickTargetTabIdForSession still rejects an explicit tab whose URL is not committed", async () => {
  for (const url of ["", "about:blank"]) {
    const api = loadBackground({
      tabs: {
        async get(tabId) {
          return { id: tabId, windowId: 5, url: url };
        },
      },
    });
    arm(api);
    await assert.rejects(
      () => api.pickTargetTabIdForSession("sess-1", { tabId: 12 }),
      /tab_id 12 is not capturable/,
    );
  }
});
