// Session page content script — register tab with background for per-session WS.
// Retries after SW cold start / extension reload so a single lost sendMessage
// does not leave the open /go page permanently unbound.
(function () {
  try {
    window.__BROWSER_AGENT_EXT__ = {
      product: "browser-agent",
      controlPort: 43761,
      version: "1.0.7",
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
  var burstTimer = null;

  function sendRegister() {
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
            if (burstTimer) {
              clearInterval(burstTimer);
              burstTimer = null;
            }
          }
        },
      );
      return true;
    } catch (e) {
      /* extension context invalidated */
      return false;
    }
  }

  if (!readSessionIdFromPage()) return;

  // Immediate + short burst: covers SW cold start and mid-reload races.
  sendRegister();
  var burstAttempt = 0;
  burstTimer = setInterval(function () {
    if (registeredAck) {
      clearInterval(burstTimer);
      burstTimer = null;
      return;
    }
    burstAttempt += 1;
    sendRegister();
    if (burstAttempt >= 12) {
      clearInterval(burstTimer);
      burstTimer = null;
    }
  }, 500);

  // While the /go page stays open, re-register periodically so a later SW death
  // or daemon restart can rebind without requiring a full page reload.
  setInterval(function () {
    sendRegister();
  }, 15000);

  try {
    window.addEventListener("pageshow", function () {
      sendRegister();
    });
  } catch (e) {
    /* ignore */
  }
  try {
    document.addEventListener("visibilitychange", function () {
      if (!document.hidden) sendRegister();
    });
  } catch (e) {
    /* ignore */
  }
})();
