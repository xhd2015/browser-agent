// Full-stack isolated E2E: real Chrome extension, browser-agent daemon, ai-critic HTTP and phone UI.
// Env: AI_CRITIC_BIN, BROWSER_AGENT_BIN, CHROME_PATH, NODE_PATH (puppeteer).
// No production browser profile, daemon, tabs, credentials, or data directory is used.
const fs = require('node:fs');
const path = require('node:path');
const os = require('node:os');
const http = require('node:http');
const net = require('node:net');
const { spawn } = require('node:child_process');
const assert = require('node:assert/strict');
const puppeteer = require('puppeteer');
const required = (key) => { if (!process.env[key]) throw Error(`${key} is required`); return process.env[key]; };
const agentBin = required('BROWSER_AGENT_BIN'), criticBin = required('AI_CRITIC_BIN'), chromeBin = required('CHROME_PATH');
const dir = fs.realpathSync(fs.mkdtempSync(path.join(os.tmpdir(), 'tab-close-stack-e2e-')));
const tmp = fs.realpathSync(os.tmpdir());
if (!dir.startsWith(tmp + path.sep + 'tab-close-stack-e2e-')) throw Error('unsafe test directory');
const ext = path.join(dir, 'extension');
fs.cpSync(path.join(__dirname, '../public'), ext, { recursive: true });
const bgPath = path.join(ext, 'background.js');
let bg = fs.readFileSync(bgPath, 'utf8');
const children = [], browsers = [];
let logs = '';
const log = (prefix, data) => { logs = (logs + prefix + data.toString()).slice(-30000); };
const bound = async (promise, ms, label) => { let timer; try { return await Promise.race([promise, new Promise((_, reject) => { timer = setTimeout(() => reject(Error(label + ' timed out')), ms); })]); } finally { clearTimeout(timer); } };
const freePort = () => new Promise((resolve, reject) => { const s = net.createServer(); s.on('error', reject); s.listen(0, '127.0.0.1', () => { const port = s.address().port; s.close(() => resolve(port)); }); });
const start = (bin, args, env) => { const child = spawn(bin, args, { env: { ...process.env, ...env }, stdio: ['ignore', 'pipe', 'pipe'] }); children.push(child); child.stdout.on('data', (s) => log('OUT ', s)); child.stderr.on('data', (s) => log('ERR ', s)); return child; };
async function waitFor(fn, label, ms = 90000) { const end = Date.now() + ms; let last; while (Date.now() < end) { try { const value = await fn(); if (value) return value; } catch (error) { last = error; } await new Promise((r) => setTimeout(r, 250)); } throw Error(`${label} timed out: ${last?.message || ''}`); }
const json = async (url, init = {}) => { const r = await fetch(url, init); const body = await r.json(); if (!r.ok) throw Error(`${r.status} ${JSON.stringify(body)}`); return body; };
const post = (url, body, headers = {}) => json(url, { method: 'POST', headers: { 'Content-Type': 'application/json', ...headers }, body: JSON.stringify(body) });
(async () => {
 let web;
 const content = http.createServer((_req, res) => { res.setHeader('Content-Type', 'text/html'); res.end(`<title>Isolated dirty tab</title><button style="width:120px;height:120px" onclick="document.body.dataset.clicked='yes'">edit</button><script>window.addEventListener('beforeunload',e=>{e.preventDefault();e.returnValue=''})</script>`); });
 try {
  const agentPort = await freePort(), criticPort = await freePort();
  await new Promise((resolve) => content.listen(0, '127.0.0.1', resolve));
  assert.ok(bg.includes('const CONTROL_PORT = 43761;')); bg = bg.replace('const CONTROL_PORT = 43761;', `const CONTROL_PORT = ${agentPort};`); fs.writeFileSync(bgPath, bg);
  const agent = `http://127.0.0.1:${agentPort}`, critic = `http://127.0.0.1:${criticPort}`;
  const dataDir = path.join(dir, 'critic-data'); fs.mkdirSync(dataDir); const credentials = path.join(dir, 'credentials'); fs.writeFileSync(credentials, 'isolated-test-token\n');
  start(agentBin, ['serve', '--host', '127.0.0.1', '--port', String(agentPort), '--base-dir', path.join(dir, 'agent'), '--no-agent-run'], {});
  await waitFor(async () => (await fetch(agent + '/v1/health')).ok, 'isolated daemon');
  await post(agent + '/v1/sessions', { session_id: 'ai-critic' });
  start(criticBin, ['--port', String(criticPort), '--no-event-bus-publish', '--dir', dir, '--credentials-file', credentials], { AI_CRITIC_HOME: dataDir, BROWSER_AGENT_PORT: String(agentPort), BROWSER_AGENT_NO_OPEN: '1' });
  await waitFor(async () => (await fetch(critic + '/browser')).ok, 'isolated ai-critic');
  const controlled = await puppeteer.launch({ headless: 'new', executablePath: chromeBin, ignoreDefaultArgs: ['--disable-extensions'], args: ['--no-sandbox', `--disable-extensions-except=${ext}`, `--load-extension=${ext}`] }); browsers.push(controlled);
  const control = await controlled.newPage(); await control.goto(agent + '/go?session=ai-critic', { waitUntil: 'domcontentloaded' });
  const target = await controlled.waitForTarget((t) => t.type() === 'service_worker' && t.url().startsWith('chrome-extension://'), { timeout: 60000 });
  const worker = await target.worker();
  await waitFor(async () => (await json(agent + '/v1/session?session=ai-critic')).extension?.connected, 'extension connection');
  const tabs = await bound(worker.evaluate(async ({ controlURL, contentURL }) => {
    const control = (await chrome.tabs.query({})).find((t) => t.url === controlURL);
    if (!control) throw Error('control tab missing');
    const clean = await chrome.tabs.create({ windowId: control.windowId, url: 'data:text/html,<title>Clean test tab</title>', active: false });
    const dirty = await chrome.tabs.create({ windowId: control.windowId, url: contentURL, active: true });
    return { clean: clean.id, dirty: dirty.id };
  }, { controlURL: agent + '/go?session=ai-critic', contentURL: `http://127.0.0.1:${content.address().port}/dirty` }), 60000, 'create isolated tabs');
  await waitFor(async () => worker.evaluate(async (id) => (await chrome.tabs.get(id)).status === 'complete', tabs.dirty), 'dirty tab load');
  const job = (method, params = {}) => post(agent + '/v1/jobs', { session_id: 'ai-critic', type: 'cdp', tab_id: tabs.dirty, timeout_ms: 30000, params: { method, params } });
  for (const type of ['mousePressed', 'mouseReleased']) { const result = await job('Input.dispatchMouseEvent', { type, x: 30, y: 30, button: 'left', clickCount: 1 }); if (!result.ok) throw Error(`click job: ${result.error}`); }
  const info = await post(agent + '/v1/jobs', { session_id: 'ai-critic', type: 'info', timeout_ms: 30000 });
  assert.ok(info.ok && info.data.features.includes('tab-close-confirm-v1'));
  web = await puppeteer.launch({ headless: 'new', executablePath: chromeBin, args: ['--no-sandbox'] }); browsers.push(web);
  const page = await web.newPage(); await page.setViewport({ width: 390, height: 844, isMobile: true, hasTouch: true });
  await page.setCookie({ name: 'ai-critic-token', value: 'isolated-test-token', url: critic });
  await page.setExtraHTTPHeaders({ Authorization: 'Bearer isolated-test-token' });
  await page.goto(critic + '/browser', { waitUntil: 'domcontentloaded', timeout: 90000 });
  await page.waitForSelector('.bwp-title', { timeout: 90000 }); await page.click('.bwp-title');
  await page.waitForFunction(() => Array.from(document.querySelectorAll('.bwp-tname')).some((e) => e.textContent.includes('Isolated dirty tab')), { timeout: 90000 });
  const clickDirtyClose = async () => { await page.evaluate(() => { const row = Array.from(document.querySelectorAll('.bwp-trow')).find((e) => e.textContent.includes('Isolated dirty tab')); if (!row) throw Error('dirty row missing'); row.querySelector('.bwp-tclose').click(); }); await page.waitForSelector('.bw-tabclose-card', { timeout: 90000 }); };
  const exists = (id) => worker.evaluate(async (id) => { try { await chrome.tabs.get(id); return true; } catch { return false; } }, id);
  await clickDirtyClose(); assert.match(await page.$eval('.bw-tabclose-card', (e) => e.textContent), /Unsaved changes/); assert.equal(await exists(tabs.dirty), true);
  await page.click('.bw-tabclose-actions button'); await page.waitForFunction(() => !document.querySelector('.bw-tabclose-card'), { timeout: 90000 }); assert.equal(await exists(tabs.dirty), true);
  await clickDirtyClose(); await page.click('.bw-tabclose-actions .is-danger'); await page.waitForFunction(() => !document.querySelector('.bw-tabclose-card'), { timeout: 90000 }); assert.equal(await exists(tabs.dirty), false);
  assert.equal(await exists(tabs.clean), true);
  console.log('PASS full stack: real extension + isolated daemon + ai-critic phone UI; keep and discard verified');
 } catch (e) {
  console.error(e); if (web) { const pages = await web.pages(); if (pages[0]) console.error('UI:', (await pages[0].evaluate(() => document.body.innerText).catch(() => '')).slice(0, 2500)); }
  console.error('PROCESS LOGS:', logs); process.exitCode = 1;
 } finally {
  for (const browser of browsers.reverse()) { try { await bound(browser.close(), 5000, 'browser cleanup'); } catch { browser.process()?.kill('SIGKILL'); } }
  for (const child of children.reverse()) if (child.exitCode === null && !child.killed) child.kill('SIGTERM');
  await Promise.all(children.map((child) => child.exitCode !== null || child.signalCode !== null ? Promise.resolve() : bound(new Promise((resolve) => child.once('exit', resolve)), 3000, 'child exit').catch(() => child.kill('SIGKILL'))));
  content.closeAllConnections(); content.close(); fs.rmSync(dir, { recursive: true, force: true });
 }
})().then(() => process.exit(process.exitCode || 0)).catch((error) => { console.error(error); process.exit(1); });
