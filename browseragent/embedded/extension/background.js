// Browser Agent MV3 service worker — per-session WS to control plane /v1/ws?session=<id>.
// Executes jobs via chrome.debugger CDP (Runtime.evaluate, Page.captureScreenshot, etc.).
// Load generated identity (version + content md5) for hello match.
try {
  importScripts("bundle-sum.js");
} catch (e) {
  // bundle-sum.js may be written by serve extract if missing from package.
  console.warn("browser-agent: bundle-sum.js not loaded", e);
}
const CONTROL_PORT = 43761;
const WS_PATH = "/v1/ws";
const EXT_VERSION =
  typeof BROWSER_AGENT_BUNDLE_VERSION === "string" && BROWSER_AGENT_BUNDLE_VERSION
    ? BROWSER_AGENT_BUNDLE_VERSION
    : "1.0.1";
const EXT_BUNDLE_MD5 =
  typeof BROWSER_AGENT_BUNDLE_MD5 === "string" ? BROWSER_AGENT_BUNDLE_MD5 : "";
const FEATURES = ["browser-agent"];
const CDP_PROTOCOL_VERSION = "1.3";

/** @type {Map<string, {ws: WebSocket|null, tabId: number|null, windowId: number|null, controlPort: number, reconnectTimer: ReturnType<typeof setTimeout>|null, reconnectAttempt: number, connectTimeoutTimer: ReturnType<typeof setTimeout>|null, keepaliveTimer: ReturnType<typeof setInterval>|null}>} */
const sessions = new Map();

/** Abort hung CONNECTING sockets (ms). */
const CONNECT_TIMEOUT_MS = 2500;
/** Cap reconnect backoff so we rejoin the control plane quickly after serve restart. */
const RECONNECT_MAX_MS = 2000;
const RECONNECT_BASE_MS = 100;
/** Keepalive ping while connected (keeps MV3 SW + WS alive). */
const KEEPALIVE_MS = 15000;
/** @type {Map<number, true>} */
const attachedTabs = new Map();
/** In-memory console log buffer for logs jobs. */
const consoleLogBuffer = [];
const MAX_LOG_ENTRIES = 500;

function wsURLForSession(sessionId, controlPort) {
  const port =
    controlPort != null && !Number.isNaN(controlPort) && controlPort > 0
      ? controlPort
      : CONTROL_PORT;
  return (
    "ws://127.0.0.1:" +
    port +
    WS_PATH +
    "?session=" +
    encodeURIComponent(sessionId)
  );
}

function getOrCreateSessionEntry(sessionId) {
  let entry = sessions.get(sessionId);
  if (!entry) {
    entry = {
      ws: null,
      tabId: null,
      windowId: null,
      controlPort: CONTROL_PORT,
      reconnectTimer: null,
      reconnectAttempt: 0,
      connectTimeoutTimer: null,
      keepaliveTimer: null,
    };
    sessions.set(sessionId, entry);
  }
  return entry;
}

function isSocketLive(entry) {
  const socket = entry && entry.ws;
  return socket && (socket.readyState === WebSocket.OPEN || socket.readyState === WebSocket.CONNECTING);
}

function clearConnectTimeout(entry) {
  if (entry.connectTimeoutTimer) {
    clearTimeout(entry.connectTimeoutTimer);
    entry.connectTimeoutTimer = null;
  }
}

function stopKeepalive(entry) {
  if (entry.keepaliveTimer) {
    clearInterval(entry.keepaliveTimer);
    entry.keepaliveTimer = null;
  }
}

function startKeepalive(sessionId, entry) {
  stopKeepalive(entry);
  entry.keepaliveTimer = setInterval(() => {
    if (entry.ws && entry.ws.readyState === WebSocket.OPEN) {
      sendJSON(entry, { v: 1, type: "ping", id: "ka-" + sessionId + "-" + Date.now() });
    } else {
      connectSession(sessionId, "keepalive");
    }
  }, KEEPALIVE_MS);
}

function sendJSON(entry, obj) {
  if (!entry.ws || entry.ws.readyState !== WebSocket.OPEN) return;
  try {
    entry.ws.send(JSON.stringify(obj));
  } catch (e) {
    /* ignore */
  }
}

async function collectSessionPageTelemetry(sessionId) {
  const pages = [];
  try {
    const tabs = await chrome.tabs.query({});
    for (const t of tabs || []) {
      const url = t.url || "";
      if (!isSessionGoPageURL(url, sessionId)) continue;
      pages.push({
        tab_id: t.id,
        url: url,
        title: t.title || "",
      });
    }
  } catch (e) {
    /* tabs permission optional */
  }
  return pages;
}

async function buildSessionTelemetry(sessionId) {
  const pages = await collectSessionPageTelemetry(sessionId);
  return {
    browser_product: "Chrome",
    session_page_count: pages.length,
    session_pages: pages,
  };
}

async function sendHello(sessionId, entry) {
  const telemetry = await buildSessionTelemetry(sessionId);
  const payload = {
    version: EXT_VERSION,
    features: FEATURES,
    browser_product: telemetry.browser_product,
    session_page_count: telemetry.session_page_count,
    session_pages: telemetry.session_pages,
  };
  if (EXT_BUNDLE_MD5) {
    payload.bundle_md5 = EXT_BUNDLE_MD5;
  }
  sendJSON(entry, {
    v: 1,
    type: "hello",
    payload: payload,
  });
}

async function sendSessionStatus(sessionId, entry) {
  if (!entry || !entry.ws || entry.ws.readyState !== WebSocket.OPEN) return;
  const telemetry = await buildSessionTelemetry(sessionId);
  sendJSON(entry, {
    v: 1,
    type: "status",
    id: "status-" + sessionId + "-" + Date.now(),
    payload: {
      browser_product: telemetry.browser_product,
      session_page_count: telemetry.session_page_count,
      session_pages: telemetry.session_pages,
    },
  });
}

function pushStatusForConnectedSessions() {
  for (const [sessionId, entry] of sessions.entries()) {
    if (entry.ws && entry.ws.readyState === WebSocket.OPEN) {
      sendSessionStatus(sessionId, entry).catch(() => {});
    }
  }
}

function scheduleReconnect(sessionId) {
  const entry = sessions.get(sessionId);
  if (!entry || entry.reconnectTimer) return;
  const delay = Math.min(RECONNECT_MAX_MS, RECONNECT_BASE_MS * Math.pow(2, entry.reconnectAttempt));
  entry.reconnectAttempt += 1;
  entry.reconnectTimer = setTimeout(() => {
    entry.reconnectTimer = null;
    connectSession(sessionId, "reconnect");
  }, delay);
}

/**
 * Open per-session WS to control plane. reason is for service-worker console diagnostics.
 */
function connectSession(sessionId, reason) {
  if (!sessionId) return;
  const entry = getOrCreateSessionEntry(sessionId);

  if (entry.ws && entry.ws.readyState === WebSocket.OPEN) {
    return;
  }
  if (entry.ws && entry.ws.readyState === WebSocket.CONNECTING) {
    return;
  }
  if (entry.ws) {
    try {
      entry.ws.close();
    } catch (e) {
      /* ignore */
    }
    entry.ws = null;
  }

  let socket;
  try {
    socket = new WebSocket(wsURLForSession(sessionId, entry.controlPort));
  } catch (e) {
    scheduleReconnect(sessionId);
    return;
  }
  entry.ws = socket;

  clearConnectTimeout(entry);
  entry.connectTimeoutTimer = setTimeout(() => {
    entry.connectTimeoutTimer = null;
    if (entry.ws && entry.ws.readyState === WebSocket.CONNECTING) {
      try {
        entry.ws.close();
      } catch (e) {
        /* ignore */
      }
    }
  }, CONNECT_TIMEOUT_MS);

  socket.onopen = () => {
    clearConnectTimeout(entry);
    entry.reconnectAttempt = 0;
    sendHello(sessionId, entry).catch(() => {});
    startKeepalive(sessionId, entry);
  };

  socket.onmessage = (ev) => {
    let msg;
    try {
      msg = JSON.parse(ev.data);
    } catch (e) {
      return;
    }
    handleMessage(msg, sessionId, entry);
  };

  socket.onclose = () => {
    clearConnectTimeout(entry);
    stopKeepalive(entry);
    entry.ws = null;
    if (sessions.has(sessionId)) {
      scheduleReconnect(sessionId);
    }
  };

  socket.onerror = () => {
    try {
      if (entry.ws) entry.ws.close();
    } catch (e) {
      /* ignore */
    }
  };
}

function unregisterSession(sessionId) {
  const entry = sessions.get(sessionId);
  if (!entry) return;
  if (entry.reconnectTimer) {
    clearTimeout(entry.reconnectTimer);
    entry.reconnectTimer = null;
  }
  clearConnectTimeout(entry);
  stopKeepalive(entry);
  if (entry.ws) {
    try {
      entry.ws.close();
    } catch (e) {
      /* ignore */
    }
    entry.ws = null;
  }
  sessions.delete(sessionId);
}

function isSessionGoPageURL(url, sessionId) {
  if (!url || typeof url !== "string") return false;
  const u = url.trim();
  if (!u.includes("/go")) return false;
  if (!sessionId) return u.includes("/go");
  return (
    u.includes("/go?session=" + sessionId) ||
    u.includes("/go?session=" + encodeURIComponent(sessionId))
  );
}

function parseGoSessionFromURL(url) {
  if (!url || typeof url !== "string") return null;
  try {
    const parsed = new URL(url);
    const host = (parsed.hostname || "").toLowerCase();
    if (host !== "127.0.0.1" && host !== "localhost") return null;
    const path = (parsed.pathname || "").toLowerCase();
    if (!path.includes("/go")) return null;
    const sessionId = parsed.searchParams.get("session");
    if (!sessionId) return null;
    let controlPort = CONTROL_PORT;
    if (parsed.port) {
      const p = parseInt(parsed.port, 10);
      if (!Number.isNaN(p) && p > 0) controlPort = p;
    } else if (parsed.protocol === "https:") {
      controlPort = 443;
    } else if (parsed.protocol === "http:") {
      controlPort = 80;
    }
    return { sessionId: sessionId, controlPort: controlPort };
  } catch (e) {
    return null;
  }
}

function maybeRegisterGoTab(tabId, url, tab) {
  const parsed = parseGoSessionFromURL(url);
  if (!parsed) return;
  handleRegisterMessage(
    {
      type: "register",
      session_id: parsed.sessionId,
      control_port: parsed.controlPort,
      tabId: tabId,
      windowId: tab && tab.windowId,
    },
    { tab: { id: tabId, windowId: tab && tab.windowId } },
  );
}

function handleRegisterMessage(msg, sender) {
  const sessionId = msg.session_id || msg.sessionId || "";
  if (!sessionId) return;
  const entry = getOrCreateSessionEntry(sessionId);
  const tabId = msg.tabId != null ? msg.tabId : sender && sender.tab && sender.tab.id;
  const windowId =
    msg.windowId != null ? msg.windowId : sender && sender.tab && sender.tab.windowId;
  const controlPort =
    msg.control_port != null
      ? msg.control_port
      : msg.controlPort != null
        ? msg.controlPort
        : entry.controlPort;
  if (tabId != null) entry.tabId = tabId;
  if (windowId != null) entry.windowId = windowId;
  if (controlPort != null && !Number.isNaN(controlPort) && controlPort > 0) {
    entry.controlPort = controlPort;
  }
  connectSession(sessionId, "register");
}

chrome.runtime.onMessage.addListener((msg, sender) => {
  if (!msg || typeof msg !== "object") return;
  if (msg.type === "register") {
    handleRegisterMessage(msg, sender);
  }
});

chrome.tabs.onRemoved.addListener((tabId) => {
  let shouldPush = false;
  for (const [sessionId, entry] of sessions.entries()) {
    if (entry.tabId === tabId) {
      unregisterSession(sessionId);
      shouldPush = true;
    }
  }
  if (shouldPush) {
    pushStatusForConnectedSessions();
  }
});

chrome.tabs.onUpdated.addListener((tabId, changeInfo, tab) => {
  const url = (changeInfo && changeInfo.url) || (tab && tab.url) || "";
  if (url) {
    if (changeInfo.status === "loading" || changeInfo.status === "complete") {
      maybeRegisterGoTab(tabId, url, tab);
    }
    for (const [sessionId, entry] of sessions.entries()) {
      if (entry.tabId !== tabId) continue;
      if (!isSessionGoPageURL(url, sessionId)) {
        unregisterSession(sessionId);
      }
    }
    if (url.includes("/go")) {
      pushStatusForConnectedSessions();
    }
  }
});

chrome.runtime.onInstalled.addListener(() => {
  console.log("Browser Agent installed; control port", CONTROL_PORT);
  for (const sessionId of sessions.keys()) {
    connectSession(sessionId, "onInstalled");
  }
});

chrome.runtime.onStartup.addListener(() => {
  for (const sessionId of sessions.keys()) {
    connectSession(sessionId, "onStartup");
  }
});

try {
  chrome.alarms.create("browser-agent-reconnect", { periodInMinutes: 1 });
  chrome.alarms.onAlarm.addListener((alarm) => {
    if (alarm && alarm.name === "browser-agent-reconnect") {
      for (const [sessionId, entry] of sessions.entries()) {
        if (!isSocketLive(entry) || (entry.ws && entry.ws.readyState !== WebSocket.OPEN)) {
          connectSession(sessionId, "alarm");
        }
      }
    }
  });
} catch (e) {
  // alarms permission optional in mini fixtures
}

function handleMessage(msg, sessionId, entry) {
  if (!msg || typeof msg !== "object") return;
  const type = msg.type;
  if (type === "job") {
    handleJob(msg, sessionId, entry);
  }
}

function sendJobResult(entry, jobId, ok, data, error) {
  sendJSON(entry, {
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

async function handleJob(msg, sessionId, entry) {
  const payload = msg.payload || {};
  const jobSessionId = payload.session_id || payload.sessionId || sessionId || "";
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
        data = await handleEvalJob(params, payload, jobSessionId);
        break;
      case "run":
        data = await handleRunJob(params, payload, jobSessionId);
        break;
      case "logs":
        data = await handleLogsJob(params, jobSessionId);
        break;
      case "screenshot":
        data = await handleScreenshotJob(params, jobSessionId);
        break;
      case "cdp":
        data = await handleCdpJob(params, jobSessionId);
        break;
      default:
        sendJobResult(entry, jobId, false, { type: jobType }, "unknown job type: " + jobType);
        return;
    }
    sendJobResult(entry, jobId, true, data, "");
  } catch (e) {
    const errMsg = e && e.message ? e.message : String(e);
    sendJobResult(entry, jobId, false, { type: jobType, stub: false }, errMsg);
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

async function handleEvalJob(params, payload, sessionId) {
  const expression =
    (params && (params.expression || params.expr || params.code)) ||
    payload.expression ||
    payload.expr ||
    "";
  const result = await withDebuggerForSession(sessionId, async (tabId) => {
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

async function handleRunJob(params, payload, sessionId) {
  const source =
    (params && (params.source || params.expression || params.expr || params.code || params.script)) ||
    payload.source ||
    payload.expression ||
    "";
  const result = await withDebuggerForSession(sessionId, async (tabId) => {
    return await sendDebuggerCommand(tabId, "Runtime.evaluate", {
      expression: String(source),
      returnByValue: true,
      awaitPromise: true,
    });
  });
  const value =
    result && result.result && "value" in result.result ? result.result.value : null;
  return {
    value: value,
    type: "run",
    cdp: "Runtime.evaluate",
    source: source,
    result: result,
  };
}

async function handleLogsJob(params, sessionId) {
  const limit = (params && params.limit) || 100;
  const level = params && params.level;
  let entries = consoleLogBuffer.slice();
  if (level) {
    entries = entries.filter((e) => e.level === level);
  }
  if (limit > 0 && entries.length > limit) {
    entries = entries.slice(entries.length - limit);
  }
  try {
    await withDebuggerForSession(sessionId, async (tabId) => {
      await sendDebuggerCommand(tabId, "Log.enable", {});
      await sendDebuggerCommand(tabId, "Runtime.enable", {});
      return { enabled: true };
    });
  } catch (e) {
    /* ignore attach failures for logs */
  }
  return { entries: entries, type: "logs" };
}

async function handleScreenshotJob(params, sessionId) {
  const format = (params && params.format) || "png";
  const fullPage = !!(params && params.full_page);
  const result = await withDebuggerForSession(sessionId, async (tabId) => {
    const cdpParams = { format: format === "jpeg" ? "jpeg" : "png" };
    if (fullPage) {
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

async function handleCdpJob(params, sessionId) {
  const method = (params && (params.method || params.cdp_method || params.cdpMethod)) || "";
  if (!method) {
    throw new Error("cdp job requires params.method");
  }
  const cdpParams = (params && params.params) || {};
  const result = await withDebuggerForSession(sessionId, async (tabId) => {
    return await sendDebuggerCommand(tabId, method, cdpParams || {});
  });
  return {
    result: result,
    method: method,
    type: "cdp",
  };
}

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

async function pickTargetTabIdForSession(sessionId) {
  const entry = sessions.get(sessionId);

  // Prefer the active capturable tab in the session window (multi-session safe).
  if (entry && entry.windowId != null) {
    try {
      const activeTabs = await chrome.tabs.query({ active: true, windowId: entry.windowId });
      const active = activeTabs && activeTabs[0];
      if (active && active.id != null && isCapturableTabURL(active.url || "")) {
        return active.id;
      }
    } catch (e) {
      /* window query optional */
    }
  }

  // Fallback: registered session control page tab.
  if (entry && entry.tabId != null) {
    try {
      const tab = await chrome.tabs.get(entry.tabId);
      if (tab && tab.id != null && isCapturableTabURL(tab.url || "")) {
        return tab.id;
      }
    } catch (e) {
      /* registered tab gone */
    }
  }

  const needles = [
    "/go?session=" + sessionId,
    "/go?session=" + encodeURIComponent(sessionId),
    "go?session=" + sessionId,
    "go?session=" + encodeURIComponent(sessionId),
  ];
  try {
    const tabs = await chrome.tabs.query({});
    for (const t of tabs || []) {
      if (t.id == null) continue;
      const url = t.url || "";
      if (!isCapturableTabURL(url)) continue;
      for (const needle of needles) {
        if (url.includes(needle)) {
          return t.id;
        }
      }
    }
  } catch (e) {
    /* ignore */
  }
  throw new Error("no capturable tab for session " + sessionId);
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
      chrome.debugger.sendCommand({ tabId: tabId }, "Runtime.enable", {}, () => {
        resolve(true);
      });
    });
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

async function withDebuggerForSession(sessionId, fn) {
  const tabId = await pickTargetTabIdForSession(sessionId);
  await attachDebugger(tabId);
  return await fn(tabId);
}

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