// MV3 popup: no inline scripts allowed (CSP script-src 'self').
// bundle-sum.js must load first (see popup.html).
(function () {
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
  var hintEl = document.getElementById("conn-hint");
  if (!statusEl || !hintEl) return;

  fetch("http://127.0.0.1:43761/v1/health")
    .then(function (r) {
      if (!r.ok) throw new Error("health " + r.status);
      return r.json();
    })
    .then(function () {
      statusEl.textContent = "reachable";
      statusEl.className = "ok";
      hintEl.textContent =
        "Serve is up. Keep a session page open so this extension can attach via WebSocket.";
    })
    .catch(function () {
      statusEl.textContent = "unreachable";
      statusEl.className = "warn";
      hintEl.textContent =
        "Start browser-agent serve, then open the session page.";
    });
})();
