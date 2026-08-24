const test = require("node:test");
const assert = require("node:assert/strict");
const { Recorder, shouldCaptureURL } = require("../public/har-recorder.js");

function requestEvent(requestId, url, extra = {}) {
  return {
    requestId,
    wallTime: 1_700_000_000,
    timestamp: 10,
    request: {
      method: extra.method || "GET",
      url,
      headers: extra.headers || { Accept: "application/json" },
      postData: extra.postData,
      hasPostData: Boolean(extra.postData),
    },
  };
}

test("records request, response, timing, and response body", async () => {
  const commands = [];
  const recorder = new Recorder(42, {
    creatorVersion: "test",
    controlPort: 43761,
    sendCommand(method, params) {
      commands.push([method, params]);
      if (method === "Network.getResponseBody") {
        return Promise.resolve({ body: '{"ok":true}', base64Encoded: false });
      }
      return Promise.resolve({});
    },
  });
  recorder.handle("Network.requestWillBeSent", requestEvent("req-1", "https://example.test/api?a=1"));
  recorder.handle("Network.responseReceived", {
    requestId: "req-1",
    timestamp: 10.1,
    response: {
      status: 200,
      statusText: "OK",
      protocol: "h2",
      headers: { "content-type": "application/json", "set-cookie": "sid=abc; Path=/" },
      mimeType: "application/json",
      encodedDataLength: 12,
      timing: { dnsStart: 0, dnsEnd: 1, connectStart: 1, connectEnd: 2, sslStart: 1, sslEnd: 2, sendStart: 2, sendEnd: 3, receiveHeadersEnd: 5 },
    },
  });
  recorder.handle("Network.loadingFinished", { requestId: "req-1", timestamp: 10.2, encodedDataLength: 12 });
  const firstFlush = await recorder.flush(1000);
  const secondFlush = await recorder.flush(1000);

  const har = recorder.buildHAR();
  assert.equal(har.log.version, "1.2");
  assert.equal(har.log.entries.length, 1);
  const entry = har.log.entries[0];
  assert.equal(entry._tabId, 42);
  assert.equal(entry.request.queryString[0].name, "a");
  assert.equal(entry.response.status, 200);
  assert.equal(entry.response.cookies[0].name, "sid");
  assert.equal(entry.response.content.text, '{"ok":true}');
  assert.equal(firstFlush.timedOut, false);
  assert.equal(secondFlush.timedOut, false);
  assert.equal(commands.filter(([method]) => method === "Network.getResponseBody").length, 1);
});

test("flush timeout marks partial and ignores late response bodies", async () => {
  let resolveBody;
  let bodyCalls = 0;
  const lateBody = new Promise((resolve) => {
    resolveBody = resolve;
  });
  const recorder = new Recorder(9, {
    sendCommand(method) {
      if (method === "Network.getResponseBody") {
        bodyCalls += 1;
        return lateBody;
      }
      return Promise.resolve({});
    },
  });

  recorder.handle("Network.requestWillBeSent", requestEvent("slow", "https://example.test/slow"));
  recorder.handle("Network.responseReceived", {
    requestId: "slow",
    timestamp: 10.1,
    response: { status: 200, statusText: "OK", headers: {}, mimeType: "text/plain" },
  });
  recorder.handle("Network.loadingFinished", {
    requestId: "slow",
    timestamp: 10.2,
    encodedDataLength: 4,
  });

  const flush = await recorder.flush(5);
  const exported = recorder.buildHAR();
  const exportedJSON = JSON.stringify(exported);
  assert.equal(flush.timedOut, true);
  assert.equal(flush.partial, true);
  assert.equal(flush.pendingBodyCount, 1);
  assert.equal(recorder.partial, true);
  assert.match(recorder.errors[0].error, /flush timed out/);

  resolveBody({ body: "late", base64Encoded: false });
  await new Promise((resolve) => setImmediate(resolve));
  const retryFlush = await recorder.flush(100);

  assert.equal(retryFlush.timedOut, false);
  assert.equal(retryFlush.partial, true);
  assert.equal(bodyCalls, 1);
  assert.equal(JSON.stringify(exported), exportedJSON);
  assert.equal(recorder.buildHAR().log.entries[0].response.content.text, undefined);
});

test("keeps identical CDP request ids isolated per tab", () => {
  const first = new Recorder(1, {});
  const second = new Recorder(2, {});
  first.handle("Network.requestWillBeSent", requestEvent("same", "https://one.test/"));
  second.handle("Network.requestWillBeSent", requestEvent("same", "https://two.test/"));
  assert.equal(first.buildHAR().log.entries[0]._tabId, 1);
  assert.equal(second.buildHAR().log.entries[0]._tabId, 2);
  assert.equal(first.buildHAR().log.entries[0].request.url, "https://one.test/");
  assert.equal(second.buildHAR().log.entries[0].request.url, "https://two.test/");
});

test("records POST data, redirects, cache, and failures", () => {
  const recorder = new Recorder(7, {});
  recorder.handle("Network.requestWillBeSent", requestEvent("post", "https://example.test/login", {
    method: "POST",
    headers: { "content-type": "application/json", cookie: "a=1; b=2" },
    postData: '{"name":"a"}',
  }));
  recorder.handle("Network.requestServedFromCache", { requestId: "post" });
  recorder.handle("Network.loadingFailed", { requestId: "post", timestamp: 10.5, errorText: "net::ERR_FAILED" });

  recorder.handle("Network.requestWillBeSent", requestEvent("redirect", "https://example.test/old"));
  recorder.handle("Network.requestWillBeSent", {
    ...requestEvent("redirect", "https://example.test/new"),
    timestamp: 11,
    redirectResponse: {
      status: 302,
      statusText: "Found",
      protocol: "h2",
      headers: { location: "https://example.test/new" },
    },
  });

  const entries = recorder.buildHAR().log.entries;
  const post = entries.find((entry) => entry.request.url.endsWith("/login"));
  assert.equal(post.request.postData.text, '{"name":"a"}');
  assert.equal(post.request.cookies.length, 2);
  assert.equal(post._error, "net::ERR_FAILED");
  assert.ok(post.cache.beforeRequest);
  assert.equal(entries.filter((entry) => entry.request.url.includes("example.test")).length, 3);
});

test("excludes Browser Agent control traffic", () => {
  assert.equal(shouldCaptureURL("http://127.0.0.1:43761/v1/jobs", 43761), false);
  assert.equal(shouldCaptureURL("http://localhost:43761/go?session=x", 43761), false);
  assert.equal(shouldCaptureURL("http://[::1]:43761/go?session=x", 43761), false);
  assert.equal(shouldCaptureURL("https://example.test/go?session=x", 43761), true);
  assert.equal(shouldCaptureURL("https://example.test/application/go?session=x", 43761), true);
  assert.equal(shouldCaptureURL("https://example.test/api", 43761), true);
  const recorder = new Recorder(1, { controlPort: 43761 });
  recorder.handle("Network.requestWillBeSent", requestEvent("control", "http://127.0.0.1:43761/v1/har/artifact"));
  assert.equal(recorder.buildHAR().log.entries.length, 0);
});
