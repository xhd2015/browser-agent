// argv: [node, script, baseURL, sessionId]
const baseURL = (process.argv[3] || '').replace(/\/$/, '');
const sessionId = process.argv[4] || '';
const USER_MARKER = 'LOOP_MARKER=active-tab-routing';
const userURL = `https://example.com/?${USER_MARKER}`;

if (!baseURL || !sessionId) {
  console.error('usage: active-tab-routing.js <baseURL> <sessionId>');
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
console.log(
  JSON.stringify({
    assert: 'extension_connected',
    ok: connected,
    session_id: sessionId,
  }),
);
if (!connected) {
  process.exit(1);
}

// Tab 2: user content — must be the ACTIVE tab when eval runs.
const userPage = await context.newPage();
await userPage.goto(userURL, { waitUntil: 'domcontentloaded' });
await userPage.bringToFront();
await page.waitForTimeout(1500);

async function postJob(type, params) {
  const r = await fetch(`${baseURL}/v1/jobs`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      session_id: sessionId,
      type,
      params,
      timeout_ms: 20000,
    }),
  });
  if (!r.ok) {
    throw new Error(`${type} job HTTP ${r.status}: ${await r.text()}`);
  }
  return r.json();
}

const infoRes = await postJob('info', {});
const tabs = (infoRes.data && infoRes.data.tabs) || [];
const activeTab = tabs.find((t) => t.active) || null;
const userTabActive = !!(activeTab && (activeTab.url || '').includes(USER_MARKER));

const evalRes = await postJob('eval', {
  expression: '({ url: location.href, title: document.title })',
  expr: '({ url: location.href, title: document.title })',
});
const evalValue = (evalRes.data && evalRes.data.value) || {};
const evalURL = typeof evalValue === 'object' && evalValue ? evalValue.url || '' : '';

const evalOnSessionPage = evalURL.includes('/go?session=');
const routingCorrect = userTabActive && evalURL.includes(USER_MARKER);
const bugPresent = userTabActive && evalOnSessionPage;

console.log(
  JSON.stringify({
    assert: 'active_tab_routing',
    ok: routingCorrect,
    session_id: sessionId,
    user_tab_active: userTabActive,
    active_tab_url: activeTab ? activeTab.url || '' : '',
    eval_url: evalURL,
    eval_on_session_page: evalOnSessionPage,
    bug_present: bugPresent,
  }),
);

process.exit(routingCorrect ? 0 : 1);