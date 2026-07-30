// argv: [node, script, baseURL, sessionId]
const baseURL = (process.argv[3] || '').replace(/\/$/, '');
const sessionId = process.argv[4] || '';
const USER_MARKER = 'LOOP_MARKER=attach-gate-open';
const userURL = `https://example.com/?${USER_MARKER}`;

if (!baseURL || !sessionId) {
  console.error('usage: session-open-eval.js <baseURL> <sessionId>');
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

// Tab 2: user content — active when eval runs.
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

const evalRes = await postJob('eval', {
  expression: '({ url: location.href, title: document.title })',
  expr: '({ url: location.href, title: document.title })',
});
const evalValue = (evalRes && evalRes.data && evalRes.data.value) || {};
const evalURL = typeof evalValue === 'object' && evalValue ? evalValue.url || '' : '';
const jobOk = !!(evalRes && evalRes.ok === true);
const routingCorrect = jobOk && evalURL.includes(USER_MARKER);

console.log(
  JSON.stringify({
    assert: 'session_open_eval',
    ok: routingCorrect,
    session_id: sessionId,
    job_ok: jobOk,
    eval_url: evalURL,
  }),
);

process.exit(routingCorrect ? 0 : 1);
