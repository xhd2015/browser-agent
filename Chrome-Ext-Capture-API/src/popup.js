import './popup.css';

(function () {
  const statusDot = document.getElementById('statusDot');
  const statusText = document.getElementById('statusText');
  const actionBtn = document.getElementById('actionBtn');
  const requestCount = document.getElementById('requestCount');
  const previewBtn = document.getElementById('previewBtn');
  const clearBtn = document.getElementById('clearBtn');
  const statsChips = document.getElementById('statsChips');
  const chipTabs = document.getElementById('chipTabs');
  const chipDomains = document.getElementById('chipDomains');
  const chipRequests = document.getElementById('chipRequests');
  const tabsSection = document.getElementById('tabsSection');
  const tabsList = document.getElementById('tabsList');

  const CLEAR_CONFIRM = 'Discard all captured requests so far?';

  function escapeHtml(s) {
    return String(s == null ? '' : s)
      .replace(/&/g, '&amp;')
      .replace(/</g, '&lt;')
      .replace(/>/g, '&gt;')
      .replace(/"/g, '&quot;');
  }

  function hideStats() {
    if (statsChips) statsChips.classList.add('hidden');
    if (tabsSection) tabsSection.classList.add('hidden');
    if (tabsList) tabsList.innerHTML = '';
    if (requestCount) requestCount.textContent = '';
  }

  function renderStats(stats) {
    if (!stats || !statsChips || !tabsSection || !tabsList) return;
    const tabsWatching = stats.tabsWatching != null ? stats.tabsWatching : 0;
    const domainCount = stats.domainCount != null ? stats.domainCount : 0;
    const count = stats.count != null ? stats.count : 0;
    const tabs = Array.isArray(stats.tabs) ? stats.tabs : [];

    chipTabs.textContent = String(tabsWatching);
    chipDomains.textContent = String(domainCount);
    chipRequests.textContent = String(count);
    statsChips.classList.remove('hidden');
    tabsSection.classList.remove('hidden');

    // Clear legacy single-line request count when chips are shown.
    if (requestCount) requestCount.textContent = '';

    tabsList.innerHTML = tabs.map((tab) => {
      const title = escapeHtml(tab.title || `Tab ${tab.tabId}`);
      const req = tab.requestCount != null ? tab.requestCount : 0;
      const dcount = tab.domainCount != null ? tab.domainCount : 0;
      const domains = Array.isArray(tab.domains) ? tab.domains : [];
      const more = dcount > domains.length ? dcount - domains.length : 0;
      const detached = !tab.attached;
      const pills = domains.map((d) => {
        const host = escapeHtml(d.host);
        const c = d.count != null ? d.count : 0;
        return `<span class="domain-pill" title="${host}">${host}<span class="dom-count">×${c}</span></span>`;
      }).join('');
      const moreEl = more > 0 ? `<span class="domain-more">+${more} more</span>` : '';
      const badge = detached ? `<span class="tab-badge">closed</span>` : '';
      return (
        `<div class="tab-row${detached ? ' detached' : ''}" data-tab-id="${tab.tabId}">` +
          `<div class="tab-row-header">` +
            `<span class="tab-title" title="${title}">${title}${badge}</span>` +
            `<span class="tab-counts">${req} req · ${dcount} dom</span>` +
          `</div>` +
          (pills || moreEl ? `<div class="tab-domains">${pills}${moreEl}</div>` : '') +
        `</div>`
      );
    }).join('');
  }

  function updateUI(state, count, hasPreview, serverSession, stats) {
    if (state === 'recording') {
      statusDot.className = 'status-dot recording';
      statusText.textContent = serverSession ? 'Recording (browser-trace)...' : 'Recording...';
      // During server-driven session: Stop only (Start hidden/disabled).
      actionBtn.textContent = serverSession ? 'Stop' : 'Stop & Download HAR';
      actionBtn.className = 'btn btn-stop';
      actionBtn.disabled = false;
      actionBtn.classList.remove('hidden');
      // Live preview + clear while recording.
      if (clearBtn) clearBtn.classList.remove('hidden');
      previewBtn.textContent = 'Preview';
      previewBtn.classList.remove('hidden');
      if (stats) {
        renderStats(stats);
      } else {
        // Fallback if background has not yet enriched getState.
        hideStats();
        requestCount.textContent = count !== undefined ? `Requests: ${count}` : '';
      }
    } else if (serverSession) {
      // Server session about to start / waiting — no manual Start.
      statusDot.className = 'status-dot idle';
      statusText.textContent = 'Waiting for browser-trace...';
      actionBtn.textContent = 'Stop';
      actionBtn.className = 'btn btn-stop';
      actionBtn.disabled = true;
      actionBtn.classList.remove('hidden');
      if (clearBtn) clearBtn.classList.add('hidden');
      previewBtn.classList.add('hidden');
      hideStats();
    } else {
      statusDot.className = 'status-dot idle';
      statusText.textContent = 'Idle';
      actionBtn.textContent = 'Start Recording';
      actionBtn.className = 'btn btn-start';
      actionBtn.disabled = false;
      actionBtn.classList.remove('hidden');
      if (clearBtn) clearBtn.classList.add('hidden');
      previewBtn.textContent = 'Preview Last Recording';
      if (hasPreview) {
        previewBtn.classList.remove('hidden');
      } else {
        previewBtn.classList.add('hidden');
      }
      hideStats();
    }
  }

  function getCurrentTabId() {
    return new Promise((resolve, reject) => {
      chrome.tabs.query({ active: true, currentWindow: true }, (tabs) => {
        if (chrome.runtime.lastError) return reject(chrome.runtime.lastError);
        if (!tabs || !tabs.length) return reject(new Error('No active tab'));
        resolve(tabs[0].id);
      });
    });
  }

  function sendMessage(action, extra) {
    return new Promise((resolve, reject) => {
      chrome.runtime.sendMessage({ action, ...extra }, (response) => {
        if (chrome.runtime.lastError) return reject(chrome.runtime.lastError);
        resolve(response);
      });
    });
  }

  function applyStateResponse(resp) {
    if (!resp) {
      updateUI('idle');
      return;
    }
    const stats = {
      count: resp.count,
      tabsWatching: resp.tabsWatching,
      domainCount: resp.domainCount,
      tabs: resp.tabs,
    };
    const hasStats = resp.tabsWatching != null || Array.isArray(resp.tabs);
    updateUI(
      resp.state,
      resp.count,
      resp.hasPreview,
      resp.serverSession,
      hasStats ? stats : null
    );
  }

  async function init() {
    try {
      const resp = await sendMessage('getState');
      applyStateResponse(resp);
    } catch (_) {
      updateUI('idle');
    }
  }

  previewBtn.addEventListener('click', async () => {
    try {
      const result = await sendMessage('openPreview');
      // Fallback if background did not open a tab.
      if (!result || result.ok === false) {
        chrome.tabs.create({ url: chrome.runtime.getURL('preview.html') });
      }
    } catch (_) {
      chrome.tabs.create({ url: chrome.runtime.getURL('preview.html') });
    }
  });

  if (clearBtn) {
    clearBtn.addEventListener('click', async () => {
      if (!confirm(CLEAR_CONFIRM)) {
        return;
      }
      clearBtn.disabled = true;
      try {
        const result = await sendMessage('clearCaptured');
        if (result && result.error) {
          statusText.textContent = 'Clear failed: ' + result.error;
        }
        try {
          applyStateResponse(await sendMessage('getState'));
        } catch (_) {}
      } catch (e) {
        statusText.textContent = 'Clear failed: ' + e.message;
      } finally {
        clearBtn.disabled = false;
      }
    });
  }

  actionBtn.addEventListener('click', async () => {
    actionBtn.disabled = true;
    try {
      const { state, serverSession } = await sendMessage('getState');
      if (state === 'recording') {
        const result = await sendMessage('stopRecording');
        if (result && result.error) throw new Error(result.error);
        // Local download only for manual sessions (server sessions save via CLI).
        if (result && result.harJson && !result.wasServer && !serverSession) {
          const blob = new Blob([result.harJson], { type: 'application/har+json' });
          const url = URL.createObjectURL(blob);
          chrome.downloads.download({ url, filename: `recording-${Date.now()}.har`, saveAs: true });
        }
        updateUI('idle', undefined, true, false, null);
      } else if (serverSession) {
        // Stop-only mode: no Start during server-driven session.
        statusText.textContent = 'Waiting for browser-trace session...';
        actionBtn.disabled = true;
      } else {
        const tabId = await getCurrentTabId();
        const result = await sendMessage('startRecording', { tabId });
        if (result && result.error) throw new Error(result.error);
        // Refresh full state (stats may still be empty right after start).
        try {
          applyStateResponse(await sendMessage('getState'));
        } catch (_) {
          updateUI('recording', 0, false, false, {
            count: 0,
            tabsWatching: 1,
            domainCount: 0,
            tabs: [],
          });
        }
      }
    } catch (e) {
      statusText.textContent = 'Error: ' + e.message;
      actionBtn.disabled = false;
    }
  });

  setInterval(async () => {
    try {
      applyStateResponse(await sendMessage('getState'));
    } catch (_) {}
  }, 1000);

  init();
})();
