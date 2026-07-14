// Page marker for browser-agent product detection + feature advertising.
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
})();
