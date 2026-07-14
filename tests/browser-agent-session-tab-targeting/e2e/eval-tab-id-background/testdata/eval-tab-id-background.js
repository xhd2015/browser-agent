// argv: [node, script, baseURL, sessionId]
const baseURL = (process.argv[3] || '').replace(/\/$/, '');
const sessionId = process.argv[4] || '';
const USER_MARKER = 'USER_MARKER=tab-id-active';
const BG_MARKER = 'BG_MARKER=tab-id-background';
const userURL = `https://example.com/?${USER_MARKER}`;
const bgURL = `https://example.org/?${BG_MARKER}`;

if (!baseURL || !sessionId) {
  console.error('usage: eval-tab-id-background.js <baseURL> <sessionId>');
  process.exit(2);
}

await context.waitForEvent('serviceworker', { timeout: 20000 }).catch(() => null);

// Tab 1: session control page (registers extension WS).
await page.goto(`${baseURL}/go?session=${encodeURIComponent(sessionId)}`);

let connected = false;
for (let i = 0; i < 40; i++) {
  const r = await fetch(`${baseURL}/v1/session?session=${encodeURIComponent(sessionId)}`);
  if (!r.ok) {
    await page.waitForTimeout(500);
    continue;
  }
  const snap = await r.json();
  if (snap && snap.extension && snap.extension.connected) {
    connected = true;
    break;
  }
  await page.waitForTimeout(500);
}
console.log(JSON.stringify({ assert: 'extension_connected', ok: connected, session_id: sessionId }));
if (!connected) {
  process.exit(1);
}

// Tab 2: user tab — bring to front (ACTIVE).
const userPage = await context.newPage();
await userPage.goto(userURL, { waitUntil: 'domcontentloaded' });
await userPage.bringToFront();

// Tab 3: background tab — open but restore user tab as ACTIVE.
const bgPage = await context.newPage();
await bgPage.goto(bgURL, { waitUntil: 'domcontentloaded' });
await userPage.bringToFront();
await page.waitForTimeout(1500);

async function postJob(type, params, tabId) {
  const body = {
    session_id: sessionId,
    type,
    params,
    timeout_ms: 25000,
  };
  if (tabId != null) {
    body.tab_id = tabId;
  }
  const r = await fetch(`${baseURL}/v1/jobs`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  });
  if (!r.ok) {
    throw new Error(`${type} job HTTP ${r.status}: ${await r.text()}`);
  }
  return r.json();
}

const infoRes = await postJob('info', {}, null);
const tabs = (infoRes.data && infoRes.data.tabs) || [];
const activeTab = tabs.find((t) => t.active) || null;
const bgTab = tabs.find((t) => (t.url || '').includes(BG_MARKER)) || null;
const userTabActive = !!(activeTab && (activeTab.url || '').includes(USER_MARKER));
const bgTabId = bgTab ? bgTab.id || bgTab.tab_id : null;

if (!bgTabId) {
  console.log(
    JSON.stringify({
      assert: 'tab_id_background_eval',
      ok: false,
      session_id: sessionId,
      reason: 'background_tab_not_found_in_info',
      tabs,
    }),
  );
  process.exit(1);
}

const evalRes = await postJob(
  'eval',
  {
    expression: '({ url: location.href, title: document.title })',
    expr: '({ url: location.href, title: document.title })',
  },
  bgTabId,
);
const evalValue = (evalRes.data && evalRes.data.value) || {};
const evalURL = typeof evalValue === 'object' && evalValue ? evalValue.url || '' : '';

const evalOnBackground = evalURL.includes(BG_MARKER);
const evalOnUser = evalURL.includes(USER_MARKER);
const routingCorrect = userTabActive && evalOnBackground && !evalOnUser;

console.log(
  JSON.stringify({
    assert: 'tab_id_background_eval',
    ok: routingCorrect,
    session_id: sessionId,
    active_tab_is_user: userTabActive,
    active_tab_url: activeTab ? activeTab.url || '' : '',
    background_tab_id: bgTabId,
    eval_url: evalURL,
    eval_on_background_tab: evalOnBackground,
    eval_on_user_tab: evalOnUser,
  }),
);

process.exit(routingCorrect ? 0 : 1);