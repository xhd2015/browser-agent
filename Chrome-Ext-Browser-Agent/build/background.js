// Browser Agent MV3 service worker — per-session WS to control plane /v1/ws?session=<id>.
// Executes jobs via chrome.debugger CDP (Runtime.evaluate, Page.captureScreenshot, etc.).
// Load generated identity (version + content md5) for hello match.
try {
  importScripts("bundle-sum.js");
} catch (e) {
  // bundle-sum.js may be written by serve extract if missing from package.
  console.warn("browser-agent: bundle-sum.js not loaded", e);
}
try {
  importScripts("har-recorder.js");
} catch (e) {
  console.error("browser-agent: har-recorder.js not loaded", e);
}
const CONTROL_PORT = 43761;
const WS_PATH = "/v1/ws";
const EXT_VERSION =
  typeof BROWSER_AGENT_BUNDLE_VERSION === "string" && BROWSER_AGENT_BUNDLE_VERSION
    ? BROWSER_AGENT_BUNDLE_VERSION
    : "1.0.18";
const EXT_BUNDLE_MD5 =
  typeof BROWSER_AGENT_BUNDLE_MD5 === "string" ? BROWSER_AGENT_BUNDLE_MD5 : "";
const FEATURES = ["browser-agent", "har-capture"];
const CDP_PROTOCOL_VERSION = "1.3";
const HAR_BODY_SETTLE_MS = 2000;
const HAR_UPLOAD_TIMEOUT_MS = 30000;
const HAR_NETWORK_BUFFER_SIZE = 10 * 1024 * 1024;

/** Active HAR capture keyed by Browser Agent session id. */
const harCaptures = new Map();
/** Active HAR owner keyed by Chrome window id. */
const harWindowOwners = new Map();
/** Capture state keyed by enrolled tab id for Network event routing. */
const harTabs = new Map();
/** Max wait for chrome.debugger.attach + Runtime.enable (and attachLock-held work). */
const ATTACH_TIMEOUT_MS = 10000;

/** @type {Map<string, {ws: WebSocket|null, tabId: number|null, windowId: number|null, controlPort: number, reconnectTimer: ReturnType<typeof setTimeout>|null, reconnectAttempt: number, connectTimeoutTimer: ReturnType<typeof setTimeout>|null, keepaliveTimer: ReturnType<typeof setInterval>|null, pendingResults: Map<string, object>, activeJobIds: Set<string>, completedJobIds: Set<string>}>} */
const sessions = new Map();

/** Abort hung CONNECTING sockets (ms). */
const CONNECT_TIMEOUT_MS = 2500;
/** Cap reconnect backoff so we rejoin the control plane quickly after serve restart. */
const RECONNECT_MAX_MS = 2000;
const RECONNECT_BASE_MS = 100;
/** Keepalive ping while connected (keeps MV3 SW + WS alive). */
const KEEPALIVE_MS = 15000;
/** Offscreen document URL for SW idle keep-alive while sessions are connected. */
const OFFSCREEN_URL = "offscreen.html";
const OFFSCREEN_PORT_NAME = "ba-offscreen-keepalive";
/**
 * Tabs with chrome.debugger currently attached (true CDP debuggees).
 *
 * Chrome UI note: while any tab is attached, Chrome may show a profile-wide
 * notice like "Browser Agent started debugging this browser" on other windows
 * and even blank New Tabs. That banner is not per-tab CDP attach; jobs still
 * target only tabs in sessionAttachState[sessionId].attachedTabIds (multi-tab
 * attach set per session — policy B). See README "Chrome debugger notice".
 */
/** @type {Map<number, true>} */
const attachedTabs = new Map();
/**
 * Per-session debugger attach state (serialize attach via attachLock).
 * Policy B multi-tab set: jobs may attach multiple tabIds; peers stay attached
 * until explicit detach or session leave (no switch-detach).
 */
/** @type {Map<string, { attachedTabIds: Set<number>, attachLock: Promise<void> }>} */
const sessionAttachState = new Map();
/** In-memory console log buffer for logs jobs. */
const consoleLogBuffer = [];
const MAX_LOG_ENTRIES = 500;

/**
 * Debug logger for extension-side job routing / navigation.
 * - Always console.log (service worker DevTools)
 * - Also buffers into consoleLogBuffer so `browser-agent session logs` can read it
 * High-signal jobs (create_tab, Page.navigate) always log; set
 * globalThis.BROWSER_AGENT_DEBUG = true for verbose.
 */
function baLog(level, msg, detail) {
  const text =
    "[browser-agent] " +
    msg +
    (detail != null && detail !== ""
      ? " " + (typeof detail === "string" ? detail : JSON.stringify(detail))
      : "");
  const entry = {
    level: level || "log",
    text: text,
    timestamp: Date.now(),
    source: "browser-agent-ext",
  };
  consoleLogBuffer.push(entry);
  if (consoleLogBuffer.length > MAX_LOG_ENTRIES) {
    consoleLogBuffer.splice(0, consoleLogBuffer.length - MAX_LOG_ENTRIES);
  }
  try {
    if (level === "error") console.error(text);
    else if (level === "warn") console.warn(text);
    else console.log(text);
  } catch (e) {
    /* ignore */
  }
  // Persist to daemon (survives SW death). Fire-and-forget; never block jobs.
  pushExtensionLogToDaemon(level, msg, detail);
}

/**
 * POST /v1/ext/log — unified extension-chrome.log.jsonl on the control server.
 * Optional session_id when a single session is bound or detail carries one.
 */
function pushExtensionLogToDaemon(level, msg, detail) {
  try {
    let controlPort = CONTROL_PORT;
    let sessionId = "";
    try {
      if (sessions && sessions.size === 1) {
        const only = sessions.values().next().value;
        if (only) {
          if (only.controlPort != null) controlPort = only.controlPort;
          if (only.sessionId) sessionId = only.sessionId;
        }
      }
    } catch (e) {
      /* sessions may not exist yet during early boot */
    }
    if (
      !sessionId &&
      detail &&
      typeof detail === "object" &&
      (detail.session_id || detail.sessionId)
    ) {
      sessionId = detail.session_id || detail.sessionId;
    }
    const body = {
      browser: "chrome",
      level: level || "log",
      msg: String(msg || ""),
      source: "extension",
      ts: new Date().toISOString(),
    };
    if (sessionId) body.session_id = sessionId;
    if (detail != null && detail !== "") body.detail = detail;
    fetch("http://127.0.0.1:" + controlPort + "/v1/ext/log", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(body),
    }).catch(function () {
      /* daemon down */
    });
  } catch (e) {
    /* ignore */
  }
}

function baDebugEnabled() {
  try {
    if (globalThis.BROWSER_AGENT_DEBUG === true) return true;
    if (globalThis.BROWSER_AGENT_DEBUG === "1") return true;
  } catch (e) {
    /* ignore */
  }
  return false;
}

function baShouldLogJob(jobType, params) {
  if (baDebugEnabled()) return true;
  // Always log attach-heavy jobs so cold-path latency is visible without DEBUG.
  if (
    jobType === "create_tab" ||
    jobType === "eval" ||
    jobType === "run" ||
    jobType === "screenshot"
  ) {
    return true;
  }
  if (jobType === "cdp") {
    const method =
      (params && (params.method || params.cdp_method || params.cdpMethod)) || "";
    return method === "Page.navigate" || String(method).startsWith("Target.");
  }
  return false;
}

/** High-res phase timing helper for attach/eval dig. */
function baNow() {
  return typeof performance !== "undefined" && performance.now
    ? performance.now()
    : Date.now();
}

/**
 * Race a promise against a timeout so attachLock cannot strand forever.
 * onTimeout is optional cleanup (sync or async); errors from it are ignored.
 */
function baWithTimeout(promise, ms, label, onTimeout) {
  let timer = null;
  const p = Promise.resolve(promise);
  // Prevent unhandled rejection if timeout wins and p settles later.
  p.catch(() => {});
  const timeoutPromise = new Promise((_, reject) => {
    timer = setTimeout(() => {
      const msg = (label || "operation") + " timed out after " + ms + "ms";
      Promise.resolve(typeof onTimeout === "function" ? onTimeout() : null)
        .catch(() => {})
        .finally(() => {
          reject(new Error(msg));
        });
    }, ms);
  });
  return Promise.race([p, timeoutPromise]).finally(() => {
    if (timer) clearTimeout(timer);
  });
}

function baSummarizeParams(jobType, params) {
  params = params || {};
  if (jobType === "create_tab") {
    return { url: params.url || params.URL || params.href || "", active: params.active };
  }
  if (jobType === "cdp") {
    const method = params.method || params.cdp_method || params.cdpMethod || "";
    const nested = params.params || {};
    return {
      method: method,
      nav_url: nested.url || "",
    };
  }
  if (jobType === "eval" || jobType === "run") {
    const expr = String(
      params.expression || params.expr || params.code || params.source || "",
    );
    return { expr: expr.slice(0, 80) };
  }
  return {};
}

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
      pendingResults: new Map(),
      activeJobIds: new Set(),
      completedJobIds: new Set(),
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
  if (!entry.ws || entry.ws.readyState !== WebSocket.OPEN) return false;
  try {
    entry.ws.send(JSON.stringify(obj));
    return true;
  } catch (e) {
    return false;
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

function pendingHAREndCaptureIDs(entry) {
  const ids = [];
  if (!entry || !entry.pendingResults) return ids;
  for (const message of entry.pendingResults.values()) {
    const result = message && message.payload;
    const data = result && result.data;
    const captureId = data && data.type === "har_end" && result.ok
      ? String(data.capture_id || "")
      : "";
    if (captureId && !ids.includes(captureId)) ids.push(captureId);
  }
  return ids;
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
  const capture = harCaptures.get(sessionId);
  if (capture && capture.captureId) {
    payload.har_capture_id = capture.captureId;
  }
  const pendingEndCaptureIDs = pendingHAREndCaptureIDs(entry);
  if (pendingEndCaptureIDs.length > 0) {
    payload.har_pending_end_capture_ids = pendingEndCaptureIDs;
  }
  if (EXT_BUNDLE_MD5) {
    payload.bundle_md5 = EXT_BUNDLE_MD5;
  }
  return sendJSON(entry, {
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
  // Also refresh last-known for the toolbar popup (session + debugger).
  writeLastKnownStatus().catch(() => {});
}

/** Storage key for popup last-known progressive status. */
const LAST_KNOWN_STATUS_KEY = "browserAgentLastKnownStatus";

/**
 * Build popup status payload: sessions connected + attach set size / armed.
 * Used by onMessage getStatus/popupStatus and chrome.storage last-known write.
 */
function buildPopupStatus() {
  let sessionsConnected = 0;
  const sessionList = [];
  for (const [sessionId, entry] of sessions.entries()) {
    const wsOpen =
      entry && entry.ws && entry.ws.readyState === WebSocket.OPEN;
    const wsConnecting =
      entry && entry.ws && entry.ws.readyState === WebSocket.CONNECTING;
    if (wsOpen) sessionsConnected += 1;
    const state = sessionAttachState.get(sessionId);
    const attachCount =
      state && state.attachedTabIds ? state.attachedTabIds.size : 0;
    sessionList.push({
      sessionId: sessionId,
      connected: !!wsOpen,
      connecting: !!wsConnecting,
      attachCount: attachCount,
      armed: attachCount > 0 || entry.tabId != null,
      tabId: entry.tabId,
      windowId: entry.windowId,
    });
  }
  let attachCount = 0;
  for (const state of sessionAttachState.values()) {
    if (state && state.attachedTabIds) {
      attachCount += state.attachedTabIds.size;
    }
  }
  // Global attachedTabs map is the true CDP attach set.
  const attachedTabCount = attachedTabs.size;
  const totalAttached = Math.max(attachCount, attachedTabCount);
  const armed = totalAttached > 0 || sessionsConnected > 0;
  return {
    ok: true,
    type: "popupStatus",
    sessionsConnected: sessionsConnected,
    session_count: sessions.size,
    sessionCount: sessions.size,
    sessions: sessionList,
    attachCount: totalAttached,
    attach_count: totalAttached,
    attachedTabCount: attachedTabCount,
    armed: armed,
    ws: sessionsConnected > 0 ? "connected" : sessions.size > 0 ? "waiting" : "none",
    debugger: totalAttached > 0 ? "armed" : "idle",
    updatedAt: Date.now(),
  };
}

function getPopupStatus() {
  return buildPopupStatus();
}

/**
 * Persist last-known status for the popup (chrome.storage.session preferred).
 */
async function writeLastKnownStatus() {
  const payload = buildPopupStatus();
  try {
    const area =
      (chrome.storage && chrome.storage.session) ||
      (chrome.storage && chrome.storage.local);
    if (!area || typeof area.set !== "function") return payload;
    await new Promise((resolve) => {
      try {
        area.set({ [LAST_KNOWN_STATUS_KEY]: payload }, () => resolve());
      } catch (e) {
        resolve();
      }
    });
  } catch (e) {
    /* storage optional in tests */
  }
  return payload;
}

function persistStatus() {
  return writeLastKnownStatus();
}

/**
 * Handle popup status queries (getStatus / popupStatus / status).
 * Returns true if handled (caller may return true for async sendResponse).
 *
 * Popup open can still feel slow on cold Chrome start (intermittent SW wake).
 * We avoid debugger attach on the popup path for that reason; residual delay is
 * often Chrome/SW cold start, not progressive status JS. Later opens are usually quick.
 */
function handleGetStatus(msg, sendResponse) {
  // Popup is opening: answer status immediately (sync), then release debugger
  // so Chrome is not stuck in "extension is debugging" (that makes the *next*
  // toolbar popup open take tens of seconds).
  const payload = buildPopupStatus();
  writeLastKnownStatus().catch(() => {});
  if (typeof sendResponse === "function") {
    sendResponse({ ok: true, type: "status", payload: payload, ...payload });
  }
  setTimeout(() => {
    const sessionIds = Array.from(sessionAttachState.keys());
    if (sessionIds.length === 0 && attachedTabs.size === 0) return;
    baLog("log", "detach debugger after popup status (popup UX)", {
      sessions: sessionIds.length,
      attached_tabs: attachedTabs.size,
    });
    Promise.all(
      sessionIds.map((id) => detachSessionDebugger(id).catch(() => {})),
    ).then(() => {
      const leftovers = Array.from(attachedTabs.keys()).filter((tid) => !isHAROwnedTab(tid));
      return Promise.all(leftovers.map((tid) => detachDebugger(tid).catch(() => {})));
    });
  }, 0);
  return true;
}

function scheduleReconnect(sessionId) {
  const entry = sessions.get(sessionId);
  if (!entry || entry.reconnectTimer) return;
  // Aggressive reconnect after prepare_reconnect: optional lower base/max from payload.
  const base =
    entry.prepareReconnectBaseMS > 0 ? entry.prepareReconnectBaseMS : RECONNECT_BASE_MS;
  const max =
    entry.prepareReconnectMaxMS > 0 ? entry.prepareReconnectMaxMS : RECONNECT_MAX_MS;
  const delay = Math.min(max, base * Math.pow(2, entry.reconnectAttempt));
  entry.reconnectAttempt += 1;
  entry.reconnectTimer = setTimeout(() => {
    entry.reconnectTimer = null;
    connectSession(sessionId, "reconnect");
  }, delay);
}

/**
 * Control plane is about to restart (daemon upgrade). Schedule close after delay_ms,
 * reset reconnectAttempt for aggressive reconnect, and apply optional retry hints.
 */
function handlePrepareReconnect(msg, sessionId, entry) {
  const payload = (msg && msg.payload) || {};
  let delayMs =
    payload.delay_ms != null
      ? Number(payload.delay_ms)
      : payload.delayMs != null
        ? Number(payload.delayMs)
        : NaN;
  if (!Number.isFinite(delayMs) || delayMs <= 0) {
    delayMs = 1000;
  }
  if (payload.retry_base_ms != null && Number(payload.retry_base_ms) > 0) {
    entry.prepareReconnectBaseMS = Number(payload.retry_base_ms);
  } else if (payload.retryBaseMs != null && Number(payload.retryBaseMs) > 0) {
    entry.prepareReconnectBaseMS = Number(payload.retryBaseMs);
  }
  if (payload.retry_max_ms != null && Number(payload.retry_max_ms) > 0) {
    entry.prepareReconnectMaxMS = Number(payload.retry_max_ms);
  } else if (payload.retryMaxMs != null && Number(payload.retryMaxMs) > 0) {
    entry.prepareReconnectMaxMS = Number(payload.retryMaxMs);
  }
  // Force aggressive reconnect after the intentional close.
  entry.reconnectAttempt = 0;
  if (entry.reconnectTimer) {
    clearTimeout(entry.reconnectTimer);
    entry.reconnectTimer = null;
  }
  baLog("log", "prepare_reconnect scheduled", {
    session_id: sessionId,
    delay_ms: delayMs,
    reason: payload.reason || "",
  });
  setTimeout(() => {
    entry.reconnectAttempt = 0;
    if (entry.ws) {
      try {
        entry.ws.close();
      } catch (e) {
        /* ignore */
      }
    } else if (sessions.has(sessionId)) {
      scheduleReconnect(sessionId);
    }
  }, delayMs);
}

/**
 * Open per-session WS to control plane. reason is for service-worker console diagnostics.
 */
function connectSession(sessionId, reason) {
  if (!sessionId) return;
  const entry = getOrCreateSessionEntry(sessionId);
  const why = reason || "";

  if (entry.ws && entry.ws.readyState === WebSocket.OPEN) {
    baLog("log", "connectSession skip", {
      session_id: sessionId,
      reason: why,
      skip: "ws_already_open",
    });
    return;
  }
  if (entry.ws && entry.ws.readyState === WebSocket.CONNECTING) {
    baLog("log", "connectSession skip", {
      session_id: sessionId,
      reason: why,
      skip: "ws_connecting",
    });
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
  const wsURL = wsURLForSession(sessionId, entry.controlPort);
  baLog("log", "connectSession start", {
    session_id: sessionId,
    reason: why,
    control_port: entry.controlPort,
    ws_url: wsURL,
  });
  try {
    socket = new WebSocket(wsURL);
  } catch (e) {
    baLog("error", "connectSession WebSocket ctor fail", {
      session_id: sessionId,
      reason: why,
      error: String(e && e.message ? e.message : e),
    });
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
    baLog("log", "connectSession ws_open", {
      session_id: sessionId,
      reason: why,
    });
    sendHello(sessionId, entry)
      .then(() => flushPendingJobResults(entry))
      .catch((e) => {
        baLog("warn", "connectSession hello fail", {
          session_id: sessionId,
          error: String(e && e.message ? e.message : e),
        });
      });
    startKeepalive(sessionId, entry);
    ensureOffscreenKeepAlive().catch(() => {});
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
  // Tear down chrome.debugger held by this session (leave / last control tab gone).
  const activeCapture = harCaptures.get(sessionId);
  if (activeCapture) {
    activeCapture.partial = true;
    activeCapture.warnings.push("session unregistered while HAR capture was active");
    rollbackHARCapture(activeCapture, true)
      .catch(() => {})
      .finally(() => detachSessionDebugger(sessionId).catch(() => {}));
  } else {
    detachSessionDebugger(sessionId);
  }
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
  sessionAttachState.delete(sessionId);
}

function parseSessionGoPageURL(url) {
  if (!url || typeof url !== "string") return null;
  try {
    const parsed = new URL(url);
    const protocol = (parsed.protocol || "").toLowerCase();
    if (protocol !== "http:" && protocol !== "https:") return null;
    const host = (parsed.hostname || "").toLowerCase();
    if (
      host !== "127.0.0.1" &&
      host !== "localhost" &&
      host !== "::1" &&
      host !== "[::1]"
    ) {
      return null;
    }
    if (parsed.pathname !== "/go") return null;
    return parsed;
  } catch (e) {
    return null;
  }
}

function isSessionGoPageURL(url, sessionId) {
  const parsed = parseSessionGoPageURL(url);
  if (!parsed) return false;
  if (!sessionId) return true;
  return parsed.searchParams.get("session") === String(sessionId);
}

function isSessionGoPageURLAtPort(url, controlPort) {
  const parsed = parseSessionGoPageURL(url);
  if (!parsed) return false;
  const port = parsed.port
    ? parseInt(parsed.port, 10)
    : parsed.protocol === "https:"
      ? 443
      : 80;
  return port === Number(controlPort);
}

function parseGoSessionFromURL(url) {
  const parsed = parseSessionGoPageURL(url);
  if (!parsed) return null;
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
}

function maybeRegisterGoTab(tabId, url, tab, source) {
  const src = source || "unknown";
  const parsed = parseGoSessionFromURL(url);
  if (!parsed) {
    // Only note near-misses (looks like control host) — avoid spam on every tab.
    if (url && (url.indexOf("/go") >= 0 || url.indexOf("43761") >= 0)) {
      baLog("log", "maybeRegisterGoTab skip", {
        source: src,
        tab_id: tabId,
        url: String(url).slice(0, 200),
        reason: "not_go_session_url",
      });
    }
    return false;
  }
  const ok = handleRegisterMessage(
    {
      type: "register",
      session_id: parsed.sessionId,
      control_port: parsed.controlPort,
      tabId: tabId,
      windowId: tab && tab.windowId,
    },
    { tab: { id: tabId, windowId: tab && tab.windowId, url: url } },
  );
  baLog(ok ? "log" : "warn", "maybeRegisterGoTab", {
    source: src,
    tab_id: tabId,
    window_id: tab && tab.windowId,
    session_id: parsed.sessionId,
    url: String(url).slice(0, 200),
    ok: !!ok,
  });
  return !!ok;
}

/**
 * List open /go?session=<sessionId> control tabs, optionally scoped to a window.
 * Used by attach gate and leave recount (not hello telemetry).
 */
async function querySessionControlTabs(sessionId, windowId) {
  if (!sessionId) return [];
  const queryInfo = {};
  if (windowId != null) queryInfo.windowId = windowId;
  try {
    const tabs = await chrome.tabs.query(queryInfo);
    return (tabs || []).filter(
      (t) => t && t.id != null && isSessionGoPageURL(t.url || "", sessionId),
    );
  } catch (e) {
    return [];
  }
}

async function countOpenSessionPages(sessionId, windowId) {
  const tabs = await querySessionControlTabs(sessionId, windowId);
  return tabs.length;
}

async function hasOpenSessionPage(sessionId, windowId) {
  return (await countOpenSessionPages(sessionId, windowId)) > 0;
}

function rebindSessionTabId(sessionId, tab) {
  const entry = sessions.get(sessionId);
  if (!entry || !tab) return;
  if (tab.id != null) entry.tabId = tab.id;
  if (tab.windowId != null) entry.windowId = tab.windowId;
}

/**
 * Detach chrome.debugger for every tab in a session's attach set
 * (last control tab leave / unregister). Serializes through attachLock so it
 * does not race multi-tab attach reuse.
 */
function detachSessionDebugger(sessionId) {
  let state = sessionAttachState.get(sessionId);
  if (!state) {
    return Promise.resolve();
  }
  const run = async () => {
    const tabIds = Array.from(state.attachedTabIds || []);
    for (const tabId of tabIds) {
      if (isHAROwnedTab(tabId)) continue;
      state.attachedTabIds.delete(tabId);
      await detachDebugger(tabId);
    }
  };
  state.attachLock = state.attachLock.then(run, run);
  return state.attachLock;
}

/**
 * Leave path: re-query remaining /go?session= control tabs.
 * remaining == 0 → detach + unregister; remaining >= 1 → rebind, stay armed.
 */
async function handleSessionControlLeave(sessionId) {
  const entry = sessions.get(sessionId);
  if (!entry) return;
  const windowId = entry.windowId != null ? entry.windowId : null;
  const remaining = await querySessionControlTabs(sessionId, windowId);
  if (remaining.length === 0) {
    // Prefer awaiting detach before WS teardown so the banner is cleared promptly.
    try {
      await detachSessionDebugger(sessionId);
    } catch (e) {
      /* ignore */
    }
    unregisterSession(sessionId);
    return;
  }
  const stillBound = remaining.some((t) => t.id === entry.tabId);
  if (!stillBound) {
    rebindSessionTabId(sessionId, remaining[0]);
  }
}

function handleRegisterMessage(msg, sender) {
  const claimedSessionId = msg.session_id || msg.sessionId || "";
  const senderURL =
    (sender && sender.tab && sender.tab.url) || (sender && sender.url) || "";
  const parsed = parseGoSessionFromURL(senderURL);
  if (!claimedSessionId) {
    baLog("warn", "register reject", { reason: "empty_session_id", sender_url: String(senderURL || "").slice(0, 200) });
    return false;
  }
  if (!parsed) {
    baLog("warn", "register reject", {
      reason: "sender_url_not_go",
      session_id: claimedSessionId,
      sender_url: String(senderURL || "").slice(0, 200),
      has_tab: !!(sender && sender.tab),
      tab_id: sender && sender.tab && sender.tab.id,
    });
    return false;
  }
  if (parsed.sessionId !== claimedSessionId) {
    baLog("warn", "register reject", {
      reason: "session_id_mismatch",
      claimed: claimedSessionId,
      url_session: parsed.sessionId,
      sender_url: String(senderURL || "").slice(0, 200),
    });
    return false;
  }
  const sessionId = parsed.sessionId;
  const senderTab = sender && sender.tab;
  if (!senderTab || senderTab.id == null || senderTab.windowId == null) {
    baLog("warn", "register reject", {
      reason: "missing_tab_or_window",
      session_id: sessionId,
      tab_id: senderTab && senderTab.id,
      window_id: senderTab && senderTab.windowId,
      sender_url: String(senderURL || "").slice(0, 200),
    });
    return false;
  }
  const entry = getOrCreateSessionEntry(sessionId);
  const tabId = senderTab.id;
  const windowId = senderTab.windowId;
  const controlPort = parsed.controlPort;
  if (tabId != null) entry.tabId = tabId;
  if (windowId != null) entry.windowId = windowId;
  if (controlPort != null && !Number.isNaN(controlPort) && controlPort > 0) {
    entry.controlPort = controlPort;
  }
  // WS only on register — never chrome.debugger.attach here.
  // Attaching on /go register (and on every SW wake heal) freezes Chrome's UI and
  // delays/blocks the toolbar default_popup for tens of seconds. Content tabs still
  // eager-attach on create_tab / navigate (P2); jobs attach on demand.
  baLog("log", "register ok", {
    session_id: sessionId,
    tab_id: tabId,
    window_id: windowId,
    control_port: entry.controlPort,
  });
  connectSession(sessionId, "register");
  ensureOffscreenKeepAlive().catch(() => {});
  return true;
}

function handleRegisterFromMessage(msg, sender, sendResponse) {
  const registered = handleRegisterMessage(msg, sender);
  // Ack so content-script / page burst retries can stop early after SW is live.
  try {
    sendResponse(registered
      ? { ok: true, type: "register_ack" }
      : { ok: false, type: "register_rejected" });
  } catch (e) {
    /* channel closed */
  }
  return true;
}

/**
 * Long-lived Ports from /go content scripts and session-page boot keep the MV3
 * service worker alive while a control tab is open. One-shot sendMessage often
 * fails with "Receiving end does not exist" after SW idle kill; connect wakes
 * and holds the worker so register + WS can complete.
 *
 * Port name: "ba-session:<sessionId>"
 */
const sessionKeepAlivePorts = new Set();

function sessionIdFromPortName(name) {
  const raw = String(name || "");
  if (raw.indexOf("ba-session:") === 0) {
    return raw.slice("ba-session:".length).trim();
  }
  return "";
}

function ackPortRegister(port, ok) {
  try {
    port.postMessage(
      ok
        ? { ok: true, type: "register_ack" }
        : { ok: false, type: "register_rejected" },
    );
  } catch (e) {
    /* channel closed */
  }
}

/**
 * Register from a keep-alive Port. When sender.tab is missing (common for
 * externally_connectable connect before the tab binding is ready), resolve the
 * /go tab via tabs.query so WS can still attach.
 */
function tryRegisterFromPort(port, msg) {
  const claimed =
    (msg && (msg.session_id || msg.sessionId)) ||
    sessionIdFromPortName(port.name) ||
    "";
  if (!claimed) {
    baLog("warn", "port register skip", {
      reason: "empty_session_id",
      port_name: port && port.name,
    });
    ackPortRegister(port, false);
    return false;
  }
  const payload = Object.assign({}, msg || {}, {
    type: "register",
    session_id: claimed,
  });
  // Prefer tabs.query over port.sender.tab — sender.tab.url is often empty on
  // the first externally_connectable connect, which caused silent register_rejected.
  if (!chrome.tabs || typeof chrome.tabs.query !== "function") {
    const sender = port.sender || {};
    const result = handleRegisterMessage(payload, sender);
    const ok = result === true || (result !== false && sessions.has(claimed));
    baLog(ok ? "log" : "warn", "port register", {
      session_id: claimed,
      via: "sender_only",
      ok: !!ok,
    });
    ackPortRegister(port, ok);
    return ok;
  }
  chrome.tabs.query({}, (tabs) => {
    if (chrome.runtime.lastError) {
      baLog("warn", "port register tabs.query fail", {
        session_id: claimed,
        error: chrome.runtime.lastError.message,
      });
      ackPortRegister(port, false);
      return;
    }
    for (const tab of tabs || []) {
      if (!tab || tab.id == null) continue;
      const parsed = parseGoSessionFromURL(tab.url || "");
      if (!parsed || parsed.sessionId !== claimed) continue;
      const tabSender = { tab: tab, url: tab.url || "" };
      const r = handleRegisterMessage(payload, tabSender);
      const success = r === true || (r !== false && sessions.has(claimed));
      baLog(success ? "log" : "warn", "port register", {
        session_id: claimed,
        via: "tabs_query",
        tab_id: tab.id,
        ok: !!success,
      });
      ackPortRegister(port, success);
      return;
    }
    // Fallback: sender.tab if query found nothing yet (tab still loading).
    const sender = port.sender || {};
    const result = handleRegisterMessage(payload, sender);
    const ok = result === true || (result !== false && sessions.has(claimed));
    baLog(ok ? "log" : "warn", "port register", {
      session_id: claimed,
      via: "sender_fallback",
      tab_id: sender.tab && sender.tab.id,
      sender_url: String(
        (sender.tab && sender.tab.url) || sender.url || "",
      ).slice(0, 200),
      ok: !!ok,
    });
    ackPortRegister(port, ok);
  });
  return false;
}

function handleSessionKeepAlivePort(port) {
  if (!port) return;
  sessionKeepAlivePorts.add(port);
  const drop = () => {
    sessionKeepAlivePorts.delete(port);
  };
  try {
    port.onDisconnect.addListener(drop);
  } catch (e) {
    /* ignore */
  }

  const portName = (port && port.name) || "";
  // Offscreen keepalive Port — hold SW idle timer; no session register.
  if (portName === OFFSCREEN_PORT_NAME || portName.indexOf("ba-offscreen") === 0) {
    baLog("log", "offscreen port connect", {});
    try {
      port.onMessage.addListener((msg) => {
        if (msg && msg.type === "keepalive") {
          /* message receipt resets SW idle */
        }
      });
    } catch (e) {
      /* ignore */
    }
    return;
  }

  baLog("log", "port connect", {
    port_name: portName,
    has_tab: !!(port.sender && port.sender.tab),
    tab_id: port.sender && port.sender.tab && port.sender.tab.id,
    sender_url: String(
      (port.sender && port.sender.tab && port.sender.tab.url) ||
        (port.sender && port.sender.url) ||
        "",
    ).slice(0, 200),
  });

  // Only register on postMessage — NOT on connect. chrome.runtime.connect()
  // delivers onConnect synchronously before the page attaches onMessage, so an
  // immediate ack was lost and the page kept spinning on register_sent.
  try {
    port.onMessage.addListener((msg) => {
      if (!msg || typeof msg !== "object") return;
      if (msg.type === "register" || msg.type === "ping") {
        tryRegisterFromPort(port, msg);
      } else if (msg.type === "keepalive") {
        /* /go page ping — resets SW idle while session connected */
      }
    });
  } catch (e) {
    /* ignore */
  }

  // Rediscover open /go tabs (tabs.onUpdated wake + register).
  scheduleHealSessionsFromTabs("port-connect", 0);
  ensureOffscreenKeepAlive().catch(() => {});
}

/**
 * Ensure a hidden offscreen document exists to ping this SW every ~20s while
 * sessions are connected (resets MV3 idle kill without a visible tab).
 */
async function ensureOffscreenKeepAlive() {
  if (!chrome.offscreen || typeof chrome.offscreen.createDocument !== "function") {
    return;
  }
  try {
    const contexts =
      chrome.runtime.getContexts &&
      (await chrome.runtime.getContexts({
        contextTypes: ["OFFSCREEN_DOCUMENT"],
      }));
    if (contexts && contexts.length > 0) return;
  } catch (e) {
    /* getContexts may be unavailable — try create anyway */
  }
  try {
    await chrome.offscreen.createDocument({
      url: OFFSCREEN_URL,
      reasons: ["WORKERS"],
      justification:
        "Keep Browser Agent service worker alive while control sessions are connected",
    });
    baLog("log", "offscreen keepalive created", {});
  } catch (e) {
    const msg = String(e && e.message ? e.message : e);
    // Already exists is fine.
    if (/already exists|only one offscreen/i.test(msg)) return;
    baLog("warn", "offscreen keepalive create failed", { error: msg });
  }
}

chrome.runtime.onMessage.addListener((msg, sender, sendResponse) => {
  if (!msg || typeof msg !== "object") return;
  if (msg.type === "register") {
    return handleRegisterFromMessage(msg, sender, sendResponse);
  }
  // Silent wakeup.html (session-new cold-start): force heal of open /go tabs.
  if (msg.type === "wakeup") {
    scheduleHealSessionsFromTabs("wakeup-page", 0);
    try {
      sendResponse({ ok: true, type: "wakeup_ack" });
    } catch (e) {
      /* channel closed */
    }
    return true;
  }
  // Popup progressive status API: status / getStatus / popupStatus.
  if (
    msg.type === "status" ||
    msg.type === "getStatus" ||
    msg.type === "get_status" ||
    msg.type === "popupStatus" ||
    msg.type === "popup_status" ||
    msg.type === "getPopupStatus" ||
    msg.type === "lastKnown" ||
    msg.type === "last_known"
  ) {
    // Cheap WS rediscover when operator opens the toolbar popup (Map may be empty).
    scheduleHealSessionsFromTabs("popup-status", 0);
    handleGetStatus(msg, sendResponse);
    return true; // keep channel open for sendResponse
  }
});

// Session page (http://127.0.0.1:43761/go) externally_connectable register.
// Survives cold-start when content scripts miss the first navigation.
try {
  chrome.runtime.onMessageExternal.addListener((msg, sender, sendResponse) => {
    if (!msg || typeof msg !== "object") return;
    if (msg.type === "register") {
      // Prefer sender.tab when the page is a real tab.
      return handleRegisterFromMessage(msg, sender, sendResponse);
    }
  });
} catch (e) {
  /* onMessageExternal unavailable in some fixtures */
}

// Port keep-alive — register listeners at top level (no try/catch around
// addListener) so Chrome includes them in MV3 service-worker wake events.
// onConnectExternal wakes from session-page chrome.runtime.connect(EXT_ID);
// onConnect wakes from content-script connect({name}).
chrome.runtime.onConnect.addListener(handleSessionKeepAlivePort);
if (chrome.runtime.onConnectExternal) {
  chrome.runtime.onConnectExternal.addListener(handleSessionKeepAlivePort);
}

chrome.tabs.onRemoved.addListener((tabId) => {
  closeHARTab(tabId);
  // Re-query remaining control tabs for every session (multi-tab safe).
  // Do not trust entry.tabId alone — last register wins and would wrongly disarm.
  const sessionIds = Array.from(sessions.keys());
  Promise.all(
    sessionIds.map((sessionId) =>
      handleSessionControlLeave(sessionId).catch(() => {}),
    ),
  ).then(() => {
    pushStatusForConnectedSessions();
  });
});

chrome.tabs.onUpdated.addListener((tabId, changeInfo, tab) => {
  updateHARTabFromChrome(tabId, changeInfo, tab).catch(() => {});
  const url = (changeInfo && changeInfo.url) || (tab && tab.url) || "";
  if (url) {
    if (changeInfo.status === "loading" || changeInfo.status === "complete") {
      // High-signal cold-start breadcrumb: did Chrome deliver onUpdated for /go?
      if (url.indexOf("/go") >= 0 || url.indexOf("session=") >= 0) {
        baLog("log", "tabs.onUpdated", {
          tab_id: tabId,
          status: changeInfo && changeInfo.status,
          change_url: !!(changeInfo && changeInfo.url),
          url: String(url).slice(0, 200),
          window_id: tab && tab.windowId,
        });
      }
      maybeRegisterGoTab(tabId, url, tab, "tabs.onUpdated");
    }
    // Navigate-away leave: when a bound control tab leaves /go?session=, recount.
    const leaveIds = [];
    for (const [sessionId, entry] of sessions.entries()) {
      if (entry.tabId !== tabId) continue;
      if (!isSessionGoPageURL(url, sessionId)) {
        leaveIds.push(sessionId);
      }
    }
    if (leaveIds.length > 0) {
      Promise.all(
        leaveIds.map((sessionId) =>
          handleSessionControlLeave(sessionId).catch(() => {}),
        ),
      ).then(() => {
        pushStatusForConnectedSessions();
      });
    } else if (url.includes("/go")) {
      pushStatusForConnectedSessions();
    }
  }
  // P2 eager arm: attach on navigation *complete* only (not loading) to avoid
  // double attach storms that freeze Chrome UI and block the toolbar popup.
  if (changeInfo && changeInfo.status === "complete") {
    const tabWindowId = tab && tab.windowId;
    const tabUrl = url || (tab && tab.url) || "";
    if (tabWindowId != null && tabUrl) {
      for (const [sessionId, entry] of sessions.entries()) {
        if (entry.windowId == null || entry.windowId !== tabWindowId) continue;
        maybeEagerAttach(sessionId, tabId, tabUrl, tabWindowId);
      }
    }
  }
});

chrome.tabs.onCreated.addListener((tab) => {
  if (tab && tab.id != null) updateHARTabFromChrome(tab.id, {}, tab).catch(() => {});
  // session new --new-window: register /go control tab as soon as it exists
  // (onUpdated alone can miss the first commit in some Chrome builds).
  if (tab && tab.id != null) {
    const earlyURL = tab.url || tab.pendingUrl || "";
    if (earlyURL) {
      if (earlyURL.indexOf("/go") >= 0 || earlyURL.indexOf("session=") >= 0) {
        baLog("log", "tabs.onCreated", {
          tab_id: tab.id,
          url: String(earlyURL).slice(0, 200),
          pending: !!(tab.pendingUrl && !tab.url),
          window_id: tab.windowId,
        });
      }
      maybeRegisterGoTab(tab.id, earlyURL, tab, "tabs.onCreated");
    }
  }
  // P2 eager arm: auto-attach newly opened capturable tabs in armed session windows.
  if (!tab || tab.id == null || tab.windowId == null) return;
  const tabUrl = tab.url || tab.pendingUrl || "";
  for (const [sessionId, entry] of sessions.entries()) {
    if (entry.windowId == null || entry.windowId !== tab.windowId) continue;
    maybeEagerAttach(sessionId, tab.id, tabUrl, tab.windowId);
  }
});

/**
 * P3 self-heal: rediscover open /go?session= tabs and reconnect **WebSocket only**.
 *
 * NEVER call chrome.debugger.attach from heal or SW top-level. Debugger attach
 * freezes Chrome UI and blocks toolbar default_popup (often 10–30s). Multi-tab
 * attach stays on create_tab / navigate / jobs only.
 */
function shouldAdoptDebuggerOnHeal(reason) {
  const r = String(reason || "");
  // Only reconcile debugger ownership after SW (re)boot — not on every Port/alarm
  // tick (that work can stall the worker while sessions are actively connected).
  return (
    r.indexOf("sw-boot") === 0 ||
    r === "onInstalled" ||
    r === "onStartup" ||
    r === "wakeup-page"
  );
}

async function healSessionsFromTabs(reason) {
  baLog("log", "healSessionsFromTabs start", { reason: reason || "" });
  let adopted = 0;
  if (shouldAdoptDebuggerOnHeal(reason)) {
    try {
      adopted = await adoptDebuggerTargetsFromChrome(reason || "heal");
    } catch (e) {
      baLog("warn", "healSessionsFromTabs adopt failed", {
        reason: reason || "",
        error: String(e && e.message ? e.message : e),
      });
    }
  }
  // Prefer known session control tabs; fall back to a full query only for /go URLs.
  let tabs = [];
  try {
    const known = [];
    for (const entry of sessions.values()) {
      if (entry && entry.tabId != null) {
        try {
          const t = await chrome.tabs.get(entry.tabId);
          if (t) known.push(t);
        } catch (e) {
          /* tab gone */
        }
      }
    }
    if (known.length > 0) {
      tabs = known;
    } else {
      tabs = await chrome.tabs.query({});
    }
  } catch (e) {
    baLog("warn", "healSessionsFromTabs tabs.query failed", {
      reason: reason || "",
      error: String(e && e.message ? e.message : e),
    });
    return;
  }
  let found = 0;
  for (const tab of tabs || []) {
    if (!tab || tab.id == null) continue;
    const url = tab.url || "";
    const parsed = parseGoSessionFromURL(url);
    if (!parsed) continue;
    found += 1;
    // Bind + WS only — no maybeEagerAttach / debugger.
    maybeRegisterGoTab(tab.id, url, tab, "heal:" + (reason || ""));
  }
  // If we only scanned known tabIds and found nothing, one full /go scan.
  if (found === 0 && sessions.size > 0) {
    try {
      const all = await chrome.tabs.query({});
      for (const tab of all || []) {
        if (!tab || tab.id == null) continue;
        const url = tab.url || "";
        if (!parseGoSessionFromURL(url)) continue;
        found += 1;
        maybeRegisterGoTab(tab.id, url, tab, "heal-full:" + (reason || ""));
      }
    } catch (e) {
      /* ignore */
    }
  }
  baLog("log", "healSessionsFromTabs done", {
    reason: reason || "",
    control_tabs: found,
    attach: "skipped_for_popup_ux",
    adopted_debugger_tabs: adopted,
    scoped: sessions.size > 0,
  });
  if (found > 0 || sessions.size > 0) {
    ensureOffscreenKeepAlive().catch(() => {});
  }
}

chrome.runtime.onInstalled.addListener(() => {
  console.log("Browser Agent installed; control port", CONTROL_PORT);
  // Defer WS rediscover; never run debugger work on the install critical path.
  setTimeout(() => {
    healSessionsFromTabs("onInstalled").catch(() => {});
  }, 1000);
});

chrome.runtime.onStartup.addListener(() => {
  setTimeout(() => {
    healSessionsFromTabs("onStartup").catch(() => {});
  }, 1000);
});

// WS-only rediscover after SW death / daemon restart.
//
// Do NOT call chrome.debugger.attach here (or from healSessionsFromTabs). A
// previous boot-time attach storm made default_popup take tens of seconds.
// healSessionsFromTabs is WS+register only — cheap tabs.query + connectSession.
//
// Why this is required: MV3 SW is memoryless after idle kill. Alarm used to
// only loop sessions.keys() → no-op when the Map is empty, so open /go tabs
// never re-bound until a full page reload. Content-script register also
// retries; alarm heal is the SW-side safety net.

/**
 * Schedule WS-only heal. delayMS=0 runs ASAP (coalesced). Positive delays are
 * independent timers so staggered boot heals (250ms / 1s / 3s / 8s) all fire.
 */
let healKickZeroTimer = null;
function scheduleHealSessionsFromTabs(reason, delayMS) {
  const ms = delayMS != null && delayMS >= 0 ? delayMS : 0;
  const run = () => {
    healSessionsFromTabs(reason || "schedule").catch(() => {});
  };
  if (ms === 0) {
    if (healKickZeroTimer) return;
    healKickZeroTimer = setTimeout(() => {
      healKickZeroTimer = null;
      run();
    }, 0);
    return;
  }
  setTimeout(run, ms);
}

// Deferred SW-boot rediscover: empty sessions Map after cold start; open /go
// tabs must be rebound. Staggered delays cover Chrome cold start where the
// /go tab exists before the SW is first evaluated (session-new race).
// First delay stays short so toolbar popup getStatus can paint first.
scheduleHealSessionsFromTabs("sw-boot", 250);
scheduleHealSessionsFromTabs("sw-boot-1s", 1000);
scheduleHealSessionsFromTabs("sw-boot-3s", 3000);
scheduleHealSessionsFromTabs("sw-boot-8s", 8000);

try {
  // periodInMinutes min is 1 in Chrome; rely on sw-boot stagger + content retry
  // for the session-new 30s window.
  chrome.alarms.create("browser-agent-reconnect", { periodInMinutes: 1 });
  chrome.alarms.onAlarm.addListener((alarm) => {
    if (alarm && alarm.name === "browser-agent-reconnect") {
      // Always rediscover open /go tabs first (Map may be empty after SW kill).
      healSessionsFromTabs("alarm").catch(() => {});
      // Also reconnect any known sessions whose socket died (daemon restart).
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

function handleHARStatus(msg, sessionId) {
  const payload = msg && msg.payload;
  const abortCaptureId = payload && String(payload.har_abort_capture_id || "");
  if (!abortCaptureId) return;
  const capture = harCaptures.get(sessionId);
  if (!capture || capture.captureId !== abortCaptureId) return;
  rollbackHARCapture(capture, false).catch(() => {});
}

function handleMessage(msg, sessionId, entry) {
  if (!msg || typeof msg !== "object") return;
  const type = msg.type;
  if (type === "result_ack") {
    handleResultAck(msg, entry);
    return;
  }
  if (type === "status") {
    handleHARStatus(msg, sessionId);
    return;
  }
  if (type === "prepare_reconnect") {
    handlePrepareReconnect(msg, sessionId, entry);
    return;
  }
  if (type === "job") {
    handleJob(msg, sessionId, entry);
  }
}

function flushPendingJobResults(entry) {
  if (!entry || !entry.pendingResults) return;
  for (const message of entry.pendingResults.values()) {
    if (!sendJSON(entry, message)) return;
  }
}

function handleResultAck(msg, entry) {
  const payload = msg && msg.payload;
  const jobId = String((payload && (payload.job_id || payload.id)) || (msg && msg.id) || "");
  if (jobId && entry && entry.pendingResults) entry.pendingResults.delete(jobId);
}

function sendJobResult(entry, jobId, ok, data, error) {
  const message = {
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
  };
  if (entry && entry.pendingResults) entry.pendingResults.set(jobId, message);
  sendJSON(entry, message);
}

function claimJobExecution(entry, jobId) {
  jobId = String(jobId || "");
  if (!entry || !jobId) return true;
  if (entry.pendingResults && entry.pendingResults.has(jobId)) {
    sendJSON(entry, entry.pendingResults.get(jobId));
    return false;
  }
  if (!entry.activeJobIds) entry.activeJobIds = new Set();
  if (!entry.completedJobIds) entry.completedJobIds = new Set();
  if (entry.activeJobIds.has(jobId) || entry.completedJobIds.has(jobId)) return false;
  entry.activeJobIds.add(jobId);
  return true;
}

function finishJobExecution(entry, jobId) {
  jobId = String(jobId || "");
  if (!entry || !jobId) return;
  if (!entry.activeJobIds) entry.activeJobIds = new Set();
  if (!entry.completedJobIds) entry.completedJobIds = new Set();
  entry.activeJobIds.delete(jobId);
  entry.completedJobIds.add(jobId);
  while (entry.completedJobIds.size > 1024) {
    entry.completedJobIds.delete(entry.completedJobIds.values().next().value);
  }
}

function isHARCapturedTab(tabId) {
  return harTabs.has(Number(tabId));
}

function isHAROwnedTab(tabId) {
  tabId = Number(tabId);
  if (harTabs.has(tabId)) return true;
  for (const capture of harCaptures.values()) {
    if (capture && capture.enrolling && capture.enrolling.has(tabId)) return true;
  }
  return false;
}

function harErrorMessage(error) {
  return error && error.message ? error.message : String(error);
}

function validateHARUploadURL(rawURL, controlPort) {
  let parsed;
  try {
    parsed = new URL(String(rawURL || ""));
  } catch (error) {
    throw new Error("HAR artifact upload URL is invalid");
  }
  const host = (parsed.hostname || "").toLowerCase();
  const port = parsed.port
    ? parseInt(parsed.port, 10)
    : parsed.protocol === "https:"
      ? 443
      : 80;
  const loopback = host === "127.0.0.1" || host === "localhost" || host === "::1" || host === "[::1]";
  if (
    parsed.protocol !== "http:" ||
    !loopback ||
    port !== Number(controlPort) ||
    parsed.pathname !== "/v1/har/artifact" ||
    parsed.username ||
    parsed.password
  ) {
    throw new Error("HAR artifact upload URL must use the session control server");
  }
  return parsed.toString();
}

function isHAREligibleTab(tab, capture) {
  if (!capture || !tab || tab.id == null || tab.windowId !== capture.windowId) return false;
  const url = tab.url || tab.pendingUrl || "";
  return isCapturableTabURL(url) && !isSessionGoPageURLAtPort(url, capture.controlPort);
}

async function enableHARNetwork(tabId) {
  await sendDebuggerCommand(tabId, "Network.enable", {
    maxPostDataSize: HAR_NETWORK_BUFFER_SIZE,
    maxTotalBufferSize: HAR_NETWORK_BUFFER_SIZE,
    maxResourceBufferSize: HAR_NETWORK_BUFFER_SIZE,
  });
}

async function enrollHARTab(capture, tab, initial) {
  if (!capture || capture.stopping || !isHAREligibleTab(tab, capture)) {
    return null;
  }
  const updateIdentity = (state) => {
    state.title = tab.title || state.title;
    state.url = tab.url || tab.pendingUrl || state.url;
    return state;
  };
  if (capture.tabs.has(tab.id)) {
    return updateIdentity(capture.tabs.get(tab.id));
  }

  let enrollment = capture.enrolling.get(tab.id);
  if (!enrollment) {
    const wasAttached = attachedTabs.has(tab.id);
    enrollment = (async () => {
      try {
        await attachDebuggerForSession(capture.sessionId, tab.id);
        await enableHARNetwork(tab.id);
        if (capture.stopping) throw new Error("capture is stopping");
        const recorder = new BrowserAgentHAR.Recorder(tab.id, {
          controlPort: capture.controlPort,
          creatorVersion: EXT_VERSION,
          sendCommand: (method, params) => sendDebuggerCommand(tab.id, method, params),
        });
        const state = {
          tabId: tab.id,
          title: tab.title || "",
          url: tab.url || tab.pendingUrl || "",
          state: "open",
          recorder,
          errors: [],
          wasAttached,
          uploaded: false,
          uploadPromise: null,
          snapshot: null,
          snapshotJSON: "",
          snapshotResult: null,
        };
        capture.tabs.set(tab.id, state);
        harTabs.set(tab.id, { capture, state, routing: true });
        return state;
      } catch (error) {
        if (!wasAttached) {
          const attachState = sessionAttachState.get(capture.sessionId);
          if (attachState && attachState.attachedTabIds) attachState.attachedTabIds.delete(tab.id);
          await detachDebugger(tab.id, true).catch(() => {});
        }
        throw new Error("tab " + tab.id + " enrollment failed: " + harErrorMessage(error));
      }
    })();
    capture.enrolling.set(tab.id, enrollment);
  }

  try {
    return updateIdentity(await enrollment);
  } catch (error) {
    const message = harErrorMessage(error);
    if (initial) throw error;
    capture.partial = true;
    if (!capture.enrollmentFailures.has(tab.id)) {
      capture.enrollmentFailures.add(tab.id);
      capture.warnings.push(message);
    }
    return null;
  } finally {
    if (capture.enrolling.get(tab.id) === enrollment) {
      capture.enrolling.delete(tab.id);
    }
  }
}

function stopHARRouting(capture) {
  if (!capture) return;
  for (const routed of harTabs.values()) {
    if (routed && routed.capture === capture) routed.routing = false;
  }
}

async function waitForHAREnrollments(capture) {
  while (capture.enrolling.size > 0) {
    await Promise.allSettled(Array.from(capture.enrolling.values()));
  }
}

async function rollbackHARCapture(capture, forceDetach) {
  if (!capture) return;
  capture.stopping = true;
  stopHARRouting(capture);
  await waitForHAREnrollments(capture);
  stopHARRouting(capture);
  const attachState = sessionAttachState.get(capture.sessionId);
  for (const state of capture.tabs.values()) {
    const routed = harTabs.get(state.tabId);
    if (routed && routed.capture === capture) harTabs.delete(state.tabId);
    if (forceDetach || !state.wasAttached) {
      if (attachState && attachState.attachedTabIds) attachState.attachedTabIds.delete(state.tabId);
      await detachDebugger(state.tabId, true).catch(() => {});
    }
  }
  capture.tabs.clear();
  harCaptures.delete(capture.sessionId);
  if (harWindowOwners.get(capture.windowId) === capture.sessionId) {
    harWindowOwners.delete(capture.windowId);
  }
}

function freezeHARSnapshot(value) {
  if (!value || typeof value !== "object" || Object.isFrozen(value)) return value;
  for (const child of Object.values(value)) freezeHARSnapshot(child);
  return Object.freeze(value);
}

function harTabResult(state) {
  if (state.snapshotResult) return state.snapshotResult;
  const har = state.recorder.buildHAR();
  const entries = (har.log && har.log.entries) || [];
  const recorderErrors = state.recorder.errors || [];
  const errors = state.errors.concat(recorderErrors.map((item) => item.error || String(item)));
  return {
    tab_id: state.tabId,
    title: state.title || "",
    url: state.url || "",
    state: state.state,
    request_count: entries.length,
    error_count: errors.length,
    errors,
  };
}

function snapshotHARTab(state) {
  if (state.snapshot) return;
  state.snapshot = freezeHARSnapshot(state.recorder.buildHAR());
  state.snapshotJSON = JSON.stringify(state.snapshot);
  const entries = (state.snapshot.log && state.snapshot.log.entries) || [];
  const recorderErrors = state.recorder.errors || [];
  const errors = state.errors.concat(recorderErrors.map((item) => item.error || String(item)));
  state.snapshotResult = freezeHARSnapshot({
    tab_id: state.tabId,
    title: state.title || "",
    url: state.url || "",
    state: state.state,
    request_count: entries.length,
    error_count: errors.length,
    errors,
  });
}

async function finalizeHARCapture(capture) {
  if (capture.snapshotsReady) return;
  if (capture.finalizePromise) return capture.finalizePromise;
  capture.stopping = true;
  stopHARRouting(capture);
  capture.finalizePromise = (async () => {
    await waitForHAREnrollments(capture);
    stopHARRouting(capture);
    const states = Array.from(capture.tabs.values());
    const flushes = await Promise.all(
      states.map((state) => state.recorder.flush(HAR_BODY_SETTLE_MS)),
    );
    for (let index = 0; index < states.length; index += 1) {
      const state = states[index];
      const flush = flushes[index] || {};
      if (flush.partial) capture.partial = true;
      if (flush.timedOut) {
        capture.warnings.push(
          "tab " + state.tabId + " response body flush timed out with " +
            (flush.pendingBodyCount || 0) + " pending request(s)",
        );
      }
      snapshotHARTab(state);
    }
    capture.snapshotsReady = true;
  })();
  try {
    await capture.finalizePromise;
  } finally {
    capture.finalizePromise = null;
  }
}

async function uploadHARTab(capture, state) {
  if (state.uploaded) return;
  if (state.uploadPromise) return state.uploadPromise;
  if (!state.snapshot || !state.snapshotJSON) {
    throw new Error("HAR snapshot for tab " + state.tabId + " is unavailable");
  }
  state.uploadPromise = (async () => {
    const url = new URL(capture.uploadURL);
    url.searchParams.set("session", capture.sessionId);
    url.searchParams.set("capture", capture.captureId);
    url.searchParams.set("token", capture.artifactToken);
    url.searchParams.set("tab_id", String(state.tabId));
    const controller = new AbortController();
    const timeout = setTimeout(() => controller.abort(), HAR_UPLOAD_TIMEOUT_MS);
    let response;
    try {
      response = await fetch(url.toString(), {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: state.snapshotJSON,
        signal: controller.signal,
      });
    } catch (error) {
      if (controller.signal.aborted) {
        throw new Error("upload HAR for tab " + state.tabId + " timed out");
      }
      throw error;
    } finally {
      clearTimeout(timeout);
    }
    if (!response.ok) {
      const detail = await response.text().catch(() => "");
      const error = new Error("upload HAR for tab " + state.tabId + " failed: HTTP " + response.status + (detail ? " " + detail.trim() : ""));
      error.harPermanentUploadFailure = response.status >= 400 && response.status < 500;
      throw error;
    }
    state.uploaded = true;
  })();
  try {
    await state.uploadPromise;
  } finally {
    state.uploadPromise = null;
  }
}

async function handleHARStartJob(params, sessionId) {
  if (typeof BrowserAgentHAR === "undefined" || typeof BrowserAgentHAR.Recorder !== "function") {
    throw new Error("HAR recorder module is unavailable");
  }
  if (harCaptures.has(sessionId)) {
    throw new Error("HAR capture is already active for this session");
  }
  const entry = sessions.get(sessionId);
  if (!entry || entry.windowId == null) {
    throw new Error("session page not bound (windowId missing); open /go?session=" + sessionId);
  }
  const owner = harWindowOwners.get(entry.windowId);
  if (owner && owner !== sessionId) {
    throw new Error("Chrome window " + entry.windowId + " already has an active HAR capture owned by session " + owner);
  }
  const captureId = String(params.capture_id || "");
  const artifactToken = String(params.artifact_token || "");
  const rawUploadURL = String(params.artifact_upload_url || "");
  if (!captureId || !artifactToken || !rawUploadURL) {
    throw new Error("HAR start is missing daemon artifact parameters");
  }
  const controlPort = entry.controlPort || CONTROL_PORT;
  const uploadURL = validateHARUploadURL(rawUploadURL, controlPort);
  const capture = {
    sessionId,
    captureId,
    artifactToken,
    uploadURL,
    controlPort,
    windowId: entry.windowId,
    startedAt: new Date().toISOString(),
    tabs: new Map(),
    enrolling: new Map(),
    enrollmentFailures: new Set(),
    warnings: [],
    partial: false,
    stopping: false,
    snapshotsReady: false,
    finalizePromise: null,
  };
  harCaptures.set(sessionId, capture);
  harWindowOwners.set(entry.windowId, sessionId);
  try {
    const tabs = await chrome.tabs.query({ windowId: entry.windowId });
    for (const tab of tabs || []) {
      if (!isHAREligibleTab(tab, capture)) continue;
      await enrollHARTab(capture, tab, true);
    }
  } catch (error) {
    await rollbackHARCapture(capture);
    throw error;
  }
  return {
    type: "har_start",
    capture_id: captureId,
    window_id: capture.windowId,
    tab_count: capture.tabs.size,
    tabs: Array.from(capture.tabs.values()).map(harTabResult),
  };
}

async function handleHAREndJob(params, sessionId) {
  const capture = harCaptures.get(sessionId);
  if (!capture) throw new Error("no active HAR capture for this session");
  if (String(params.capture_id || "") !== capture.captureId) {
    throw new Error("HAR capture id does not match the active capture");
  }
  capture.stopping = true;
  stopHARRouting(capture);
  await finalizeHARCapture(capture);
  try {
    for (const state of capture.tabs.values()) {
      await uploadHARTab(capture, state);
    }
  } catch (error) {
    if (error && error.harPermanentUploadFailure) {
      error.harCaptureAborted = true;
      error.harCaptureId = capture.captureId;
      await rollbackHARCapture(capture, false);
    }
    throw error;
  }
  const tabs = Array.from(capture.tabs.values()).map(harTabResult);
  const result = {
    type: "har_end",
    capture_id: capture.captureId,
    window_id: capture.windowId,
    tab_count: tabs.length,
    partial: capture.partial || tabs.some((tab) => tab.error_count > 0),
    warnings: capture.warnings.slice(),
    tabs,
  };
  await rollbackHARCapture(capture);
  return result;
}

async function updateHARTabFromChrome(tabId, changeInfo, tab) {
  for (const capture of harCaptures.values()) {
    if (!tab || tab.windowId !== capture.windowId || capture.stopping) continue;
    const state = capture.tabs.get(tabId);
    if (state) {
      if (tab.title) state.title = tab.title;
      if (tab.url || tab.pendingUrl) state.url = tab.url || tab.pendingUrl;
      continue;
    }
    if (isHAREligibleTab(tab, capture)) {
      await enrollHARTab(capture, tab, false);
    }
  }
}

function closeHARTab(tabId) {
  const routed = harTabs.get(tabId);
  if (!routed) return;
  routed.state.state = "closed";
  harTabs.delete(tabId);
}

async function handleJob(msg, sessionId, entry) {
  const payload = msg.payload || {};
  const jobSessionId = payload.session_id || payload.sessionId || sessionId || "";
  const jobId = payload.job_id || payload.id || msg.id || "";
  const jobType = payload.type || payload.job_type || "eval";
  const params = payload.params || {};

  if (!claimJobExecution(entry, jobId)) return;

  const targetOpts = {
    tabId: payload.tab_id != null ? payload.tab_id : params.tab_id,
    tabIndex: payload.tab_index != null ? payload.tab_index : params.tab_index,
  };

  const t0 = Date.now();
  if (baShouldLogJob(jobType, params)) {
    baLog("log", "job start", {
      job_id: jobId,
      session_id: jobSessionId,
      type: jobType,
      tab_id: targetOpts.tabId,
      tab_index: targetOpts.tabIndex,
      window_id: entry && entry.windowId,
      params: baSummarizeParams(jobType, params),
    });
  }

  try {
    let data;
    switch (jobType) {
      case "info":
        data = await handleInfoJob(params, jobSessionId);
        break;
      case "eval":
        data = await handleEvalJob(params, payload, jobSessionId, targetOpts);
        break;
      case "run":
        data = await handleRunJob(params, payload, jobSessionId, targetOpts);
        break;
      case "logs":
        data = await handleLogsJob(params, jobSessionId, targetOpts);
        break;
      case "screenshot":
        data = await handleScreenshotJob(params, jobSessionId, targetOpts);
        break;
      case "cdp":
        data = await handleCdpJob(params, jobSessionId, targetOpts);
        break;
      case "create_tab":
        data = await handleCreateTabJob(params, jobSessionId);
        break;
      case "har_start":
        data = await handleHARStartJob(params, jobSessionId);
        break;
      case "har_end":
        data = await handleHAREndJob(params, jobSessionId);
        break;
      default:
        baLog("error", "job unknown type", { job_id: jobId, type: jobType });
        sendJobResult(entry, jobId, false, { type: jobType }, "unknown job type: " + jobType);
        return;
    }
    if (baShouldLogJob(jobType, params)) {
      baLog("log", "job ok", {
        job_id: jobId,
        type: jobType,
        elapsed_ms: Date.now() - t0,
        data: {
          type: data && data.type,
          tab_id: data && data.tab_id,
          url: data && data.url,
          method: data && data.method,
        },
      });
    }
    sendJobResult(entry, jobId, true, data, "");
  } catch (e) {
    const errMsg = e && e.message ? e.message : String(e);
    baLog("error", "job fail", {
      job_id: jobId,
      type: jobType,
      elapsed_ms: Date.now() - t0,
      error: errMsg,
      params: baSummarizeParams(jobType, params),
    });
    const failureData = { type: jobType, stub: false };
    if (e && e.harCaptureAborted) {
      failureData.capture_aborted = true;
      failureData.capture_id = e.harCaptureId || String(params.capture_id || "");
    }
    sendJobResult(entry, jobId, false, failureData, errMsg);
  } finally {
    finishJobExecution(entry, jobId);
  }
}

async function handleInfoJob(_params, sessionId) {
  const entry = sessions.get(sessionId);
  const windowId = entry && entry.windowId != null ? entry.windowId : null;
  const capturableTabs = await listCapturableTabsInSessionWindow(sessionId, entry);
  const tabsOut = capturableTabs.map((t, i) => ({
    index: i + 1,
    id: t.id,
    url: t.url || "",
    title: t.title || "",
    active: !!t.active,
    role: isSessionGoPageURL(t.url || "", sessionId) ? "session_page" : "user",
  }));

  let jobTarget = { tab_id: null, tab_index: null, reason: "session_page_fallback" };
  let recommendedCLI = "";
  try {
    const tabId = await pickTargetTabIdForSession(sessionId, {});
    const idx = tabsOut.findIndex((t) => t.id === tabId);
    let reason = "active_in_session_window";
    const activeTab = tabsOut.find((t) => t.active);
    if (activeTab && activeTab.id === tabId) {
      reason = "active_in_session_window";
    } else if (idx >= 0 && tabsOut[idx].role === "session_page") {
      reason = "session_page_fallback";
    }
    jobTarget = {
      tab_id: tabId,
      tab_index: idx >= 0 ? idx + 1 : null,
      reason: reason,
    };
    recommendedCLI = "browser-agent session eval --tab-id " + tabId + " '...'";
  } catch (e) {
    /* no capturable tab yet */
  }

  return {
    version: EXT_VERSION,
    features: FEATURES,
    controlPort: entry && entry.controlPort != null ? entry.controlPort : CONTROL_PORT,
    window_id: windowId,
    tabs: tabsOut,
    job_target: jobTarget,
    recommended_cli: recommendedCLI,
  };
}

async function handleEvalJob(params, payload, sessionId, targetOpts) {
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
  }, targetOpts);
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

async function handleRunJob(params, payload, sessionId, targetOpts) {
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
  }, targetOpts);
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

async function handleLogsJob(params, sessionId, targetOpts) {
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
    }, targetOpts);
  } catch (e) {
    /* ignore attach failures for logs */
  }
  return { entries: entries, type: "logs" };
}

async function handleScreenshotJob(params, sessionId, targetOpts) {
  const format = (params && params.format) || "png";
  const fullPage = !!(params && params.full_page);
  const result = await withDebuggerForSession(sessionId, async (tabId) => {
    const cdpParams = { format: format === "jpeg" ? "jpeg" : "png" };
    if (fullPage) {
      cdpParams.captureBeyondViewport = true;
    }
    return await sendDebuggerCommand(tabId, "Page.captureScreenshot", cdpParams);
  }, targetOpts);
  return {
    base64: (result && result.data) || "",
    format: format,
    type: "screenshot",
    cdp: "Page.captureScreenshot",
  };
}

async function handleCdpJob(params, sessionId, targetOpts) {
  const method = (params && (params.method || params.cdp_method || params.cdpMethod)) || "";
  if (!method) {
    throw new Error("cdp job requires params.method");
  }
  const cdpParams = (params && params.params) || {};
  if (baShouldLogJob("cdp", params)) {
    baLog("log", "cdp begin", {
      session_id: sessionId,
      method: method,
      nav_url: cdpParams && cdpParams.url,
      tab_id: targetOpts && (targetOpts.tabId != null ? targetOpts.tabId : targetOpts.tab_id),
    });
  }
  // Intercept all Target.* — polyfill via chrome.tabs; never raw debugger sendCommand.
  if (typeof method === "string" && method.startsWith("Target.")) {
    const polyfilled = await polyfillTargetMethod(method, cdpParams || {}, sessionId, targetOpts);
    return {
      result: polyfilled,
      method: method,
      type: "cdp",
      polyfilled: true,
    };
  }
  const result = await withDebuggerForSession(sessionId, async (tabId) => {
    if (method === "Page.navigate") {
      baLog("log", "cdp Page.navigate target", {
        tab_id: tabId,
        url: cdpParams && cdpParams.url,
      });
    } else if (baDebugEnabled()) {
      baLog("log", "cdp sendCommand", { tab_id: tabId, method: method });
    }
    return await sendDebuggerCommand(tabId, method, cdpParams || {});
  }, targetOpts);
  return {
    result: result,
    method: method,
    type: "cdp",
  };
}

/**
 * Shared create path for job create_tab and Target.createTarget.
 * Always scopes to the session window (entry.windowId). Default active: true.
 * Public identity: tab_id only (no targetId).
 */
async function createTabInSession(sessionId, opts) {
  opts = opts || {};
  const entry = sessions.get(sessionId);
  if (!entry || entry.windowId == null) {
    baLog("error", "create_tab no windowId", { session_id: sessionId });
    throw new Error(
      "session page not bound (windowId missing); open /go?session=" + (sessionId || ""),
    );
  }
  const createProps = {
    windowId: entry.windowId,
  };
  // Default active:true when unspecified.
  if (opts.active === false || opts.active === "false" || opts.active === 0) {
    createProps.active = false;
  } else {
    createProps.active = true;
  }
  const url = opts.url != null ? String(opts.url).trim() : "";
  if (url) {
    createProps.url = url;
  }
  baLog("log", "create_tab chrome.tabs.create", {
    session_id: sessionId,
    window_id: entry.windowId,
    url: url,
    active: createProps.active,
  });
  const tab = await new Promise((resolve, reject) => {
    chrome.tabs.create(createProps, (created) => {
      if (chrome.runtime.lastError) {
        reject(new Error(chrome.runtime.lastError.message || "chrome.tabs.create failed"));
        return;
      }
      resolve(created);
    });
  });
  baLog("log", "create_tab created", {
    tab_id: tab && tab.id,
    url: (tab && tab.url) || url || "",
    pendingUrl: tab && tab.pendingUrl,
    status: tab && tab.status,
  });
  // P2 eager arm: attach new capturable tab in the session window immediately.
  if (tab && tab.id != null) {
    const createdUrl = (tab && tab.url) || url || (tab && tab.pendingUrl) || "";
    maybeEagerAttach(sessionId, tab.id, createdUrl, entry.windowId);
  }
  return {
    type: "create_tab",
    tab_id: tab && tab.id != null ? tab.id : null,
    url: (tab && tab.url) || url || "",
    window_id: entry.windowId,
  };
}

async function handleCreateTabJob(params, sessionId) {
  return await createTabInSession(sessionId, {
    url: params && (params.url || params.URL || params.href),
    active: params && params.active,
  });
}

/** Resolve tab_id from extension-shaped params or decimal CDP targetId. */
function resolveTabIdFromParams(params) {
  params = params || {};
  if (params.tab_id != null && params.tab_id !== "") {
    const n = parseInt(params.tab_id, 10);
    if (!Number.isNaN(n) && n > 0) return n;
  }
  if (params.tabId != null && params.tabId !== "") {
    const n = parseInt(params.tabId, 10);
    if (!Number.isNaN(n) && n > 0) return n;
  }
  // Inbound CDP may send targetId as decimal string of a chrome tab id.
  if (params.targetId != null && params.targetId !== "") {
    const n = parseInt(String(params.targetId), 10);
    if (!Number.isNaN(n) && n > 0 && String(n) === String(params.targetId).trim()) {
      return n;
    }
  }
  throw new Error(
    "unable to resolve tab_id from params (need tab_id or decimal targetId string)",
  );
}

function requireSessionWindowId(sessionId) {
  const entry = sessions.get(sessionId);
  if (!entry || entry.windowId == null) {
    throw new Error(
      "session page not bound (windowId missing); open /go?session=" + (sessionId || ""),
    );
  }
  return entry;
}

/**
 * Polyfill dispatch for all Target.* CDP methods.
 * Tier A: full chrome.tabs implementation.
 * Tier B: soft no-op / debugger attach map.
 * Tier C: explicit polyfill-unsupported product error (never Chrome -32000 fallthrough).
 */
async function polyfillTargetMethod(method, params, sessionId, targetOpts) {
  params = params || {};
  switch (method) {
    case "Target.createTarget": {
      const data = await createTabInSession(sessionId, {
        url: params.url || params.URL,
        active: params.active != null ? params.active : params.background === true ? false : undefined,
      });
      return data;
    }
    case "Target.closeTarget": {
      return await polyfillCloseTarget(params, sessionId);
    }
    case "Target.activateTarget": {
      const tabId = resolveTabIdFromParams(params);
      await new Promise((resolve, reject) => {
        chrome.tabs.update(tabId, { active: true }, (tab) => {
          if (chrome.runtime.lastError) {
            reject(new Error(chrome.runtime.lastError.message || "chrome.tabs.update failed"));
            return;
          }
          resolve(tab);
        });
      });
      return { tab_id: tabId, active: true };
    }
    case "Target.getTargets": {
      return await polyfillGetTargets(sessionId);
    }
    case "Target.getTargetInfo": {
      return await polyfillGetTargetInfo(params, sessionId);
    }
    case "Target.setDiscoverTargets":
    case "Target.setAutoAttach":
      return { polyfilled: true, method: method };
    case "Target.attachToTarget": {
      const entry = requireSessionWindowId(sessionId);
      const tabId = resolveTabIdFromParams(params);
      const tab = await chrome.tabs.get(tabId);
      if (entry.windowId != null && tab.windowId !== entry.windowId) {
        throw new Error("tab_id " + tabId + " is not in session window " + entry.windowId);
      }
      await attachDebuggerForSession(sessionId, tabId);
      return { tab_id: tabId, polyfilled: true };
    }
    case "Target.detachFromTarget": {
      let tabId = null;
      try {
        tabId = resolveTabIdFromParams(params);
      } catch (e) {
        // If no tab id, detach one tab from the session attach set if any.
        const state = sessionAttachState.get(sessionId);
        if (state && state.attachedTabIds && state.attachedTabIds.size > 0) {
          tabId = state.attachedTabIds.values().next().value;
        }
      }
      if (tabId != null) {
        await detachDebugger(tabId);
        const state = sessionAttachState.get(sessionId);
        if (state && state.attachedTabIds) {
          state.attachedTabIds.delete(tabId);
        }
      }
      return { tab_id: tabId, polyfilled: true };
    }
    default:
      throw new Error(
        "Target method polyfill unsupported: " + method + " (not implemented via chrome.tabs)",
      );
  }
}

async function polyfillCloseTarget(params, sessionId) {
  const tabId = resolveTabIdFromParams(params);
  const entry = sessions.get(sessionId);
  let tab = null;
  try {
    tab = await chrome.tabs.get(tabId);
  } catch (e) {
    throw new Error("closeTarget: tab_id " + tabId + " not found");
  }
  // Never close the session control page (/go?session=<id>).
  if (isSessionGoPageURL(tab.url || "", sessionId)) {
    throw new Error("refusing to close session control page (/go?session=)");
  }
  if (entry && entry.tabId != null && entry.tabId === tabId) {
    throw new Error("refusing to close session-page tab_id " + tabId);
  }
  if (entry && entry.windowId != null && tab.windowId !== entry.windowId) {
    throw new Error("tab_id " + tabId + " is not in session window " + entry.windowId);
  }
  await new Promise((resolve, reject) => {
    chrome.tabs.remove(tabId, () => {
      if (chrome.runtime.lastError) {
        reject(new Error(chrome.runtime.lastError.message || "chrome.tabs.remove failed"));
        return;
      }
      resolve();
    });
  });
  return { tab_id: tabId, success: true };
}

async function polyfillGetTargets(sessionId) {
  const entry = requireSessionWindowId(sessionId);
  const tabs = await new Promise((resolve, reject) => {
    chrome.tabs.query({ windowId: entry.windowId }, (list) => {
      if (chrome.runtime.lastError) {
        reject(new Error(chrome.runtime.lastError.message || "chrome.tabs.query failed"));
        return;
      }
      resolve(list || []);
    });
  });
  const targetInfos = (tabs || []).map((t) => ({
    tab_id: t.id,
    url: t.url || "",
    title: t.title || "",
    active: !!t.active,
    role: isSessionGoPageURL(t.url || "", sessionId) ? "session_page" : "user",
  }));
  return { targetInfos: targetInfos, tab_id: null };
}

async function polyfillGetTargetInfo(params, sessionId) {
  const tabId = resolveTabIdFromParams(params);
  const entry = sessions.get(sessionId);
  const tab = await chrome.tabs.get(tabId);
  if (entry && entry.windowId != null && tab.windowId !== entry.windowId) {
    throw new Error("tab_id " + tabId + " is not in session window " + entry.windowId);
  }
  return {
    tab_id: tab.id,
    url: tab.url || "",
    title: tab.title || "",
    active: !!tab.active,
    role: isSessionGoPageURL(tab.url || "", sessionId) ? "session_page" : "user",
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

async function listCapturableTabsInSessionWindow(sessionId, entry) {
  if (entry && entry.windowId != null) {
    try {
      const tabs = await chrome.tabs.query({ windowId: entry.windowId });
      return (tabs || []).filter((t) => t.id != null && isCapturableTabURL(t.url || ""));
    } catch (e) {
      /* window query optional */
    }
  }
  try {
    const tabs = await chrome.tabs.query({});
    return (tabs || []).filter((t) => t.id != null && isCapturableTabURL(t.url || ""));
  } catch (e) {
    return [];
  }
}

/**
 * Resolve job target tab for a session.
 * Priority: payload.tab_id -> tab_index (1-based capturable) -> active tab -> session page.
 */
async function pickTargetTabIdForSession(sessionId, opts) {
  opts = opts || {};
  const entry = sessions.get(sessionId);
  const explicitTabId =
    opts.tabId != null ? opts.tabId : opts.tab_id != null ? opts.tab_id : null;
  const explicitTabIndex =
    opts.tabIndex != null ? opts.tabIndex : opts.tab_index != null ? opts.tab_index : null;

  // 1. Explicit tab_id from job payload (--tab-id flag).
  if (explicitTabId != null) {
    const tabId = parseInt(explicitTabId, 10);
    const tab = await chrome.tabs.get(tabId);
    if (entry && entry.windowId != null && tab.windowId !== entry.windowId) {
      throw new Error("tab_id " + tabId + " is not in session window " + entry.windowId);
    }
    if (!isCapturableTabURL(tab.url || "")) {
      throw new Error("tab_id " + tabId + " is not capturable");
    }
    return tab.id;
  }

  // 2. tab_index: 1-based index over capturable tabs in session window (left-to-right).
  if (explicitTabIndex != null) {
    const capturable = await listCapturableTabsInSessionWindow(sessionId, entry);
    const idx = parseInt(explicitTabIndex, 10);
    if (Number.isNaN(idx) || idx < 1 || idx > capturable.length) {
      throw new Error(
        "tab_index " + explicitTabIndex + " out of range (1-" + capturable.length + ")",
      );
    }
    return capturable[idx - 1].id;
  }

  // 3. Active capturable tab in the session window (multi-session safe).
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

  // 4. Fallback: registered session control page tab.
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

  try {
    const tabs = await chrome.tabs.query({});
    for (const t of tabs || []) {
      if (t.id == null) continue;
      const url = t.url || "";
      if (!isCapturableTabURL(url)) continue;
      if (isSessionGoPageURL(url, sessionId)) return t.id;
    }
  } catch (e) {
    /* ignore */
  }
  throw new Error("no capturable tab for session " + sessionId);
}

function isAnotherDebuggerAttachedError(err) {
  const msg = String(err && err.message ? err.message : err || "");
  return /another debugger is already attached/i.test(msg);
}

function anotherDebuggerAttachedHint(tabId) {
  return (
    "Another debugger is already attached to the tab with id: " +
    tabId +
    ". Close DevTools on that tab, or reload the Browser Agent extension, then retry."
  );
}

/** List tab ids this extension currently owns via chrome.debugger (survives SW restart). */
function debuggerGetTargets() {
  return new Promise((resolve) => {
    if (!chrome.debugger || typeof chrome.debugger.getTargets !== "function") {
      resolve([]);
      return;
    }
    chrome.debugger.getTargets((targets) => {
      if (chrome.runtime.lastError) {
        resolve([]);
        return;
      }
      resolve(Array.isArray(targets) ? targets : []);
    });
  });
}

/**
 * Adopt Chrome-owned debugger targets into attachedTabs.
 * After MV3 SW death the in-memory map is empty while Chrome still holds attaches
 * for this extension — jobs then fail with "Another debugger is already attached".
 */
async function adoptDebuggerTargetsFromChrome(reason) {
  const targets = await debuggerGetTargets();
  let adopted = 0;
  for (const t of targets) {
    if (!t || !t.attached || t.tabId == null) continue;
    if (!attachedTabs.has(t.tabId)) {
      attachedTabs.set(t.tabId, true);
      adopted += 1;
    }
  }
  if (adopted > 0) {
    baLog("log", "adoptDebuggerTargetsFromChrome", {
      reason: reason || "",
      adopted,
      attached_tabs: attachedTabs.size,
      targets: targets.length,
    });
    // Keep idle-detach armed so adopted orphans do not freeze popup UX forever.
    touchDebuggerActivity();
  }
  return adopted;
}

function detachDebugger(tabId, force) {
  return new Promise((resolve) => {
    if (isHARCapturedTab(tabId) && !force) {
      resolve(false);
      return;
    }
    // force=true: detach even when attachedTabs lost the id (SW orphan / reclaim).
    if (!force && !attachedTabs.has(tabId)) {
      resolve(false);
      return;
    }
    chrome.debugger.detach({ tabId: tabId }, () => {
      attachedTabs.delete(tabId);
      resolve(true);
    });
  });
}

/**
 * Attach chrome.debugger to a single tabId (CDP transport for jobs).
 * Side effect: Chrome may show a browser-wide "debugging this browser" notice
 * on unrelated tabs/windows while this attach remains live — not multi-tab CDP.
 *
 * Reclaim: if attach fails with "Another debugger is already attached", adopt
 * targets we already own (SW restart) or best-effort detach+retry once. If still
 * blocked (e.g. DevTools), fail with an actionable hint.
 */
function attachDebugger(tabId) {
  if (attachedTabs.has(tabId)) {
    baLog("log", "attachDebugger reuse", { tab_id: tabId, already_attached: true });
    return Promise.resolve(true);
  }
  const t0 = baNow();
  let cancelled = false;
  baLog("log", "attachDebugger start", {
    tab_id: tabId,
    protocol: CDP_PROTOCOL_VERSION,
    timeout_ms: ATTACH_TIMEOUT_MS,
  });

  const enableRuntime = () =>
    new Promise((resolve) => {
      const t1 = baNow();
      chrome.debugger.sendCommand({ tabId: tabId }, "Runtime.enable", {}, () => {
        if (cancelled) {
          chrome.debugger.detach({ tabId: tabId }, () => {
            attachedTabs.delete(tabId);
          });
          return;
        }
        const enableMs = Math.round(baNow() - t1);
        const totalMs = Math.round(baNow() - t0);
        const enableErr =
          chrome.runtime.lastError && chrome.runtime.lastError.message
            ? chrome.runtime.lastError.message
            : "";
        baLog("log", "attachDebugger done", {
          tab_id: tabId,
          attach_ms: Math.round(baNow() - t0),
          runtime_enable_ms: enableMs,
          total_ms: totalMs,
          runtime_enable_error: enableErr || null,
        });
        // Still resolve so a non-fatal Runtime.enable glitch does not strand attachLock.
        resolve(true);
      });
    });

  const tryAttachOnce = () =>
    new Promise((resolve, reject) => {
      chrome.debugger.attach({ tabId: tabId }, CDP_PROTOCOL_VERSION, () => {
        if (cancelled) {
          // Late attach after timeout: drop it so we do not leave a stuck debugger.
          chrome.debugger.detach({ tabId: tabId }, () => {
            attachedTabs.delete(tabId);
          });
          return;
        }
        const attachMs = Math.round(baNow() - t0);
        if (chrome.runtime.lastError) {
          baLog("error", "attachDebugger attach fail", {
            tab_id: tabId,
            attach_ms: attachMs,
            error: chrome.runtime.lastError.message,
          });
          reject(new Error(chrome.runtime.lastError.message || "chrome.debugger.attach failed"));
          return;
        }
        attachedTabs.set(tabId, true);
        enableRuntime().then(resolve, reject);
      });
    });

  const attachPromise = (async () => {
    // Adopt orphans before the first attach attempt (cheap getTargets).
    await adoptDebuggerTargetsFromChrome("pre-attach");
    if (cancelled) return true;
    if (attachedTabs.has(tabId)) {
      baLog("log", "attachDebugger reuse after adopt", { tab_id: tabId });
      await enableRuntime();
      return true;
    }
    try {
      await tryAttachOnce();
      return true;
    } catch (err) {
      if (cancelled) return true;
      if (!isAnotherDebuggerAttachedError(err)) throw err;

      // We may already own this target after SW restart but missed it on pre-attach.
      await adoptDebuggerTargetsFromChrome("reclaim");
      if (attachedTabs.has(tabId)) {
        baLog("log", "attachDebugger reclaim via getTargets", { tab_id: tabId });
        await enableRuntime();
        return true;
      }

      // Best-effort detach + one retry (stale attach we can release).
      await detachDebugger(tabId, true);
      if (cancelled) return true;
      try {
        await tryAttachOnce();
        baLog("log", "attachDebugger reclaim via detach+retry", { tab_id: tabId });
        return true;
      } catch (err2) {
        if (isAnotherDebuggerAttachedError(err2)) {
          throw new Error(anotherDebuggerAttachedHint(tabId));
        }
        throw err2;
      }
    }
  })();

  return baWithTimeout(attachPromise, ATTACH_TIMEOUT_MS, "chrome.debugger.attach", () => {
    cancelled = true;
    baLog("error", "attachDebugger timeout cleanup", {
      tab_id: tabId,
      timeout_ms: ATTACH_TIMEOUT_MS,
      had_partial_attach: attachedTabs.has(tabId),
    });
    // Best-effort detach if attach partially succeeded but Runtime.enable hung.
    return detachDebugger(tabId);
  });
}

/** Tabs with eager/job attach already in-flight (dedupe stampede). */
const eagerAttachInFlight = new Set();
/**
 * Idle detach: holding chrome.debugger slows the toolbar default_popup.
 * Keep content-tab attach warm for multi-step agent SPA work; only detach after
 * 5 minutes without debugger activity (was 3s — thrash / mid-job detach).
 */
const DEBUGGER_IDLE_DETACH_MS = 5 * 60 * 1000;
/** @type {ReturnType<typeof setTimeout>|null} */
let debuggerIdleDetachTimer = null;

function idleDebuggerLeftoverTabIDs() {
  return Array.from(attachedTabs.keys()).filter((tabId) => !isHAROwnedTab(tabId));
}

function scheduleDebuggerIdleDetach() {
  if (debuggerIdleDetachTimer) {
    clearTimeout(debuggerIdleDetachTimer);
  }
  debuggerIdleDetachTimer = setTimeout(() => {
    debuggerIdleDetachTimer = null;
    // Detach all session debuggers so toolbar default_popup is instant again.
    // WS sessions stay; next job/create_tab re-attaches.
    const sessionIds = Array.from(sessionAttachState.keys());
    if (sessionIds.length === 0 && attachedTabs.size === 0) return;
    baLog("log", "idle detach debugger (popup UX)", {
      sessions: sessionIds.length,
      attached_tabs: attachedTabs.size,
    });
    Promise.all(
      sessionIds.map((id) => detachSessionDebugger(id).catch(() => {})),
    ).then(() => {
      // Safety: clear any orphans not tracked in session state.
      const leftovers = idleDebuggerLeftoverTabIDs();
      return Promise.all(leftovers.map((tid) => detachDebugger(tid).catch(() => {})));
    });
  }, DEBUGGER_IDLE_DETACH_MS);
}

function touchDebuggerActivity() {
  scheduleDebuggerIdleDetach();
}

/**
 * P2 eager arm helper: auto-attach a capturable **content** tab in an armed window.
 * Never attaches the /go session control page (attach there freezes popup UX).
 * Skips non-capturable URLs and other windows. Dedupe in-flight/already attached.
 */
function maybeEagerAttach(sessionId, tabId, url, windowId) {
  if (!sessionId || tabId == null) return Promise.resolve();
  // Never debugger-attach the session control page.
  if (typeof url === "string" && url.length > 0 && isSessionGoPageURL(url, sessionId)) {
    return Promise.resolve();
  }
  // Skip non-capturable when URL is known (chrome://, chrome-extension://, etc.).
  if (typeof url === "string" && url.length > 0 && !isCapturableTabURL(url)) {
    return Promise.resolve();
  }
  const entry = sessions.get(sessionId);
  if (!entry || entry.windowId == null) return Promise.resolve();
  // Scope to session window — never auto-attach tabs in other windows.
  if (windowId != null && entry.windowId !== windowId) {
    return Promise.resolve();
  }
  // Already a live CDP debuggee — nothing to do.
  if (attachedTabs.has(tabId)) {
    const state = sessionAttachState.get(sessionId);
    if (state && state.attachedTabIds) state.attachedTabIds.add(tabId);
    touchDebuggerActivity();
    return Promise.resolve();
  }
  if (eagerAttachInFlight.has(tabId)) {
    return Promise.resolve();
  }
  eagerAttachInFlight.add(tabId);
  touchDebuggerActivity();
  return attachDebuggerForSession(sessionId, tabId)
    .catch(() => {})
    .finally(() => {
      eagerAttachInFlight.delete(tabId);
      touchDebuggerActivity();
    });
}

/**
 * Session-scoped attach: gate on open /go?session= in the target tab's window,
 * then add tabId to the multi-tab attach set (policy B — keep peer attaches).
 * Chrome's "debugging this browser" banner can still appear outside that window;
 * only tabs in the session attach set receive sendCommand / eval / screenshot.
 */
async function attachDebuggerForSession(sessionId, tabId) {
  let state = sessionAttachState.get(sessionId);
  if (!state) {
    state = { attachedTabIds: new Set(), attachLock: Promise.resolve() };
    sessionAttachState.set(sessionId, state);
  }
  if (!state.attachedTabIds) {
    state.attachedTabIds = new Set();
  }
  const queuedAt = baNow();
  baLog("log", "attachDebuggerForSession enqueue", {
    session_id: sessionId,
    tab_id: tabId,
    attach_set_size: state.attachedTabIds.size,
    will_wait_lock: true,
    timeout_ms: ATTACH_TIMEOUT_MS,
  });
  const run = async () => {
    const lockWaitMs = Math.round(baNow() - queuedAt);
    // Cap lock-hold work so a stuck attach cannot block the session forever.
    const work = async () => {
      // Attach gate: require ≥1 open /go?session=<id> control tab in the same window.
      let windowId = null;
      let tabUrl = "";
      try {
        const t = await chrome.tabs.get(tabId);
        if (t && t.windowId != null) windowId = t.windowId;
        tabUrl = (t && t.url) || "";
      } catch (e) {
        /* tab may be gone */
      }
      if (windowId == null) {
        const entry = sessions.get(sessionId);
        if (entry && entry.windowId != null) windowId = entry.windowId;
      }
      if (windowId == null || !(await hasOpenSessionPage(sessionId, windowId))) {
        // Drop all leftover attaches so jobs cannot reuse orphaned debuggers.
        const leftover = Array.from(state.attachedTabIds);
        state.attachedTabIds.clear();
        for (const id of leftover) {
          await detachDebugger(id);
        }
        baLog("error", "attachDebuggerForSession gate fail", {
          session_id: sessionId,
          tab_id: tabId,
          lock_wait_ms: lockWaitMs,
          window_id: windowId,
        });
        throw new Error(
          "no open session page for attach; session control tab required in the same window",
        );
      }

      // Policy B: keep peer tabs attached; only attach/add the requested tabId.
      // Same tab already attached → attachDebugger reuses via attachedTabs.has.
      const tAttach = baNow();
      await attachDebugger(tabId);
      state.attachedTabIds.add(tabId);
      baLog("log", "attachDebuggerForSession done", {
        session_id: sessionId,
        tab_id: tabId,
        tab_url: String(tabUrl).slice(0, 120),
        lock_wait_ms: lockWaitMs,
        attach_call_ms: Math.round(baNow() - tAttach),
        attach_set_size: state.attachedTabIds.size,
        total_ms: Math.round(baNow() - queuedAt),
      });
    };

    try {
      await baWithTimeout(work(), ATTACH_TIMEOUT_MS + 2000, "attachDebuggerForSession", async () => {
        baLog("error", "attachDebuggerForSession work timeout cleanup", {
          session_id: sessionId,
          tab_id: tabId,
          attach_set_size: state.attachedTabIds ? state.attachedTabIds.size : 0,
        });
        // Only drop the tab that timed out; keep peer attaches (policy B).
        if (state.attachedTabIds) {
          state.attachedTabIds.delete(tabId);
        }
        if (attachedTabs.has(tabId)) {
          await detachDebugger(tabId);
        }
      });
    } catch (e) {
      if (state.attachedTabIds && state.attachedTabIds.has(tabId) && !attachedTabs.has(tabId)) {
        state.attachedTabIds.delete(tabId);
      }
      baLog("error", "attachDebuggerForSession fail", {
        session_id: sessionId,
        tab_id: tabId,
        lock_wait_ms: lockWaitMs,
        error: e && e.message ? e.message : String(e),
      });
      throw e;
    }
  };
  // Chain must always settle so the next waiter is not stranded.
  state.attachLock = state.attachLock.then(run, run);
  return state.attachLock;
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

/** CDP -32000 when the tab navigates/closes while a debugger command is in flight. */
function isTargetNavigatedOrClosedError(err) {
  const msg = String((err && err.message) || err || "");
  return /navigated or closed|Inspected target navigated|Cannot access|target closed|-32000/i.test(
    msg,
  );
}

/**
 * Wait until the tab reaches status=complete (or timeout). Used after a
 * mid-command navigation so we can re-attach and retry once.
 */
function waitTabComplete(tabId, timeoutMs) {
  const ms = timeoutMs != null && timeoutMs > 0 ? timeoutMs : 15000;
  const deadline = Date.now() + ms;
  return new Promise((resolve, reject) => {
    const tick = async () => {
      try {
        const t = await chrome.tabs.get(tabId);
        if (!t) {
          reject(new Error("tab closed after navigation"));
          return;
        }
        if (t.status === "complete" || Date.now() >= deadline) {
          resolve(t);
          return;
        }
      } catch (e) {
        reject(new Error("tab closed after navigation"));
        return;
      }
      setTimeout(tick, 200);
    };
    tick();
  });
}

function clearDebuggerAttachBookkeeping(sessionId, tabId) {
  attachedTabs.delete(tabId);
  const state = sessionAttachState.get(sessionId);
  if (state && state.attachedTabIds) state.attachedTabIds.delete(tabId);
}

/**
 * Fail-fast: abort in-flight debugger work as soon as the target tab navigates,
 * is removed, or the debugger detaches — instead of waiting for CLI job timeout.
 * Returns { promise, cancel }.
 */
function createTabTargetAbort(tabId) {
  let cleaned = false;
  let onUpdated = null;
  let onRemoved = null;
  let onDetach = null;
  const cleanup = () => {
    if (cleaned) return;
    cleaned = true;
    try {
      if (onUpdated) chrome.tabs.onUpdated.removeListener(onUpdated);
    } catch (e) {
      /* ignore */
    }
    try {
      if (onRemoved) chrome.tabs.onRemoved.removeListener(onRemoved);
    } catch (e) {
      /* ignore */
    }
    try {
      if (onDetach) chrome.debugger.onDetach.removeListener(onDetach);
    } catch (e) {
      /* ignore */
    }
  };
  const promise = new Promise((_, reject) => {
    const abort = (why) => {
      cleanup();
      reject(
        new Error(
          "Inspected target navigated or closed" + (why ? " (" + why + ")" : ""),
        ),
      );
    };
    onUpdated = (id, changeInfo) => {
      if (id !== tabId) return;
      // URL change or navigation start — CDP target is about to die / already dead.
      if (changeInfo && (changeInfo.url != null || changeInfo.status === "loading")) {
        abort(changeInfo.url != null ? "url changed" : "loading");
      }
    };
    onRemoved = (id) => {
      if (id === tabId) abort("tab removed");
    };
    onDetach = (source, reason) => {
      if (source && source.tabId === tabId) {
        abort("debugger detached" + (reason ? ": " + reason : ""));
      }
    };
    chrome.tabs.onUpdated.addListener(onUpdated);
    chrome.tabs.onRemoved.addListener(onRemoved);
    try {
      chrome.debugger.onDetach.addListener(onDetach);
    } catch (e) {
      /* ignore */
    }
  });
  return { promise: promise, cancel: cleanup };
}

async function runWithTabTargetWatch(tabId, fn) {
  const watch = createTabTargetAbort(tabId);
  try {
    return await Promise.race([fn(tabId), watch.promise]);
  } finally {
    watch.cancel();
  }
}

async function withDebuggerForSession(sessionId, fn, opts) {
  const tabId = await pickTargetTabIdForSession(sessionId, opts || {});
  touchDebuggerActivity();
  if (baDebugEnabled() || (opts && (opts.tabId != null || opts.tab_id != null))) {
    let tabUrl = "";
    try {
      const t = await chrome.tabs.get(tabId);
      tabUrl = (t && t.url) || "";
    } catch (e) {
      /* ignore */
    }
    baLog("log", "debugger target", {
      session_id: sessionId,
      tab_id: tabId,
      url: tabUrl,
      opts: {
        tabId: opts && (opts.tabId != null ? opts.tabId : opts.tab_id),
        tabIndex: opts && (opts.tabIndex != null ? opts.tabIndex : opts.tab_index),
      },
    });
  }
  await attachDebuggerForSession(sessionId, tabId);
  touchDebuggerActivity();
  try {
    try {
      // Race CDP work against tab nav/close/detach so agents fail in ms, not at --timeout.
      return await runWithTabTargetWatch(tabId, fn);
    } catch (err) {
      // Eval/run that does location.href / form submit kills the CDP target.
      // Fail fast (above), then wait for load, re-attach, retry the job once.
      if (!isTargetNavigatedOrClosedError(err)) throw err;
      baLog("warn", "debugger target navigated; wait+reattach+retry once", {
        session_id: sessionId,
        tab_id: tabId,
        error: err && err.message ? err.message : String(err),
      });
      clearDebuggerAttachBookkeeping(sessionId, tabId);
      try {
        await waitTabComplete(tabId, 15000);
      } catch (e2) {
        throw new Error(
          "Inspected target navigated or closed (tab gone after navigation): " +
            (e2 && e2.message ? e2.message : String(e2)),
        );
      }
      await attachDebuggerForSession(sessionId, tabId);
      touchDebuggerActivity();
      return await runWithTabTargetWatch(tabId, fn);
    }
  } finally {
    touchDebuggerActivity();
  }
}

try {
  chrome.debugger.onEvent.addListener((source, method, params) => {
    if (source && source.tabId != null && typeof method === "string" && method.startsWith("Network.")) {
      const routed = harTabs.get(source.tabId);
      if (routed && routed.routing) routed.state.recorder.handle(method, params || {});
    }
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
  chrome.debugger.onDetach.addListener((source, reason) => {
    if (source && source.tabId != null) {
      const routed = harTabs.get(source.tabId);
      if (routed && !routed.capture.snapshotsReady) {
        const message = "debugger detached while recording" + (reason ? ": " + reason : "");
        routed.capture.partial = true;
        routed.capture.warnings.push("tab " + source.tabId + " " + message);
        routed.state.errors.push(message);
        routed.state.state = "detached";
      }
      if (routed) harTabs.delete(source.tabId);
      attachedTabs.delete(source.tabId);
      for (const state of sessionAttachState.values()) {
        if (state && state.attachedTabIds) {
          state.attachedTabIds.delete(source.tabId);
        }
      }
    }
  });
} catch (e) {
  /* debugger events optional in test stubs */
}