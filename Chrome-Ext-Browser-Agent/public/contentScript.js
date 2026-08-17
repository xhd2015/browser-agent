// Session page content script — register tab with background for per-session WS.
// All attach retries live in-page (this script + session-page boot). CLI must not
// re-open the session URL.
(function () {
  try {
    window.__BROWSER_AGENT_EXT__ = {
      product: "browser-agent",
      controlPort: 43761,
      version: "1.0.13",
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

  function sendRegister() {
    if (connected) return false;
    const session_id = readSessionIdFromPage();
    if (!session_id) return false;

    const control_port = readControlPortFromPage();
    try {
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
            return;
          }
          if (resp && resp.ok) {
            registeredAck = true;
          }
        },
      );
      return true;
    } catch (e) {
      /* extension context invalidated */
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
        })
        .catch(function () {
          /* ignore */
        });
    } catch (e) {
      /* ignore */
    }
  }

  function kick() {
    sendRegister();
    pollConnected();
  }

  if (!readSessionIdFromPage()) return;

  // Immediate + burst while SW may still be starting.
  kick();
  var burstAttempt = 0;
  burstTimer = setInterval(function () {
    if (connected) {
      clearInterval(burstTimer);
      burstTimer = null;
      return;
    }
    burstAttempt += 1;
    kick();
    if (burstAttempt >= 40) {
      clearInterval(burstTimer);
      burstTimer = null;
    }
  }, 500);

  // While the /go page stays open: re-register if daemon/SW dies later.
  slowTimer = setInterval(function () {
    if (!connected) {
      kick();
      return;
    }
    // Connected: only re-check; resume register if status drops.
    pollConnected();
    setTimeout(function () {
      if (!connected) sendRegister();
    }, 200);
  }, 10000);

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
