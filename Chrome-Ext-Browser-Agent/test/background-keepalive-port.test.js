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
    _emit(...args) {
      for (const fn of listeners) fn(...args);
    },
    _listeners: listeners,
  };
}

function loadBackground(overrides = {}) {
  const source = fs.readFileSync(
    path.join(__dirname, "../public/background.js"),
    "utf8",
  );
  const onConnect = event();
  const onConnectExternal = event();
  const chromeApi = {
    runtime: {
      lastError: null,
      onMessage: event(),
      onMessageExternal: event(),
      onConnect,
      onConnectExternal,
      onInstalled: event(),
      onStartup: event(),
    },
    tabs: {
      query: (_opts, cb) => {
        const list = overrides.tabs || [];
        if (typeof cb === "function") cb(list);
        return Promise.resolve(list);
      },
      onRemoved: event(),
      onUpdated: event(),
      onCreated: event(),
    },
    alarms: { create() {}, onAlarm: event() },
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
  const sockets = [];
  const context = {
    URL,
    AbortController,
    WebSocket: class {
      static OPEN = 1;
      static CONNECTING = 0;
      constructor(url) {
        this.url = url;
        this.readyState = WebSocket.CONNECTING;
        sockets.push(this);
        setImmediate(() => {
          this.readyState = WebSocket.OPEN;
          if (typeof this.onopen === "function") this.onopen();
        });
      }
      close() {
        this.readyState = 3;
      }
      send() {}
    },
    console: { log() {}, warn() {}, error() {} },
    importScripts() {},
    setTimeout(fn, ms) {
      if (typeof fn !== "function") return 1;
      if (ms && ms >= 100) return 1;
      return setImmediate(fn);
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
      ";globalThis.__backgroundTest = {" +
      "sessions, sessionKeepAlivePorts, handleSessionKeepAlivePort, sessionIdFromPortName" +
      "};",
    context,
  );
  return {
    api: context.__backgroundTest,
    chrome: chromeApi,
    onConnect,
    onConnectExternal,
    sockets,
  };
}

function mockPort({ name, sender, onMessageFns }) {
  const disconnectListeners = [];
  const messageListeners = [];
  const posted = [];
  return {
    name,
    sender,
    posted,
    postMessage(msg) {
      posted.push(msg);
    },
    onMessage: {
      addListener(fn) {
        messageListeners.push(fn);
        if (onMessageFns) onMessageFns.push(fn);
      },
    },
    onDisconnect: {
      addListener(fn) {
        disconnectListeners.push(fn);
      },
    },
    _disconnect() {
      for (const fn of disconnectListeners) fn();
    },
    _message(msg) {
      for (const fn of messageListeners) fn(msg);
    },
  };
}

test("sessionIdFromPortName parses ba-session prefix", () => {
  const { api } = loadBackground();
  assert.equal(api.sessionIdFromPortName("ba-session:sess-abc"), "sess-abc");
  assert.equal(api.sessionIdFromPortName("other"), "");
});

test("onConnect alone does not register; postMessage does", async () => {
  const { api, onConnect, sockets } = loadBackground({
    tabs: [
      {
        id: 42,
        windowId: 7,
        url: "http://127.0.0.1:43761/go?session=sess-port1",
      },
    ],
  });
  const port = mockPort({
    name: "ba-session:sess-port1",
    sender: {
      tab: {
        id: 42,
        windowId: 7,
        url: "http://127.0.0.1:43761/go?session=sess-port1",
      },
    },
  });
  onConnect._emit(port);
  assert.equal(api.sessionKeepAlivePorts.has(port), true);
  assert.equal(api.sessions.has("sess-port1"), false);
  port._message({
    type: "register",
    session_id: "sess-port1",
    control_port: 43761,
  });
  await new Promise((r) => setImmediate(r));
  assert.ok(api.sessions.has("sess-port1"));
  assert.equal(api.sessions.get("sess-port1").tabId, 42);
  assert.ok(port.posted.some((m) => m && m.ok && m.type === "register_ack"));
  await new Promise((r) => setImmediate(r));
  assert.ok(sockets.some((s) => String(s.url).includes("session=sess-port1")));
});

test("Port postMessage register works when name lacks session", async () => {
  const { api, onConnect } = loadBackground({
    tabs: [
      {
        id: 99,
        windowId: 3,
        url: "http://127.0.0.1:43761/go?session=sess-port2",
      },
    ],
  });
  const port = mockPort({
    name: "ba-keepalive",
    sender: {
      tab: {
        id: 99,
        windowId: 3,
        url: "http://127.0.0.1:43761/go?session=sess-port2",
      },
    },
  });
  onConnect._emit(port);
  assert.equal(api.sessions.has("sess-port2"), false);
  port._message({
    type: "register",
    session_id: "sess-port2",
    control_port: 43761,
  });
  await new Promise((r) => setImmediate(r));
  assert.ok(api.sessions.has("sess-port2"));
  assert.ok(port.posted.some((m) => m && m.ok));
});

test("Port disconnect drops keep-alive hold", () => {
  const { api, onConnect } = loadBackground();
  const port = mockPort({
    name: "ba-session:sess-x",
    sender: {
      tab: {
        id: 1,
        windowId: 1,
        url: "http://127.0.0.1:43761/go?session=sess-x",
      },
    },
  });
  onConnect._emit(port);
  assert.equal(api.sessionKeepAlivePorts.has(port), true);
  port._disconnect();
  assert.equal(api.sessionKeepAlivePorts.has(port), false);
});
