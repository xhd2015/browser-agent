// argv: [node, script, baseURL, sessionId]
const baseURL = (process.argv[3] || '').replace(/\/$/, '');
const sessionId = process.argv[4] || '';

if (!baseURL || !sessionId) {
  console.error('usage: warning-banner.js <baseURL> <sessionId>');
  process.exit(2);
}

if (context) {
  const sw =
    context.serviceWorkers()[0] ||
    (await context.waitForEvent('serviceworker', { timeout: 15000 }).catch(() => null));
}

await page.goto(`${baseURL}/go?session=${encodeURIComponent(sessionId)}`);

const banner = page.locator('[data-browser-agent-session-warning]').first();
const present = (await banner.count()) > 0;
const dataSessionId = present ? await banner.getAttribute('data-session-id') : null;
const idMatches = dataSessionId === sessionId;

console.log(
  JSON.stringify({
    assert: 'warning_banner_present',
    ok: present,
    session_id: sessionId,
  }),
);
console.log(
  JSON.stringify({
    assert: 'warning_banner_session_id',
    ok: idMatches,
    session_id: sessionId,
    data_session_id: dataSessionId,
  }),
);

if (!present || !idMatches) {
  process.exit(1);
}