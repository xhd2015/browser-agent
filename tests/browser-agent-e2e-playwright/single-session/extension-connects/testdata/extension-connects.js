// argv: [node, script, baseURL, sessionId]
const baseURL = (process.argv[3] || '').replace(/\/$/, '');
const sessionId = process.argv[4] || '';

if (!baseURL || !sessionId) {
  console.error('usage: extension-connects.js <baseURL> <sessionId>');
  process.exit(2);
}

// Wait for MV3 service worker.
const sw =
  (context && context.serviceWorkers && context.serviceWorkers()[0]) ||
  (context &&
    (await context.waitForEvent('serviceworker', { timeout: 15000 }).catch(() => null)));

await page.goto(`${baseURL}/go?session=${encodeURIComponent(sessionId)}`);

const deadline = Date.now() + 15000;
let connected = false;
while (Date.now() < deadline) {
  const snap = await page.evaluate(async (sid) => {
    const r = await fetch(`/v1/session?session=${encodeURIComponent(sid)}`);
    if (!r.ok) return null;
    return r.json();
  }, sessionId);
  if (snap && snap.extension && snap.extension.connected) {
    connected = true;
    break;
  }
  await page.waitForTimeout(500);
}

const line = {
  assert: 'extension_connected',
  ok: connected,
  session_id: sessionId,
};
console.log(JSON.stringify(line));

if (!connected) {
  process.exit(1);
}