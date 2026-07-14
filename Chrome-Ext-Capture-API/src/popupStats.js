// Pure popup recording stats builder — mirrors Go package popupstats rules.
// Used by background getState enrichment (no Chrome APIs inside).

const OPAQUE_HOST = 'opaque';

/**
 * Parse hostname from request URL; invalid/empty → "opaque".
 * @param {string} raw
 * @returns {string}
 */
function hostFromURL(raw) {
  if (!raw) return OPAQUE_HOST;
  try {
    // Relative / scheme-less strings: URL() may throw or yield empty host.
    const u = new URL(raw);
    const host = u.hostname;
    if (!host) return OPAQUE_HOST;
    return host;
  } catch (_) {
    return OPAQUE_HOST;
  }
}

/**
 * Preferred entry key: `${tabId}:${requestId}` — tabId before first ':'.
 * @param {string} key
 * @returns {number|null}
 */
function tabIDFromKey(key) {
  if (typeof key !== 'string') return null;
  const idx = key.indexOf(':');
  if (idx <= 0) return null;
  const n = Number.parseInt(key.slice(0, idx), 10);
  if (!Number.isFinite(n) || String(n) !== key.slice(0, idx)) return null;
  return n;
}

/**
 * Top-N domains by count desc, host asc on ties.
 * @param {Map<string, number>|Object<string, number>} hosts
 * @param {number} n
 * @returns {{host: string, count: number}[]}
 */
function topDomains(hosts, n) {
  const list = [];
  if (hosts instanceof Map) {
    for (const [host, count] of hosts) list.push({ host, count });
  } else {
    for (const host of Object.keys(hosts || {})) {
      list.push({ host, count: hosts[host] });
    }
  }
  list.sort((a, b) => {
    if (a.count !== b.count) return b.count - a.count;
    return a.host < b.host ? -1 : a.host > b.host ? 1 : 0;
  });
  if (n > 0 && list.length > n) return list.slice(0, n);
  return list;
}

/**
 * Build enriched popup stats from in-memory capture state.
 *
 * @param {{
 *   entries: Object<string, {request?: {url?: string}, url?: string}|string>,
 *   attachedTabIds: Iterable<number>|number[],
 *   tabMeta?: Object<number, {title?: string, url?: string, active?: boolean}>
 * }} input
 * @returns {{
 *   count: number,
 *   tabsWatching: number,
 *   domainCount: number,
 *   tabs: Array<{
 *     tabId: number,
 *     title: string,
 *     url: string,
 *     active: boolean,
 *     attached: boolean,
 *     requestCount: number,
 *     domainCount: number,
 *     domains: {host: string, count: number}[]
 *   }>
 * }}
 */
export function buildPopupStats(input) {
  const entries = input && input.entries ? input.entries : {};
  const tabMeta = (input && input.tabMeta) || {};
  const attachedSet = new Set();
  const attachedList = input && input.attachedTabIds != null
    ? Array.from(input.attachedTabIds)
    : [];
  for (const id of attachedList) {
    if (id != null && Number.isFinite(Number(id))) attachedSet.add(Number(id));
  }

  /** @type {Map<number, {hosts: Map<string, number>, reqCount: number}>} */
  const byTab = new Map();
  const globalHosts = new Set();

  const entryKeys = Object.keys(entries);
  for (const key of entryKeys) {
    const raw = entries[key];
    let url = '';
    if (typeof raw === 'string') {
      url = raw;
    } else if (raw && typeof raw === 'object') {
      url = (raw.request && raw.request.url) || raw.url || '';
    }
    const host = hostFromURL(url);
    globalHosts.add(host);

    const tabId = tabIDFromKey(key);
    // Fallback: HAR-style _tabId on entry object.
    let tid = tabId;
    if (tid == null && raw && typeof raw === 'object' && raw._tabId != null) {
      const n = Number(raw._tabId);
      if (Number.isFinite(n)) tid = n;
    }
    if (tid == null) continue;

    let acc = byTab.get(tid);
    if (!acc) {
      acc = { hosts: new Map(), reqCount: 0 };
      byTab.set(tid, acc);
    }
    acc.reqCount += 1;
    acc.hosts.set(host, (acc.hosts.get(host) || 0) + 1);
  }

  const memberIds = new Set(attachedSet);
  for (const id of byTab.keys()) memberIds.add(id);

  const tabs = [];
  for (const id of memberIds) {
    const attached = attachedSet.has(id);
    const meta = tabMeta[id] || tabMeta[String(id)] || {};
    const acc = byTab.get(id);
    const reqCount = acc ? acc.reqCount : 0;
    const hosts = acc ? acc.hosts : new Map();
    const domainCount = hosts.size;
    const domains = topDomains(hosts, 3);

    let title = meta.title || '';
    if (!title) {
      title = attached ? `Tab ${id}` : 'Closed tab';
    }

    tabs.push({
      tabId: id,
      title,
      url: meta.url || '',
      active: !!meta.active,
      attached,
      requestCount: reqCount,
      domainCount,
      domains,
    });
  }

  tabs.sort((a, b) => {
    if (a.requestCount !== b.requestCount) return b.requestCount - a.requestCount;
    return a.tabId - b.tabId;
  });

  return {
    count: entryKeys.length,
    tabsWatching: attachedSet.size,
    domainCount: globalHosts.size,
    tabs,
  };
}

export { OPAQUE_HOST, hostFromURL, tabIDFromKey, topDomains };
