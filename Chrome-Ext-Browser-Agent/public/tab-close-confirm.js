/* Remote, user-confirmed beforeunload handling for one Chrome tab. */
(function (root) {
  function createTabCloseController({ chrome, validate, attach, send, newId, waitMs = 10000, decisionMs = 60000 }) {
    const operations = new Map();
    const byTab = new Map();
    async function within(promise, ms, fallback) {
      let timer;
      try { return await Promise.race([promise, new Promise((resolve) => { timer = setTimeout(() => resolve(fallback), ms); })]); }
      finally { clearTimeout(timer); }
    }

    function signal(op, value) {
      if (op.wake) { const wake = op.wake; op.wake = null; wake(value); }
    }
    function release(op) {
      if (op.released) return;
      op.released = true;
      clearTimeout(op.expiry);
      chrome.debugger.onEvent.removeListener(op.onEvent);
      chrome.debugger.onDetach.removeListener(op.onDetach);
      chrome.tabs.onRemoved.removeListener(op.onRemoved);
      if (byTab.get(op.tabId) === op.id) byTab.delete(op.tabId);
    }
    function forgetLater(op) {
      setTimeout(() => { if (operations.get(op.id) === op) operations.delete(op.id); }, decisionMs);
    }
    async function tabExists(tabId) {
      const query = chrome.tabs.get(tabId).then(() => true, (error) =>
        /no tab with id|not found/i.test(String(error && error.message)) ? false : null);
      return within(query, waitMs, null);
    }
    async function observed(op, timeout) {
      if (op.status === 'closed' || op.status === 'dialog' || op.status === 'error') return op.status;
      return within(new Promise((resolve) => { op.wake = resolve; }), timeout, 'timeout');
    }
    async function start(sessionId, tabId) {
      tabId = Number(tabId);
      if (!Number.isSafeInteger(tabId) || tabId <= 0) throw new Error('invalid tab id');
      if (byTab.has(tabId)) throw new Error('a close operation is already pending for this tab');
      const op = { id: newId(), sessionId, tabId, status: 'starting', released: false, choice: null, decision: null };
      operations.set(op.id, op); byTab.set(tabId, op.id);
      op.onEvent = (source, method, params) => {
        if (!source || source.tabId !== tabId) return;
        if (method === 'Page.javascriptDialogOpening') {
          if (params && params.type === 'beforeunload') {
            op.status = 'dialog';
            signal(op, 'dialog');
            if (!op.expiry) op.expiry = setTimeout(() => { void decide(sessionId, op.id, false).catch(() => {}); }, decisionMs);
          } else {
            op.status = 'error'; op.error = 'A different Chrome dialog is blocking this tab; handle it on the Mac.';
            signal(op, 'error');
          }
        } else if (method === 'Page.javascriptDialogClosed' && op.status === 'dialog' && op.choice === null) {
          op.status = 'externally_handled';
        }
      };
      op.onRemoved = (id) => { if (id === tabId) { op.status = 'closed'; signal(op, 'closed'); release(op); forgetLater(op); } };
      op.onDetach = (source, reason) => {
        if (source && source.tabId === tabId && op.status !== 'closed' && reason !== 'target_closed') {
          op.status = 'error'; op.error = 'Chrome debugger disconnected; tab close outcome is unknown.'; signal(op, 'error');
          release(op); forgetLater(op);
        }
      };
      chrome.debugger.onEvent.addListener(op.onEvent);
      chrome.debugger.onDetach.addListener(op.onDetach);
      chrome.tabs.onRemoved.addListener(op.onRemoved);
      try {
        await validate(sessionId, tabId);
        await attach(sessionId, tabId);
        if (!(await within(send(tabId, 'Page.enable', {}).then(() => true), waitMs, false))) {
          throw new Error('Chrome debugger did not respond; the tab was not closed');
        }
        // Page.close may acknowledge before beforeunload is answered; only a
        // dialog event or tabs.onRemoved determines the result, not that ack.
        void send(tabId, 'Page.close', {}).catch((err) => {
          if (op.status === 'starting') { op.status = 'error'; op.error = err.message || String(err); signal(op, 'error'); }
        });
        const state = await observed(op, waitMs);
        if (state === 'dialog') return { status: 'confirmation_required', operation_id: op.id, tab_id: tabId };
        if (state === 'closed') { forgetLater(op); return { status: 'closed', tab_id: tabId }; }
        if (state === 'error') throw new Error(op.error || 'Chrome could not close the tab');
        // The Page.close command cannot be retracted. Keep listeners long enough
        // to safely cancel a late beforeunload; never auto-confirm or retry it.
        op.wake = null;
        op.expiry = setTimeout(() => { if (op.status === 'dialog') void decide(sessionId, op.id, false).catch(() => {}); else release(op); }, decisionMs);
        forgetLater(op);
        return { status: 'unknown', tab_id: tabId };
      } catch (error) {
        release(op); operations.delete(op.id); throw error;
      }
    }
    async function decide(sessionId, operationId, accept) {
      const op = operations.get(operationId);
      if (!op || op.sessionId !== sessionId) throw new Error('close confirmation expired or not found');
      if (typeof accept !== 'boolean') throw new Error('close confirmation requires an explicit choice');
      if (op.choice !== null) {
        if (op.choice !== accept) throw new Error('a different close decision is already in progress');
        return op.decision;
      }
      op.choice = accept;
      op.decision = (async () => {
        if (op.status === 'closed') return { status: 'closed', tab_id: op.tabId };
        if (op.status === 'externally_handled') {
          const exists = await tabExists(op.tabId);
          release(op); forgetLater(op);
          return { status: exists === true ? 'kept' : exists === false ? 'closed' : 'unknown', tab_id: op.tabId };
        }
        if (op.status !== 'dialog') {
          release(op); forgetLater(op);
          throw new Error('close confirmation is no longer active; check the tab');
        }
        try {
          clearTimeout(op.expiry);
          op.status = 'deciding';
          const handled = await within(send(op.tabId, 'Page.handleJavaScriptDialog', { accept }).then(() => true), waitMs, false);
          if (!handled) return { status: 'unknown', tab_id: op.tabId };
          if (accept) {
            // An acknowledgement is not proof the tab disappeared.
            const outcome = await observed(op, waitMs);
            if (outcome === 'closed' || (await tabExists(op.tabId)) === false) return { status: 'closed', tab_id: op.tabId };
            return { status: 'unknown', tab_id: op.tabId };
          }
          return { status: (await tabExists(op.tabId)) === true ? 'kept' : 'unknown', tab_id: op.tabId };
        } finally { release(op); forgetLater(op); }
      })();
      return op.decision;
    }
    return { start, decide };
  }
  root.createTabCloseController = createTabCloseController;
  if (typeof module !== 'undefined' && module.exports) module.exports = { createTabCloseController };
})(globalThis);
