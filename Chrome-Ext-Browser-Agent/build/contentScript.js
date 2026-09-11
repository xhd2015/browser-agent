// Session page content script — register tab with background for per-session WS.
// All attach retries live in-page (this script + session-page boot). CLI must not
// re-open the session URL.
(function () {
  try {
    window.__BROWSER_AGENT_EXT__ = {
      product: "browser-agent",
      controlPort: 43761,
      version: "1.0.18",
      features: ["browser-agent"],
    };
  } catch (e) {
    /* ignore */
  }

  function readSessionIdFromPage() {
    const root = document.documentElement || document.body;
    if (root && root.getAttribute) {
      const dataSessionId = root.getAttribute("data-session-id");
      if (dataSessionId) return dataSessionId;
    }

    const path = (location.pathname || "").toLowerCase();
    if (!path.includes("/go")) return null;

    const params = new URLSearchParams(location.search);
    const session = params.get("session");
    if (session) return session;

    return null;
  }

  function readControlPortFromPage() {
    // Prefer the live page origin port (E2E harness uses 127.0.0.1:0).
    if (location.port) {
      const parsed = parseInt(location.port, 10);
      if (!Number.isNaN(parsed) && parsed > 0) return parsed;
    }

    const root = document.documentElement || document.body;
    if (root && root.getAttribute) {
      const dataPort = root.getAttribute("data-control-port");
      if (dataPort) {
        const parsed = parseInt(dataPort, 10);
        if (!Number.isNaN(parsed) && parsed > 0) return parsed;
      }
    }

    if (location.protocol === "https:") return 443;
    return 80;
  }

  var registeredAck = false;
  var connected = false;
  var burstTimer = null;
  var slowTimer = null;
  var holdTimer = null;
  var registerAttempts = 0;
  var keepAlivePort = null;
  var keepAliveRetryTimer = null;
  var PORT_PING_MS = 20000;

  function reportAttach(stage, lastError) {
    const session_id = readSessionIdFromPage();
    if (!session_id) return;
    try {
      var body = {
        session_id: session_id,
        stage: stage,
        register_attempts: registerAttempts,
      };
      if (lastError) body.last_error = String(lastError);
      fetch(location.origin + "/v1/ext/attach", {
        method: "POST",
        credentials: "same-origin",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(body),
      }).catch(function () {});
    } catch (e) {
      /* ignore */
    }
  }

  function clearKeepAliveRetry() {
    if (keepAliveRetryTimer) {
      clearTimeout(keepAliveRetryTimer);
      keepAliveRetryTimer = null;
    }
  }

  function scheduleKeepAliveRetry(ms) {
    clearKeepAliveRetry();
    keepAliveRetryTimer = setTimeout(function () {
      keepAliveRetryTimer = null;
      ensureKeepAlivePort();
    }, ms || 500);
  }

  /**
   * Long-lived Port holds the MV3 service worker while /go stays open —
   * including AFTER extension.connected (critical for agent job reliability).
   */
  function ensureKeepAlivePort() {
    const session_id = readSessionIdFromPage();
    if (!session_id) return;
    if (keepAlivePort) {
      try {
        keepAlivePort.postMessage({ type: "keepalive", session_id: session_id });
      } catch (e) {
        keepAlivePort = null;
      }
      if (keepAlivePort) return;
    }
    try {
      if (!chrome || !chrome.runtime || typeof chrome.runtime.connect !== "function") {
        return;
      }
      const port = chrome.runtime.connect({
        name: "ba-session:" + session_id,
      });
      keepAlivePort = port;
      port.onMessage.addListener(function (resp) {
        if (resp && resp.ok) {
          registeredAck = true;
          if (!connected) reportAttach("sw_ack", null);
        } else if (resp && resp.ok === false && !connected) {
          reportAttach("register_sent", "register_rejected");
        }
      });
      setTimeout(function () {
        if (!keepAlivePort) return;
        try {
          keepAlivePort.postMessage({
            type: connected ? "keepalive" : "register",
            session_id: session_id,
            control_port: readControlPortFromPage(),
          });
        } catch (e) {
          /* ignore */
        }
      }, 50);
      port.onDisconnect.addListener(function () {
        keepAlivePort = null;
        if (!connected && chrome.runtime && chrome.runtime.lastError) {
          reportAttach(
            "register_sent",
            chrome.runtime.lastError.message ||
              String(chrome.runtime.lastError),
          );
        }
        // Always retry — Port must stay up while this /go tab is open.
        scheduleKeepAliveRetry(connected ? 1000 : 500);
      });
      if (!connected) {
        registerAttempts += 1;
        reportAttach("register_sent", null);
      }
      port.postMessage({
        type: connected ? "keepalive" : "register",
        session_id: session_id,
        control_port: readControlPortFromPage(),
      });
    } catch (e) {
      keepAlivePort = null;
      scheduleKeepAliveRetry(connected ? 2000 : 1000);
    }
  }

  function sendRegister() {
    if (connected) return false;
    const session_id = readSessionIdFromPage();
    if (!session_id) return false;

    ensureKeepAlivePort();

    const control_port = readControlPortFromPage();
    try {
      registerAttempts += 1;
      reportAttach("register_sent", null);
      chrome.runtime.sendMessage(
        {
          type: "register",
          session_id: session_id,
          control_port: control_port,
          tabId: null,
          windowId: null,
        },
        function (resp) {
          if (chrome.runtime && chrome.runtime.lastError) {
            reportAttach(
              "register_sent",
              chrome.runtime.lastError.message ||
                String(chrome.runtime.lastError),
            );
            // Port path is the SW wake; retry connect when one-shot message fails.
            if (!keepAlivePort) scheduleKeepAliveRetry(250);
            return;
          }
          if (resp && resp.ok) {
            registeredAck = true;
            reportAttach("sw_ack", null);
          } else if (resp && resp.ok === false) {
            reportAttach("register_sent", "register_rejected");
          } else {
            reportAttach("register_sent", "register_no_response");
          }
        },
      );
      return true;
    } catch (e) {
      /* extension context invalidated */
      if (!keepAlivePort) scheduleKeepAliveRetry(500);
      return false;
    }
  }

  function pollConnected() {
    const session_id = readSessionIdFromPage();
    if (!session_id) return;
    try {
      fetch(
        location.origin +
          "/v1/session?session=" +
          encodeURIComponent(session_id),
        { credentials: "same-origin" },
      )
        .then(function (r) {
          return r.ok ? r.json() : null;
        })
        .then(function (data) {
          if (!data) return;
          const ext = data.extension || {};
          const now =
            !!(ext.connected || data.phase === "extension_connected");
          connected = now;
          if (now && burstTimer) {
            clearInterval(burstTimer);
            burstTimer = null;
          }
          // Keep Port alive after connect — do not stop hold timer.
          if (now) ensureKeepAlivePort();
        })
        .catch(function () {
          /* ignore */
        });
    } catch (e) {
      /* ignore */
    }
  }

  function kick() {
    ensureKeepAlivePort();
    if (!connected) {
      sendRegister();
      pollConnected();
    }
  }

  if (!readSessionIdFromPage()) return;

  reportAttach("page_open", null);
  // Immediate + burst until connected (covers SW cold start for adaptive wait).
  kick();
  burstTimer = setInterval(function () {
    if (connected) {
      clearInterval(burstTimer);
      burstTimer = null;
      return;
    }
    kick();
  }, 500);

  // While the /go page stays open: re-register if daemon/SW dies later.
  slowTimer = setInterval(function () {
    ensureKeepAlivePort();
    if (!connected) {
      kick();
      return;
    }
    // Connected: hold Port + re-check; resume register if status drops.
    pollConnected();
    setTimeout(function () {
      if (!connected) sendRegister();
    }, 200);
  }, 10000);

  // Dedicated Port ping while connected (resets MV3 SW idle).
  holdTimer = setInterval(function () {
    ensureKeepAlivePort();
  }, PORT_PING_MS);

  try {
    window.addEventListener("pageshow", function () {
      connected = false;
      registeredAck = false;
      kick();
    });
  } catch (e) {
    /* ignore */
  }
  try {
    document.addEventListener("visibilitychange", function () {
      if (!document.hidden) kick();
    });
  } catch (e) {
    /* ignore */
  }
})();
