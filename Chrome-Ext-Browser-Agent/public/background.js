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
try {
  importScripts("tab-close-confirm.js");
} catch (e) {
  console.error("browser-agent: tab-close-confirm.js not loaded", e);
}
const CONTROL_PORT = 43761;
const WS_PATH = "/v1/ws";
const EXT_VERSION =
  typeof BROWSER_AGENT_BUNDLE_VERSION === "string" && BROWSER_AGENT_BUNDLE_VERSION
    ? BROWSER_AGENT_BUNDLE_VERSION
    : "1.0.30";
const EXT_BUNDLE_MD5 =
  typeof BROWSER_AGENT_BUNDLE_MD5 === "string" ? BROWSER_AGENT_BUNDLE_MD5 : "";
const FEATURES = ["browser-agent", "har-capture", "all-windows", "window-close", "direct-window-v1", "all-windows-v2", "js-dialog"];
if (typeof globalThis.createTabCloseController === "function") FEATURES.push("tab-close-confirm-v1");
const CDP_PROTOCOL_VERSION = "1.3";
const HAR_BODY_SETTLE_MS = 2000;
const HAR_UPLOAD_TIMEOUT_MS = 30000;
const HAR_NETWORK_BUFFER_SIZE = 10 * 1024 * 1024;

// --- Direct mode: address windows/tabs by id with no session and no /go page. ---
// One instance-level WS (per extension, not per session) rides the same envelope
// protocol; the daemon routes jobs carrying window_id/tab_id to it. Consent is an
// explicit popup toggle (default off), so the extension never debugs windows the
// user did not opt in; capturable-URL checks still apply per tab.
const DIRECT_SESSION_KEY = "__direct__";
const DIRECT_ENABLED_KEY = "directWindowControlEnabled";
const DIRECT_INSTANCE_KEY = "directInstanceId";
const DIRECT_PORT_KEY = "directControlPort";
const DIRECT_HEALTH_TIMEOUT_MS = 1500;

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
 * Open JS dialogs (alert / confirm / prompt / beforeunload) keyed by tab id.
 * The page is suspended while a dialog is open, so this state is the only way
 * an agent or the mirror can see and answer it. Cleared on
 * javascriptDialogClosed, debugger detach, and explicit handle jobs.
 */
const tabDialogs = new Map();
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
    return method === "Page.navigate" || String(method).startsWith("Target.") || String(method).startsWith("BrowserAgent.");
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
    } else if (entry.kind === "direct") {
      connectInstance("keepalive").catch(() => {});
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
  const directWs = directEntry.ws ? directEntry.ws.readyState : WebSocket.CLOSED;
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
    direct:
      directWs === WebSocket.OPEN
        ? "connected"
        : directWs === WebSocket.CONNECTING
          ? "connecting"
          : "none",
    directSupported: directSupported,
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

// =============================================================================
// Direct channel: instance-level WS for window_id/tab_id jobs (no session).
//
// The link is opt-in (popup toggle), negotiated against /v1/health so an old
// daemon is never sent an ?instance= URL it would misread, and reconnects with
// the same aggressive backoff as session links. Jobs arrive as regular job
// envelopes with payload.direct=true and are executed through the shared
// handleJob path with a windowId-capable target resolver.
// =============================================================================

const directEntry = {
  kind: "direct",
  ws: null,
  tabId: null,
  windowId: null,
  controlPort: CONTROL_PORT,
  reconnectTimer: null,
  reconnectAttempt: 0,
  connectTimeoutTimer: null,
  keepaliveTimer: null,
  prepareReconnectBaseMS: 0,
  prepareReconnectMaxMS: 0,
  pendingResults: new Map(),
  activeJobIds: new Set(),
  completedJobIds: new Set(),
};
/** Cached positive health probe so repeated schedules do not re-probe. */
let directSupported = false;
let directConnecting = false;

function directStorageGet(key) {
  return new Promise((resolve) => {
    try {
      chrome.storage.local.get([key], (items) => {
        resolve(items ? items[key] : undefined);
      });
    } catch (e) {
      resolve(undefined);
    }
  });
}

function directStorageSet(key, value) {
  return new Promise((resolve) => {
    try {
      chrome.storage.local.set({ [key]: value }, () => resolve());
    } catch (e) {
      resolve();
    }
  });
}

async function directModeEnabled() {
  const enabled = await directStorageGet(DIRECT_ENABLED_KEY);
  // Default ON: only an explicit popup opt-out (false) disables the direct
  // rail, so an untouched install connects the instance channel right away.
  // A failed storage read behaves like an untouched install.
  return enabled !== false;
}

/** Stable per-install identity so the daemon can tell extensions apart. */
async function directInstanceId() {
  let id = await directStorageGet(DIRECT_INSTANCE_KEY);
  if (typeof id === "string" && id) return id;
  const bytes = new Uint8Array(8);
  try {
    crypto.getRandomValues(bytes);
  } catch (e) {
    for (let i = 0; i < bytes.length; i++) bytes[i] = Math.floor(Math.random() * 256);
  }
  id = "inst-" + Array.from(bytes, (b) => b.toString(16).padStart(2, "0")).join("");
  await directStorageSet(DIRECT_INSTANCE_KEY, id);
  return id;
}

async function directPort() {
  const saved = await directStorageGet(DIRECT_PORT_KEY);
  const port = Number(saved);
  if (Number.isFinite(port) && port > 0) return port;
  return CONTROL_PORT;
}

/**
 * Ask the daemon whether it speaks direct-window-v1. Only a positive probe
 * opens an ?instance= link: an old daemon resolves any unknown query to a
 * session and could otherwise adopt our socket into a session slot.
 */
async function directDaemonSupports() {
  if (directSupported) return true;
  const port = await directPort();
  let ok = false;
  try {
    const controller = new AbortController();
    const timer = setTimeout(() => controller.abort(), DIRECT_HEALTH_TIMEOUT_MS);
    const response = await fetch("http://127.0.0.1:" + port + "/v1/health", {
      signal: controller.signal,
    });
    clearTimeout(timer);
    if (response.ok) {
      const body = await response.json().catch(() => null);
      ok = !!(body && body.direct_supported === true);
    }
  } catch (e) {
    ok = false;
  }
  if (ok) {
    directSupported = true;
    await directStorageSet(DIRECT_PORT_KEY, port);
  }
  return ok;
}

function isDirectSocketLive() {
  return directEntry.ws && directEntry.ws.readyState === WebSocket.OPEN;
}

function scheduleDirectConnect(reason) {
  if (directEntry.reconnectTimer) return;
  if (isDirectSocketLive()) return;
  const base = RECONNECT_BASE_MS;
  const delay = Math.min(RECONNECT_MAX_MS, base * Math.pow(2, directEntry.reconnectAttempt));
  directEntry.reconnectAttempt += 1;
  directEntry.reconnectTimer = setTimeout(() => {
    directEntry.reconnectTimer = null;
    connectInstance(reason || "reconnect");
  }, delay);
}

async function sendDirectHello() {
  return sendJSON(directEntry, {
    v: 1,
    type: "hello",
    payload: {
      version: EXT_VERSION,
      features: FEATURES,
      browser_product: "Chrome",
      direct: true,
    },
  });
}

async function connectInstance(reason) {
  if (directConnecting) return;
  if (isDirectSocketLive()) return;
  if (!(await directModeEnabled())) return;
  if (!(await directDaemonSupports())) {
    // Old daemon or serve down; retry on the next alarm/heal kick.
    return;
  }
  directConnecting = true;
  try {
    const port = await directPort();
    const instanceId = await directInstanceId();
    if (directEntry.ws) {
      try {
        directEntry.ws.close();
      } catch (e) {
        /* ignore */
      }
      directEntry.ws = null;
    }
    directEntry.controlPort = port;
    const wsURL =
      "ws://127.0.0.1:" + port + WS_PATH + "?instance=" + encodeURIComponent(instanceId);
    baLog("log", "connectInstance start", { reason: reason || "", ws_url: wsURL });
    let socket;
    try {
      socket = new WebSocket(wsURL);
    } catch (e) {
      baLog("warn", "connectInstance ctor fail", { error: String(e && e.message ? e.message : e) });
      scheduleDirectConnect(reason);
      return;
    }
    directEntry.ws = socket;
    directEntry.connectTimeoutTimer = setTimeout(() => {
      directEntry.connectTimeoutTimer = null;
      if (directEntry.ws && directEntry.ws.readyState === WebSocket.CONNECTING) {
        try {
          directEntry.ws.close();
        } catch (e) {
          /* ignore */
        }
      }
    }, CONNECT_TIMEOUT_MS);

    socket.onopen = () => {
      if (directEntry.connectTimeoutTimer) {
        clearTimeout(directEntry.connectTimeoutTimer);
        directEntry.connectTimeoutTimer = null;
      }
      directEntry.reconnectAttempt = 0;
      baLog("log", "connectInstance ws_open", { reason: reason || "" });
      notifyDirectChannelStatus("connected");
      sendDirectHello()
        .then(() => flushPendingJobResults(directEntry))
        .catch((e) => {
          baLog("warn", "connectInstance hello fail", {
            error: String(e && e.message ? e.message : e),
          });
        });
      startKeepalive(DIRECT_SESSION_KEY, directEntry);
      ensureOffscreenKeepAlive().catch(() => {});
    };
    socket.onmessage = (ev) => {
      let msg;
      try {
        msg = JSON.parse(ev.data);
      } catch (e) {
        return;
      }
      handleMessage(msg, DIRECT_SESSION_KEY, directEntry);
    };
    socket.onclose = () => {
      if (directEntry.connectTimeoutTimer) {
        clearTimeout(directEntry.connectTimeoutTimer);
        directEntry.connectTimeoutTimer = null;
      }
      stopKeepalive(directEntry);
      directEntry.ws = null;
      notifyDirectChannelStatus("disconnected");
      scheduleDirectConnect("close");
    };
    socket.onerror = () => {
      try {
        if (directEntry.ws && directEntry.ws !== socket) return;
        socket.close();
      } catch (e) {
        /* ignore */
      }
    };
  } finally {
    directConnecting = false;
  }
}

/** Broadcast direct channel state to any open popup. */
function notifyDirectChannelStatus(state) {
  try {
    chrome.runtime.sendMessage({ type: "direct_channel_status", state: state }, () => {
      void chrome.runtime.lastError; // popup may be closed; ignore
    });
  } catch (e) {
    /* runtime optional in test stubs */
  }
}

/** startKeepalive/stopKeepalive are keyed by session id; direct uses the reserved key. */
function directAlarmKick() {
  if (isDirectSocketLive()) return;
  connectInstance("alarm").catch(() => {});
}

// Boot kicks: same staggering rationale as session heal (SW may cold-start
// before serve is reachable, and storage reads are async).
[250, 1000, 3000, 8000].forEach((delay, index) => {
  setTimeout(() => {
    connectInstance("sw-boot-" + (index + 1)).catch(() => {});
  }, delay);
});

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
  // Popup toggle flipped: open or tear down the direct channel right away.
  if (msg.type === "direct_mode_changed") {
    directSupported = false; // re-probe on next enable
    if (msg.enabled === true) {
      directEntry.reconnectAttempt = 0;
      connectInstance("popup-toggle").catch(() => {});
    } else if (directEntry.ws) {
      try {
        directEntry.ws.close();
      } catch (e) {
        /* ignore */
      }
    }
    try {
      sendResponse({ ok: true, type: "direct_mode_ack" });
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
    connectInstance("onInstalled").catch(() => {});
  }, 1000);
});

chrome.runtime.onStartup.addListener(() => {
  setTimeout(() => {
    healSessionsFromTabs("onStartup").catch(() => {});
    connectInstance("onStartup").catch(() => {});
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
      // Direct rail rides the same alarm cadence.
      directAlarmKick();
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
  const isDirect = payload.direct === true || (entry && entry.kind === "direct");
  const jobSessionId = isDirect ? "" : payload.session_id || payload.sessionId || sessionId || "";
  const jobId = payload.job_id || payload.id || msg.id || "";
  const jobType = payload.type || payload.job_type || "eval";
  const params = payload.params || {};

  if (!claimJobExecution(entry, jobId)) return;

  const targetOpts = {
    tabId: payload.tab_id != null ? payload.tab_id : params.tab_id,
    tabIndex: payload.tab_index != null ? payload.tab_index : params.tab_index,
  };
  // Direct jobs address windows without a session page; window_id rides the
  // payload (never params) so session callers cannot be confused by it.
  if (isDirect && payload.window_id != null) {
    targetOpts.windowId = payload.window_id;
  }
  targetOpts.direct = isDirect;
  targetOpts.entry = entry;

  const t0 = Date.now();
  if (baShouldLogJob(jobType, params)) {
    baLog("log", "job start", {
      job_id: jobId,
      session_id: jobSessionId,
      type: jobType,
      tab_id: targetOpts.tabId,
      tab_index: targetOpts.tabIndex,
      window_id: (targetOpts.windowId != null ? targetOpts.windowId : (entry && entry.windowId)),
      direct: isDirect,
      params: baSummarizeParams(jobType, params),
    });
  }

  try {
    let data;
    if (isDirect && (jobType === "har_start" || jobType === "har_end")) {
      // HAR is a session spool feature; direct has no session dir to write to.
      throw new Error("HAR capture requires a session (direct mode has no session spool)");
    }
    switch (jobType) {
      case "info":
        if (isDirect) {
          data = await handleInfoJobDirect(params, targetOpts);
        } else {
          data = await handleInfoJob(params, jobSessionId);
        }
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
      case "dialog":
        data = await handleDialogJob(params, payload, jobSessionId, targetOpts);
        break;
      case "create_tab":
        data = await handleCreateTabJob(params, jobSessionId, entry, targetOpts);
        break;
      case "window":
        data = await handleWindowJob(params, jobSessionId, entry, targetOpts);
        break;
      case "windows":
        data = await handleWindowsJob();
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
    dialog: tabDialogs.get(t.id) || null,
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

/**
 * Answer a JS dialog (alert / confirm / prompt / beforeunload) on the
 * session's target tab. The page is suspended while the dialog is open, so
 * nothing else on that tab runs until this lands.
 */
async function handleDialogJob(params, payload, sessionId, targetOpts) {
  const accept = !!(params && (params.accept != null ? params.accept : params.ok));
  const hasPrompt = params && params.prompt_text != null;
  const promptText = hasPrompt ? String(params.prompt_text) : undefined;
  const result = await withDebuggerForSession(sessionId, async (tabId) => {
    const response = await sendDebuggerCommand(tabId, "Page.handleJavaScriptDialog", {
      accept: accept,
      promptText: promptText,
    });
    tabDialogs.delete(tabId);
    return response;
  }, targetOpts);
  return {
    type: "dialog",
    op: "handle",
    accept: accept,
    result: result,
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
  const qualityRaw = params && params.quality != null ? Math.round(Number(params.quality)) : null;
  const result = await withDebuggerForSession(sessionId, async (tabId) => {
    const cdpParams = { format: format === "jpeg" ? "jpeg" : "png" };
    if (fullPage) {
      cdpParams.captureBeyondViewport = true;
    }
    if (format === "jpeg" && qualityRaw != null && qualityRaw >= 1 && qualityRaw <= 100) {
      cdpParams.quality = qualityRaw;
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

const tabCloseController = typeof globalThis.createTabCloseController === "function"
  ? globalThis.createTabCloseController({
      chrome,
      validate: validateRemoteCloseTab,
      attach: attachDebuggerForSession,
      send: sendDebuggerCommand,
      newId: () => crypto.randomUUID(),
    })
  : null;

async function validateRemoteCloseTab(sessionId, tabId) {
  const entry = requireSessionWindowId(sessionId);
  const tab = await chrome.tabs.get(tabId);
  if (!tab || tab.windowId !== entry.windowId) throw new Error("tab is not in the session window");
  // No control-page carve-out: /go closes like any tab. The confirmation flow
  // still guards pages that carry unsaved state.
}

async function handleCdpJob(params, sessionId, targetOpts) {
  const method = (params && (params.method || params.cdp_method || params.cdpMethod)) || "";
  if (!method) {
    throw new Error("cdp job requires params.method");
  }
  const cdpParams = (params && params.params) || {};
  if (method === "BrowserAgent.closeTab" || method === "BrowserAgent.resolveCloseTab") {
    if (!tabCloseController) throw new Error("remote tab close confirmation is not supported");
    const result = method === "BrowserAgent.closeTab"
      ? await tabCloseController.start(sessionId, cdpParams.tab_id)
      : await tabCloseController.decide(sessionId, cdpParams.operation_id, cdpParams.accept);
    return { result, method, type: "cdp", polyfilled: true };
  }
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

const CREATE_TAB_CAPTURABLE_TIMEOUT_MS = 5000;
const CREATE_TAB_CAPTURABLE_POLL_MS = 50;

function createTabURLText(url) {
  const text = url != null ? String(url).trim() : "";
  return text || "(empty)";
}

// Empty and about: are the gap between tabs.create and a committed page URL.
function isTransitionalTabURL(url) {
  const text = url != null ? String(url).trim() : "";
  if (!text) return true;
  return text.toLowerCase().startsWith("about:");
}

function sleepMs(ms) {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

async function waitForCapturableTabURL(tabId, timeoutMs) {
  const ms =
    timeoutMs != null && timeoutMs >= 0 ? timeoutMs : CREATE_TAB_CAPTURABLE_TIMEOUT_MS;
  const deadline = Date.now() + ms;
  let lastURL = "";
  for (;;) {
    let tab;
    try {
      tab = await chrome.tabs.get(tabId);
    } catch (e) {
      throw new Error("create-tab tab closed before URL was capturable (tab " + tabId + ")");
    }
    if (!tab) {
      throw new Error("create-tab tab closed before URL was capturable (tab " + tabId + ")");
    }
    lastURL = tab.url || "";
    if (isCapturableTabURL(lastURL)) {
      return tab;
    }
    const timedOut = Date.now() >= deadline;
    if (!isTransitionalTabURL(lastURL) || timedOut) {
      const why = !isTransitionalTabURL(lastURL)
        ? "create-tab URL is not capturable"
        : "create-tab timed out waiting for a capturable URL";
      throw new Error(
        why + " (tab " + tabId + ", last url " + createTabURLText(lastURL) + ")",
      );
    }
    await sleepMs(CREATE_TAB_CAPTURABLE_POLL_MS);
  }
}

/**
 * Shared create path for job create_tab and Target.createTarget.
 * Defaults to the session window (entry.windowId); opts.window_id may target
 * another existing window for explicit window adoption. Default active: true.
 * Public identity: tab_id only (no targetId).
 * When a URL is requested, the result waits until tab.url is capturable.
 */
async function createTabInSession(sessionId, opts, createOpts) {
  opts = opts || {};
  const direct = !!(createOpts && createOpts.direct);
  const entry = sessions.get(sessionId);
  if (!direct && (!entry || entry.windowId == null)) {
    baLog("error", "create_tab no windowId", { session_id: sessionId });
    throw new Error(
      "session page not bound (windowId missing); open /go?session=" + (sessionId || ""),
    );
  }
  let targetWindowId = entry && entry.windowId != null ? entry.windowId : null;
  const requestedWindow = opts.window_id != null ? opts.window_id : opts.windowId;
  if (requestedWindow != null && requestedWindow !== "") {
    const parsedWindowId = parseInt(requestedWindow, 10);
    if (!Number.isNaN(parsedWindowId) && parsedWindowId > 0) {
      targetWindowId = parsedWindowId;
    }
  }
  if (targetWindowId == null) {
    // Direct create_tab must name a window: never guess into the user's windows.
    throw new Error("create_tab needs window_id in direct mode");
  }
  const createProps = {
    windowId: targetWindowId,
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
    window_id: targetWindowId,
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
  let ready = tab;
  if (url && ready && ready.id != null && !isCapturableTabURL(ready.url || "")) {
    ready = await waitForCapturableTabURL(ready.id);
  }
  // P2 eager arm: attach only after the committed URL is known.
  if (ready && ready.id != null) {
    const createdUrl = (ready && ready.url) || url || (ready && ready.pendingUrl) || "";
    if (direct) {
      maybeEagerAttachDirect(ready.id, createdUrl, targetWindowId);
    } else {
      maybeEagerAttach(sessionId, ready.id, createdUrl, targetWindowId);
    }
  }
  return {
    type: "create_tab",
    tab_id: ready && ready.id != null ? ready.id : null,
    url: (ready && ready.url) || url || "",
    window_id: targetWindowId,
    direct: direct || undefined,
  };
}

async function handleCreateTabJob(params, sessionId, entry, targetOpts) {
  const direct = !!(targetOpts && targetOpts.direct);
  if (direct) {
    // window_id comes from the job payload (targetOpts), params.window_id is
    // honored for explicit overrides.
    return await createTabInSession(sessionId, {
      url: params && (params.url || params.URL || params.href),
      active: params && params.active,
      window_id: params && params.window_id != null
        ? params.window_id
        : targetOpts && targetOpts.windowId,
    }, { direct: true });
  }
  return await createTabInSession(sessionId, {
    url: params && (params.url || params.URL || params.href),
    active: params && params.active,
    window_id: params && (params.window_id != null ? params.window_id : params.windowId),
  });
}

function baChromeWindowsGet(windowId) {
  return new Promise((resolve, reject) => {
    chrome.windows.get(windowId, { populate: false }, (w) => {
      if (chrome.runtime.lastError) {
        reject(new Error(chrome.runtime.lastError.message || "chrome.windows.get failed"));
        return;
      }
      resolve(w || null);
    });
  });
}

function baChromeWindowsUpdate(windowId, info) {
  return new Promise((resolve, reject) => {
    chrome.windows.update(windowId, info || {}, (w) => {
      if (chrome.runtime.lastError) {
        reject(new Error(chrome.runtime.lastError.message || "chrome.windows.update failed"));
        return;
      }
      resolve(w || null);
    });
  });
}

function baChromeWindowsRemove(windowId) {
  return new Promise((resolve, reject) => {
    if (!chrome.windows || typeof chrome.windows.remove !== "function") {
      reject(new Error("chrome.windows.remove is unavailable"));
      return;
    }
    chrome.windows.remove(windowId, () => {
      if (chrome.runtime.lastError) {
        reject(new Error(chrome.runtime.lastError.message || "chrome.windows.remove failed"));
        return;
      }
      resolve(true);
    });
  });
}

function baChromeWindowsGetAll(populate) {
  return new Promise((resolve) => {
    chrome.windows.getAll({ populate: !!populate }, (ws) => {
      if (chrome.runtime.lastError) {
        resolve([]);
        return;
      }
      resolve(Array.isArray(ws) ? ws : []);
    });
  });
}

function baChromeWindowsGetAllPopulated() {
  return new Promise((resolve, reject) => {
    if (!chrome.windows || typeof chrome.windows.getAll !== "function") {
      reject(new Error("chrome.windows.getAll is unavailable"));
      return;
    }
    chrome.windows.getAll({ populate: true }, (ws) => {
      if (chrome.runtime.lastError) {
        reject(new Error(chrome.runtime.lastError.message || "chrome.windows.getAll failed"));
        return;
      }
      resolve(Array.isArray(ws) ? ws : []);
    });
  });
}

function baWindowInfo(w) {
  if (!w) return null;
  return {
    id: w.id,
    left: w.left,
    top: w.top,
    width: w.width,
    height: w.height,
    state: w.state,
    focused: w.focused,
    type: w.type,
    incognito: w.incognito,
  };
}

/** Outer window minus page viewport, in CSS px (title bar + tab strip + borders). */
async function baMeasureViewportMetrics(tabId) {
  const res = await sendDebuggerCommand(tabId, "Runtime.evaluate", {
    expression:
      "JSON.stringify({iw:innerWidth,ih:innerHeight,ow:outerWidth,oh:outerHeight,dpr:devicePixelRatio,sx:screenX,sy:screenY,sw:screen.width,sh:screen.height})",
    returnByValue: true,
  });
  const raw = res && res.result && res.result.value;
  if (!raw) return null;
  try {
    return JSON.parse(raw);
  } catch (e) {
    return null;
  }
}

async function baActiveTabIdInWindow(windowId) {
  const tabs = await chrome.tabs.query({ windowId: windowId, active: true });
  const t = Array.isArray(tabs) && tabs.length ? tabs[0] : null;
  return t && t.id != null ? t.id : null;
}

async function baResolveWindowIdForJob(params, sessionId, entry, targetOpts) {
  params = params || {};
  const explicit = params.window_id != null ? params.window_id : params.windowId;
  if (explicit != null && explicit !== "") {
    const n = parseInt(explicit, 10);
    if (!Number.isNaN(n) && n > 0) return n;
  }
  if (entry && entry.windowId != null) return entry.windowId;
  const tabId = targetOpts && targetOpts.tabId != null ? targetOpts.tabId : null;
  if (tabId != null) {
    const t = await chrome.tabs.get(tabId).catch(() => null);
    if (t && t.windowId != null) return t.windowId;
  }
  const wins = await baChromeWindowsGetAll(false);
  const chosen = wins.find((w) => w && w.focused) || wins[0];
  if (!chosen) throw new Error("no chrome window available");
  return chosen.id;
}

/**
 * Workareas of every connected display, in window-API coordinates. Returns
 * null when display info is unavailable (Firefox, or the permission is
 * missing) so callers fall back to the unclamped legacy behavior.
 */
async function baDisplayWorkareas() {
  try {
    const api = chrome.system && chrome.system.display && chrome.system.display.getInfo;
    if (!api) return null;
    const infos = await new Promise((resolve) => {
      try {
        api.call(chrome.system.display, (items) => resolve(items || []));
      } catch (e) {
        resolve(null);
      }
    });
    if (!infos || !infos.length) return null;
    return infos.map((d) => {
      const r = d.workArea || d.bounds || {};
      return {
        left: Number(r.left) || 0,
        top: Number(r.top) || 0,
        width: Number(r.width) || 0,
        height: Number(r.height) || 0,
      };
    });
  } catch (e) {
    return null;
  }
}

/**
 * Chrome rejects chrome.windows.update when the resulting bounds keep less
 * than half the window inside the visible screen ("Bounds must be at least
 * 50% within visible screen space."). A window dragged mostly off-screen — or
 * remembered bounds captured in that state — makes every later API restore or
 * resize fail with that raw error, with no way back through the same API.
 *
 * Clamp the target into the workarea of the display hosting the window:
 * keep the requested size and move the window first, shrink only when the
 * size itself cannot fit. Non-geometry keys (focused/state/drawAttention)
 * pass through; without display info the info is returned unchanged.
 */
async function baClampWindowBounds(windowId, info) {
  const geometry = {};
  const other = {};
  for (const k of ["left", "top", "width", "height"]) {
    if (info[k] != null) geometry[k] = info[k];
  }
  for (const k of ["focused", "state", "drawAttention"]) {
    if (info[k] != null) other[k] = info[k];
  }
  if (Object.keys(geometry).length === 0) return info;
  let current = null;
  try {
    current = await baChromeWindowsGet(windowId);
  } catch (e) {
    return info;
  }
  if (!current) return info;
  const displays = await baDisplayWorkareas();
  const target = {
    left: Math.round(Number(geometry.left != null ? geometry.left : current.left) || 0),
    top: Math.round(Number(geometry.top != null ? geometry.top : current.top) || 0),
    width: Math.round(Number(geometry.width != null ? geometry.width : current.width) || 0),
    height: Math.round(Number(geometry.height != null ? geometry.height : current.height) || 0),
  };
  if (!(target.width > 0) || !(target.height > 0)) return info;
  if (!displays) return Object.assign({}, target, other);
  let display = displays[0];
  let bestOverlap = -1;
  for (const d of displays) {
    const ox = Math.max(0, Math.min(current.left + current.width, d.left + d.width) - Math.max(current.left, d.left));
    const oy = Math.max(0, Math.min(current.top + current.height, d.top + d.height) - Math.max(current.top, d.top));
    if (ox * oy > bestOverlap) {
      bestOverlap = ox * oy;
      display = d;
    }
  }
  const wa = display;
  if (!(wa.width > 0) || !(wa.height > 0)) return Object.assign({}, target, other);
  const width = Math.min(target.width, wa.width);
  const height = Math.min(target.height, wa.height);
  let left = target.left;
  let top = target.top;
  if (left + width > wa.left + wa.width) left = wa.left + wa.width - width;
  if (left < wa.left) left = wa.left;
  if (top + height > wa.top + wa.height) top = wa.top + wa.height - height;
  if (top < wa.top) top = wa.top;
  return Object.assign(
    { left: Math.round(left), top: Math.round(top), width: Math.round(width), height: Math.round(height) },
    other,
  );
}

/** Augment Chrome's raw bounds rejection with an actionable hint. */
function baBoundsError(e) {
  const msg = e && e.message ? e.message : String(e);
  if (/bounds/i.test(msg) && /50%|visible screen/i.test(msg)) {
    return new Error(
      msg + " — the window sits mostly off-screen; it could not be auto-fitted, drag it back into view and retry",
    );
  }
  return e;
}

/**
 * Window jobs: chrome.windows.* (Browser-domain window control).
 *
 * `resize` targets the *page viewport* rather than the outer window: it
 * measures the outer-minus-inner chrome delta on the target tab, sets the
 * outer size accordingly, then reports the achieved viewport (macOS clamps
 * windows taller than the screen, so the achieved size may be smaller).
 */
async function handleWindowJob(params, sessionId, entry, targetOpts) {
  params = params || {};
  const op = String(params.op || params.operation || "get").toLowerCase();
  const windowId = await baResolveWindowIdForJob(params, sessionId, entry, targetOpts);

  if (op === "get") {
    const w = await baChromeWindowsGet(windowId);
    return { type: "window", op: op, window_id: windowId, window: baWindowInfo(w) };
  }

  if (op === "close") {
    const wins = await baChromeWindowsGetAll(false);
    const normal = wins.filter((w) => w && w.type === "normal");
    if (!normal.some((w) => w.id === windowId)) {
      throw new Error("that Chrome window is already closed");
    }
    if (normal.length <= 1) {
      throw new Error("cannot close the last Chrome window");
    }
    await baChromeWindowsRemove(windowId);
    return { type: "window", op: op, window_id: windowId, closed: true };
  }

  if (op === "update") {
    const info = {};
    for (const k of ["left", "top", "width", "height", "focused", "state", "drawAttention"]) {
      if (params[k] != null) info[k] = params[k];
    }
    const current = await baChromeWindowsGet(windowId);
    if (current && current.state && current.state !== "normal" && (info.width || info.height)) {
      await baChromeWindowsUpdate(windowId, { state: "normal" });
    }
    // Stale remembered bounds (e.g. the Mac preset restoring a window that has
    // since been dragged half off-screen) would be rejected outright; clamp
    // the target into the workarea so the restore self-heals.
    let w;
    try {
      w = await baChromeWindowsUpdate(windowId, await baClampWindowBounds(windowId, info));
    } catch (e) {
      throw baBoundsError(e);
    }
    return { type: "window", op: op, window_id: windowId, window: baWindowInfo(w) };
  }

  if (op !== "resize") {
    throw new Error("unknown window op: " + op + " (want get|update|resize|close)");
  }

  const wantW = Math.round(Number(params.width || params.viewport_width || 0));
  const wantH = Math.round(Number(params.height || params.viewport_height || 0));
  if (!(wantW > 0) || !(wantH > 0)) {
    throw new Error("window resize requires params.width and params.height (viewport CSS px)");
  }
  const settleMs = Math.max(0, Math.round(Number(params.settle_ms != null ? params.settle_ms : 200)));

  const tabId =
    (targetOpts && targetOpts.tabId != null ? targetOpts.tabId : null) ||
    (await baActiveTabIdInWindow(windowId));
  if (tabId == null) throw new Error("window resize needs a tab in window " + windowId);

  return await withDebuggerForSession(
    sessionId,
    async (attachedTabId) => {
      const passes = [];
      let metrics = await baMeasureViewportMetrics(attachedTabId);
      if (!metrics) throw new Error("could not measure viewport on tab " + attachedTabId);
      passes.push(metrics);

      const current = await baChromeWindowsGet(windowId);
      // A tab driven through a debugger can report a viewport frozen in an
      // older rendering state (inner larger than its own outer, or larger
      // than the window Chrome reports). Delta math on such a reading aims
      // the resize below Chrome's window floor and the window never moves;
      // re-measure once, then refuse rather than mis-size.
      const insane = (m, w) => !w || m.iw > w.width || m.ih > w.height || m.ow < m.iw || m.oh < m.ih;
      if (current && insane(metrics, current)) {
        await new Promise((r) => setTimeout(r, settleMs));
        const remeasured = await baMeasureViewportMetrics(attachedTabId);
        if (remeasured) {
          metrics = remeasured;
          passes.push(metrics);
        }
        if (insane(metrics, current)) {
          throw new Error(
            `viewport measurement (inner ${metrics.iw}x${metrics.ih}) disagrees with the window (outer ${current.width}x${current.height}); the tab may be driven by another debugger. Retry once it settles.`,
          );
        }
      }
      if (current && current.state && current.state !== "normal") {
        await baChromeWindowsUpdate(windowId, { state: "normal" });
        await new Promise((r) => setTimeout(r, settleMs));
        metrics = await baMeasureViewportMetrics(attachedTabId);
        if (metrics) passes.push(metrics);
      }

      // Chrome UI adds outer size on top of the page viewport; a negative
      // delta is always a bad measurement, never a real window.
      let deltaW = Math.max(0, metrics.ow - metrics.iw);
      let deltaH = Math.max(0, metrics.oh - metrics.ih);
      let applied = null;
      for (let attempt = 0; attempt < 3; attempt += 1) {
        const targetW = Math.max(200, wantW + deltaW);
        const targetH = Math.max(200, wantH + deltaH);
        // A size-only update on a window parked half off-screen would be
        // rejected; clamp to the workarea (moving the window up/left as
        // needed) so viewport resizes always land somewhere valid.
        let appliedInfo;
        try {
          appliedInfo = await baChromeWindowsUpdate(
            windowId,
            await baClampWindowBounds(windowId, { width: targetW, height: targetH }),
          );
        } catch (e) {
          throw baBoundsError(e);
        }
        applied = appliedInfo;
        await new Promise((r) => setTimeout(r, settleMs));
        const after = await baMeasureViewportMetrics(attachedTabId);
        if (!after) break;
        passes.push(after);
        const offW = wantW - after.iw;
        const offH = wantH - after.ih;
        // chrome.windows.get reports the achieved (possibly clamped) outer size;
        // when it differs from what we asked for, the OS refused that axis
        // (macOS Chrome floors windows at 500x~200 and caps them at the display
        // work area), and retrying cannot help.
        const clampedW = !!(applied && Math.abs(applied.width - targetW) > 1);
        const clampedH = !!(applied && Math.abs(applied.height - targetH) > 1);
        if ((Math.abs(offW) <= 2 || clampedW) && (Math.abs(offH) <= 2 || clampedH)) break;
        // Re-derive the chrome delta from the achieved pair, then retry.
        deltaW = Math.max(0, after.ow - after.iw);
        deltaH = Math.max(0, after.oh - after.ih);
      }

      const finalMetrics = passes[passes.length - 1];
      return {
        type: "window",
        op: "resize",
        window_id: windowId,
        requested: { width: wantW, height: wantH },
        viewport: { width: finalMetrics.iw, height: finalMetrics.ih },
        outer: { width: finalMetrics.ow, height: finalMetrics.oh },
        delta: { width: finalMetrics.ow - finalMetrics.iw, height: finalMetrics.oh - finalMetrics.ih },
        display: { width: finalMetrics.sw, height: finalMetrics.sh },
        dpr: finalMetrics.dpr,
        window: baWindowInfo(await baChromeWindowsGet(windowId)),
        passes: passes.length,
        clamped: finalMetrics.iw !== wantW || finalMetrics.ih !== wantH,
      };
    },
    { tabId: tabId },
  );
}

function baCleanWindowTitle(title) {
  return String(title || "").replace(/\s+/g, " ").trim();
}

function baIsTitlePageURL(url) {
  if (!url) return false;
  try {
    const parsed = new URL(String(url));
    return (
      parsed.protocol === "http:" ||
      parsed.protocol === "https:" ||
      parsed.protocol === "file:"
    );
  } catch (e) {
    return false;
  }
}

function baFallbackTitleForURL(url) {
  try {
    const parsed = new URL(String(url || ""));
    if (parsed.hostname) return parsed.hostname;
    if (parsed.protocol === "file:") {
      const name = decodeURIComponent(parsed.pathname || "").split("/").filter(Boolean).pop();
      if (name) return name;
    }
    return parsed.protocol ? parsed.protocol.replace(":", "") : "";
  } catch (e) {
    return "";
  }
}

/**
 * A window's title comes from its first (lowest-index) real page, never from
 * the active tab. Session control pages are skipped; if a window only contains
 * a session page, the session id is the best available label.
 */
function baWindowTitleFromTabs(tabs) {
  const ordered = (tabs || []).slice().sort((a, b) => {
    const ai = a && a.index != null ? a.index : Number.MAX_SAFE_INTEGER;
    const bi = b && b.index != null ? b.index : Number.MAX_SAFE_INTEGER;
    return ai - bi;
  });
  const isSessionTab = (tab) => parseSessionGoPageURL((tab && (tab.url || tab.pendingUrl)) || "") != null;
  const firstPage = ordered.find(
    (tab) => tab && tab.id != null && !isSessionTab(tab) && baIsTitlePageURL(tab.url || tab.pendingUrl),
  );
  if (firstPage) {
    const url = firstPage.url || firstPage.pendingUrl || "";
    return {
      title: baCleanWindowTitle(firstPage.title) || baFallbackTitleForURL(url) || "Untitled window",
      title_source: "tab",
      title_tab_id: firstPage.id,
    };
  }
  const firstTitled = ordered.find(
    (tab) => tab && tab.id != null && !isSessionTab(tab) && baCleanWindowTitle(tab.title),
  );
  if (firstTitled) {
    return {
      title: baCleanWindowTitle(firstTitled.title),
      title_source: "tab",
      title_tab_id: firstTitled.id,
    };
  }
  const firstSession = ordered.find((tab) => tab && isSessionTab(tab));
  if (firstSession) {
    const parsed = parseSessionGoPageURL(firstSession.url || firstSession.pendingUrl || "");
    if (parsed && parsed.searchParams.get("session")) {
      return {
        title: parsed.searchParams.get("session"),
        title_source: "session",
        title_tab_id: null,
      };
    }
  }
  return { title: "Untitled window", title_source: "none", title_tab_id: null };
}

async function handleWindowsJob() {
  const wins = await baChromeWindowsGetAllPopulated();
  const windows = [];
  for (const w of wins) {
    // Every window type is reported (normal, popup, devtools, app, panel):
    // a browser is half an operating system, and consumers filter by type.
    if (!w || w.id == null) continue;
    const tabs = (Array.isArray(w.tabs) ? w.tabs : []).slice().sort((a, b) => {
      const ai = a && a.index != null ? a.index : Number.MAX_SAFE_INTEGER;
      const bi = b && b.index != null ? b.index : Number.MAX_SAFE_INTEGER;
      return ai - bi;
    });
    const sessionIds = [];
    for (const tab of tabs) {
      const parsed = parseSessionGoPageURL((tab && (tab.url || tab.pendingUrl)) || "");
      const sessionId = parsed && parsed.searchParams.get("session");
      if (sessionId && !sessionIds.includes(sessionId)) sessionIds.push(sessionId);
    }
    const title = baWindowTitleFromTabs(tabs);
    windows.push({
      window_id: w.id,
      type: w.type,
      focused: !!w.focused,
      width: w.width,
      height: w.height,
      state: w.state,
      title: title.title,
      title_source: title.title_source,
      title_tab_id: title.title_tab_id,
      session_ids: sessionIds,
      tabs: tabs.map((tab) => {
        const url = (tab && (tab.url || tab.pendingUrl)) || "";
        const isSessionPage = parseSessionGoPageURL(url) != null;
        return {
          tab_id: tab && tab.id != null ? tab.id : null,
          index: tab && tab.index != null ? tab.index : null,
          title: baCleanWindowTitle(tab && tab.title),
          url: url,
          active: !!(tab && tab.active),
          capturable: isCapturableTabURL(url),
          role: isSessionPage ? "session" : "user",
          opener_tab_id: tab && tab.openerTabId != null ? tab.openerTabId : null,
        };
      }),
    });
  }
  windows.sort((a, b) => a.window_id - b.window_id);
  return {
    type: "windows",
    version: EXT_VERSION,
    features: FEATURES,
    windows: windows,
  };
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
  // A tab is a tab: the session control page (/go?session=<id>) closes like any
  // other tab. Chrome's tabs.onRemoved recount cleans the session up, and the
  // caller owns the "this drops the mirror" warning.
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
  // Direct jobs (window_id/tab_id addressing, no session page) resolve in their
  // own branch: explicit tab_id, then the active capturable tab of the named
  // window. No /go page is required or consulted.
  if (opts.direct) {
    return pickTargetTabIdDirect(opts);
  }
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

/**
 * Direct-mode target resolution. Consent (popup toggle, default on) is
 * enforced here so direct jobs fail cleanly when the user opted out.
 * chrome:// and extension pages are excluded by isCapturableTabURL.
 */
async function pickTargetTabIdDirect(opts) {
  if (!(await directModeEnabled())) {
    throw new Error("direct mode disabled in extension (enable it from the toolbar popup)");
  }
  const explicitTabId = opts.tabId != null ? opts.tabId : opts.tab_id != null ? opts.tab_id : null;
  if (explicitTabId != null) {
    const tabId = parseInt(explicitTabId, 10);
    const tab = await chrome.tabs.get(tabId);
    if (!isCapturableTabURL(tab.url || "")) {
      throw new Error("tab_id " + tabId + " is not capturable");
    }
    return tab.id;
  }
  const windowIdRaw = opts.windowId != null ? opts.windowId : opts.window_id;
  if (windowIdRaw == null) {
    throw new Error("direct job needs tab_id or window_id to pick a target tab");
  }
  const windowId = parseInt(windowIdRaw, 10);
  if (Number.isNaN(windowId) || windowId <= 0) {
    throw new Error("invalid window_id " + windowIdRaw);
  }
  const activeTabs = await chrome.tabs.query({ active: true, windowId: windowId });
  const active = activeTabs && activeTabs[0];
  if (active && active.id != null && isCapturableTabURL(active.url || "")) {
    return active.id;
  }
  const wins = await chrome.windows.getAll({ populate: true });
  const win = (wins || []).find((w) => w && w.id === windowId);
  const fallback = win && (win.tabs || []).find((t) => t.id != null && isCapturableTabURL(t.url || ""));
  if (fallback) return fallback.id;
  throw new Error("no capturable tab in window " + windowId);
}

/** info for direct callers: window + tabs of the addressed window (no session). */
async function handleInfoJobDirect(_params, targetOpts) {
  const windowIdRaw = targetOpts && targetOpts.windowId != null
    ? targetOpts.windowId
    : targetOpts && targetOpts.window_id;
  if (windowIdRaw == null) {
    throw new Error("direct info job needs window_id");
  }
  const windowId = parseInt(windowIdRaw, 10);
  const w = await new Promise((resolve) => {
    try {
      chrome.windows.get(windowId, { populate: true }, (win) => {
        if (chrome.runtime.lastError || !win) {
          resolve(null);
          return;
        }
        resolve(win);
      });
    } catch (e) {
      resolve(null);
    }
  });
  if (!w) {
    throw new Error("window " + windowId + " not found (already closed?)");
  }
  const tabsOut = (w.tabs || []).map((t, i) => ({
    index: i + 1,
    id: t.id,
    url: t.url || "",
    title: t.title || "",
    active: !!t.active,
    role: "user",
    dialog: tabDialogs.get(t.id) || null,
  }));
  const active = (w.tabs || []).find((t) => t.active);
  return {
    type: "info",
    window_id: windowId,
    direct: true,
    focused: !!w.focused,
    tabs: tabsOut,
    job_target: active && active.id != null
      ? { tab_id: active.id, tab_index: tabsOut.findIndex((t) => t.id === active.id) + 1, reason: "active_in_window" }
      : { tab_id: null, tab_index: null, reason: "no_active_tab" },
  };
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

function isDebuggerNotAttachedError(err) {
  const msg = String((err && err.message) || err || "");
  return /debugger is not attached to the tab/i.test(msg);
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
      // "adopted" is unverified: getTargets.attached does not prove this
      // extension can still sendCommand after an MV3 worker restart.
      attachedTabs.set(t.tabId, "adopted");
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
  if (attachedTabs.get(tabId) === "verified") {
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
        // Page events (javascriptDialogOpening) ride the same attach; a
        // failure only costs dialog awareness, never the attach itself.
        chrome.debugger.sendCommand({ tabId: tabId }, "Page.enable", {}, () => {
          void chrome.runtime.lastError;
        });
        baLog("log", "attachDebugger done", {
          tab_id: tabId,
          attach_ms: Math.round(baNow() - t0),
          runtime_enable_ms: enableMs,
          total_ms: totalMs,
          runtime_enable_error: enableErr || null,
        });
        // Resolve the error instead of rejecting so a non-fatal Runtime.enable
        // glitch does not strand attachLock. Callers retry only when the target
        // is genuinely not attached to this worker.
        resolve(enableErr ? { ok: false, error: enableErr } : { ok: true });
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
        attachedTabs.set(tabId, "verified");
        enableRuntime().then((enabled) => {
          if (enabled && !enabled.ok && isDebuggerNotAttachedError(enabled.error)) {
            // Attach succeeded but Chrome cannot route commands: drop the
            // bookkeeping so a later attempt performs a real attach.
            attachedTabs.delete(tabId);
            reject(new Error(enabled.error));
            return;
          }
          resolve(true);
        }, reject);
      });
    });

  const attachPromise = (async () => {
    // Adopt orphans before the first attach attempt (cheap getTargets).
    await adoptDebuggerTargetsFromChrome("pre-attach");
    if (cancelled) return true;
    if (attachedTabs.has(tabId)) {
      baLog("log", "attachDebugger reuse after adopt", { tab_id: tabId });
      const enabled = await enableRuntime();
      if (enabled && !enabled.ok && isDebuggerNotAttachedError(enabled.error)) {
        // getTargets said attached, but sendCommand says otherwise (stale
        // target or a worker that no longer owns the debugger session).
        baLog("warn", "attachDebugger adopted target stale; reattach", {
          tab_id: tabId,
          error: enabled.error,
        });
        attachedTabs.delete(tabId);
        await detachDebugger(tabId, true);
        try {
          await tryAttachOnce();
        } catch (err) {
          if (isAnotherDebuggerAttachedError(err)) {
            throw new Error(anotherDebuggerAttachedHint(tabId));
          }
          throw err;
        }
        return true;
      }
      if (enabled && enabled.ok) attachedTabs.set(tabId, "verified");
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
        const enabled = await enableRuntime();
        if (!enabled || enabled.ok || !isDebuggerNotAttachedError(enabled.error)) {
          attachedTabs.set(tabId, "verified");
          return true;
        }
        attachedTabs.delete(tabId);
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

/** Direct variant: warm the debugger on a created tab; no session checks. */
function maybeEagerAttachDirect(tabId, url, windowId) {
  if (tabId == null) return Promise.resolve();
  if (typeof url === "string" && url.length > 0 && !isCapturableTabURL(url)) {
    return Promise.resolve();
  }
  if (attachedTabs.has(tabId)) {
    const state = sessionAttachState.get(DIRECT_SESSION_KEY);
    if (state && state.attachedTabIds) state.attachedTabIds.add(tabId);
    touchDebuggerActivity();
    return Promise.resolve();
  }
  if (eagerAttachInFlight.has(tabId)) {
    return Promise.resolve();
  }
  eagerAttachInFlight.add(tabId);
  touchDebuggerActivity();
  return attachDebuggerForSession(DIRECT_SESSION_KEY, tabId, { direct: true })
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
async function attachDebuggerForSession(sessionId, tabId, opts) {
  const direct = !!(opts && opts.direct);
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
    direct: direct,
    attach_set_size: state.attachedTabIds.size,
    will_wait_lock: true,
    timeout_ms: ATTACH_TIMEOUT_MS,
  });
  const run = async () => {
    const lockWaitMs = Math.round(baNow() - queuedAt);
    // Cap lock-hold work so a stuck attach cannot block the session forever.
    const work = async () => {
      // Attach gate: require ≥1 open /go?session=<id> control tab in the same window.
      // Direct jobs skip the gate: they were explicitly addressed by window/tab id
      // and the popup toggle plus capturable-URL checks provide the consent model.
      let windowId = null;
      let tabUrl = "";
      try {
        const t = await chrome.tabs.get(tabId);
        if (t && t.windowId != null) windowId = t.windowId;
        tabUrl = (t && t.url) || "";
      } catch (e) {
        /* tab may be gone */
      }
      if (!direct) {
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
  // Direct jobs bypass the /go-page attach gate but keep the same attach lock,
  // timeout cleanup, and idle-detach bookkeeping (stored under DIRECT_SESSION_KEY).
  const attach = opts && opts.direct
    ? (sid, tid) => attachDebuggerForSession(sid, tid, { direct: true })
    : attachDebuggerForSession;
  await attach(sessionId, tabId);
  touchDebuggerActivity();
  try {
    try {
      // Race CDP work against tab nav/close/detach so agents fail in ms, not at --timeout.
      return await runWithTabTargetWatch(tabId, fn);
    } catch (err) {
      if (isDebuggerNotAttachedError(err)) {
        // A stale adopted target can fail on the first command even though
        // attachDebugger reported success. Force a real attach and retry once.
        baLog("warn", "debugger not attached; force reattach and retry once", {
          session_id: sessionId,
          tab_id: tabId,
          error: err && err.message ? err.message : String(err),
        });
        clearDebuggerAttachBookkeeping(sessionId, tabId);
        await detachDebugger(tabId, true);
        await attach(sessionId, tabId);
        touchDebuggerActivity();
        return await runWithTabTargetWatch(tabId, fn);
      }
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
    if (method === "Page.javascriptDialogOpening" && params) {
      const dialogTabId = source && source.tabId != null ? source.tabId : null;
      if (dialogTabId != null) {
        tabDialogs.set(dialogTabId, {
          type: params.type || "alert",
          message: params.message || "",
          default_prompt: params.defaultPrompt || "",
          url: params.url || "",
          opened_at: Date.now(),
        });
      }
    }
    if (method === "Page.javascriptDialogClosed" && params) {
      const closedTabId = source && source.tabId != null ? source.tabId : null;
      if (closedTabId != null) tabDialogs.delete(closedTabId);
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
      tabDialogs.delete(source.tabId);
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