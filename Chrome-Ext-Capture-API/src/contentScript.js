// Content script on the browser-trace control origin (session page).
// Marks the page so the dashboard can detect the extension and wakes the agent.

(function () {
  const EXT_VERSION = '1.2.0';
  const FEATURES = ['browser-trace', 'multi-tab-window'];

  try {
    window.__BROWSER_TRACE_EXT__ = {
      version: EXT_VERSION,
      features: FEATURES.slice()
    };
  } catch (_) {
    // Page may be cross-origin sandboxed; ignore.
  }

  try {
    chrome.runtime.sendMessage({ action: 'wakeAgent' }, () => {
      // Ignore response / missing SW; chrome.runtime.lastError is expected if unloaded.
      void chrome.runtime.lastError;
    });
  } catch (_) {}
})();
