import { useEffect, useState } from "react";
import type { ProductConfig } from "../products/types";
import { browserAgentProduct } from "../products/browser-agent";
import { InstallGuideline } from "./InstallGuideline";

export interface SessionPageAppProps {
  product?: ProductConfig;
  sessionId?: string;
}

interface SessionSnap {
  session_id?: string;
  phase?: string;
  hint?: string;
  extension_install_path?: string;
  extension_match?: string;
  bundled_extension?: {
    version?: string;
    md5?: string;
    path?: string;
  };
  extension?: {
    connected?: boolean;
    version?: string;
    bundle_md5?: string;
    supports_browser_agent?: boolean;
  };
}

function dash(v?: string) {
  return v && v.length ? v : "—";
}

export function SessionPageApp({
  product = browserAgentProduct,
  sessionId,
}: SessionPageAppProps) {
  const [snap, setSnap] = useState<SessionSnap | null>(null);
  const sid =
    sessionId ||
    (typeof window !== "undefined"
      ? new URLSearchParams(window.location.search).get("session") || ""
      : "");

  useEffect(() => {
    if (!sid) return;
    let cancelled = false;
    const poll = () => {
      fetch(`/v1/session?session=${encodeURIComponent(sid)}`)
        .then((r) => r.json())
        .then((j) => {
          if (!cancelled) setSnap(j);
        })
        .catch(() => {
          /* ignore */
        });
    };
    poll();
    const t = setInterval(poll, 500);
    return () => {
      cancelled = true;
      clearInterval(t);
    };
  }, [sid]);

  const connected = !!snap?.extension?.connected;
  const match = snap?.extension_match || "not_connected";
  const bundled = snap?.bundled_extension;
  const loaded = snap?.extension;
  const installPath =
    snap?.extension_install_path || bundled?.path || "";

  return (
    <div className="session-page" data-product={product.id} data-control-port={product.controlPort}>
      <h1>{product.displayName}</h1>
      <p>
        Session <code>{sid || snap?.session_id || "…"}</code>
      </p>
      <p className="muted">
        Control port <strong>{product.controlPort}</strong> · product{" "}
        <code>{product.id}</code>
      </p>
      <div data-browser-agent-status>
        <div>
          <strong>Phase:</strong> {snap?.phase || "…"}
        </div>
        <div>
          <strong>Extension:</strong>{" "}
          {connected ? "connected" : "not connected"}
        </div>
        <div className="hint">{snap?.hint || "Loading status…"}</div>
      </div>

      <section
        className="ext-identity"
        data-browser-agent-ext-identity
        style={{
          border: "1px solid #ccc",
          borderRadius: 8,
          padding: "0.75rem 1rem",
          margin: "1rem 0",
        }}
      >
        <h2 style={{ fontSize: "1.05rem", margin: "0 0 0.5rem" }}>
          Extension package
        </h2>
        <div>
          <strong>Bundled (this serve)</strong> version{" "}
          <code>{dash(bundled?.version)}</code> md5{" "}
          <code style={{ wordBreak: "break-all" }}>{dash(bundled?.md5)}</code>
        </div>
        <div>
          <strong>Loaded (Chrome)</strong> version{" "}
          <code>{connected ? dash(loaded?.version) : "—"}</code> md5{" "}
          <code style={{ wordBreak: "break-all" }}>
            {connected ? dash(loaded?.bundle_md5) : "—"}
          </code>
        </div>
        <div>
          <strong>Match:</strong>{" "}
          <span
            style={{
              fontWeight: 600,
              color:
                match === "ok"
                  ? "#0a7a2f"
                  : match === "not_connected"
                    ? "#666"
                    : "#c45c00",
            }}
          >
            {match}
          </span>
        </div>
        <p
          className="muted"
          style={{ fontSize: "0.85rem" }}
          data-browser-agent-ext-install-path
        >
          Load unpacked:{" "}
          <code style={{ wordBreak: "break-all" }}>
            {installPath || "…"}
          </code>
        </p>
      </section>

      {!connected ? (
        <>
          <InstallGuideline
            product={product}
            installPath={installPath}
            defaultOpen
          />
          <details
            className="troubleshoot-panel"
            data-browser-agent-troubleshoot
            style={{
              border: "1px solid #ddd",
              borderRadius: 8,
              padding: "0.75rem 1rem",
              margin: "1rem 0",
            }}
          >
            <summary>Troubleshoot extension connection</summary>
            <p style={{ margin: "0.5rem 0 0", fontSize: "0.9rem" }}>
              Chrome 137+ ignores <code>--load-extension</code>. Load unpacked
              once from the path above (chrome://extensions → Developer mode →
              Load unpacked).
            </p>
            <p className="muted" style={{ fontSize: "0.85rem" }}>
              Or run: <code>browser-agent install-chrome-extension</code>
            </p>
          </details>
        </>
      ) : null}
    </div>
  );
}

export default SessionPageApp;
