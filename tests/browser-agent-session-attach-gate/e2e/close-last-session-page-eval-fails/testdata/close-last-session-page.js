// argv: [node, script, baseURL, sessionId]
const baseURL = (process.argv[3] || '').replace(/\/$/, '');
const sessionId = process.argv[4] || '';
const USER_MARKER = 'LOOP_MARKER=attach-gate-close-last';
const userURL = `https://example.com/?${USER_MARKER}`;

if (!baseURL || !sessionId) {
  console.error('usage: close-last-session-page.js <baseURL> <sessionId>');
  process.exit(2);
}

await context.waitForEvent('serviceworker', { timeout: 20000 }).catch(() => null);

const sessionPage = page;
await sessionPage.goto(`${baseURL}/go?session=${encodeURIComponent(sessionId)}`);

let connected = false;
for (let i = 0; i < 40; i++) {
  const r = await fetch(`${baseURL}/v1/session?session=${encodeURIComponent(sessionId)}`);
  if (!r.ok) {
    await sessionPage.waitForTimeout(500);
    continue;
  }
  const snap = await r.json();
  if (snap && snap.extension && snap.extension.connected) {
    connected = true;
    break;
  }
  await sessionPage.waitForTimeout(500);
}
console.log(JSON.stringify({ assert: 'extension_connected', ok: connected, session_id: sessionId }));
if (!connected) {
  process.exit(1);
}

const userPage = await context.newPage();
await userPage.goto(userURL, { waitUntil: 'domcontentloaded' });
await userPage.bringToFront();
await sessionPage.waitForTimeout(1500);

async function postJobMaybe(type, params, timeoutMs) {
  const controller = new AbortController();
  const timer = setTimeout(() => controller.abort(), timeoutMs || 15000);
  try {
    const r = await fetch(`${baseURL}/v1/jobs`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        session_id: sessionId,
        type,
        params,
        timeout_ms: timeoutMs || 15000,
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
  if (res.body.ok === true) {
    const v = (res.body.data && res.body.data.value) || null;
    // Successful eval typically returns a value object with url/title.
    if (v && typeof v === 'object' && (v.url || v.title)) return true;
    if (v != null && v !== false) return true;
    // ok:true with empty data still counts as job success for attach purposes.
    return true;
  }
  return false;
}

const before = await postJobMaybe(
  'eval',
  {
    expression: '({ url: location.href, title: document.title })',
    expr: '({ url: location.href, title: document.title })',
  },
  20000,
);
const beforeOk = jobSucceeded(before);
const beforeValue = before.body && before.body.data && before.body.data.value;
const beforeURL =
  beforeValue && typeof beforeValue === 'object' ? beforeValue.url || '' : '';
console.log(
  JSON.stringify({
    assert: 'eval_before_close',
    ok: beforeOk && beforeURL.includes(USER_MARKER),
    session_id: sessionId,
    job_ok: beforeOk,
    eval_url: beforeURL,
  }),
);
if (!beforeOk) {
  process.exit(1);
}

// Close the last session control tab.
await sessionPage.close();
// Allow leave handlers (unregister / detach / gate) to settle.
await userPage.waitForTimeout(2500);

const after = await postJobMaybe(
  'eval',
  {
    expression: '({ url: location.href, title: document.title })',
    expr: '({ url: location.href, title: document.title })',
  },
  12000,
);
const afterOk = jobSucceeded(after);
// Desired: must NOT succeed after last control tab is gone.
const failAssertOk = !afterOk;

console.log(
  JSON.stringify({
    assert: 'eval_fails_after_last_session_page_closed',
    ok: failAssertOk,
    session_id: sessionId,
    after_job_ok: afterOk,
    after_http_ok: after.httpOk,
    after_status: after.status,
    after_error: after.error || (after.body && after.body.error) || '',
  }),
);

process.exit(failAssertOk ? 0 : 1);
