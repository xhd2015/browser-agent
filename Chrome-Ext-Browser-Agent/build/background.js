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
    : "1.0.10";
const EXT_BUNDLE_MD5 =
  typeof BROWSER_AGENT_BUNDLE_MD5 === "string" ? BROWSER_AGENT_BUNDLE_MD5 : "";
const FEATURES = ["browser-agent"];
const CDP_PROTOCOL_VERSION = "1.3";
/** Max wait for chrome.debugger.attach + Runtime.enable (and attachLock-held work). */
const ATTACH_TIMEOUT_MS = 10000;

/** @type {Map<string, {ws: WebSocket|null, tabId: number|null, windowId: number|null, controlPort: number, reconnectTimer: ReturnType<typeof setTimeout>|null, reconnectAttempt: number, connectTimeoutTimer: ReturnType<typeof setTimeout>|null, keepaliveTimer: ReturnType<typeof setInterval>|null}>} */
const sessions = new Map();

/** Abort hung CONNECTING sockets (ms). */
const CONNECT_TIMEOUT_MS = 2500;
/** Cap reconnect backoff so we rejoin the control plane quickly after serve restart. */
const RECONNECT_MAX_MS = 2000;
const RECONNECT_BASE_MS = 100;
/** Keepalive ping while connected (keeps MV3 SW + WS alive). */
const KEEPALIVE_MS = 15000;
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
      const leftovers = Array.from(attachedTabs.keys());
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
  // Tear down chrome.debugger held by this session (leave / last control tab gone).
  detachSessionDebugger(sessionId);
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
    state.attachedTabIds.clear();
    for (const tabId of tabIds) {
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
  // WS only on register — never chrome.debugger.attach here.
  // Attaching on /go register (and on every SW wake heal) freezes Chrome's UI and
  // delays/blocks the toolbar default_popup for tens of seconds. Content tabs still
  // eager-attach on create_tab / navigate (P2); jobs attach on demand.
  connectSession(sessionId, "register");
}

function handleRegisterFromMessage(msg, sender, sendResponse) {
  handleRegisterMessage(msg, sender);
  // Ack so content-script / page burst retries can stop early after SW is live.
  try {
    sendResponse({ ok: true, type: "register_ack" });
  } catch (e) {
    /* channel closed */
  }
  return true;
}

chrome.runtime.onMessage.addListener((msg, sender, sendResponse) => {
  if (!msg || typeof msg !== "object") return;
  if (msg.type === "register") {
    return handleRegisterFromMessage(msg, sender, sendResponse);
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

chrome.tabs.onRemoved.addListener((tabId) => {
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
  const url = (changeInfo && changeInfo.url) || (tab && tab.url) || "";
  if (url) {
    if (changeInfo.status === "loading" || changeInfo.status === "complete") {
      maybeRegisterGoTab(tabId, url, tab);
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
async function healSessionsFromTabs(reason) {
  baLog("log", "healSessionsFromTabs start", { reason: reason || "" });
  let tabs = [];
  try {
    tabs = await chrome.tabs.query({});
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
    // Bind + WS only — no maybeEagerAttach.
    maybeRegisterGoTab(tab.id, url, tab);
  }
  baLog("log", "healSessionsFromTabs done", {
    reason: reason || "",
    control_tabs: found,
    attach: "skipped_for_popup_ux",
  });
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

function handleMessage(msg, sessionId, entry) {
  if (!msg || typeof msg !== "object") return;
  const type = msg.type;
  if (type === "prepare_reconnect") {
    handlePrepareReconnect(msg, sessionId, entry);
    return;
  }
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
    sendJobResult(entry, jobId, false, { type: jobType, stub: false }, errMsg);
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

function detachDebugger(tabId) {
  return new Promise((resolve) => {
    if (!attachedTabs.has(tabId)) {
      resolve();
      return;
    }
    chrome.debugger.detach({ tabId: tabId }, () => {
      attachedTabs.delete(tabId);
      resolve();
    });
  });
}

/**
 * Attach chrome.debugger to a single tabId (CDP transport for jobs).
 * Side effect: Chrome may show a browser-wide "debugging this browser" notice
 * on unrelated tabs/windows while this attach remains live — not multi-tab CDP.
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
  const attachPromise = new Promise((resolve, reject) => {
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
          attach_ms: attachMs,
          runtime_enable_ms: enableMs,
          total_ms: totalMs,
          runtime_enable_error: enableErr || null,
        });
        // Still resolve so a non-fatal Runtime.enable glitch does not strand attachLock.
        resolve(true);
      });
    });
  });
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
      const leftovers = Array.from(attachedTabs.keys());
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
    return await fn(tabId);
  } finally {
    touchDebuggerActivity();
  }
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