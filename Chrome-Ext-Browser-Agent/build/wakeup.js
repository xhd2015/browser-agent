/**
 * Silent SW wake page: opened by session-new wait when attach stalls at
 * register_sent. Loading an extension page starts the MV3 service worker
 * (same path as the toolbar popup). Then heal open /go tabs and close.
 */
(function () {
  function closeSelf() {
    try {
      chrome.tabs.getCurrent(function (tab) {
        if (chrome.runtime.lastError) return;
        if (tab && tab.id != null) {
          chrome.tabs.remove(tab.id, function () {
            /* ignore */
          });
        }
      });
    } catch (e) {
      /* ignore */
    }
  }

  try {
    chrome.runtime.sendMessage({ type: "wakeup" }, function () {
      /* lastError ok — SW still started by loading this page */
      setTimeout(closeSelf, 300);
    });
  } catch (e) {
    setTimeout(closeSelf, 300);
  }
})();
