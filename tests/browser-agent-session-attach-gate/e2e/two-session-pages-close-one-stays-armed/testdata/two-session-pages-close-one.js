// argv: [node, script, baseURL, sessionId]
const baseURL = (process.argv[3] || '').replace(/\/$/, '');
const sessionId = process.argv[4] || '';
const USER_MARKER = 'LOOP_MARKER=attach-gate-two-pages';
const userURL = `https://example.com/?${USER_MARKER}`;

if (!baseURL || !sessionId) {
  console.error('usage: two-session-pages-close-one.js <baseURL> <sessionId>');
  process.exit(2);
}

await context.waitForEvent('serviceworker', { timeout: 20000 }).catch(() => null);

// Control tab 1 (first register).
const control1 = page;
await control1.goto(`${baseURL}/go?session=${encodeURIComponent(sessionId)}`);

let connected = false;
for (let i = 0; i < 40; i++) {
  const r = await fetch(`${baseURL}/v1/session?session=${encodeURIComponent(sessionId)}`);
  if (!r.ok) {
    await control1.waitForTimeout(500);
    continue;
  }
  const snap = await r.json();
  if (snap && snap.extension && snap.extension.connected) {
    connected = true;
    break;
  }
  await control1.waitForTimeout(500);
}
console.log(JSON.stringify({ assert: 'extension_connected', ok: connected, session_id: sessionId }));
if (!connected) {
  process.exit(1);
}

// Control tab 2 — becomes last-registered entry.tabId under current master.
const control2 = await context.newPage();
await control2.goto(`${baseURL}/go?session=${encodeURIComponent(sessionId)}`, {
  waitUntil: 'domcontentloaded',
});
await control2.waitForTimeout(1000);

// User tab (job target).
const userPage = await context.newPage();
await userPage.goto(userURL, { waitUntil: 'domcontentloaded' });
await userPage.bringToFront();
await control1.waitForTimeout(1500);

async function postJobMaybe(type, params, timeoutMs) {
  const controller = new AbortController();
  const timer = setTimeout(() => controller.abort(), timeoutMs || 20000);
  try {
    const r = await fetch(`${baseURL}/v1/jobs`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        session_id: sessionId,
        type,
        params,
        timeout_ms: timeoutMs || 20000,
      }),
      signal: controller.signal,
    });
    const text = await r.text();
    let body = null;
    try {
      body = JSON.parse(text);
    } catch (_) {
      body = { raw: text };
    }
    return { httpOk: r.ok, status: r.status, body };
  } catch (e) {
    return { httpOk: false, status: 0, body: null, error: String(e && e.message ? e.message : e) };
  } finally {
    clearTimeout(timer);
  }
}

function jobSucceeded(res) {
  if (!res || !res.httpOk || !res.body) return false;
  return res.body.ok === true;
}

function evalURLOf(res) {
  const v = res && res.body && res.body.data && res.body.data.value;
  if (v && typeof v === 'object') return v.url || '';
  return '';
}

const before = await postJobMaybe(
  'eval',
  {
    expression: '({ url: location.href, title: document.title })',
    expr: '({ url: location.href, title: document.title })',
  },
  20000,
);
const beforeOk = jobSucceeded(before) && evalURLOf(before).includes(USER_MARKER);
console.log(
  JSON.stringify({
    assert: 'eval_before_partial_close',
    ok: beforeOk,
    session_id: sessionId,
    eval_url: evalURLOf(before),
  }),
);
if (!beforeOk) {
  process.exit(1);
}

// Close the second control tab (last register under master) while control1 remains.
await control2.close();
await userPage.waitForTimeout(2500);
// Keep control1 focused enough that extension can rebind if needed.
await control1.bringToFront().catch(() => null);
await userPage.bringToFront();
await userPage.waitForTimeout(1000);

const after = await postJobMaybe(
  'eval',
  {
    expression: '({ url: location.href, title: document.title })',
    expr: '({ url: location.href, title: document.title })',
  },
  20000,
);
const afterURL = evalURLOf(after);
const staysArmed = jobSucceeded(after) && afterURL.includes(USER_MARKER);

console.log(
  JSON.stringify({
    assert: 'stays_armed_after_one_of_two_closed',
    ok: staysArmed,
    session_id: sessionId,
    after_job_ok: jobSucceeded(after),
    eval_url: afterURL,
    after_http_ok: after.httpOk,
    after_error: after.error || (after.body && after.body.error) || '',
  }),
);

process.exit(staysArmed ? 0 : 1);
