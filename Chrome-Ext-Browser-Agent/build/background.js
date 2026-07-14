// Browser Agent MV3 service worker — connects to control plane /v1/ws.
// Executes jobs via chrome.debugger CDP (Runtime.evaluate, Page.captureScreenshot, etc.).
// Load generated identity (version + content md5) for hello match.
try {
  importScripts("bundle-sum.js");
} catch (e) {
  // bundle-sum.js may be written by serve extract if missing from package.
  console.warn("browser-agent: bundle-sum.js not loaded", e);
}
const CONTROL_PORT = 43761;
const WS_URL = "ws://127.0.0.1:" + CONTROL_PORT + "/v1/ws";
const EXT_VERSION =
  typeof BROWSER_AGENT_BUNDLE_VERSION === "string" && BROWSER_AGENT_BUNDLE_VERSION
    ? BROWSER_AGENT_BUNDLE_VERSION
    : "1.0.1";
const EXT_BUNDLE_MD5 =
  typeof BROWSER_AGENT_BUNDLE_MD5 === "string" ? BROWSER_AGENT_BUNDLE_MD5 : "";
const FEATURES = ["browser-agent"];
const CDP_PROTOCOL_VERSION = "1.3";

let socket = null;
let reconnectTimer = null;
let reconnectAttempt = 0;
/** Abort hung CONNECTING sockets (ms). */
const CONNECT_TIMEOUT_MS = 2500;
/** Cap reconnect backoff so we rejoin the control plane quickly after serve restart. */
const RECONNECT_MAX_MS = 2000;
const RECONNECT_BASE_MS = 100;
/** Keepalive ping while connected (keeps MV3 SW + WS alive). */
const KEEPALIVE_MS = 15000;
let connectTimeoutTimer = null;
let keepaliveTimer = null;
/** @type {Map<number, true>} */
const attachedTabs = new Map();
/** In-memory console log buffer for logs jobs. */
const consoleLogBuffer = [];
const MAX_LOG_ENTRIES = 500;

function isSocketLive() {
  return socket && (socket.readyState === WebSocket.OPEN || socket.readyState === WebSocket.CONNECTING);
}

function clearConnectTimeout() {
  if (connectTimeoutTimer) {
    clearTimeout(connectTimeoutTimer);
    connectTimeoutTimer = null;
  }
}

function stopKeepalive() {
  if (keepaliveTimer) {
    clearInterval(keepaliveTimer);
    keepaliveTimer = null;
  }
}

function startKeepalive() {
  stopKeepalive();
  keepaliveTimer = setInterval(() => {
    if (socket && socket.readyState === WebSocket.OPEN) {
      sendJSON({ v: 1, type: "ping", id: "ka-" + Date.now() });
    } else {
      connect("keepalive");
    }
  }, KEEPALIVE_MS);
}

chrome.runtime.onInstalled.addListener(() => {
  console.log("Browser Agent installed; control port", CONTROL_PORT);
  connect("onInstalled");
});

chrome.runtime.onStartup.addListener(() => {
  connect("onStartup");
});

// Alarm-based reconnect backup when service worker restarts (Chrome min period ~1 min).
try {
  chrome.alarms.create("browser-agent-reconnect", { periodInMinutes: 1 });
  chrome.alarms.onAlarm.addListener((alarm) => {
    if (alarm && alarm.name === "browser-agent-reconnect") {
      if (!isSocketLive() || (socket && socket.readyState !== WebSocket.OPEN)) {
        connect("alarm");
      }
    }
  });
} catch (e) {
  // alarms permission optional in mini fixtures
}

/**
 * Open WS to control plane. reason is for service-worker console diagnostics.
 * Aborts hung CONNECTING after CONNECT_TIMEOUT_MS and uses short reconnect backoff.
 */
function connect(reason) {
  if (socket && socket.readyState === WebSocket.OPEN) {
    return;
  }
  // Drop a hung CONNECTING socket so we can retry immediately.
  if (socket && socket.readyState === WebSocket.CONNECTING) {
    // leave existing timeout to abort; avoid parallel sockets
    return;
  }
  if (socket) {
    try {
      socket.close();
    } catch (e) {
      /* ignore */
    }
    socket = null;
  }

  try {
    socket = new WebSocket(WS_URL);
  } catch (e) {
    scheduleReconnect();
    return;
  }

  clearConnectTimeout();
  connectTimeoutTimer = setTimeout(() => {
    connectTimeoutTimer = null;
    if (socket && socket.readyState === WebSocket.CONNECTING) {
      try {
        socket.close();
      } catch (e) {
        /* ignore */
      }
      // onclose will schedule reconnect
    }
  }, CONNECT_TIMEOUT_MS);

  socket.onopen = () => {
    clearConnectTimeout();
    reconnectAttempt = 0;
    sendHello();
    startKeepalive();
  };

  socket.onmessage = (ev) => {
    let msg;
    try {
      msg = JSON.parse(ev.data);
    } catch (e) {
      return;
    }
    handleMessage(msg);
  };

  socket.onclose = () => {
    clearConnectTimeout();
    stopKeepalive();
    socket = null;
    scheduleReconnect();
  };

  socket.onerror = () => {
    try {
      if (socket) socket.close();
    } catch (e) {
      /* ignore */
    }
  };
}

function scheduleReconnect() {
  if (reconnectTimer) return;
  // Fast, bounded backoff: 100, 200, 400, … cap 2s (was up to 30s).
  const delay = Math.min(RECONNECT_MAX_MS, RECONNECT_BASE_MS * Math.pow(2, reconnectAttempt));
  reconnectAttempt += 1;
  reconnectTimer = setTimeout(() => {
    reconnectTimer = null;
    connect("reconnect");
  }, delay);
}

function sendJSON(obj) {
  if (!socket || socket.readyState !== WebSocket.OPEN) return;
  try {
    socket.send(JSON.stringify(obj));
  } catch (e) {
    /* ignore */
  }
}

function sendHello() {
  const payload = {
    version: EXT_VERSION,
    features: FEATURES,
  };
  if (EXT_BUNDLE_MD5) {
    payload.bundle_md5 = EXT_BUNDLE_MD5;
  }
  sendJSON({
    v: 1,
    type: "hello",
    payload: payload,
  });
}

function handleMessage(msg) {
  if (!msg || typeof msg !== "object") return;
  const type = msg.type;
  if (type === "job") {
    handleJob(msg);
  }
}

function sendJobResult(jobId, ok, data, error) {
  sendJSON({
    v: 1,
    type: "result",
    id: jobId,
    payload: {
      job_id: jobId,
      id: jobId,
      ok: !!ok,
      error: error || "",
      data: data || {},
    },
  });
}

async function handleJob(msg) {
  const payload = msg.payload || {};
  const jobId = payload.job_id || payload.id || msg.id || "";
  const jobType = payload.type || payload.job_type || "eval";
  const params = payload.params || {};

  try {
    let data;
    switch (jobType) {
      case "info":
        data = await handleInfoJob(params);
        break;
      case "eval":
        data = await handleEvalJob(params, payload);
        break;
      case "run":
        data = await handleRunJob(params, payload);
        break;
      case "logs":
        data = await handleLogsJob(params);
        break;
      case "screenshot":
        data = await handleScreenshotJob(params);
        break;
      case "cdp":
        data = await handleCdpJob(params);
        break;
      default:
        sendJobResult(jobId, false, { type: jobType }, "unknown job type: " + jobType);
        return;
    }
    sendJobResult(jobId, true, data, "");
  } catch (e) {
    const errMsg = (e && e.message) ? e.message : String(e);
    sendJobResult(jobId, false, { type: jobType, stub: false }, errMsg);
  }
}

async function handleInfoJob(_params) {
  let tabs = [];
  try {
    const all = await chrome.tabs.query({});
    tabs = (all || []).map((t) => ({
      id: t.id,
      url: t.url || "",
      title: t.title || "",
      active: !!t.active,
      windowId: t.windowId,
    }));
  } catch (e) {
    /* tabs permission optional */
  }
  return {
    version: EXT_VERSION,
    features: FEATURES,
    controlPort: CONTROL_PORT,
    tabs: tabs,
  };
}

async function handleEvalJob(params, payload) {
  const expression =
    (params && (params.expression || params.expr || params.code)) ||
    payload.expression ||
    payload.expr ||
    "";
  const result = await withDebugger(async (tabId) => {
    return await sendDebuggerCommand(tabId, "Runtime.evaluate", {
      expression: String(expression),
      returnByValue: true,
      awaitPromise: true,
    });
  });
  const value =
    result && result.result && "value" in result.result
      ? result.result.value
      : result && result.result
        ? result.result.description
        : null;
  return {
    value: value,
    type: "eval",
    cdp: "Runtime.evaluate",
    expression: expression,
    result: result,
  };
}

async function handleRunJob(params, payload) {
  const source =
    (params && (params.source || params.expression || params.expr || params.code || params.script)) ||
    payload.source ||
    payload.expression ||
    "";
  const result = await withDebugger(async (tabId) => {
    return await sendDebuggerCommand(tabId, "Runtime.evaluate", {
      expression: String(source),
      returnByValue: true,
      awaitPromise: true,
    });
  });
  const value =
    result && result.result && "value" in result.result
      ? result.result.value
      : null;
  return {
    value: value,
    type: "run",
    cdp: "Runtime.evaluate",
    source: source,
    result: result,
  };
}

async function handleLogsJob(params) {
  const limit = (params && params.limit) || 100;
  const level = params && params.level;
  let entries = consoleLogBuffer.slice();
  if (level) {
    entries = entries.filter((e) => e.level === level);
  }
  if (limit > 0 && entries.length > limit) {
    entries = entries.slice(entries.length - limit);
  }
  // Best-effort: enable Log domain to capture future entries.
  try {
    await withDebugger(async (tabId) => {
      await sendDebuggerCommand(tabId, "Log.enable", {});
      await sendDebuggerCommand(tabId, "Runtime.enable", {});
      return { enabled: true };
    });
  } catch (e) {
    /* ignore attach failures for logs */
  }
  return { entries: entries, type: "logs" };
}

async function handleScreenshotJob(params) {
  const format = (params && params.format) || "png";
  const fullPage = !!(params && params.full_page);
  const result = await withDebugger(async (tabId) => {
    const cdpParams = { format: format === "jpeg" ? "jpeg" : "png" };
    if (fullPage) {
      // Capture beyond viewport when possible.
      cdpParams.captureBeyondViewport = true;
    }
    return await sendDebuggerCommand(tabId, "Page.captureScreenshot", cdpParams);
  });
  return {
    base64: (result && result.data) || "",
    format: format,
    type: "screenshot",
    cdp: "Page.captureScreenshot",
  };
}

async function handleCdpJob(params) {
  const method = (params && (params.method || params.cdp_method || params.cdpMethod)) || "";
  if (!method) {
    throw new Error("cdp job requires params.method");
  }
  const cdpParams = (params && params.params) || {};
  const result = await withDebugger(async (tabId) => {
    return await sendDebuggerCommand(tabId, method, cdpParams || {});
  });
  return {
    result: result,
    method: method,
    type: "cdp",
  };
}

// --- chrome.debugger attach / sendCommand / detach ---

function isCapturableTabURL(url) {
  if (!url || typeof url !== "string") return false;
  const u = url.trim().toLowerCase();
  if (!u) return false;
  if (
    u.startsWith("chrome://") ||
    u.startsWith("chrome-extension://") ||
    u.startsWith("devtools://") ||
    u.startsWith("edge://") ||
    u.startsWith("about:")
  ) {
    return false;
  }
  return true;
}

async function pickTargetTabId() {
  // Prefer active tab in last focused window.
  try {
    const [active] = await chrome.tabs.query({ active: true, lastFocusedWindow: true });
    if (active && active.id != null && isCapturableTabURL(active.url || "")) {
      return active.id;
    }
  } catch (e) {
    /* ignore */
  }
  try {
    const tabs = await chrome.tabs.query({});
    for (const t of tabs || []) {
      if (t.id != null && isCapturableTabURL(t.url || "")) {
        return t.id;
      }
    }
  } catch (e) {
    /* ignore */
  }
  throw new Error("no capturable tab for chrome.debugger attach");
}

function attachDebugger(tabId) {
  return new Promise((resolve, reject) => {
    if (attachedTabs.has(tabId)) {
      resolve(true);
      return;
    }
    chrome.debugger.attach({ tabId: tabId }, CDP_PROTOCOL_VERSION, () => {
      if (chrome.runtime.lastError) {
        reject(new Error(chrome.runtime.lastError.message || "chrome.debugger.attach failed"));
        return;
      }
      attachedTabs.set(tabId, true);
      // Enable Runtime for console / evaluate convenience.
      chrome.debugger.sendCommand({ tabId: tabId }, "Runtime.enable", {}, () => {
        resolve(true);
      });
    });
  });
}

function detachDebugger(tabId) {
  return new Promise((resolve) => {
    if (!attachedTabs.has(tabId)) {
      resolve();
      return;
    }
    try {
      chrome.debugger.detach({ tabId: tabId }, () => {
        attachedTabs.delete(tabId);
        resolve();
      });
    } catch (e) {
      attachedTabs.delete(tabId);
      resolve();
    }
  });
}

function sendDebuggerCommand(tabId, method, params) {
  return new Promise((resolve, reject) => {
    chrome.debugger.sendCommand({ tabId: tabId }, method, params || {}, (result) => {
      if (chrome.runtime.lastError) {
        reject(new Error(chrome.runtime.lastError.message || "sendCommand failed: " + method));
        return;
      }
      resolve(result || {});
    });
  });
}

async function withDebugger(fn) {
  const tabId = await pickTargetTabId();
  await attachDebugger(tabId);
  try {
    return await fn(tabId);
  } finally {
    // Keep attach for subsequent jobs on the same tab (faster); detach on SW suspend is OK.
    // Detach is available for cleanup if needed:
    // await detachDebugger(tabId);
  }
}

// Buffer Runtime.consoleAPICalled / Log.entryAdded when debugger is attached.
try {
  chrome.debugger.onEvent.addListener((source, method, params) => {
    if (method === "Runtime.consoleAPICalled" && params) {
      const entry = {
        level: params.type || "log",
        text: (params.args || [])
          .map((a) => (a && (a.value != null ? String(a.value) : a.description)) || "")
          .join(" "),
        timestamp: params.timestamp || Date.now(),
      };
      consoleLogBuffer.push(entry);
      if (consoleLogBuffer.length > MAX_LOG_ENTRIES) {
        consoleLogBuffer.splice(0, consoleLogBuffer.length - MAX_LOG_ENTRIES);
      }
    }
    if (method === "Log.entryAdded" && params && params.entry) {
      const e = params.entry;
      consoleLogBuffer.push({
        level: e.level || "log",
        text: e.text || "",
        timestamp: e.timestamp || Date.now(),
      });
      if (consoleLogBuffer.length > MAX_LOG_ENTRIES) {
        consoleLogBuffer.splice(0, consoleLogBuffer.length - MAX_LOG_ENTRIES);
      }
    }
  });
  chrome.debugger.onDetach.addListener((source) => {
    if (source && source.tabId != null) {
      attachedTabs.delete(source.tabId);
    }
  });
} catch (e) {
  /* debugger events optional in test stubs */
}

// Kick off initial connect as soon as the service worker evaluates.
connect("sw-load");
