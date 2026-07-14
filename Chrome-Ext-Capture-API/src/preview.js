import './preview.css';

(function () {
  let entries = [];

  const entriesBody = document.getElementById('entriesBody');
  const entryCount = document.getElementById('entryCount');
  const filterInput = document.getElementById('filterInput');
  const statusFilter = document.getElementById('statusFilter');
  const detailPanel = document.getElementById('detailPanel');
  const detailTitle = document.getElementById('detailTitle');
  const detailRequest = document.getElementById('detailRequest');
  const detailResponse = document.getElementById('detailResponse');
  const detailHeaders = document.getElementById('detailHeaders');
  const closeDetail = document.getElementById('closeDetail');
  const emptyState = document.getElementById('emptyState');

  function methodClass(method) {
    const cls = { GET: 'get', POST: 'post', PUT: 'put', PATCH: 'patch', DELETE: 'delete', HEAD: 'head', OPTIONS: 'options' };
    return cls[method] || '';
  }

  function statusClass(code) {
    if (code >= 200 && code < 300) return 'status-2xx';
    if (code >= 300 && code < 400) return 'status-3xx';
    if (code >= 400 && code < 500) return 'status-4xx';
    if (code >= 500) return 'status-5xx';
    return '';
  }

  function formatBytes(bytes) {
    if (!bytes || bytes <= 0) return '-';
    if (bytes < 1024) return bytes + ' B';
    if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB';
    return (bytes / (1024 * 1024)).toFixed(1) + ' MB';
  }

  function formatTime(ms) {
    if (ms == null || ms < 0) return '-';
    if (ms < 1000) return ms.toFixed(0) + ' ms';
    return (ms / 1000).toFixed(2) + ' s';
  }

  function getTypeFromURL(url) {
    const ext = url.split('?')[0].split('.').pop().toLowerCase();
    const types = {
      js: 'Script', css: 'Stylesheet', png: 'Image', jpg: 'Image',
      jpeg: 'Image', gif: 'Image', webp: 'Image', svg: 'Image',
      ico: 'Image', woff: 'Font', woff2: 'Font', ttf: 'Font',
      mp4: 'Media', webm: 'Media', json: 'Fetch', xml: 'Fetch',
      html: 'Document', htm: 'Document'
    };
    return types[ext] || 'XHR';
  }

  function render() {
    const filterText = filterInput.value.toLowerCase();
    const statusVal = statusFilter.value;

    const filtered = entries.filter(e => {
      const url = e.request.url.toLowerCase();
      const method = e.request.method.toLowerCase();
      const status = e.response.status;
      if (filterText && !url.includes(filterText) && !method.includes(filterText)) return false;
      if (statusVal === '2xx' && (status < 200 || status >= 300)) return false;
      if (statusVal === '3xx' && (status < 300 || status >= 400)) return false;
      if (statusVal === '4xx' && (status < 400 || status >= 500)) return false;
      if (statusVal === '5xx' && (status < 500 || status >= 600)) return false;
      return true;
    });

    entryCount.textContent = filtered.length + ' / ' + entries.length + ' entries';

    if (filtered.length === 0) {
      entriesBody.innerHTML = '';
      emptyState.classList.remove('hidden');
      return;
    }
    emptyState.classList.add('hidden');

    const rows = filtered.map(e => {
      const method = e.request.method;
      const url = e.request.url;
      const status = e.response.status || '-';
      const time = e.time;
      const size = e.response.content ? e.response.content.size : 0;
      const truncatedURL = url.length > 100 ? url.substring(0, 97) + '...' : url;
      const type = e.response.content.mimeType ? e.response.content.mimeType.split('/')[1] || e.response.content.mimeType : getTypeFromURL(url);

      return `<tr data-idx="${entries.indexOf(e)}">
        <td class="col-method"><span class="method-badge ${methodClass(method)}">${method}</span></td>
        <td class="col-url" title="${url}">${truncatedURL}</td>
        <td class="col-status"><span class="status-badge ${statusClass(status)}">${status}</span></td>
        <td class="col-type">${type}</td>
        <td class="col-time">${formatTime(time)}</td>
        <td class="col-size">${formatBytes(size)}</td>
      </tr>`;
    }).join('');
    entriesBody.innerHTML = rows;

    entriesBody.querySelectorAll('tr').forEach(row => {
      row.addEventListener('click', () => {
        const idx = parseInt(row.dataset.idx);
        showDetail(entries[idx]);
      });
    });
  }

  function showDetail(entry) {
    const req = entry.request;
    const res = entry.response;
    detailTitle.textContent = req.method + ' ' + req.url;
    detailRequest.textContent = JSON.stringify({
      method: req.method,
      url: req.url,
      httpVersion: req.httpVersion,
      headers: req.headers,
      queryString: req.queryString,
      postData: req.postData,
      cookies: req.cookies,
      headersSize: req.headersSize,
      bodySize: req.bodySize
    }, null, 2);
    detailResponse.textContent = JSON.stringify({
      status: res.status,
      statusText: res.statusText,
      httpVersion: res.httpVersion,
      cookies: res.cookies,
      redirectURL: res.redirectURL,
      headersSize: res.headersSize,
      bodySize: res.bodySize,
      content: {
        size: res.content ? res.content.size : 0,
        mimeType: res.content ? res.content.mimeType : '',
        encoding: res.content ? res.content.encoding : undefined,
        text: res.content ? res.content.text : undefined
      }
    }, null, 2);
    detailHeaders.textContent = JSON.stringify({
      requestHeaders: req.headers,
      responseHeaders: res.headers
    }, null, 2);
    detailPanel.classList.remove('hidden');
  }

  closeDetail.addEventListener('click', () => {
    detailPanel.classList.add('hidden');
  });

  filterInput.addEventListener('input', render);
  statusFilter.addEventListener('change', render);

  function applyHarJson(harJson) {
    if (!harJson) {
      entries = [];
      emptyState.textContent = 'No requests captured yet (empty / cleared).';
      emptyState.classList.remove('hidden');
      entriesBody.innerHTML = '';
      entryCount.textContent = '0 entries';
      return;
    }
    try {
      const har = typeof harJson === 'string' ? JSON.parse(harJson) : harJson;
      entries = (har && har.log && har.log.entries) ? har.log.entries : [];
      render();
      if (!entries.length) {
        emptyState.textContent = 'No requests captured yet (empty / cleared).';
        emptyState.classList.remove('hidden');
      }
    } catch (e) {
      emptyState.textContent = 'Failed to parse HAR data';
      emptyState.classList.remove('hidden');
    }
  }

  function loadPreview() {
    // Prefer live in-memory entries from background (recording or last snapshot).
    try {
      chrome.runtime.sendMessage({ action: 'getPreview' }, (resp) => {
        if (chrome.runtime.lastError) {
          // Fall back to storage.
          chrome.storage.local.get('lastHarJson', (result) => {
            if (result.lastHarJson) {
              applyHarJson(result.lastHarJson);
            } else {
              emptyState.textContent = 'No recording data found. Capture some traffic first.';
              emptyState.classList.remove('hidden');
            }
          });
          return;
        }
        if (resp && resp.harJson) {
          applyHarJson(resp.harJson);
        } else {
          chrome.storage.local.get('lastHarJson', (result) => {
            if (result.lastHarJson) {
              applyHarJson(result.lastHarJson);
            } else {
              emptyState.textContent = 'No recording data found. Capture some traffic first.';
              emptyState.classList.remove('hidden');
            }
          });
        }
      });
    } catch (_) {
      chrome.storage.local.get('lastHarJson', (result) => {
        if (result.lastHarJson) {
          applyHarJson(result.lastHarJson);
        } else {
          emptyState.textContent = 'No recording data found. Capture some traffic first.';
          emptyState.classList.remove('hidden');
        }
      });
    }
  }

  loadPreview();
  // Refresh while viewing live capture (extension fallback preview).
  setInterval(loadPreview, 1000);

  document.addEventListener('keydown', (e) => {
    if (e.key === 'Escape') detailPanel.classList.add('hidden');
  });
})();
