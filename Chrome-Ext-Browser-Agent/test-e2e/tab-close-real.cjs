// Isolated real-Chromium test. Run with NODE_PATH=<puppeteer-node_modules>
// and CHROME_PATH=<Chrome for Testing>; no normal Chrome profile is touched.
const assert = require('node:assert/strict');
const fs = require('node:fs');
const os = require('node:os');
const path = require('node:path');
const http = require('node:http');
const puppeteer = require('puppeteer');
const root = fs.mkdtempSync(path.join(os.tmpdir(), 'browser-agent-tab-close-e2e-'));
const checked = fs.realpathSync(root);
const tmp = fs.realpathSync(os.tmpdir());
if (!checked.startsWith(tmp + path.sep + 'browser-agent-tab-close-e2e-')) throw Error('unsafe test directory');
const ext = path.join(checked, 'extension');
fs.mkdirSync(ext);
fs.copyFileSync(path.join(__dirname, '../public/tab-close-confirm.js'), path.join(ext, 'tab-close-confirm.js'));
fs.writeFileSync(path.join(ext, 'manifest.json'), JSON.stringify({ manifest_version: 3, name: 'Isolated tab close test', version: '1.0.0', background: { service_worker: 'background.js' }, permissions: ['tabs', 'debugger'], host_permissions: ['<all_urls>'] }));
fs.writeFileSync(path.join(ext, 'background.js'), `
importScripts('tab-close-confirm.js');
const attached = new Set();
const controller = createTabCloseController({chrome, waitMs: 20000, decisionMs: 60000,
 newId: () => crypto.randomUUID(),
 validate: async (session, id) => { if (session !== 'test-session') throw Error('wrong session'); await chrome.tabs.get(id); },
 attach: async (_session, id) => { if (!attached.has(id)) { await chrome.debugger.attach({tabId:id},'1.3'); attached.add(id); } },
 send: (id, method, params) => chrome.debugger.sendCommand({tabId:id},method,params)});
globalThis.testController = controller;
globalThis.testPrepare = async (url, dirty) => {
 const tab = await chrome.tabs.create({url,active:true});
 for(let i=0;i<100;i++){const t=await chrome.tabs.get(tab.id);if(t.status==='complete')break;await new Promise(r=>setTimeout(r,100));}
 await chrome.debugger.attach({tabId:tab.id},'1.3'); attached.add(tab.id);
 await chrome.debugger.sendCommand({tabId:tab.id},'Page.enable',{});
 if(dirty){await chrome.debugger.sendCommand({tabId:tab.id},'Input.dispatchMouseEvent',{type:'mousePressed',x:30,y:30,button:'left',clickCount:1});await chrome.debugger.sendCommand({tabId:tab.id},'Input.dispatchMouseEvent',{type:'mouseReleased',x:30,y:30,button:'left',clickCount:1});}
 return tab.id;
};
globalThis.testExists = async id => {try{await chrome.tabs.get(id);return true}catch{return false}};
`);
const bound = async (promise, ms, name) => {
  let timer; try { return await Promise.race([promise, new Promise((_, reject) => { timer = setTimeout(() => reject(Error(name + ' timed out')), ms); })]); }
  finally { clearTimeout(timer); }
};
(async () => {
  const server = http.createServer((req, res) => {
    res.setHeader('Content-Type', 'text/html');
    res.end(`<button style="width:100px;height:100px" onclick="document.body.dataset.clicked='yes'">edit</button>${req.url === '/dirty' ? "<script>window.addEventListener('beforeunload',e=>{e.preventDefault();e.returnValue=''})</script>" : ''}`);
  });
  await new Promise((resolve) => server.listen(0, '127.0.0.1', resolve));
  let browser;
  try {
    browser = await puppeteer.launch({ headless: 'new', executablePath: process.env.CHROME_PATH || undefined,
      ignoreDefaultArgs: ['--disable-extensions'], args: ['--no-sandbox', `--disable-extensions-except=${ext}`, `--load-extension=${ext}`] });
    const target = await browser.waitForTarget((t) => t.type() === 'service_worker' && t.url().startsWith('chrome-extension://'), { timeout: 60000 });
    const worker = await target.worker();
    const origin = `http://127.0.0.1:${server.address().port}`;
    const prepare = (url, dirty) => bound(worker.evaluate(({ url, dirty }) => globalThis.testPrepare(url, dirty), { url, dirty }), 90000, 'prepare tab');
    const start = (id) => bound(worker.evaluate((id) => globalThis.testController.start('test-session', id), id), 90000, 'start close');
    const choose = (id, accept) => bound(worker.evaluate(({ id, accept }) => globalThis.testController.decide('test-session', id, accept), { id, accept }), 90000, 'decide');
    const exists = (id) => bound(worker.evaluate((id) => globalThis.testExists(id), id), 30000, 'inspect tab');
    const dirty = await prepare(origin + '/dirty', true);
    const first = await start(dirty);
    assert.equal(first.status, 'confirmation_required'); assert.equal(await exists(dirty), true);
    assert.equal((await choose(first.operation_id, false)).status, 'kept'); assert.equal(await exists(dirty), true);
    const second = await start(dirty);
    assert.equal(second.status, 'confirmation_required'); assert.equal(await exists(dirty), true);
    assert.equal((await choose(second.operation_id, true)).status, 'closed'); assert.equal(await exists(dirty), false);
    const clean = await prepare(origin + '/clean', false);
    assert.equal((await start(clean)).status, 'closed'); assert.equal(await exists(clean), false);
    console.log('PASS real Chromium: dirty/keep, dirty/discard and clean close');
  } finally {
    if (browser) {
      const child = browser.process();
      try { await bound(browser.close(), 5000, 'browser cleanup'); } catch { if (child && !child.killed) child.kill('SIGKILL'); }
    }
    server.close();
    fs.rmSync(checked, { recursive: true, force: true });
  }
})().catch((error) => { console.error(error); process.exitCode = 1; });
