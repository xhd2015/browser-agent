(function (root, factory) {
  const api = factory();
  if (typeof module === "object" && module.exports) module.exports = api;
  root.BrowserAgentHAR = api;
})(typeof globalThis !== "undefined" ? globalThis : this, function () {
  "use strict";

  function headersToHAR(headers) {
    if (!headers) return [];
    if (Array.isArray(headers)) {
      return headers.map((h) => ({ name: String(h.name || ""), value: String(h.value || "") }));
    }
    return Object.entries(headers).map(([name, value]) => ({
      name: String(name),
      value: Array.isArray(value) ? value.join("\n") : String(value == null ? "" : value),
    }));
  }

  function queryToHAR(rawURL) {
    try {
      const url = new URL(rawURL);
      return Array.from(url.searchParams.entries()).map(([name, value]) => ({ name, value }));
    } catch (_) {
      return [];
    }
  }

  function cookieHeaderToHAR(headers) {
    const out = [];
    for (const header of headers || []) {
      if (String(header.name).toLowerCase() !== "cookie") continue;
      for (const part of String(header.value || "").split(";")) {
        const index = part.indexOf("=");
        if (index < 0) continue;
        out.push({ name: part.slice(0, index).trim(), value: part.slice(index + 1).trim() });
      }
    }
    return out;
  }

  function setCookieHeadersToHAR(headers) {
    const out = [];
    for (const header of headers || []) {
      if (String(header.name).toLowerCase() !== "set-cookie") continue;
      const first = String(header.value || "").split(";", 1)[0];
      const index = first.indexOf("=");
      if (index < 0) continue;
      out.push({ name: first.slice(0, index).trim(), value: first.slice(index + 1).trim() });
    }
    return out;
  }

  function headerValue(headers, wanted) {
    const lower = String(wanted).toLowerCase();
    const found = (headers || []).find((h) => String(h.name).toLowerCase() === lower);
    return found ? String(found.value || "") : "";
  }

  function timingDelta(end, start) {
    if (!Number.isFinite(end) || !Number.isFinite(start) || end < 0 || start < 0 || end < start) {
      return -1;
    }
    return Math.round(end - start);
  }

  function responseTimings(response) {
    const out = { blocked: -1, dns: -1, ssl: -1, connect: -1, send: 0, wait: 0, receive: 0 };
    const timing = response && response.timing;
    if (!timing) return out;
    out.dns = timingDelta(timing.dnsEnd, timing.dnsStart);
    out.connect = timingDelta(timing.connectEnd, timing.connectStart);
    out.ssl = timingDelta(timing.sslEnd, timing.sslStart);
    out.send = timingDelta(timing.sendEnd, timing.sendStart);
    out.wait = timingDelta(timing.receiveHeadersEnd, timing.sendEnd);
    return out;
  }

  function sumTimings(timings) {
    return Object.values(timings || {}).reduce((total, value) => {
      return total + (Number.isFinite(value) && value > 0 ? value : 0);
    }, 0);
  }

  function shouldCaptureURL(rawURL, controlPort) {
    if (!rawURL || typeof rawURL !== "string") return false;
    try {
      const url = new URL(rawURL);
      const protocol = url.protocol.toLowerCase();
      if (protocol !== "http:" && protocol !== "https:") return false;
      const host = url.hostname.toLowerCase();
      const port = url.port || (protocol === "https:" ? "443" : "80");
      if (
        (host === "127.0.0.1" || host === "localhost" || host === "::1" || host === "[::1]") &&
        port === String(controlPort || 43761)
      ) {
        return false;
      }
      return true;
    } catch (_) {
      return false;
    }
  }

  function baseEntry(tabID, params) {
    const request = params.request || {};
    const headers = headersToHAR(request.headers);
    const postText = typeof request.postData === "string" ? request.postData : "";
    const entry = {
      startedDateTime: params.wallTime
        ? new Date(params.wallTime * 1000).toISOString()
        : new Date().toISOString(),
      time: 0,
      request: {
        method: request.method || "GET",
        url: request.url || "",
        httpVersion: "HTTP/1.1",
        headers,
        queryString: queryToHAR(request.url || ""),
        cookies: cookieHeaderToHAR(headers),
        headersSize: -1,
        bodySize: postText ? postText.length : -1,
      },
      response: {
        status: 0,
        statusText: "",
        httpVersion: "",
        headers: [],
        cookies: [],
        content: { size: 0, mimeType: "" },
        redirectURL: "",
        headersSize: -1,
        bodySize: -1,
        _transferSize: 0,
      },
      cache: {},
      timings: { blocked: -1, dns: -1, ssl: -1, connect: -1, send: 0, wait: 0, receive: 0 },
      serverIPAddress: "",
      connection: "",
      pageref: "",
      _tabId: tabID,
      _requestId: params.requestId,
      _requestTimestamp: params.timestamp,
    };
    if (postText || request.hasPostData) {
      entry.request.postData = {
        mimeType: headerValue(headers, "content-type") || "application/octet-stream",
        text: postText,
      };
    }
    return entry;
  }

  function applyResponse(entry, response) {
    if (!entry || !response) return;
    const headers = headersToHAR(response.headers);
    const encoded = Number(response.encodedDataLength) || 0;
    entry.response = {
      status: Number(response.status) || 0,
      statusText: response.statusText || "",
      httpVersion: response.protocol || response.httpVersion || "HTTP/1.1",
      headers,
      cookies: setCookieHeadersToHAR(headers),
      content: { size: encoded, mimeType: response.mimeType || "" },
      redirectURL: response.redirectURL || headerValue(headers, "location") || "",
      headersSize: -1,
      bodySize: encoded || -1,
      _transferSize: encoded,
    };
    entry.request.httpVersion = entry.response.httpVersion || entry.request.httpVersion;
    entry.serverIPAddress = response.remoteIPAddress || "";
    entry.connection = response.connectionId != null ? String(response.connectionId) : "";
    entry.timings = responseTimings(response);
  }

  class Recorder {
    constructor(tabID, options) {
      this.tabID = Number(tabID);
      this.options = options || {};
      this.entries = new Map();
      this.current = new Map();
      this.pending = new Set();
      this.bodyStates = new WeakMap();
      this.bodyResultsClosed = false;
      this.partial = false;
      this.flushTimeoutReported = false;
      this.errors = [];
      this.redirectSequence = 0;
    }

    handle(method, params) {
      params = params || {};
      switch (method) {
        case "Network.requestWillBeSent":
          this.requestWillBeSent(params);
          break;
        case "Network.responseReceived":
          this.responseReceived(params);
          break;
        case "Network.loadingFinished":
          this.loadingFinished(params);
          break;
        case "Network.loadingFailed":
          this.loadingFailed(params);
          break;
        case "Network.requestServedFromCache":
          this.requestServedFromCache(params);
          break;
      }
    }

    requestWillBeSent(params) {
      const request = params.request || {};
      if (!shouldCaptureURL(request.url, this.options.controlPort)) return;
      const requestID = String(params.requestId || "");
      const previous = this.current.get(requestID);
      if (params.redirectResponse && previous) {
        applyResponse(previous, params.redirectResponse);
        previous.response.redirectURL = request.url || previous.response.redirectURL;
        previous.time = sumTimings(previous.timings);
        this.entries.set(requestID + ":redirect:" + this.redirectSequence++, previous);
      }
      const entry = baseEntry(this.tabID, params);
      this.current.set(requestID, entry);
      this.entries.set(requestID, entry);
    }

    responseReceived(params) {
      const requestID = String(params.requestId || "");
      const entry = this.current.get(requestID);
      if (!entry) return;
      applyResponse(entry, params.response || {});
      entry._responseTimestamp = params.timestamp;
    }

    finalize(params) {
      const requestID = String(params.requestId || "");
      const entry = this.current.get(requestID);
      if (!entry) return null;
      if (Number.isFinite(entry._responseTimestamp) && Number.isFinite(params.timestamp)) {
        entry.timings.receive = Math.max(0, Math.round((params.timestamp - entry._responseTimestamp) * 1000));
      }
      const encoded = Number(params.encodedDataLength) || 0;
      if (encoded > 0) {
        entry.response.content.size = encoded;
        entry.response.bodySize = encoded;
        entry.response._transferSize = encoded;
      }
      entry.time = sumTimings(entry.timings);
      delete entry._responseTimestamp;
      delete entry._requestTimestamp;
      return entry;
    }

    loadingFinished(params) {
      const entry = this.finalize(params);
      if (entry) this.queueBodies(String(params.requestId || ""), entry);
    }

    loadingFailed(params) {
      const entry = this.finalize(params);
      if (!entry) return;
      entry._error = params.errorText || "request failed";
      if (params.canceled) entry._canceled = true;
      this.errors.push({ request_id: String(params.requestId || ""), error: entry._error });
    }

    requestServedFromCache(params) {
      const entry = this.current.get(String(params.requestId || ""));
      if (entry) entry.cache = { beforeRequest: {}, afterRequest: {} };
    }

    bodyState(entry) {
      let state = this.bodyStates.get(entry);
      if (!state) {
        state = {
          postDataAttempted: false,
          postDataComplete: false,
          responseAttempted: false,
          responseComplete: false,
          complete: false,
          promise: null,
        };
        this.bodyStates.set(entry, state);
      }
      return state;
    }

    async captureBodies(requestID, entry, state) {
      const send = this.options.sendCommand;
      const tasks = [];
      if (this.bodyResultsClosed || typeof send !== "function") {
        state.postDataComplete = true;
        state.responseComplete = true;
        state.complete = true;
        return;
      }
      if (entry.request.postData && !entry.request.postData.text) {
        if (!state.postDataAttempted) {
          state.postDataAttempted = true;
          tasks.push(
            Promise.resolve()
              .then(() => send("Network.getRequestPostData", { requestId: requestID }))
              .then((result) => {
                if (this.bodyResultsClosed) return;
                if (result && typeof result.postData === "string") {
                  entry.request.postData.text = result.postData;
                  entry.request.bodySize = result.postData.length;
                }
              })
              .catch(() => {})
              .finally(() => {
                state.postDataComplete = true;
              }),
          );
        }
      } else {
        state.postDataComplete = true;
      }
      const method = String(entry.request.method || "").toUpperCase();
      const status = Number(entry.response.status) || 0;
      if (status > 0 && method !== "HEAD" && status !== 204 && status !== 304) {
        if (!state.responseAttempted) {
          state.responseAttempted = true;
          tasks.push(
            Promise.resolve()
              .then(() => send("Network.getResponseBody", { requestId: requestID }))
              .then((result) => {
                if (this.bodyResultsClosed || !result || result.body == null) return;
                entry.response.content.text = String(result.body);
                if (result.base64Encoded) {
                  entry.response.content.encoding = "base64";
                } else {
                  entry.response.content.size = String(result.body).length;
                }
              })
              .catch((error) => {
                if (this.bodyResultsClosed) return;
                const message = error && error.message ? error.message : String(error);
                entry._bodyError = message;
                this.errors.push({ request_id: requestID, error: "response body unavailable: " + message });
              })
              .finally(() => {
                state.responseComplete = true;
              }),
          );
        }
      } else {
        state.responseComplete = true;
      }
      await Promise.all(tasks);
      state.complete = state.postDataComplete && state.responseComplete;
    }

    queueBodies(requestID, entry) {
      const state = this.bodyState(entry);
      if (state.promise) return state.promise;
      if (state.complete) return Promise.resolve();
      const pending = this.captureBodies(requestID, entry, state).finally(() => {
        state.promise = null;
        this.pending.delete(pending);
      });
      state.promise = pending;
      this.pending.add(pending);
      return pending;
    }

    async flush(timeoutMS) {
      for (const [requestID, entry] of this.current.entries()) {
        this.queueBodies(requestID, entry);
      }
      const pending = Array.from(this.pending);
      if (pending.length === 0) {
        return { timedOut: false, partial: this.partial, pendingBodyCount: 0 };
      }

      let timer = null;
      const timeout = new Promise((resolve) => {
        timer = setTimeout(() => resolve("timeout"), Math.max(0, timeoutMS || 0));
      });
      const completed = Promise.all(pending).then(() => "complete");
      const outcome = await Promise.race([completed, timeout]);
      if (timer) clearTimeout(timer);
      if (outcome === "timeout") {
        this.partial = true;
        this.bodyResultsClosed = true;
        const pendingBodyCount = this.pending.size;
        if (!this.flushTimeoutReported) {
          this.flushTimeoutReported = true;
          this.errors.push({
            request_id: "",
            error: "response body flush timed out with " + pendingBodyCount + " pending request(s)",
          });
        }
        return { timedOut: true, partial: true, pendingBodyCount };
      }
      return { timedOut: false, partial: this.partial, pendingBodyCount: 0 };
    }

    buildHAR() {
      const entries = Array.from(this.entries.values())
        .filter((entry) => entry && entry.request && entry.request.url)
        .map((entry) => {
          const copy = JSON.parse(JSON.stringify(entry));
          delete copy._requestId;
          delete copy._responseTimestamp;
          delete copy._requestTimestamp;
          return copy;
        })
        .sort((a, b) => String(a.startedDateTime).localeCompare(String(b.startedDateTime)));
      return {
        log: {
          version: "1.2",
          creator: {
            name: "Browser Agent HAR Recorder",
            version: String(this.options.creatorVersion || "1.0.0"),
          },
          entries,
        },
      };
    }
  }

  return {
    Recorder,
    shouldCaptureURL,
    headersToHAR,
    queryToHAR,
  };
});
