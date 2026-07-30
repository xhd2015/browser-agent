// Session page content script — register tab with background for per-session WS.
(function () {
  try {
    window.__BROWSER_AGENT_EXT__ = {
      product: "browser-agent",
      controlPort: 43761,
      version: "1.0.1",
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

  const session_id = readSessionIdFromPage();
  if (!session_id) return;

  const control_port = readControlPortFromPage();

  try {
    chrome.runtime.sendMessage({
      type: "register",
      session_id: session_id,
      control_port: control_port,
      tabId: null,
      windowId: null,
    });
  } catch (e) {
    /* extension context invalidated */
  }
})();