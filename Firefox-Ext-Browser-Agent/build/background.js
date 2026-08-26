// Browser Agent Firefox background — HTTP poll transport to control plane
// (POST /v1/ext/hello|poll|result) as primary; WebSocket /v1/ws kept as secondary.
// Phase 1: register + hello + keepalive/reconnect.
// Phase 2: core jobs via WebExtensions APIs (no chrome.debugger).
// Phase 3: partial CDP job matrix polyfilled with tabs/scripting (no chrome.debugger).
// Load via about:debugging → This Firefox → Load Temporary Add-on.
// Prefer browser.* (Firefox); fall back to chrome.* for shared API shapes.
const api =
  typeof browser !== "undefined" && browser.runtime
    ? browser
    : typeof chrome !== "undefined"
      ? chrome
      : null;

// bundle-sum.js identity (version + md5). Firefox event pages do not support
// importScripts (service-worker API); load via fetch when needed.
let EXT_VERSION =
  typeof BROWSER_AGENT_BUNDLE_VERSION === "string" && BROWSER_AGENT_BUNDLE_VERSION
    ? BROWSER_AGENT_BUNDLE_VERSION
    : "1.0.16";
let EXT_BUNDLE_MD5 =
  typeof BROWSER_AGENT_BUNDLE_MD5 === "string" ? BROWSER_AGENT_BUNDLE_MD5 : "";
let bundleSumLoadPromise = null;

try {
  if (typeof importScripts === "function") {
    importScripts("bundle-sum.js");
    if (typeof BROWSER_AGENT_BUNDLE_VERSION === "string" && BROWSER_AGENT_BUNDLE_VERSION) {
      EXT_VERSION = BROWSER_AGENT_BUNDLE_VERSION;
    }
    if (typeof BROWSER_AGENT_BUNDLE_MD5 === "string" && BROWSER_AGENT_BUNDLE_MD5) {
      EXT_BUNDLE_MD5 = BROWSER_AGENT_BUNDLE_MD5;
    }
  }
} catch (e) {
  console.warn("browser-agent-firefox: importScripts bundle-sum.js failed", e);
}

/** Ensure bundle-sum.js is applied before hello (fetch for event pages). */
function ensureBundleSumLoaded() {
  if (EXT_BUNDLE_MD5) {
    return Promise.resolve();
  }
  if (bundleSumLoadPromise) {
    return bundleSumLoadPromise;
  }
  bundleSumLoadPromise = (async () => {
    try {
      const url =
        api && api.runtime && typeof api.runtime.getURL === "function"
          ? api.runtime.getURL("bundle-sum.js")
          : "bundle-sum.js";
      const res = await fetch(url);
      if (!res.ok) {
        throw new Error("HTTP " + res.status);
      }
      const text = await res.text();
      const vm = text.match(
        /BROWSER_AGENT_BUNDLE_VERSION\s*=\s*["']([^"']*)["']/
      );
      const mm = text.match(/BROWSER_AGENT_BUNDLE_MD5\s*=\s*["']([^"']*)["']/);
      if (vm && vm[1]) {
        EXT_VERSION = vm[1];
      }
      if (mm && mm[1]) {
        EXT_BUNDLE_MD5 = mm[1];
      }
    } catch (e) {
      console.warn("browser-agent-firefox: fetch bundle-sum.js failed", e);
    }
  })();
  return bundleSumLoadPromise;
}

const CONTROL_PORT = 43761;
const WS_PATH = "/v1/ws";
/** HTTP poll attach / long-poll / job completion (primary Firefox transport). */
const EXT_HELLO_PATH = "/v1/ext/hello";
const EXT_POLL_PATH = "/v1/ext/poll";
const EXT_RESULT_PATH = "/v1/ext/result";
/** Long-poll wait_ms sent to POST /v1/ext/poll (server normalizes default/cap). */
const POLL_WAIT_MS = 25000;
/** transportMode for HTTP poll sessions. */
const TRANSPORT_HTTP_POLL = "http_poll";
const TRANSPORT_WS = "ws";
/**
 * Firefox uses HTTP poll as primary transport (avoid WS/wss mixed-content noise).
 * WebSocket path remains available as secondary (connectSessionViaWebSocket).
 */
const USE_HTTP_POLL_PRIMARY = true;
const FEATURES = ["browser-agent"];

/**
 * @type {Map<string, {
 *   sessionId: string,
 *   ws: WebSocket|null,
 *   tabId: number|null,
 *   windowId: number|null,
 *   controlPort: number,
 *   reconnectTimer: ReturnType<typeof setTimeout>|null,
 *   reconnectAttempt: number,
 *   connectTimeoutTimer: ReturnType<typeof setTimeout>|null,
 *   keepaliveTimer: ReturnType<typeof setInterval>|null,
 *   transport: string|null,
 *   pollAbort: AbortController|null,
 *   pollLoopRunning: boolean
 * }>}
 */
const sessions = new Map();

/** Abort hung CONNECTING sockets (ms). */
const CONNECT_TIMEOUT_MS = 2500;
/** Cap reconnect backoff so we rejoin the control plane quickly after serve restart. */
const RECONNECT_MAX_MS = 2000;
const RECONNECT_BASE_MS = 100;
/** Keepalive ping while connected (keeps background + WS alive). */
const KEEPALIVE_MS = 15000;

console.log(
  "[browser-agent-firefox] background loaded; version=" +
    EXT_VERSION +
    " control_port=" +
    CONTROL_PORT +
    " transport=http_poll"
);

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

function httpBaseURL(controlPort) {
  const port =
    controlPort != null && !Number.isNaN(controlPort) && controlPort > 0
      ? controlPort
      : CONTROL_PORT;
  return "http://127.0.0.1:" + port;
}

function getOrCreateSessionEntry(sessionId) {
  let entry = sessions.get(sessionId);
  if (!entry) {
    entry = {
      sessionId: sessionId,
      ws: null,
      tabId: null,
      windowId: null,
      controlPort: CONTROL_PORT,
      reconnectTimer: null,
      reconnectAttempt: 0,
      connectTimeoutTimer: null,
      keepaliveTimer: null,
      transport: null,
      pollAbort: null,
      pollLoopRunning: false,
    };
    sessions.set(sessionId, entry);
  } else if (!entry.sessionId) {
    entry.sessionId = sessionId;
  }
  return entry;
}

function isSocketLive(entry) {
  const socket = entry && entry.ws;
  return (
    socket &&
    (socket.readyState === WebSocket.OPEN ||
      socket.readyState === WebSocket.CONNECTING)
  );
}

/** True when HTTP poll loop is active or a live WebSocket is open/connecting. */
function isTransportLive(entry) {
  if (!entry) return false;
  if (entry.transport === TRANSPORT_HTTP_POLL) {
    return !!(
      entry.pollLoopRunning &&
      entry.pollAbort &&
      entry.pollAbort.signal &&
      !entry.pollAbort.signal.aborted
    );
  }
  return isSocketLive(entry);
}

function stopHttpPoll(entry) {
  if (!entry) return;
  if (entry.pollAbort) {
    try {
      entry.pollAbort.abort();
    } catch (e) {
      /* ignore */
    }
    entry.pollAbort = null;
  }
  entry.pollLoopRunning = false;
  if (entry.transport === TRANSPORT_HTTP_POLL) {
    entry.transport = null;
  }
}

/**
 * POST JSON to control plane path; returns parsed JSON body.
 * Uses fetch (WebExtensions HTTP client) with method POST.
 */
async function postExtJSON(entry, path, body, signal) {
  const url = httpBaseURL(entry && entry.controlPort) + path;
  const res = await fetch(url, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body || {}),
    signal: signal,
  });
  let data = null;
  try {
    data = await res.json();
  } catch (e) {
    data = null;
  }
  if (!res.ok) {
    const errMsg =
      (data && (data.error || data.message)) ||
      "HTTP " + res.status + " " + path;
    const err = new Error(errMsg);
    err.status = res.status;
    err.body = data;
    throw err;
  }
  return data;
}

async function postExtHello(sessionId, entry, signal) {
  await ensureBundleSumLoaded();
  const telemetry = await buildSessionTelemetry(sessionId);
  const body = {
    session_id: sessionId,
    version: EXT_VERSION,
    features: FEATURES,
    browser_product: telemetry.browser_product || "firefox",
  };
  if (EXT_BUNDLE_MD5) {
    body.bundle_md5 = EXT_BUNDLE_MD5;
  }
  return postExtJSON(entry, EXT_HELLO_PATH, body, signal);
}

async function postExtResult(sessionId, entry, jobId, ok, data, error, signal) {
  const body = {
    session_id: sessionId,
    job_id: jobId,
    id: jobId,
    ok: !!ok,
    error: error || "",
    data: data || {},
  };
  return postExtJSON(entry, EXT_RESULT_PATH, body, signal);
}

/**
 * Normalize a flat poll job into the WS-like job envelope expected by handleJob.
 */
function pollJobToMessage(job, sessionId) {
  const jobId = (job && (job.job_id || job.id)) || "";
  const jobType = (job && job.type) || "eval";
  return {
    type: "job",
    id: jobId,
    payload: {
      job_id: jobId,
      id: jobId,
      type: jobType,
      session_id: (job && job.session_id) || sessionId || "",
      params: (job && job.params) || {},
      tab_id: job && job.tab_id,
      timeout_ms: job && job.timeout_ms,
    },
  };
}

async function handlePollJob(job, sessionId, entry) {
  const msg = pollJobToMessage(job, sessionId);
  await handleJob(msg, sessionId, entry);
}

/**
 * Long-poll loop: POST /v1/ext/poll with wait_ms, drain jobs + events.
 */
async function httpPollLoop(sessionId, entry, signal) {
  while (!signal.aborted && sessions.has(sessionId)) {
    let data;
    try {
      data = await postExtJSON(
        entry,
        EXT_POLL_PATH,
        { session_id: sessionId, wait_ms: POLL_WAIT_MS },
        signal
      );
    } catch (e) {
      if (signal.aborted) return;
      throw e;
    }
    if (signal.aborted || !sessions.has(sessionId)) return;

    const events = (data && data.events) || [];
    const jobs = (data && data.jobs) || [];

    // Drain poll events first (e.g. prepare_reconnect) before handling jobs.
    for (let i = 0; i < events.length; i++) {
      const ev = events[i];
      if (!ev || typeof ev !== "object") continue;
      handleMessage(ev, sessionId, entry);
      if (signal.aborted || !sessions.has(sessionId)) return;
    }

    for (let i = 0; i < jobs.length; i++) {
      const job = jobs[i];
      if (!job || typeof job !== "object") continue;
      try {
        await handlePollJob(job, sessionId, entry);
      } catch (e) {
        /* per-job errors already reported via sendJobResult */
      }
      if (signal.aborted || !sessions.has(sessionId)) return;
    }
  }
}

/**
 * Start HTTP poll transport: POST /v1/ext/hello then loop POST /v1/ext/poll.
 * Named starter used by register / connectSession (primary Firefox path).
 */
function startHttpPoll(sessionId, reason) {
  if (!sessionId) return;
  const entry = getOrCreateSessionEntry(sessionId);

  // Already running a live poll loop for this session.
  if (
    entry.transport === TRANSPORT_HTTP_POLL &&
    entry.pollLoopRunning &&
    entry.pollAbort &&
    entry.pollAbort.signal &&
    !entry.pollAbort.signal.aborted
  ) {
    return;
  }

  // Tear down any previous poll or leftover WS before starting.
  stopHttpPoll(entry);
  stopKeepalive(entry);
  clearConnectTimeout(entry);
  if (entry.ws) {
    try {
      entry.ws.close();
    } catch (e) {
      /* ignore */
    }
    entry.ws = null;
  }

  entry.transport = TRANSPORT_HTTP_POLL;
  const controller =
    typeof AbortController !== "undefined" ? new AbortController() : null;
  entry.pollAbort = controller;
  entry.pollLoopRunning = true;
  const signal = controller ? controller.signal : undefined;

  try {
    console.log("[browser-agent-firefox] startHttpPoll", {
      session_id: sessionId,
      reason: reason || "",
      control_port: entry.controlPort,
    });
  } catch (e) {
    /* ignore */
  }

  (async () => {
    try {
      await postExtHello(sessionId, entry, signal);
      entry.reconnectAttempt = 0;
      await httpPollLoop(sessionId, entry, signal || { aborted: false });
    } catch (e) {
      if (controller && controller.signal.aborted) return;
      try {
        console.warn("[browser-agent-firefox] http poll error", e && e.message ? e.message : e);
      } catch (e2) {
        /* ignore */
      }
      if (sessions.has(sessionId)) {
        scheduleReconnect(sessionId);
      }
    } finally {
      if (entry.pollAbort === controller) {
        entry.pollLoopRunning = false;
        if (entry.transport === TRANSPORT_HTTP_POLL) {
          entry.transport = null;
        }
      }
    }
  })();
}

/** Alias for poll-loop naming / register→poll markers. */
function startPollLoop(sessionId, reason) {
  startHttpPoll(sessionId, reason);
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
  // HTTP poll long-poll acts as keepalive; only ping on live WebSocket.
  if (entry && entry.transport === TRANSPORT_HTTP_POLL) {
    return;
  }
  entry.keepaliveTimer = setInterval(() => {
    if (entry.ws && entry.ws.readyState === WebSocket.OPEN) {
      sendJSON(entry, {
        v: 1,
        type: "ping",
        id: "ka-" + sessionId + "-" + Date.now(),
      });
    } else if (entry.transport === TRANSPORT_HTTP_POLL) {
      /* poll loop is the keepalive */
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
  if (!api || !api.tabs || !api.tabs.query) return pages;
  try {
    const tabs = await api.tabs.query({});
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
    browser_product: "firefox",
    session_page_count: pages.length,
    session_pages: pages,
  };
}

async function sendHello(sessionId, entry) {
  await ensureBundleSumLoaded();
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
  try {
    console.log("[browser-agent-firefox] prepare_reconnect scheduled", {
      session_id: sessionId,
      delay_ms: delayMs,
      reason: payload.reason || "",
    });
  } catch (e) {
    /* ignore */
  }
  setTimeout(() => {
    entry.reconnectAttempt = 0;
    // Stop HTTP poll transport if active, then close any WS.
    stopHttpPoll(entry);
    if (entry.ws) {
      try {
        entry.ws.close();
      } catch (e) {
        /* ignore */
      }
      entry.ws = null;
    }
    if (sessions.has(sessionId)) {
      scheduleReconnect(sessionId);
    }
  }, delayMs);
}

/**
 * Connect session transport. Firefox primary: HTTP poll (startHttpPoll).
 * WebSocket remains as secondary via connectSessionViaWebSocket.
 */
function connectSession(sessionId, reason) {
  if (!sessionId) return;
  if (USE_HTTP_POLL_PRIMARY) {
    startHttpPoll(sessionId, reason || "connect");
    return;
  }
  connectSessionViaWebSocket(sessionId, reason);
}

/**
 * Secondary WebSocket path to control plane /v1/ws?session=.
 * Kept for dual-mode / fallback; not the primary Firefox transport.
 */
function connectSessionViaWebSocket(sessionId, reason) {
  if (!sessionId) return;
  const entry = getOrCreateSessionEntry(sessionId);

  stopHttpPoll(entry);
  entry.transport = TRANSPORT_WS;

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
    // On WS failure, fall back to HTTP poll transport.
    startHttpPoll(sessionId, "ws-fail-fallback");
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
    entry.transport = TRANSPORT_WS;
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
    if (entry.transport === TRANSPORT_WS) {
      entry.transport = null;
    }
    if (sessions.has(sessionId)) {
      // Prefer HTTP poll after WS drops (Firefox fallback).
      startHttpPoll(sessionId, "ws-close-fallback");
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
  stopHttpPoll(entry);
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
    { tab: { id: tabId, windowId: tab && tab.windowId } }
  );
}

function handleRegisterMessage(msg, sender) {
  const sessionId = msg.session_id || msg.sessionId || "";
  if (!sessionId) return;
  const entry = getOrCreateSessionEntry(sessionId);
  const tabId =
    msg.tabId != null ? msg.tabId : sender && sender.tab && sender.tab.id;
  const windowId =
    msg.windowId != null
      ? msg.windowId
      : sender && sender.tab && sender.tab.windowId;
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

if (api && api.runtime && api.runtime.onMessage) {
  api.runtime.onMessage.addListener((msg, sender) => {
    if (!msg || typeof msg !== "object") return;
    if (msg.type === "register") {
      handleRegisterMessage(msg, sender);
    }
  });
}

if (api && api.tabs && api.tabs.onRemoved) {
  api.tabs.onRemoved.addListener((tabId) => {
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
}

if (api && api.tabs && api.tabs.onUpdated) {
  api.tabs.onUpdated.addListener((tabId, changeInfo, tab) => {
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
}

if (api && api.runtime && api.runtime.onInstalled) {
  api.runtime.onInstalled.addListener(() => {
    console.log(
      "[browser-agent-firefox] installed; control port",
      CONTROL_PORT
    );
    for (const sessionId of sessions.keys()) {
      connectSession(sessionId, "onInstalled");
    }
  });
}

if (api && api.runtime && api.runtime.onStartup) {
  api.runtime.onStartup.addListener(() => {
    for (const sessionId of sessions.keys()) {
      connectSession(sessionId, "onStartup");
    }
  });
}

// Alarm reconnect for dead sockets (not a no-op-only keepalive).
try {
  if (api && api.alarms) {
    api.alarms.create("browser-agent-firefox-reconnect", {
      periodInMinutes: 1,
    });
    api.alarms.onAlarm.addListener((alarm) => {
      if (alarm && alarm.name === "browser-agent-firefox-reconnect") {
        for (const [sessionId, entry] of sessions.entries()) {
          if (!isTransportLive(entry)) {
            connectSession(sessionId, "alarm");
          }
        }
      }
    });
  }
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
  if (type === "ping") {
    sendJSON(entry, {
      v: 1,
      type: "pong",
      id: msg.id || "pong-" + Date.now(),
    });
    return;
  }
  if (type === "pong") {
    // keepalive response from control plane — ignore
    return;
  }
  if (type === "job") {
    handleJob(msg, sessionId, entry);
  }
}

function sendJobResult(entry, jobId, ok, data, error) {
  // HTTP poll transport: complete jobs via POST /v1/ext/result.
  if (entry && entry.transport === TRANSPORT_HTTP_POLL) {
    const sessionId = entry.sessionId || "";
    const signal =
      entry.pollAbort && entry.pollAbort.signal ? entry.pollAbort.signal : undefined;
    postExtResult(sessionId, entry, jobId, ok, data, error, signal).catch(() => {});
    return;
  }
  // WebSocket transport: type=result envelope.
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

/** Best-effort console buffer for logs job (no debugger Log.enable in Phase 2). */
const consoleLogBuffer = [];

function isCapturableTabURL(url) {
  if (!url || typeof url !== "string") return false;
  const u = url.trim().toLowerCase();
  if (!u) return false;
  if (
    u.startsWith("chrome://") ||
    u.startsWith("chrome-extension://") ||
    u.startsWith("devtools://") ||
    u.startsWith("edge://") ||
    u.startsWith("about:") ||
    u.startsWith("moz-extension://") ||
    u.startsWith("resource://") ||
    u.startsWith("view-source:")
  ) {
    return false;
  }
  return true;
}

async function listCapturableTabsInSessionWindow(sessionId, entry) {
  if (!api || !api.tabs || !api.tabs.query) return [];
  if (entry && entry.windowId != null) {
    try {
      const tabs = await api.tabs.query({ windowId: entry.windowId });
      return (tabs || []).filter((t) => t.id != null && isCapturableTabURL(t.url || ""));
    } catch (e) {
      /* window query optional */
    }
  }
  try {
    const tabs = await api.tabs.query({});
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

  if (explicitTabId != null) {
    const tabId = parseInt(explicitTabId, 10);
    const tab = await api.tabs.get(tabId);
    if (entry && entry.windowId != null && tab.windowId !== entry.windowId) {
      throw new Error("tab_id " + tabId + " is not in session window " + entry.windowId);
    }
    if (!isCapturableTabURL(tab.url || "")) {
      throw new Error("tab_id " + tabId + " is not capturable");
    }
    return tab.id;
  }

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

  if (entry && entry.windowId != null) {
    try {
      const activeTabs = await api.tabs.query({ active: true, windowId: entry.windowId });
      const active = activeTabs && activeTabs[0];
      if (active && active.id != null && isCapturableTabURL(active.url || "")) {
        return active.id;
      }
    } catch (e) {
      /* window query optional */
    }
  }

  if (entry && entry.tabId != null) {
    try {
      const tab = await api.tabs.get(entry.tabId);
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
    const tabs = await api.tabs.query({});
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

/**
 * Inject and evaluate expression in page via scripting.executeScript (preferred)
 * or tabs.executeScript fallback. No chrome.debugger / Runtime.evaluate.
 */
async function executeScriptInTab(tabId, expression) {
  const expr = String(expression == null ? "" : expression);
  if (api && api.scripting && typeof api.scripting.executeScript === "function") {
    const results = await api.scripting.executeScript({
      target: { tabId: tabId },
      world: "MAIN",
      func: (code) => {
        // Indirect eval runs in page global scope.
        return (0, eval)(code);
      },
      args: [expr],
    });
    const first = results && results[0];
    if (first && first.error) {
      throw new Error(first.error.message || String(first.error));
    }
    return first ? first.result : undefined;
  }
  // Legacy tabs.executeScript (MV2-style / older Firefox).
  if (api && api.tabs && typeof api.tabs.executeScript === "function") {
    const results = await api.tabs.executeScript(tabId, {
      code: expr,
    });
    return results && results.length ? results[0] : undefined;
  }
  throw new Error("executeScript not available (need scripting or tabs.executeScript)");
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
    browser_product: "firefox",
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
  const tabId = await pickTargetTabIdForSession(sessionId, targetOpts || {});
  const value = await executeScriptInTab(tabId, expression);
  return {
    value: value,
    type: "eval",
    expression: expression,
    tab_id: tabId,
  };
}

async function handleRunJob(params, payload, sessionId, targetOpts) {
  const source =
    (params && (params.source || params.expression || params.expr || params.code || params.script)) ||
    payload.source ||
    payload.expression ||
    "";
  const tabId = await pickTargetTabIdForSession(sessionId, targetOpts || {});
  const value = await executeScriptInTab(tabId, source);
  return {
    value: value,
    type: "run",
    source: source,
    tab_id: tabId,
  };
}

async function handleLogsJob(params, _sessionId, _targetOpts) {
  const limit = (params && params.limit) || 100;
  const level = params && params.level;
  let entries = consoleLogBuffer.slice();
  if (level) {
    entries = entries.filter((e) => e.level === level);
  }
  if (limit > 0 && entries.length > limit) {
    entries = entries.slice(entries.length - limit);
  }
  // Best-effort: empty entries OK without debugger Log.enable.
  return { entries: entries, type: "logs" };
}

async function handleScreenshotJob(params, sessionId, targetOpts) {
  const format = (params && params.format) || "png";
  const tabId = await pickTargetTabIdForSession(sessionId, targetOpts || {});
  const tab = await api.tabs.get(tabId);
  const windowId = tab && tab.windowId != null ? tab.windowId : null;
  // captureVisibleTab captures the active tab in the window — activate first.
  if (!tab.active) {
    await api.tabs.update(tabId, { active: true });
  }
  const captureOpts = {
    format: format === "jpeg" || format === "jpg" ? "jpeg" : "png",
  };
  if (params && params.quality != null && captureOpts.format === "jpeg") {
    captureOpts.quality = params.quality;
  }
  let dataUrl;
  if (windowId != null) {
    dataUrl = await api.tabs.captureVisibleTab(windowId, captureOpts);
  } else {
    dataUrl = await api.tabs.captureVisibleTab(captureOpts);
  }
  let base64 = "";
  if (typeof dataUrl === "string") {
    const comma = dataUrl.indexOf(",");
    base64 = comma >= 0 ? dataUrl.slice(comma + 1) : dataUrl;
  }
  return {
    base64: base64,
    data_url: typeof dataUrl === "string" ? dataUrl : "",
    format: captureOpts.format,
    type: "screenshot",
    tab_id: tabId,
  };
}

/**
 * Shared create path for job create_tab.
 * Always scopes to the session window (entry.windowId). Default active: true.
 */
async function createTabInSession(sessionId, opts) {
  opts = opts || {};
  const entry = sessions.get(sessionId);
  if (!entry || entry.windowId == null) {
    throw new Error(
      "session page not bound (windowId missing); open /go?session=" + (sessionId || ""),
    );
  }
  if (!api || !api.tabs || typeof api.tabs.create !== "function") {
    throw new Error("tabs.create not available");
  }
  const createProps = {
    windowId: entry.windowId,
  };
  if (opts.active === false || opts.active === "false" || opts.active === 0) {
    createProps.active = false;
  } else {
    createProps.active = true;
  }
  const url = opts.url != null ? String(opts.url).trim() : "";
  if (url) {
    createProps.url = url;
  }
  const tab = await api.tabs.create(createProps);
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

/**
 * Partial CDP matrix for Firefox (no chrome.debugger).
 * Supported: Page.navigate → tabs.update, Runtime.evaluate → eval path,
 * optional Target.createTarget → create_tab. Other methods fail clearly.
 */
async function handleCdpJob(params, sessionId, targetOpts) {
  const method = (params && (params.method || params.cdp_method || params.cdpMethod)) || "";
  if (!method) {
    throw new Error("cdp job requires params.method");
  }
  const cdpParams = (params && params.params) || {};

  if (method === "Page.navigate") {
    const tabId = await pickTargetTabIdForSession(sessionId, targetOpts || {});
    const url = cdpParams.url != null ? String(cdpParams.url) : "";
    if (!url) {
      throw new Error("Page.navigate requires params.url");
    }
    if (!api || !api.tabs || typeof api.tabs.update !== "function") {
      throw new Error("tabs.update not available");
    }
    const tab = await api.tabs.update(tabId, { url: url });
    return {
      result: {
        frameId: "0",
        loaderId: "",
        tab_id: tab && tab.id != null ? tab.id : tabId,
        url: (tab && tab.url) || url,
      },
      method: method,
      type: "cdp",
      polyfilled: true,
      tab_id: tab && tab.id != null ? tab.id : tabId,
    };
  }

  if (method === "Runtime.evaluate") {
    // Reuse Phase-2 eval path (handleEvalJob → executeScriptInTab).
    const evalParams = {
      expression:
        (cdpParams && (cdpParams.expression || cdpParams.expr || cdpParams.code)) || "",
    };
    const evalResult = await handleEvalJob(evalParams, {}, sessionId, targetOpts || {});
    const value = evalResult && evalResult.value;
    return {
      result: {
        result: {
          type: value === null ? "object" : typeof value,
          value: value,
        },
      },
      method: method,
      type: "cdp",
      polyfilled: true,
      tab_id: evalResult && evalResult.tab_id,
    };
  }

  if (method === "Target.createTarget") {
    const created = await createTabInSession(sessionId, {
      url: cdpParams && (cdpParams.url || cdpParams.URL || cdpParams.href),
      active: cdpParams && cdpParams.active,
    });
    return {
      result: {
        targetId: created && created.tab_id != null ? String(created.tab_id) : "",
        tab_id: created && created.tab_id,
        url: (created && created.url) || "",
      },
      method: method,
      type: "cdp",
      polyfilled: true,
      tab_id: created && created.tab_id,
    };
  }

  throw new Error(
    "CDP method " + method + " is not supported on firefox (partial polyfill only)",
  );
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
        // Unknown job types remain unimplemented.
        sendJobResult(
          entry,
          jobId,
          false,
          { type: jobType, stub: true, browser_product: "firefox" },
          "not implemented: firefox job type=" + jobType,
        );
        return;
    }
    sendJobResult(entry, jobId, true, data, "");
  } catch (e) {
    const errMsg = e && e.message ? e.message : String(e);
    sendJobResult(entry, jobId, false, { type: jobType, stub: false }, errMsg);
  }
}
