// API Capture - HAR Recorder (background service worker)
// Supports manual single/multi-tab recording and browser-trace control-server sessions.

import { buildPopupStats } from './popupStats';

const CONTROL_BASE = 'http://127.0.0.1:43759';
const CONTROL_PORT = '43759';
const EXT_VERSION = '1.2.0';
const EXT_FEATURES = ['browser-trace', 'multi-tab-window'];
const AGENT_ALARM = 'browser-trace-agent';
const ENTRIES_PUSH_INTERVAL_MS = 1000;

let state = 'idle'; // idle | recording
let mode = 'manual'; // manual | server
let targetTabId = null; // legacy single-tab (manual start with one tab)
let targetWindowId = null; // pinned window for multi-tab / server session
let attachedTabs = new Set(); // tabIds with debugger attached
let entries = {}; // key: `${tabId}:${requestId}` or requestId for single-tab
const pendingBodyFetches = new Map();

let agentRunning = false;
let agentAbort = null;
let sessionId = null;
let stopReasonPending = null;
let entriesPushTimer = null;

/**
 * Whether a request URL should enter the capture buffer.
 * Mirrors Go browsertrace.ShouldCaptureURL: exclude control hosts
 * 127.0.0.1:43759 and localhost:43759 (any path/query).
 */
function shouldCaptureURL(rawUrl) {
  if (!rawUrl || typeof rawUrl !== 'string') return true;
  try {
    const u = new URL(rawUrl);
    const host = (u.hostname || '').toLowerCase();
    let port = u.port;
    if (!port) {
      port = (u.protocol === 'https:') ? '443' : '80';
    }
    if (port !== CONTROL_PORT) return true;
    if (host === '127.0.0.1' || host === 'localhost') return false;
    return true;
  } catch (_) {
    return true;
  }
}

/**
 * Whether a tab URL is eligible for chrome.debugger attach.
 * Mirrors Go browsertrace.IsCapturableTabURL.
 * false: empty, chrome://, chrome-extension://, devtools://, about:blank
 * true:  http(s):// including control page http://127.0.0.1:43759/go
 */
function isCapturableTabURL(rawUrl) {
  if (!rawUrl || typeof rawUrl !== 'string') return false;
  const u = rawUrl.trim();
  if (!u) return false;
  const lower = u.toLowerCase();
  if (
    lower.startsWith('chrome://') ||
    lower.startsWith('chrome-extension://') ||
    lower.startsWith('devtools://') ||
    lower === 'about:blank' ||
    lower.startsWith('about:blank?') ||
    lower.startsWith('about:blank#')
  ) {
    return false;
  }
  if (lower.startsWith('http://') || lower.startsWith('https://')) {
    return true;
  }
  return false;
}

/**
 * Whether the agent should attempt debugger attach for a tab now.
 * Mirrors Go browsertrace.ShouldAttemptAttach.
 * true iff recording && windowMatch && !alreadyAttached && isCapturableTabURL(url)
 */
function shouldAttemptAttach(recording, windowMatch, alreadyAttached, rawUrl) {
  return !!(recording && windowMatch && !alreadyAttached && isCapturableTabURL(rawUrl));
}

const MAX_POST_DATA_SIZE = 10 * 1024 * 1024;
const MAX_BUFFER_SIZE = 10 * 1024 * 1024;

const IDLE_ICON = {
  16: 'icons/icon_16.png',
  32: 'icons/icon_32.png',
  48: 'icons/icon_48.png',
  128: 'icons/icon_128.png'
};
const RECORDING_ICON = {
  16: 'icons/recording_16.png',
  32: 'icons/recording_32.png',
  48: 'icons/recording_48.png',
  128: 'icons/recording_128.png'
};

chrome.storage.local.get(['state', 'targetTabId', 'mode'], (result) => {
  if (result.state === 'recording') {
    state = 'idle';
    targetTabId = null;
    targetWindowId = null;
    attachedTabs = new Set();
    chrome.storage.local.set({ state: 'idle', targetTabId: null, mode: 'manual', serverSession: false });
    chrome.action.setIcon({ path: IDLE_ICON });
  }
});

function formatHeaders(headers) {
  if (!headers) return [];
  if (Array.isArray(headers)) {
    return headers.map(h => ({ name: h.name, value: h.value }));
  }
  return Object.entries(headers).map(([name, value]) => ({ name, value }));
}

function buildQueryString(parsedURL) {
  if (!parsedURL || !parsedURL.query) return [];
  return parsedURL.query.map(q => ({ name: q.name, value: q.value }));
}

function buildPostData(request) {
  if (!request.postData) return undefined;
  return {
    mimeType: request.postData.mimeType || '',
    text: request.postData.text || ''
  };
}

function timingDelta(end, start) {
  if (end <= 0 || start < 0) return -1;
  return Math.round(end - start);
}

function buildTimings(response) {
  const timings = {
    blocked: -1,
    dns: -1,
    connect: -1,
    ssl: -1,
    send: 0,
    wait: 0,
    receive: 0
  };
  if (response && response.timing) {
    const t = response.timing;
    timings.dns = timingDelta(t.dnsEnd, t.dnsStart);
    timings.connect = timingDelta(t.connectEnd, t.connectStart);
    timings.ssl = timingDelta(t.sslEnd, t.sslStart);
    timings.send = timingDelta(t.sendEnd, t.sendStart);
    timings.wait = timingDelta(t.receiveHeadersEnd, t.sendEnd);
  }
  return timings;
}

function entryKey(tabId, requestId) {
  return `${tabId}:${requestId}`;
}

function sendDebuggerCommand(tabId, method, params = {}) {
  return new Promise((resolve, reject) => {
    chrome.debugger.sendCommand({ tabId }, method, params, (result) => {
      if (chrome.runtime.lastError) {
        reject(new Error(chrome.runtime.lastError.message));
        return;
      }
      resolve(result);
    });
  });
}

function requestHasBody(entry) {
  const method = (entry.request.method || '').toUpperCase();
  if (['POST', 'PUT', 'PATCH', 'DELETE'].includes(method)) return true;
  return Boolean(entry.request.postData);
}

function inferRequestMimeType(entry) {
  const contentType = (entry.request.headers || []).find(
    h => h.name.toLowerCase() === 'content-type'
  );
  return contentType?.value || 'application/octet-stream';
}

async function captureRequestBody(tabId, requestId, entry) {
  if (entry.request.postData?.text) return;
  if (!requestHasBody(entry)) return;

  try {
    const result = await sendDebuggerCommand(tabId, 'Network.getRequestPostData', { requestId });
    if (!result?.postData) return;

    if (!entry.request.postData) {
      entry.request.postData = {
        mimeType: inferRequestMimeType(entry),
        text: ''
      };
    }
    entry.request.postData.text = result.postData;
    entry.request.bodySize = result.postData.length;
  } catch (_) {
    // No request body for this request id.
  }
}

async function captureResponseBody(tabId, requestId, entry) {
  if (entry.response.content?.text) return;
  if (!entry.response.status) return;

  try {
    const result = await sendDebuggerCommand(tabId, 'Network.getResponseBody', { requestId });
    if (result?.body == null) return;

    entry.response.content.text = result.body;
    if (result.base64Encoded) {
      entry.response.content.encoding = 'base64';
    } else {
      entry.response.content.size = result.body.length;
      delete entry.response.content.encoding;
    }
  } catch (_) {
    // Body unavailable (cached, cancelled, or too large).
  }
}

async function captureBodies(tabId, requestId, entry) {
  await Promise.all([
    captureRequestBody(tabId, requestId, entry),
    captureResponseBody(tabId, requestId, entry)
  ]);
}

function queueBodyCapture(tabId, requestId, entry) {
  if (!entry) return;
  const key = entryKey(tabId, requestId);
  const existing = pendingBodyFetches.get(key);
  if (existing) return;

  const promise = captureBodies(tabId, requestId, entry).finally(() => {
    pendingBodyFetches.delete(key);
  });
  pendingBodyFetches.set(key, promise);
}

async function flushPendingBodies(capturedEntries) {
  await Promise.all([...pendingBodyFetches.values()]);

  const tasks = Object.entries(capturedEntries).map(([key, entry]) => {
    const tabId = entry._tabId;
    const requestId = entry._requestId || key.split(':').slice(1).join(':');
    if (tabId == null) return Promise.resolve();
    return captureBodies(tabId, requestId, entry).catch(() => {});
  });
  await Promise.all(tasks);
}

function handleRequestWillBeSent(tabId, params) {
  const { requestId, request, wallTime, redirectResponse } = params;
  // Auto-exclude control-server traffic so hello/status/entries/preview do not pollute HAR.
  if (!shouldCaptureURL(request && request.url)) {
    return;
  }
  const key = entryKey(tabId, requestId);
  if (redirectResponse) {
    const redirectedId = key + ':redirect:' + Date.now();
    const redirectEntry = {
      startedDateTime: new Date(wallTime * 1000).toISOString(),
      time: 0,
      request: {
        method: request.method,
        url: redirectResponse.headers['location'] || request.url,
        httpVersion: request.httpVersion || 'HTTP/1.1',
        headers: formatHeaders(request.headers),
        queryString: buildQueryString(request.urlDetails || request.parsedURL),
        cookies: [],
        headersSize: -1,
        bodySize: -1
      },
      response: {
        status: redirectResponse.status,
        statusText: redirectResponse.statusText,
        httpVersion: redirectResponse.httpVersion || 'HTTP/1.1',
        headers: formatHeaders(redirectResponse.headers),
        cookies: [],
        content: { size: 0, mimeType: redirectResponse.mimeType || '' },
        redirectURL: request.url,
        headersSize: -1,
        bodySize: -1,
        _transferSize: redirectResponse.encodedDataLength || 0
      },
      cache: {},
      timings: { blocked: -1, dns: -1, ssl: -1, connect: -1, send: 0, wait: 0, receive: 0 },
      serverIPAddress: redirectResponse.remoteIPAddress || '',
      connection: redirectResponse.connectionId ? String(redirectResponse.connectionId) : '',
      pageref: '',
      _tabId: tabId,
      _requestId: requestId
    };
    entries[redirectedId] = redirectEntry;
  }
  const entry = {
    startedDateTime: new Date((wallTime || params.timestamp) * 1000).toISOString(),
    time: 0,
    request: {
      method: request.method,
      url: request.url,
      httpVersion: request.httpVersion || 'HTTP/1.1',
      headers: formatHeaders(request.headers),
      queryString: buildQueryString(params.urlDetails || request.parsedURL || {}),
      cookies: [],
      headersSize: -1,
      bodySize: request.postData ? (request.postData.text || '').length : -1
    },
    response: {
      status: 0,
      statusText: '',
      httpVersion: '',
      headers: [],
      cookies: [],
      content: { size: 0, mimeType: '' },
      redirectURL: '',
      headersSize: -1,
      bodySize: -1,
      _transferSize: 0
    },
    cache: {},
    timings: { blocked: -1, dns: -1, ssl: -1, connect: -1, send: 0, wait: 0, receive: 0 },
    serverIPAddress: '',
    connection: '',
    pageref: '',
    _tabId: tabId,
    _requestId: requestId
  };
  if (request.postData) {
    entry.request.postData = buildPostData(request);
  }
  entries[key] = entry;
}

function handleResponseReceived(tabId, params) {
  const { requestId, response } = params;
  const entry = entries[entryKey(tabId, requestId)];
  if (!entry) return;
  entry.response = {
    status: response.status,
    statusText: response.statusText,
    httpVersion: response.httpVersion || 'HTTP/1.1',
    headers: formatHeaders(response.headers),
    cookies: [],
    content: {
      size: response.encodedDataLength || 0,
      mimeType: response.mimeType || ''
    },
    redirectURL: response.redirectURL || '',
    headersSize: -1,
    bodySize: response.encodedDataLength || 0,
    _transferSize: response.encodedDataLength || 0
  };
  if (response.remoteIPAddress) {
    entry.serverIPAddress = response.remoteIPAddress;
  }
  if (response.connectionId) {
    entry.connection = String(response.connectionId);
  }
  entry.timings = buildTimings(response);
  entry._responseTimestamp = params.timestamp;
}

function finalizeEntryTiming(entry, timestamp, encodedDataLength) {
  if (entry._responseTimestamp) {
    entry.timings.receive = Math.round((timestamp - entry._responseTimestamp) * 1000);
  }
  entry.response.content.size = encodedDataLength || entry.response.content.size || 0;
  entry.response._transferSize = encodedDataLength || entry.response._transferSize || 0;
  const totalTime = entry.timings.dns + entry.timings.connect + entry.timings.ssl +
    entry.timings.send + entry.timings.wait + entry.timings.receive;
  entry.time = Math.max(0, totalTime);
  delete entry._responseTimestamp;
}

function handleLoadingFinished(tabId, params) {
  const { requestId, timestamp, encodedDataLength } = params;
  const entry = entries[entryKey(tabId, requestId)];
  if (!entry) return;
  finalizeEntryTiming(entry, timestamp, encodedDataLength);
  queueBodyCapture(tabId, requestId, entry);
}

function handleLoadingFailed(tabId, params) {
  const { requestId, timestamp, encodedDataLength } = params;
  const entry = entries[entryKey(tabId, requestId)];
  if (!entry) return;
  finalizeEntryTiming(entry, timestamp, encodedDataLength || 0);
  queueBodyCapture(tabId, requestId, entry);
}

function handleRequestServedFromCache(tabId, params) {
  const { requestId } = params;
  const entry = entries[entryKey(tabId, requestId)];
  if (entry) {
    entry.cache = { beforeRequest: {}, afterRequest: {} };
  }
}

chrome.debugger.onEvent.addListener((debuggeeId, method, params) => {
  if (state !== 'recording') return;
  const tabId = debuggeeId.tabId;
  if (!attachedTabs.has(tabId) && tabId !== targetTabId) return;
  switch (method) {
    case 'Network.requestWillBeSent': handleRequestWillBeSent(tabId, params); break;
    case 'Network.responseReceived': handleResponseReceived(tabId, params); break;
    case 'Network.loadingFinished': handleLoadingFinished(tabId, params); break;
    case 'Network.loadingFailed': handleLoadingFailed(tabId, params); break;
    case 'Network.requestServedFromCache': handleRequestServedFromCache(tabId, params); break;
  }
});

chrome.debugger.onDetach.addListener((debuggeeId) => {
  const tabId = debuggeeId.tabId;
  attachedTabs.delete(tabId);
  if (state === 'recording' && attachedTabs.size === 0 && mode === 'manual') {
    // Last tab detached during manual recording — stop cleanly without download.
    state = 'idle';
    targetTabId = null;
    targetWindowId = null;
    pendingBodyFetches.clear();
    chrome.storage.local.set({ state: 'idle', targetTabId: null, serverSession: false });
    chrome.action.setIcon({ path: IDLE_ICON });
  }
});

function buildHAR(captured) {
  const entriesArray = Object.values(captured)
    .filter(e => e && e.request && e.request.url)
    .map(e => {
      // Keep _tabId for multi-tab merge; strip internal fields.
      const { _requestId, _responseTimestamp, ...rest } = e;
      return rest;
    });
  entriesArray.sort((a, b) => (a.startedDateTime || '').localeCompare(b.startedDateTime || ''));
  return {
    log: {
      version: '1.2',
      creator: { name: 'API Capture - HAR Recorder', version: EXT_VERSION },
      entries: entriesArray
    }
  };
}

function attachTab(tabId) {
  return new Promise((resolve) => {
    if (attachedTabs.has(tabId)) {
      resolve(true);
      return;
    }
    chrome.debugger.attach({ tabId }, '1.3', () => {
      if (chrome.runtime.lastError) {
        resolve(false);
        return;
      }
      chrome.debugger.sendCommand(
        { tabId },
        'Network.enable',
        {
          maxPostDataSize: MAX_POST_DATA_SIZE,
          maxTotalBufferSize: MAX_BUFFER_SIZE,
          maxResourceBufferSize: MAX_BUFFER_SIZE
        },
        () => {
          if (chrome.runtime.lastError) {
            try { chrome.debugger.detach({ tabId }); } catch (_) {}
            resolve(false);
            return;
          }
          attachedTabs.add(tabId);
          resolve(true);
        }
      );
    });
  });
}

async function detachTab(tabId) {
  try {
    await sendDebuggerCommand(tabId, 'Network.disable');
  } catch (_) {}
  try {
    await new Promise((resolve) => {
      chrome.debugger.detach({ tabId }, () => resolve());
    });
  } catch (_) {}
  attachedTabs.delete(tabId);
}

async function attachAllTabsInWindow(windowId) {
  const tabs = await chrome.tabs.query({ windowId });
  const results = [];
  for (const tab of tabs) {
    if (tab.id == null) continue;
    // Align skip list with isCapturableTabURL / IsCapturableTabURL.
    const url = tab.url || '';
    if (!isCapturableTabURL(url)) {
      results.push({ tabId: tab.id, ok: false, skipped: true });
      continue;
    }
    const ok = await attachTab(tab.id);
    results.push({ tabId: tab.id, ok });
  }
  return results;
}

// Watch new tabs in the pinned window during server/multi-tab recording.
// Create-time URL is often empty or chrome://newtab/ — skip attach without
// permanent give-up; tabs.onUpdated will re-attempt when URL becomes capturable.
chrome.tabs.onCreated.addListener(async (tab) => {
  if (tab.id == null) return;
  const recording = state === 'recording';
  const windowMatch = targetWindowId != null && tab.windowId === targetWindowId;
  const alreadyAttached = attachedTabs.has(tab.id);
  const url = tab.url || '';
  if (!shouldAttemptAttach(recording, windowMatch, alreadyAttached, url)) return;
  await attachTab(tab.id);
});

// Re-attach when a tab navigates to a capturable URL (new tab → https).
// Fires on url change and/or status complete while recording in the pinned window.
chrome.tabs.onUpdated.addListener(async (tabId, changeInfo, tab) => {
  // Existing: wake agent when control page finishes loading.
  if (changeInfo.status === 'complete' && tab.url && tab.url.startsWith(CONTROL_BASE)) {
    startAgent();
  }

  // Only re-evaluate attach when URL or load status changes.
  if (changeInfo.url == null && changeInfo.status !== 'complete' && changeInfo.status !== 'loading') {
    return;
  }
  if (tabId == null) return;
  const recording = state === 'recording';
  const windowMatch = targetWindowId != null && tab && tab.windowId === targetWindowId;
  const alreadyAttached = attachedTabs.has(tabId);
  // Prefer changeInfo.url when present (navigation), else current tab.url.
  const url = (changeInfo.url != null ? changeInfo.url : (tab && tab.url)) || '';
  if (!shouldAttemptAttach(recording, windowMatch, alreadyAttached, url)) return;
  await attachTab(tabId);
});

chrome.tabs.onRemoved.addListener((tabId) => {
  attachedTabs.delete(tabId);
});

// Best-effort: when the pinned capture window is closed, stop recording and
// complete the control-server session with stop_reason=window_closed.
chrome.windows.onRemoved.addListener((windowId) => {
  if (state !== 'recording' || targetWindowId == null) return;
  if (windowId !== targetWindowId) return;
  (async () => {
    try {
      const result = await stopRecording('window_closed');
      if (result.wasServer && result.har) {
        try {
          await agentComplete(result);
        } catch (e) {
          console.warn('window_closed complete failed', e);
        }
      }
    } catch (e) {
      console.warn('window_closed stop failed', e);
    }
  })();
});

async function startRecordingForWindow(windowId, opts = {}) {
  targetWindowId = windowId;
  targetTabId = opts.seedTabId || null;
  mode = opts.mode || 'manual';
  entries = {};
  pendingBodyFetches.clear();
  attachedTabs = new Set();

  const attachResults = await attachAllTabsInWindow(windowId);
  const attached = attachResults.filter(r => r.ok).length;
  if (attached === 0 && opts.seedTabId) {
    const ok = await attachTab(opts.seedTabId);
    if (!ok) {
      return { error: 'Failed to attach debugger to any tab' };
    }
  } else if (attached === 0) {
    return { error: 'Failed to attach debugger to any tab in window' };
  }

  state = 'recording';
  chrome.action.setIcon({ path: RECORDING_ICON });
  chrome.storage.local.set({
    state,
    targetTabId,
    targetWindowId,
    mode,
    serverSession: mode === 'server'
  });
  startEntriesPushLoop();
  // Immediate push so server preview can open with empty/current snapshot.
  pushEntriesSnapshot().catch(() => {});
  return { state, attached, windowId, attachResults };
}

async function startRecordingSingleTab(tabId) {
  const tab = await chrome.tabs.get(tabId);
  return startRecordingForWindow(tab.windowId, { seedTabId: tabId, mode: 'manual' });
}

async function stopRecording(reason = 'extension') {
  if (state !== 'recording') {
    return { state: 'idle', harJson: null, stop_reason: reason };
  }
  stopEntriesPushLoop();
  const capturedEntries = { ...entries };
  const tabs = [...attachedTabs];
  state = 'idle';
  const wasServer = mode === 'server';
  mode = 'manual';
  targetTabId = null;
  const winId = targetWindowId;
  targetWindowId = null;
  entries = {};
  chrome.storage.local.set({ state: 'idle', targetTabId: null, mode: 'manual', serverSession: false });
  chrome.action.setIcon({ path: IDLE_ICON });

  try {
    await flushPendingBodies(capturedEntries);
  } catch (_) {}

  for (const tabId of tabs) {
    await detachTab(tabId);
  }
  pendingBodyFetches.clear();
  attachedTabs = new Set();

  const har = buildHAR(capturedEntries);
  const harJson = JSON.stringify(har, null, 2);
  chrome.storage.local.set({ lastHarJson: harJson });
  return {
    state: 'idle',
    harJson,
    har,
    stop_reason: reason,
    window_id: winId,
    wasServer,
    stats: {
      entry_count: har.log.entries.length,
      tabs: tabs.length
    }
  };
}

/**
 * Discard captured entries so far while staying in recording state.
 * Immediately POSTs an empty snapshot so the server preview resets.
 */
async function clearCaptured() {
  if (state !== 'recording') {
    return { ok: false, error: 'not recording' };
  }
  entries = {};
  pendingBodyFetches.clear();
  try {
    await pushEntriesSnapshot();
  } catch (_) {}
  return { ok: true, count: 0, state };
}

/** Strip internal debugger fields before wire push. */
function sanitizeEntryForPush(entry) {
  if (!entry || typeof entry !== 'object') return entry;
  const { _tabId, _requestId, ...rest } = entry;
  return rest;
}

function currentEntriesArray() {
  return Object.values(entries)
    .map(sanitizeEntryForPush)
    .sort((a, b) => (a.startedDateTime || '').localeCompare(b.startedDateTime || ''));
}

/**
 * POST current entries snapshot to the control server (best-effort).
 * Used for live preview while recording (server sessions).
 */
async function pushEntriesSnapshot() {
  if (!sessionId) return false;
  const list = currentEntriesArray();
  try {
    const res = await postJSON('/v1/entries', {
      session_id: sessionId,
      entries: list,
      count: list.length,
    });
    return !!(res && res.ok);
  } catch (_) {
    return false;
  }
}

function startEntriesPushLoop() {
  stopEntriesPushLoop();
  entriesPushTimer = setInterval(() => {
    if (state !== 'recording') return;
    if (!sessionId) return;
    pushEntriesSnapshot().catch(() => {});
  }, ENTRIES_PUSH_INTERVAL_MS);
}

function stopEntriesPushLoop() {
  if (entriesPushTimer != null) {
    clearInterval(entriesPushTimer);
    entriesPushTimer = null;
  }
}

/**
 * Open live preview: prefer control-server /preview when session is known;
 * otherwise fall back to extension preview.html (live in-memory or last HAR).
 */
async function openPreviewTab() {
  // Preferred: server live preview when we have a control session id.
  if (sessionId) {
    const url = CONTROL_BASE + '/preview?session=' + encodeURIComponent(sessionId);
    try {
      // Probe health so we do not open a dead tab when server is down.
      const health = await fetch(CONTROL_BASE + '/v1/health');
      if (health.ok) {
        await chrome.tabs.create({ url });
        return { ok: true, url, source: 'server' };
      }
    } catch (_) {
      // fall through to extension preview
    }
  }
  // Also try server when recording/manual if health is up and we have session from storage.
  // Fallback: extension-local preview (live entries via getPreview).
  const extUrl = chrome.runtime.getURL('preview.html');
  await chrome.tabs.create({ url: extUrl });
  return { ok: true, url: extUrl, source: 'extension' };
}

// --- browser-trace control server agent ---

async function postJSON(path, body) {
  const res = await fetch(CONTROL_BASE + path, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body)
  });
  return res;
}

async function getJSON(path) {
  const res = await fetch(CONTROL_BASE + path, { method: 'GET' });
  if (!res.ok) throw new Error('HTTP ' + res.status);
  return res.json();
}

async function agentHello() {
  const res = await postJSON('/v1/hello', {
    version: EXT_VERSION,
    features: EXT_FEATURES.slice()
  });
  return res.ok;
}

async function agentStatus() {
  const body = {
    state: state === 'recording' ? 'recording' : 'waiting_extension',
    entry_count: Object.keys(entries).length,
    window_id: targetWindowId || 0
  };
  try {
    await postJSON('/v1/status', body);
  } catch (_) {}
}

async function agentComplete(result) {
  const payload = {
    har: result.har || JSON.parse(result.harJson || '{}'),
    stop_reason: result.stop_reason || 'extension',
    window_id: result.window_id || targetWindowId || 0,
    stats: result.stats || { entry_count: 0, tabs: 0 }
  };
  const res = await postJSON('/v1/complete', payload);
  return res.ok;
}

async function pollCommands(waitSec = 2) {
  const res = await fetch(CONTROL_BASE + '/v1/commands?wait=' + waitSec, { method: 'GET' });
  if (!res.ok) throw new Error('commands HTTP ' + res.status);
  return res.json();
}

async function findSessionWindow() {
  // Prefer a tab on the control server session page.
  const tabs = await chrome.tabs.query({ url: CONTROL_BASE + '/*' });
  if (tabs && tabs.length) {
    return tabs[0].windowId;
  }
  // Fallback: last focused window.
  const win = await chrome.windows.getLastFocused();
  return win.id;
}

async function handleStartCommand(cmd) {
  sessionId = cmd.session_id || sessionId;
  const windowId = await findSessionWindow();
  const result = await startRecordingForWindow(windowId, { mode: 'server' });
  if (result.error) {
    console.warn('browser-trace start failed:', result.error);
    return;
  }
  await agentStatus();
}

async function handleStopCommand() {
  if (state !== 'recording') return;
  const result = await stopRecording(stopReasonPending || 'cli');
  stopReasonPending = null;
  if (result.har) {
    try {
      await agentComplete(result);
    } catch (e) {
      console.warn('complete failed', e);
    }
  }
}

async function agentLoop(signal) {
  // Hello once at start of loop iteration series.
  try {
    await agentHello();
  } catch (_) {
    return false; // server not up
  }

  while (!signal.aborted) {
    try {
      await agentStatus();
      const cmd = await pollCommands(2);
      if (signal.aborted) break;
      const type = (cmd && cmd.type) || '';
      if (type === 'start') {
        await handleStartCommand(cmd || {});
      } else if (type === 'stop') {
        await handleStopCommand();
        // After server-driven stop+complete, leave server mode.
        break;
      }
    } catch (e) {
      // Server gone or network error — back off.
      await new Promise(r => setTimeout(r, 500));
      try {
        const health = await fetch(CONTROL_BASE + '/v1/health');
        if (!health.ok) break;
        await agentHello();
      } catch (_) {
        break;
      }
    }
  }
  return true;
}

function startAgent() {
  if (agentRunning) return;
  agentRunning = true;
  agentAbort = { aborted: false };
  const signal = agentAbort;
  (async () => {
    while (!signal.aborted) {
      const ok = await agentLoop(signal);
      if (signal.aborted) break;
      // If server not available, wait and retry (alarm also wakes us).
      await new Promise(r => setTimeout(r, ok ? 200 : 1500));
    }
    agentRunning = false;
  })();
}

function stopAgent() {
  if (agentAbort) agentAbort.aborted = true;
  agentRunning = false;
}

// Wake agent periodically so ready deadline is fair even if SW sleeps.
chrome.alarms.create(AGENT_ALARM, { periodInMinutes: 0.5 });
chrome.alarms.onAlarm.addListener((alarm) => {
  if (alarm.name === AGENT_ALARM) {
    startAgent();
  }
});

// Also start on install/startup and when a control-server page loads.
chrome.runtime.onInstalled.addListener(() => startAgent());
chrome.runtime.onStartup.addListener(() => startAgent());
startAgent();

// Note: tabs.onUpdated for attach re-try + control-page agent wake is registered
// near onCreated (shouldAttemptAttach). Do not add a second onUpdated listener here.

/**
 * Collect chrome.tabs metadata for candidate tab ids (best-effort; closed tabs omit).
 */
async function collectTabMeta(tabIds) {
  const meta = {};
  const ids = [...new Set(tabIds.filter((id) => id != null))];
  await Promise.all(ids.map(async (id) => {
    try {
      const tab = await chrome.tabs.get(id);
      if (!tab) return;
      meta[id] = {
        title: tab.title || '',
        url: tab.url || '',
        active: !!tab.active,
      };
    } catch (_) {
      // Tab closed or inaccessible — leave meta missing (title → "Closed tab").
    }
  }));
  return meta;
}

/**
 * Build enriched getState payload: chips + per-tab breakdown (popupstats rules).
 */
async function buildGetStatePayload(storageResult) {
  const attachedTabIds = [...attachedTabs];
  // Also seed meta lookup for tabs that only appear via entry keys.
  const entryTabIds = [];
  for (const key of Object.keys(entries)) {
    const idx = key.indexOf(':');
    if (idx > 0) {
      const n = Number.parseInt(key.slice(0, idx), 10);
      if (Number.isFinite(n)) entryTabIds.push(n);
    } else if (entries[key] && entries[key]._tabId != null) {
      entryTabIds.push(Number(entries[key]._tabId));
    }
  }
  const tabMeta = await collectTabMeta([...attachedTabIds, ...entryTabIds]);
  const stats = buildPopupStats({
    entries,
    attachedTabIds,
    tabMeta,
  });
  const hasLive = state === 'recording' && Object.keys(entries).length > 0;
  return {
    state,
    tabId: targetTabId,
    windowId: targetWindowId,
    count: stats.count,
    tabsWatching: stats.tabsWatching,
    domainCount: stats.domainCount,
    tabs: stats.tabs,
    // Live while recording, or last saved HAR when idle.
    hasPreview: hasLive || !!storageResult.lastHarJson || state === 'recording',
    serverSession: mode === 'server' || !!storageResult.serverSession,
    sessionId: sessionId || null,
    mode,
  };
}

chrome.runtime.onMessage.addListener((msg, sender, sendResponse) => {
  if (msg.action === 'wakeAgent') {
    startAgent();
    sendResponse({ ok: true });
    return;
  } else if (msg.action === 'getState') {
    chrome.storage.local.get(['lastHarJson', 'serverSession'], (result) => {
      buildGetStatePayload(result || {})
        .then(sendResponse)
        .catch(() => {
          sendResponse({
            state,
            tabId: targetTabId,
            windowId: targetWindowId,
            count: Object.keys(entries).length,
            tabsWatching: attachedTabs.size,
            domainCount: 0,
            tabs: [],
            hasPreview: state === 'recording' || !!(result && result.lastHarJson),
            serverSession: mode === 'server' || !!(result && result.serverSession),
            sessionId: sessionId || null,
            mode,
          });
        });
    });
    return true;
  } else if (msg.action === 'getPreview') {
    // Prefer live in-memory entries while recording; else last saved HAR.
    if (state === 'recording') {
      const list = currentEntriesArray();
      const har = {
        log: {
          version: '1.2',
          creator: { name: 'API Capture', version: EXT_VERSION },
          entries: list,
        },
      };
      sendResponse({ harJson: JSON.stringify(har), live: true, count: list.length });
      return;
    }
    chrome.storage.local.get('lastHarJson', (result) => {
      sendResponse({ harJson: result.lastHarJson || null, live: false });
    });
    return true;
  } else if (msg.action === 'clearCaptured') {
    clearCaptured()
      .then(sendResponse)
      .catch((e) => sendResponse({ ok: false, error: e.message }));
    return true;
  } else if (msg.action === 'openPreview') {
    openPreviewTab()
      .then(sendResponse)
      .catch((e) => sendResponse({ ok: false, error: e.message }));
    return true;
  } else if (msg.action === 'startRecording') {
    if (mode === 'server' && state === 'recording') {
      sendResponse({ error: 'Server-driven session active; use Stop only' });
      return;
    }
    if (!msg.tabId) { sendResponse({ error: 'No tabId' }); return; }
    startRecordingSingleTab(msg.tabId)
      .then(sendResponse)
      .catch(e => sendResponse({ error: e.message }));
    return true;
  } else if (msg.action === 'stopRecording') {
    const reason = mode === 'server' ? 'extension' : 'extension';
    stopRecording(reason).then(async (result) => {
      if (result.wasServer && result.har) {
        try {
          await agentComplete(result);
        } catch (e) {
          // Still return HAR for local download fallback.
          result.completeError = e.message;
        }
      }
      sendResponse(result);
    }).catch(e => sendResponse({ error: e.message }));
    return true;
  }
});
