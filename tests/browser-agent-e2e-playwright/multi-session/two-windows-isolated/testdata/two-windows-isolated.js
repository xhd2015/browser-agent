// argv: [node, script, baseURL, sessionIdA, sessionIdB]
const baseURL = (process.argv[3] || '').replace(/\/$/, '');
const sessionIdA = process.argv[4] || '';
const sessionIdB = process.argv[5] || '';

if (!baseURL || !sessionIdA || !sessionIdB) {
  console.error('usage: two-windows-isolated.js <baseURL> <sessionIdA> <sessionIdB>');
  process.exit(2);
}

if (context) {
  const sw =
    context.serviceWorkers()[0] ||
    (await context.waitForEvent('serviceworker', { timeout: 15000 }).catch(() => null));
}

const pageA = page;
const pageB = await context.newPage();

await pageA.goto(`${baseURL}/go?session=${encodeURIComponent(sessionIdA)}`);
await pageB.goto(`${baseURL}/go?session=${encodeURIComponent(sessionIdB)}`);

async function waitConnected(pg, sid) {
  const deadline = Date.now() + 15000;
  while (Date.now() < deadline) {
    const connected = await pg.evaluate(async (id) => {
      const r = await fetch(`/v1/session?session=${encodeURIComponent(id)}`);
      if (!r.ok) return false;
      const snap = await r.json();
      return !!(snap.extension && snap.extension.connected);
    }, sid);
    if (connected) return true;
    await pg.waitForTimeout(500);
  }
  return false;
}

const connA = await waitConnected(pageA, sessionIdA);
const connB = await waitConnected(pageB, sessionIdB);

console.log(
  JSON.stringify({
    assert: 'session_a_connected',
    ok: connA,
    session_id: sessionIdA,
  }),
);
console.log(
  JSON.stringify({
    assert: 'session_b_connected',
    ok: connB,
    session_id: sessionIdB,
  }),
);

if (!connA || !connB) {
  process.exit(1);
}