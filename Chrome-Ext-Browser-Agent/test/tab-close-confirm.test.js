const test = require('node:test');
const assert = require('node:assert/strict');
const { createTabCloseController } = require('../public/tab-close-confirm.js');

function event() {
  const listeners = new Set();
  return { addListener: (fn) => listeners.add(fn), removeListener: (fn) => listeners.delete(fn), emit: (...args) => { for (const fn of [...listeners]) fn(...args); } };
}
function fixture({ dirty = true, waitMs = 50, decisionMs = 150, late = false, attachError = false } = {}) {
  const onEvent = event(), onDetach = event(), onRemoved = event(), actions = [];
  let exists = true;
  const chrome = {
    debugger: { onEvent, onDetach },
    tabs: { onRemoved, get: async () => { if (!exists) throw Error('No tab with id: 7'); return { id: 7 }; } },
  };
  const controller = createTabCloseController({
    chrome, waitMs, decisionMs, newId: () => 'operation-1',
    validate: async (session, tab) => { assert.equal(session, 'session-1'); assert.equal(tab, 7); },
    attach: async () => { if (attachError) throw Error('debugger unavailable'); },
    send: async (tab, method, params) => {
      actions.push({ tab, method, params });
      if (method === 'Page.close') {
        const fire = () => {
          if (dirty) onEvent.emit({ tabId: tab }, 'Page.javascriptDialogOpening', { type: 'beforeunload' });
          else { exists = false; onRemoved.emit(tab); }
        };
        if (late) setTimeout(fire, waitMs + 10); else queueMicrotask(fire);
      }
      if (method === 'Page.handleJavaScriptDialog') {
        onEvent.emit({ tabId: tab }, 'Page.javascriptDialogClosed', { result: params.accept });
        if (params.accept) { exists = false; onRemoved.emit(tab); }
      }
      return {};
    },
  });
  return { controller, actions, exists: () => exists, onEvent, onDetach, onRemoved };
}

test('clean tab closes only after Chrome reports removal', async () => {
  const f = fixture({ dirty: false });
  assert.deepEqual(await f.controller.start('session-1', 7), { status: 'closed', tab_id: 7 });
  assert.equal(f.exists(), false);
  assert.deepEqual(f.actions.map((a) => a.method), ['Page.enable', 'Page.close']);
});

test('dirty tab remains open until explicit keep or discard decision', async () => {
  for (const accept of [false, true]) {
    const f = fixture();
    assert.deepEqual(await f.controller.start('session-1', 7), { status: 'confirmation_required', operation_id: 'operation-1', tab_id: 7 });
    assert.equal(f.exists(), true);
    const result = await f.controller.decide('session-1', 'operation-1', accept);
    assert.deepEqual(result, { status: accept ? 'closed' : 'kept', tab_id: 7 });
    assert.equal(f.exists(), !accept);
    assert.deepEqual(f.actions.at(-1).params, { accept });
    assert.deepEqual(await f.controller.decide('session-1', 'operation-1', accept), result);
    await assert.rejects(f.controller.decide('session-1', 'operation-1', !accept), /different close decision/);
  }
});

test('another session cannot decide and duplicate starts are refused', async () => {
  const f = fixture();
  await f.controller.start('session-1', 7);
  await assert.rejects(f.controller.decide('other-session', 'operation-1', true), /not found/);
  await assert.rejects(f.controller.start('session-1', 7), /already pending/);
  assert.equal(f.exists(), true);
  await f.controller.decide('session-1', 'operation-1', false);
});

test('unanswered dialog is cancelled without discarding changes', async () => {
  const f = fixture({ decisionMs: 30 });
  await f.controller.start('session-1', 7);
  await new Promise((r) => setTimeout(r, 60));
  assert.equal(f.exists(), true);
  assert.deepEqual(f.actions.at(-1).params, { accept: false });
});

test('unknown outcome never auto-confirms a late dialog', async () => {
  const f = fixture({ waitMs: 10, decisionMs: 60, late: true });
  assert.deepEqual(await f.controller.start('session-1', 7), { status: 'unknown', tab_id: 7 });
  await new Promise((r) => setTimeout(r, 100));
  assert.equal(f.exists(), true);
  assert.deepEqual(f.actions.at(-1).params, { accept: false });
});

test('debugger disconnect releases the operation without accepting the dialog', async () => {
  const f = fixture();
  await f.controller.start('session-1', 7);
  f.onDetach.emit({ tabId: 7 }, 'canceled_by_user');
  await assert.rejects(f.controller.decide('session-1', 'operation-1', true), /no longer active/);
  assert.equal(f.exists(), true);
  assert.equal(f.actions.some((a) => a.method === 'Page.handleJavaScriptDialog'), false);
});

test('a decision made on the Mac is reconciled instead of repeated', async () => {
  const f = fixture();
  await f.controller.start('session-1', 7);
  f.onEvent.emit({ tabId: 7 }, 'Page.javascriptDialogClosed', { result: false });
  assert.deepEqual(await f.controller.decide('session-1', 'operation-1', true), { status: 'kept', tab_id: 7 });
  assert.equal(f.actions.some((a) => a.method === 'Page.handleJavaScriptDialog'), false);
});

test('attach failure never sends Page.close', async () => {
  const f = fixture({ attachError: true });
  await assert.rejects(f.controller.start('session-1', 7), /debugger unavailable/);
  assert.equal(f.exists(), true);
  assert.equal(f.actions.length, 0);
});
