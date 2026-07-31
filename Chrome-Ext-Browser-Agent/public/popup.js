// MV3 popup: no inline scripts allowed (CSP script-src 'self').
// bundle-sum.js must load first (see popup.html).
// Instant shell: package version paints sync from BROWSER_AGENT_BUNDLE_*.
// Progressive status: last-known from storage / SW, then live health + status.
// "reachable" = HTTP /v1/health only — not the same as session WebSocket attach.
//
// Toolbar default_popup: after a cold Chrome start, the *first* open can sometimes
// be slow (MV3 service worker cold start). Not always; later opens are usually
// quick. CLI/session page do not depend on the popup being snappy.
(function () {
  var HEALTH_URL = "http://127.0.0.1:43761/v1/health";
  var HEALTH_TIMEOUT_MS = 1500;
  var LAST_KNOWN_KEY = "browserAgentLastKnownStatus";

  // --- Instant shell: package identity (no network) ---
  var verEl = document.getElementById("pkg-version");
  var md5El = document.getElementById("pkg-md5");
  var ver =
    typeof BROWSER_AGENT_BUNDLE_VERSION === "string"
      ? BROWSER_AGENT_BUNDLE_VERSION
      : "";
  var md5 =
    typeof BROWSER_AGENT_BUNDLE_MD5 === "string" ? BROWSER_AGENT_BUNDLE_MD5 : "";
  if (verEl) verEl.textContent = ver || "—";
  if (md5El) md5El.textContent = md5 || "—";

  var statusEl = document.getElementById("ctrl-status");
  var sessionEl = document.getElementById("status-session");
  var debuggerEl = document.getElementById("status-debugger");
  var hintEl = document.getElementById("conn-hint");
  if (!statusEl) return;

  function setPhase(el, text, kind) {
    if (!el) return;
    el.textContent = text;
    if (kind === "ok") el.className = "ok";
    else if (kind === "warn") el.className = "warn";
    else el.className = "muted";
  }

  function applyStatusPayload(payload, fromLive) {
    if (!payload || typeof payload !== "object") return;

    // Daemon/control (only overwrite from live health when present)
    if (payload.daemon === "reachable" || payload.control === "reachable" || payload.health === "reachable") {
      setPhase(statusEl, "reachable", "ok");
    } else if (payload.daemon === "unreachable" || payload.control === "unreachable" || payload.health === "unreachable") {
      setPhase(statusEl, "unreachable", "warn");
    } else if (payload.daemon || payload.control || payload.health) {
      setPhase(statusEl, String(payload.daemon || payload.control || payload.health), "muted");
    }

    // Session / WS
    if (sessionEl) {
      var sessionsConnected =
        payload.sessionsConnected != null
          ? payload.sessionsConnected
          : payload.sessions_connected != null
            ? payload.sessions_connected
            : payload.connected != null
              ? payload.connected
              : null;
      var sessionCount =
        payload.sessionCount != null
          ? payload.sessionCount
          : payload.session_count != null
            ? payload.session_count
            : Array.isArray(payload.sessions)
              ? payload.sessions.length
              : null;
      if (sessionsConnected != null || sessionCount != null) {
        var nConn = sessionsConnected != null ? Number(sessionsConnected) : 0;
        var nTotal = sessionCount != null ? Number(sessionCount) : nConn;
        if (nConn > 0) {
          setPhase(
            sessionEl,
            "connected (" + nConn + (nTotal > nConn ? "/" + nTotal : "") + ")",
            "ok",
          );
        } else if (nTotal > 0) {
          setPhase(sessionEl, "waiting (" + nTotal + " session)", "warn");
        } else {
          setPhase(sessionEl, "none", "muted");
        }
      } else if (payload.ws || payload.session || payload.wsStatus) {
        var wsLabel = String(payload.ws || payload.session || payload.wsStatus);
        var okWs =
          wsLabel === "connected" ||
          wsLabel === "open" ||
          wsLabel.indexOf("connected") >= 0;
        setPhase(sessionEl, wsLabel, okWs ? "ok" : "muted");
      }
    }

    // Debugger / armed
    if (debuggerEl) {
      var attachCount =
        payload.attachCount != null
          ? payload.attachCount
          : payload.attach_count != null
            ? payload.attach_count
            : payload.attachedTabCount != null
              ? payload.attachedTabCount
              : null;
      var armed =
        payload.armed === true ||
        payload.armed === "true" ||
        (attachCount != null && Number(attachCount) > 0);
      if (attachCount != null) {
        var n = Number(attachCount);
        if (n > 0) {
          setPhase(debuggerEl, "armed (" + n + " tab" + (n === 1 ? "" : "s") + ")", "ok");
        } else if (armed) {
          setPhase(debuggerEl, "armed (0 attached)", "warn");
        } else {
          setPhase(debuggerEl, "idle", "muted");
        }
      } else if (payload.debugger || payload.attach || payload.attachStatus) {
        var dLabel = String(payload.debugger || payload.attach || payload.attachStatus);
        setPhase(
          debuggerEl,
          dLabel,
          dLabel === "armed" || dLabel.indexOf("attached") >= 0 ? "ok" : "muted",
        );
      }
    }

    if (hintEl && fromLive) {
      var parts = [];
      if (statusEl && statusEl.textContent === "reachable") {
        parts.push("Control server is up (HTTP).");
      } else if (statusEl && statusEl.textContent === "unreachable") {
        parts.push("Start browser-agent serve (or session new).");
      }
      if (sessionEl && sessionEl.textContent.indexOf("connected") >= 0) {
        parts.push("Session WebSocket attached.");
      } else {
        parts.push("Keep a /go?session= page open so the service worker can connect via WebSocket.");
      }
      if (debuggerEl && debuggerEl.textContent.indexOf("armed") >= 0) {
        parts.push("Debugger armed.");
      }
      hintEl.textContent = parts.join(" ");
    }
  }

  // --- Last-known paint (storage first; non-blocking) ---
  function readLastKnownFromStorage() {
    try {
      if (!chrome || !chrome.storage) return;
      var area = chrome.storage.session || chrome.storage.local;
      if (!area || typeof area.get !== "function") {
        if (chrome.storage.local && typeof chrome.storage.local.get === "function") {
          area = chrome.storage.local;
        } else {
          return;
        }
      }
      area.get([LAST_KNOWN_KEY, "lastKnownStatus", "popupStatus"], function (result) {
        if (chrome.runtime && chrome.runtime.lastError) return;
        result = result || {};
        var payload =
          result[LAST_KNOWN_KEY] ||
          result.lastKnownStatus ||
          result.popupStatus ||
          null;
        if (payload) applyStatusPayload(payload, false);
      });
    } catch (e) {
      /* storage optional in tests */
    }
  }

  function requestStatusFromSW() {
    try {
      if (!chrome || !chrome.runtime || typeof chrome.runtime.sendMessage !== "function") {
        return;
      }
      // Bound wait: if SW is busy with debugger attach, do not hang UI forever.
      var settled = false;
      var timer = setTimeout(function () {
        if (settled) return;
        settled = true;
        if (sessionEl && (!sessionEl.textContent || sessionEl.textContent === "…")) {
          setPhase(sessionEl, "sw busy…", "muted");
        }
      }, 2000);
      chrome.runtime.sendMessage({ type: "getStatus" }, function (resp) {
        if (settled) return;
        settled = true;
        clearTimeout(timer);
        if (chrome.runtime && chrome.runtime.lastError) return;
        if (resp && typeof resp === "object") {
          applyStatusPayload(resp.payload || resp, true);
        }
      });
    } catch (e) {
      /* ignore */
    }
  }

  // --- Health with timeout (AbortController) ---
  function fetchHealthWithTimeout() {
    var controller = new AbortController();
    var timedOut = false;
    var timer = setTimeout(function () {
      timedOut = true;
      try {
        controller.abort();
      } catch (e) {
        /* ignore */
      }
    }, HEALTH_TIMEOUT_MS);

    fetch(HEALTH_URL, { signal: controller.signal })
      .then(function (r) {
        clearTimeout(timer);
        if (!r.ok) throw new Error("health " + r.status);
        return r.json().catch(function () {
          return {};
        });
      })
      .then(function () {
        setPhase(statusEl, "reachable", "ok");
        if (hintEl) {
          hintEl.textContent =
            "Control server is up (HTTP). Session attach is separate: keep a /go?session= page open so the service worker can connect via WebSocket.";
        }
      })
      .catch(function () {
        clearTimeout(timer);
        setPhase(statusEl, "unreachable", "warn");
        if (hintEl) {
          hintEl.textContent = timedOut
            ? "Control server health timed out. Start browser-agent serve (or session new), then open the session page."
            : "Start browser-agent serve (or session new), then open the session page.";
        }
      });
  }

  // Progressive flow: last-known → SW status + live health in parallel (never block paint).
  readLastKnownFromStorage();
  requestStatusFromSW();
  fetchHealthWithTimeout();
})();
