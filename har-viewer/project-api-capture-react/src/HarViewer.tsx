import { useState, useEffect } from "react";
import "./HarViewer.css";

const ResourceTypes = {
  All: "all",
  XHR: "xhr",
  Fetch: "fetch",
  Script: "script",
  Stylesheet: "stylesheet",
  Image: "image",
  Font: "font",
  Other: "other",
} as const;

type ResourceType = (typeof ResourceTypes)[keyof typeof ResourceTypes];

interface EntrySummary {
  index: number;
  startedDateTime: string;
  time: number;
  method: string;
  url: string;
  host: string;
  path: string;
  status: number;
  statusText: string;
  mimeType: string;
  size: number;
  resourceType: string;
}

interface NameValue {
  name: string;
  value: string;
}

interface EntryDetail {
  startedDateTime: string;
  time: number;
  request: {
    method: string;
    url: string;
    httpVersion: string;
    headers: NameValue[];
    queryString: NameValue[];
    postData?: {
      mimeType: string;
      text: string;
    };
    headersSize: number;
    bodySize: number;
  };
  response: {
    status: number;
    statusText: string;
    httpVersion: string;
    headers: NameValue[];
    content: {
      size: number;
      mimeType: string;
      text?: string;
      encoding?: string;
    };
    headersSize: number;
    bodySize: number;
  };
  timings: {
    blocked: number;
    dns: number;
    connect: number;
    send: number;
    wait: number;
    receive: number;
    ssl: number;
  };
  _resourceType: string;
  serverIPAddress?: string;
}

const DetailTabs = {
  Headers: "headers",
  Query: "query",
  Request: "request",
  Response: "response",
  Timing: "timing",
} as const;

type DetailTab = (typeof DetailTabs)[keyof typeof DetailTabs];

function formatSize(bytes: number): string {
  if (bytes < 0) return "—";
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
}

function formatTime(ms: number): string {
  if (ms < 0) return "—";
  if (ms < 1) return `${(ms * 1000).toFixed(0)} µs`;
  if (ms < 1000) return `${ms.toFixed(1)} ms`;
  return `${(ms / 1000).toFixed(2)} s`;
}

function statusClass(status: number): string {
  if (status >= 200 && status < 300) return "status-ok";
  if (status >= 300 && status < 400) return "status-redirect";
  if (status >= 400 && status < 500) return "status-client-error";
  if (status >= 500) return "status-server-error";
  return "";
}

function methodClass(method: string): string {
  switch (method) {
    case "GET":
      return "method-get";
    case "POST":
      return "method-post";
    case "PUT":
      return "method-put";
    case "DELETE":
      return "method-delete";
    case "PATCH":
      return "method-patch";
    case "OPTIONS":
      return "method-options";
    default:
      return "";
  }
}

function tryFormatJSON(text: string): string {
  try {
    return JSON.stringify(JSON.parse(text), null, 2);
  } catch {
    return text;
  }
}

function HeadersTable({ headers }: { headers: NameValue[] }) {
  if (!headers || headers.length === 0) {
    return <div className="har-empty">No headers</div>;
  }
  return (
    <table className="har-kv-table">
      <thead>
        <tr>
          <th>Name</th>
          <th>Value</th>
        </tr>
      </thead>
      <tbody>
        {headers.map((h, i) => (
          <tr key={i}>
            <td className="har-kv-name">{h.name}</td>
            <td className="har-kv-value">{h.value}</td>
          </tr>
        ))}
      </tbody>
    </table>
  );
}

function TimingBar({ timings }: { timings: EntryDetail["timings"] }) {
  const segments = [
    { key: "blocked", label: "Blocked", value: timings.blocked, color: "#aaa" },
    { key: "dns", label: "DNS", value: timings.dns, color: "#6ecb63" },
    { key: "connect", label: "Connect", value: timings.connect, color: "#f5a623" },
    { key: "ssl", label: "SSL", value: timings.ssl, color: "#9b59b6" },
    { key: "send", label: "Send", value: timings.send, color: "#3498db" },
    { key: "wait", label: "Wait (TTFB)", value: timings.wait, color: "#2ecc71" },
    { key: "receive", label: "Receive", value: timings.receive, color: "#e74c3c" },
  ].filter((s) => s.value > 0);

  const total = segments.reduce((sum, s) => sum + s.value, 0);

  return (
    <div className="har-timing">
      <div className="har-timing-bar">
        {segments.map((s) => (
          <div
            key={s.key}
            className="har-timing-segment"
            style={{
              width: `${(s.value / total) * 100}%`,
              backgroundColor: s.color,
            }}
            title={`${s.label}: ${formatTime(s.value)}`}
          />
        ))}
      </div>
      <div className="har-timing-legend">
        {segments.map((s) => (
          <div key={s.key} className="har-timing-legend-item">
            <span className="har-timing-dot" style={{ backgroundColor: s.color }} />
            <span className="har-timing-label">{s.label}</span>
            <span className="har-timing-value">{formatTime(s.value)}</span>
          </div>
        ))}
        <div className="har-timing-legend-item har-timing-total">
          <span className="har-timing-label">Total</span>
          <span className="har-timing-value">{formatTime(total)}</span>
        </div>
      </div>
    </div>
  );
}

function EntryDetailPanel({
  entry,
  onClose,
}: {
  entry: EntryDetail;
  onClose: () => void;
}) {
  const [activeTab, setActiveTab] = useState<DetailTab>(DetailTabs.Headers);

  const tabs: { key: DetailTab; label: string; visible: boolean }[] = [
    { key: DetailTabs.Headers, label: "Headers", visible: true },
    {
      key: DetailTabs.Query,
      label: "Query",
      visible: entry.request.queryString.length > 0,
    },
    {
      key: DetailTabs.Request,
      label: "Request Body",
      visible: !!entry.request.postData,
    },
    { key: DetailTabs.Response, label: "Response", visible: true },
    { key: DetailTabs.Timing, label: "Timing", visible: true },
  ];

  return (
    <div className="har-detail-panel">
      <div className="har-detail-summary">
        <span className={`har-method ${methodClass(entry.request.method)}`}>
          {entry.request.method}
        </span>
        <span className="har-detail-url">{entry.request.url}</span>
        <button className="har-detail-close" onClick={onClose} title="Close">
          ×
        </button>
      </div>

      <div className="har-detail-meta">
        <span>
          Status:{" "}
          <span className={statusClass(entry.response.status)}>
            {entry.response.status} {entry.response.statusText}
          </span>
        </span>
        <span>Time: {formatTime(entry.time)}</span>
        <span>Size: {formatSize(entry.response.content.size)}</span>
        {entry.serverIPAddress && <span>IP: {entry.serverIPAddress}</span>}
      </div>

      <div className="har-detail-tabs">
        {tabs
          .filter((t) => t.visible)
          .map((t) => (
            <button
              key={t.key}
              className={`har-tab ${activeTab === t.key ? "active" : ""}`}
              onClick={() => setActiveTab(t.key)}
            >
              {t.label}
            </button>
          ))}
      </div>

      <div className="har-detail-content">
        {activeTab === DetailTabs.Headers && (
          <div>
            <h4>Request Headers</h4>
            <HeadersTable headers={entry.request.headers} />
            <h4>Response Headers</h4>
            <HeadersTable headers={entry.response.headers} />
          </div>
        )}
        {activeTab === DetailTabs.Query && (
          <div>
            <h4>Query Parameters</h4>
            <HeadersTable headers={entry.request.queryString} />
          </div>
        )}
        {activeTab === DetailTabs.Request && entry.request.postData && (
          <div>
            <div className="har-body-meta">
              MIME Type: {entry.request.postData.mimeType}
            </div>
            <pre className="har-body-content">
              {tryFormatJSON(entry.request.postData.text)}
            </pre>
          </div>
        )}
        {activeTab === DetailTabs.Response && (
          <div>
            <div className="har-body-meta">
              MIME Type: {entry.response.content.mimeType}
              {entry.response.content.encoding &&
                ` (${entry.response.content.encoding})`}
            </div>
            {entry.response.content.text ? (
              entry.response.content.encoding === "base64" ? (
                <div className="har-empty">
                  Binary content ({formatSize(entry.response.content.size)})
                </div>
              ) : (
                <pre className="har-body-content">
                  {tryFormatJSON(entry.response.content.text)}
                </pre>
              )
            ) : (
              <div className="har-empty">No response body</div>
            )}
          </div>
        )}
        {activeTab === DetailTabs.Timing && <TimingBar timings={entry.timings} />}
      </div>
    </div>
  );
}

function getFileFromURL(): string {
  const params = new URLSearchParams(window.location.search);
  return params.get("file") || "";
}

function setFileInURL(file: string) {
  const url = new URL(window.location.href);
  if (file) {
    url.searchParams.set("file", file);
  } else {
    url.searchParams.delete("file");
  }
  window.history.replaceState({}, "", url.toString());
}

export default function HarViewer() {
  const [files, setFiles] = useState<string[]>([]);
  const [activeFile, setActiveFile] = useState<string>("");
  const [entries, setEntries] = useState<EntrySummary[]>([]);
  const [selectedIndex, setSelectedIndex] = useState<number | null>(null);
  const [detail, setDetail] = useState<EntryDetail | null>(null);
  const [filter, setFilter] = useState("");
  const [typeFilter, setTypeFilter] = useState<ResourceType>(ResourceTypes.All);
  const [loading, setLoading] = useState(true);

  const selectFile = (file: string) => {
    setActiveFile(file);
    setFileInURL(file);
  };

  useEffect(() => {
    fetch("/api/har/files")
      .then((r) => r.json())
      .then((data: string[]) => {
        setFiles(data);
        const urlFile = getFileFromURL();
        if (urlFile && data.includes(urlFile)) {
          setActiveFile(urlFile);
        } else if (data.length > 0) {
          selectFile(data[0]);
        }
        setLoading(false);
      })
      .catch((err) => {
        console.error("Failed to load HAR files:", err);
        setLoading(false);
      });
  }, []);

  useEffect(() => {
    if (!activeFile) return;
    setLoading(true);
    setSelectedIndex(null);
    setDetail(null);
    fetch(`/api/har/entries?file=${encodeURIComponent(activeFile)}`)
      .then((r) => r.json())
      .then((data) => {
        setEntries(data);
        setLoading(false);
      })
      .catch((err) => {
        console.error("Failed to load HAR entries:", err);
        setLoading(false);
      });
  }, [activeFile]);

  useEffect(() => {
    if (selectedIndex === null || !activeFile) {
      setDetail(null);
      return;
    }
    fetch(
      `/api/har/entry?file=${encodeURIComponent(activeFile)}&index=${selectedIndex}`
    )
      .then((r) => r.json())
      .then(setDetail)
      .catch((err) => console.error("Failed to load entry detail:", err));
  }, [selectedIndex, activeFile]);

  const filtered = entries.filter((e) => {
    const matchesText =
      !filter ||
      e.url.toLowerCase().includes(filter.toLowerCase()) ||
      e.method.toLowerCase().includes(filter.toLowerCase());
    const matchesType =
      typeFilter === ResourceTypes.All ||
      e.resourceType === typeFilter ||
      (typeFilter === ResourceTypes.Other &&
        !(
          [
            ResourceTypes.XHR,
            ResourceTypes.Fetch,
            ResourceTypes.Script,
            ResourceTypes.Stylesheet,
            ResourceTypes.Image,
            ResourceTypes.Font,
          ] as string[]
        ).includes(e.resourceType));
    return matchesText && matchesType;
  });

  if (loading && entries.length === 0) {
    return <div className="har-loading">Loading HAR data...</div>;
  }

  if (files.length === 0 && !loading) {
    return (
      <div className="har-loading">
        No .har files found in the current directory.
      </div>
    );
  }

  return (
    <div className="har-viewer">
      <div className="har-toolbar">
        {files.length > 1 && (
          <select
            className="har-file-select"
            value={activeFile}
            onChange={(e) => selectFile(e.target.value)}
          >
            {files.map((f) => (
              <option key={f} value={f}>
                {f}
              </option>
            ))}
          </select>
        )}
        {files.length === 1 && (
          <span className="har-file-name">{activeFile}</span>
        )}
        <input
          className="har-search"
          type="text"
          placeholder="Filter by URL or method..."
          value={filter}
          onChange={(e) => setFilter(e.target.value)}
        />
        <div className="har-type-filters">
          {Object.entries(ResourceTypes).map(([label, value]) => (
            <button
              key={value}
              className={`har-type-btn ${typeFilter === value ? "active" : ""}`}
              onClick={() => setTypeFilter(value)}
            >
              {label}
            </button>
          ))}
        </div>
        <span className="har-count">
          {filtered.length} / {entries.length} requests
        </span>
      </div>

      <div className="har-content">
        <div className="har-list">
          <table className="har-table">
            <thead>
              <tr>
                <th className="col-status">Status</th>
                <th className="col-method">Method</th>
                <th className="col-host">Host</th>
                <th className="col-path">Path</th>
                <th className="col-type">Type</th>
                <th className="col-size">Size</th>
                <th className="col-time">Time</th>
              </tr>
            </thead>
            <tbody>
              {filtered.map((entry) => (
                <tr
                  key={entry.index}
                  className={`har-row ${selectedIndex === entry.index ? "selected" : ""}`}
                  onClick={() =>
                    setSelectedIndex(
                      selectedIndex === entry.index ? null : entry.index
                    )
                  }
                >
                  <td className={`col-status ${statusClass(entry.status)}`}>
                    {entry.status}
                  </td>
                  <td className={`col-method ${methodClass(entry.method)}`}>
                    {entry.method}
                  </td>
                  <td className="col-host" title={entry.host}>
                    {entry.host}
                  </td>
                  <td className="col-path" title={entry.path}>
                    {entry.path}
                  </td>
                  <td className="col-type">{entry.resourceType}</td>
                  <td className="col-size">{formatSize(entry.size)}</td>
                  <td className="col-time">{formatTime(entry.time)}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>

        {detail && (
          <div className="har-detail">
            <EntryDetailPanel
              entry={detail}
              onClose={() => {
                setSelectedIndex(null);
                setDetail(null);
              }}
            />
          </div>
        )}
      </div>
    </div>
  );
}
