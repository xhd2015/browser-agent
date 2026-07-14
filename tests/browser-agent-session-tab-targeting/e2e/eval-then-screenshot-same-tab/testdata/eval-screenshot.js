// argv: [node, script, baseURL, sessionId]
const baseURL = (process.argv[3] || '').replace(/\/$/, '');
const sessionId = process.argv[4] || '';
const PIN_MARKER = 'PIN_MARKER=eval-screenshot-same-tab';
const ACTIVE_MARKER = 'ACTIVE_MARKER=eval-screenshot-active';
const pinURL = `https://example.com/?${PIN_MARKER}`;
const activeURL = `https://example.org/?${ACTIVE_MARKER}`;

if (!baseURL || !sessionId) {
  console.error('usage: eval-screenshot.js <baseURL> <sessionId>');
  process.exit(2);
}

await context.waitForEvent('serviceworker', { timeout: 20000 }).catch(() => null);

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

// Pin tab (target) — not focused.
const pinPage = await context.newPage();
await pinPage.goto(pinURL, { waitUntil: 'domcontentloaded' });

// Active tab must differ from pin tab so ignored tab_id fails RED pre-implementation.
const activePage = await context.newPage();
await activePage.goto(activeURL, { waitUntil: 'domcontentloaded' });
await activePage.bringToFront();
await page.waitForTimeout(1200);

async function postJob(type, params, tabId) {
  const body = {
    session_id: sessionId,
    type,
    params,
    timeout_ms: 30000,
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
const pinTab = tabs.find((t) => (t.url || '').includes(PIN_MARKER)) || null;
const pinTabId = pinTab ? pinTab.id || pinTab.tab_id : null;
const activeIsPin = !!(activeTab && (activeTab.url || '').includes(ACTIVE_MARKER));

if (!pinTabId) {
  console.log(
    JSON.stringify({
      assert: 'eval_then_screenshot_same_tab',
      ok: false,
      session_id: sessionId,
      reason: 'pin_tab_not_found',
      tabs,
    }),
  );
  process.exit(1);
}

let evalOK = false;
let screenshotOK = false;
let evalURL = '';
let shotFormat = '';
let shotBase64 = '';

try {
  const evalRes = await postJob(
    'eval',
    {
      expression: '({ url: location.href, marker: ' + JSON.stringify(PIN_MARKER) + ' })',
      expr: '({ url: location.href, marker: ' + JSON.stringify(PIN_MARKER) + ' })',
    },
    pinTabId,
  );
  evalOK = !!(evalRes && evalRes.ok);
  const evalValue = (evalRes.data && evalRes.data.value) || {};
  evalURL = typeof evalValue === 'object' && evalValue ? evalValue.url || '' : '';
  if (evalURL.includes(PIN_MARKER)) {
    evalOK = evalOK && true;
  } else {
    evalOK = false;
  }
} catch (e) {
  evalOK = false;
}

try {
  const shotRes = await postJob('screenshot', { format: 'png', full_page: false }, pinTabId);
  screenshotOK = !!(shotRes && shotRes.ok);
  const data = (shotRes.data && shotRes.data) || {};
  shotFormat = data.format || '';
  shotBase64 = data.base64 || data.data || '';
  if (!shotBase64) {
    screenshotOK = false;
  }
} catch (e) {
  screenshotOK = false;
}

const ok = activeIsPin && evalOK && screenshotOK && evalURL.includes(PIN_MARKER);

console.log(
  JSON.stringify({
    assert: 'eval_then_screenshot_same_tab',
    ok,
    session_id: sessionId,
    pin_tab_id: pinTabId,
    same_tab_id: true,
    active_tab_is_pin_target: activeIsPin,
    active_tab_url: activeTab ? activeTab.url || '' : '',
    eval_ok: evalOK,
    screenshot_ok: screenshotOK,
    eval_url: evalURL,
    eval_on_pin_tab: evalURL.includes(PIN_MARKER),
    screenshot_format: shotFormat,
    screenshot_has_base64: !!shotBase64,
  }),
);

process.exit(ok ? 0 : 1);