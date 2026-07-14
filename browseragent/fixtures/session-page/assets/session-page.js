// fixture session-page asset
(function () {
  var el = document.getElementById("browser-agent-boot");
  var sid = "";
  try { sid = JSON.parse(el && el.textContent || "{}").session_id || ""; } catch (e) {}
  fetch("/v1/session" + (sid ? "?session=" + encodeURIComponent(sid) : "")).catch(function () {});
})();
