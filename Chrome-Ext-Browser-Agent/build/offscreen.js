/**
 * Hidden offscreen document: pings the MV3 service worker every 20s so it is
 * not idle-killed while sessions are connected (Chrome resets SW idle on
 * messages from an offscreen document).
 */
(function () {
  var PORT_NAME = "ba-offscreen-keepalive";
  var PING_MS = 20000;
  var port = null;
  var timer = null;

  function connect() {
    try {
      if (port) return;
      port = chrome.runtime.connect({ name: PORT_NAME });
      port.onDisconnect.addListener(function () {
        port = null;
        setTimeout(connect, 1000);
      });
      port.postMessage({ type: "keepalive", source: "offscreen" });
    } catch (e) {
      port = null;
      setTimeout(connect, 2000);
    }
  }

  function ping() {
    try {
      if (!port) {
        connect();
        return;
      }
      port.postMessage({ type: "keepalive", source: "offscreen" });
    } catch (e) {
      port = null;
      connect();
    }
  }

  connect();
  timer = setInterval(ping, PING_MS);
})();
