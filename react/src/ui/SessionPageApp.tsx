import { useEffect, useState } from "react";
import type { ProductConfig } from "../products/types";
import { browserAgentProduct } from "../products/browser-agent";
import { InstallGuideline, type InstallBrowser } from "./InstallGuideline";
import {
  formatUpgradeHintHeadline,
  resolveExtensionUpgradeHint,
} from "./extensionUpgradeHint";

export interface SessionPageAppProps {
  product?: ProductConfig;
  sessionId?: string;
  /** Force install browser path (chrome | firefox). Auto-detected when omitted. */
  browser?: InstallBrowser;
}

interface SessionSnap {
  session_id?: string;
  phase?: string;
  hint?: string;
  extension_install_path?: string;
  firefox_xpi_path?: string;
  firefox_xpi_url?: string;
  firefox_xpi_http_url?: string;
  extension_match?: string;
  browsers?: string[];
  browser?: string;
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
  attach?: {
    stage?: string;
    last_error?: string;
    register_attempts?: number;
  };
}

function isReceivingEndError(err?: string): boolean {
  return String(err || "")
    .toLowerCase()
    .includes("receiving end does not exist");
}

function dash(v?: string) {
  return v && v.length ? v : "—";
}

/**
 * Pure UA helper: Firefox product token → firefox, else chrome.
 * Prefer exported for unit extract / contract name.
 */
export function detectRuntimeBrowser(userAgent?: string): InstallBrowser {
  const ua =
    userAgent ??
    (typeof navigator !== "undefined" ? navigator.userAgent : "");
  if (ua.includes("Firefox/")) {
    return "firefox";
  }
  return "chrome";
}

/**
 * Install browser priority (highest → lowest):
 * 1. forced prop
 * 2. content-script marker window.__BROWSER_AGENT_EXT__.browser
 * 3. runtime UA via detectRuntimeBrowser()
 * 4. snap / boot / path (legacy)
 * 5. default chrome
 *
 * UA / EXT must win over chrome-stamped boot so Firefox tabs show
 * about:debugging install steps even when Create stamped chrome.
 */
export function resolveInstallBrowser(
  forced: InstallBrowser | undefined,
  snap: SessionSnap | null,
  pathHint: string,
): InstallBrowser {
  if (forced === "firefox" || forced === "chrome") {
    return forced;
  }

  // Content-script marker (live extension inject) over chrome-stamped boot.
  if (typeof window !== "undefined") {
    const ext = (window as unknown as {
      __BROWSER_AGENT_EXT__?: { browser?: string };
    }).__BROWSER_AGENT_EXT__;
    if (String(ext?.browser || "").toLowerCase() === "firefox") {
      return "firefox";
    }
  }

  // Live tab UA — Firefox tab with chrome-stamped meta still shows firefox install.
  if (detectRuntimeBrowser() === "firefox") {
    return "firefox";
  }

  // Session snap browsers array (hello telemetry / create path).
  const browsers = snap?.browsers;
  if (Array.isArray(browsers)) {
    for (const b of browsers) {
      if (String(b).toLowerCase() === "firefox") {
        return "firefox";
      }
    }
  }
  if (String(snap?.browser || "").toLowerCase() === "firefox") {
    return "firefox";
  }
  // Path segment browser-agent-firefox (canonical Firefox extract tree).
  // Use installPath local after UA/EXT so source-order contracts see UA first.
  const installPath = pathHint;
  if (installPath.includes("browser-agent-firefox")) {
    return "firefox";
  }
  // Boot / window.__BROWSER_AGENT (injected by injectSessionBoot).
  if (typeof window !== "undefined") {
    const ba = (window as unknown as {
      __BROWSER_AGENT?: { browser?: string };
    }).__BROWSER_AGENT;
    if (String(ba?.browser || "").toLowerCase() === "firefox") {
      return "firefox";
    }
    // browser-agent-boot JSON may carry browser.
    try {
      const el = document.getElementById("browser-agent-boot");
      if (el?.textContent) {
        const boot = JSON.parse(el.textContent) as {
          browser?: string;
          browsers?: string[];
        };
        if (String(boot.browser || "").toLowerCase() === "firefox") {
          return "firefox";
        }
        if (Array.isArray(boot.browsers)) {
          for (const b of boot.browsers) {
            if (String(b).toLowerCase() === "firefox") {
              return "firefox";
            }
          }
        }
      }
    } catch {
      /* ignore boot parse */
    }
  }
  return "chrome";
}

export function SessionPageApp({
  product = browserAgentProduct,
  sessionId,
  browser: browserProp,
}: SessionPageAppProps) {
  const [snap, setSnap] = useState<SessionSnap | null>(null);
  const sid =
    sessionId ||
    (typeof window !== "undefined"
      ? new URLSearchParams(window.location.search).get("session") || ""
      : "");

  useEffect(() => {
    if (!sid) return;
    document.title = `${sid} - Browser Agent`;
  }, [sid]);

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
  const xpiPath = snap?.firefox_xpi_path || "";
  const xpiFileURL = snap?.firefox_xpi_url || "";
  const xpiHttpURL = snap?.firefox_xpi_http_url || "/v1/firefox-xpi";
  const installBrowser = resolveInstallBrowser(browserProp, snap, installPath);
  const isFirefox = installBrowser === "firefox";
  const upgradeHint = resolveExtensionUpgradeHint({
    match,
    bundledVersion: bundled?.version,
    loadedVersion: connected ? loaded?.version : undefined,
    browser: installBrowser,
  });
  const showMd5Details = match !== "ok" && match !== "not_connected";

  return (
    <div className="session-page" data-product={product.id} data-control-port={product.controlPort}>
      <h1>{product.displayName}</h1>
      <p>
        Session <code>{sid || snap?.session_id || "…"}</code>
      </p>
      <p className="muted">
        Control port <strong>{product.controlPort}</strong> · product{" "}
        <code>{product.id}</code>
        {isFirefox ? (
          <>
            {" "}
            · browser <code>firefox</code>
          </>
        ) : null}
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
          <strong>Bundled (this serve)</strong>{" "}
          <code>{dash(bundled?.version)}</code>
        </div>
        <div>
          <strong>Loaded ({isFirefox ? "Firefox" : "Chrome"})</strong>{" "}
          <code>{connected ? dash(loaded?.version) : "—"}</code>
        </div>
        {upgradeHint ? (
          <div
            data-browser-agent-ext-upgrade
            data-upgrade-kind={upgradeHint.kind}
            style={{
              margin: "0.65rem 0 0.35rem",
              padding: "0.55rem 0.7rem",
              borderRadius: 6,
              background: "#fff6eb",
              border: "1px solid #f0c48a",
              color: "#8a4b00",
            }}
          >
            <div style={{ fontWeight: 600 }}>
              ⚠ {formatUpgradeHintHeadline(upgradeHint)}
            </div>
            <div style={{ marginTop: "0.35rem", fontSize: "0.9rem" }}>
              Run:{" "}
              <code style={{ wordBreak: "break-all" }}>{upgradeHint.command}</code>
            </div>
          </div>
        ) : null}
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
        {showMd5Details ? (
          <details
            data-browser-agent-ext-md5
            style={{ marginTop: "0.5rem", fontSize: "0.85rem" }}
          >
            <summary className="muted" style={{ cursor: "pointer" }}>
              Details (md5)
            </summary>
            <div style={{ marginTop: "0.35rem" }}>
              Bundled md5{" "}
              <code style={{ wordBreak: "break-all" }}>{dash(bundled?.md5)}</code>
            </div>
            <div>
              Loaded md5{" "}
              <code style={{ wordBreak: "break-all" }}>
                {connected ? dash(loaded?.bundle_md5) : "—"}
              </code>
            </div>
          </details>
        ) : (
          <div className="muted" style={{ fontSize: "0.85rem", marginTop: "0.35rem" }}>
            Bundled md5{" "}
            <code style={{ wordBreak: "break-all" }}>{dash(bundled?.md5)}</code>
            {" · "}
            Loaded md5{" "}
            <code style={{ wordBreak: "break-all" }}>
              {connected ? dash(loaded?.bundle_md5) : "—"}
            </code>
          </div>
        )}
        <p
          className="muted"
          style={{ fontSize: "0.85rem" }}
          data-browser-agent-ext-install-path
        >
          {isFirefox ? "Load Temporary Add-on from: " : "Load unpacked: "}
          <code style={{ wordBreak: "break-all" }}>
            {installPath || "…"}
          </code>
        </p>
        {isFirefox && (xpiFileURL || xpiPath) ? (
          <p
            className="muted"
            style={{ fontSize: "0.85rem", wordBreak: "break-all" }}
            data-browser-agent-firefox-xpi
          >
            Signed .xpi:{" "}
            {xpiFileURL ? (
              <a href={xpiFileURL} data-firefox-xpi-file>
                {xpiFileURL}
              </a>
            ) : (
              <code>{xpiPath}</code>
            )}
            {" · "}
            <a href={xpiHttpURL} data-firefox-xpi-http>
              {xpiHttpURL}
            </a>
          </p>
        ) : null}
      </section>

      {!connected ? (
        <>
          <InstallGuideline
            product={product}
            installPath={installPath}
            xpiPath={xpiPath}
            xpiFileURL={xpiFileURL}
            xpiHttpURL={isFirefox ? xpiHttpURL : undefined}
            defaultOpen
            browser={installBrowser}
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
            {isFirefox ? (
              <>
                <p style={{ margin: "0.5rem 0 0", fontSize: "0.9rem" }}>
                  Prefer the signed <code>.xpi</code> (download link above /{" "}
                  <a href={xpiHttpURL}>/v1/firefox-xpi</a>
                  {xpiFileURL ? (
                    <>
                      {" "}
                      or <code style={{ wordBreak: "break-all" }}>{xpiFileURL}</code>
                    </>
                  ) : null}
                  ). Temporary add-ons unload on restart: open{" "}
                  <code>about:debugging#/runtime/this-firefox</code>, click{" "}
                  <strong>Load Temporary Add-on…</strong>, and select{" "}
                  <code>manifest.json</code> under the{" "}
                  <code>browser-agent-firefox</code> extract path above.
                </p>
                <p className="muted" style={{ fontSize: "0.85rem" }}>
                  Or run: <code>browser-agent install-firefox-extension</code>
                </p>
              </>
            ) : (
              <>
                {isReceivingEndError(snap?.attach?.last_error) ? (
                  <p
                    style={{ margin: "0.5rem 0 0", fontSize: "0.9rem" }}
                    data-browser-agent-sw-asleep
                  >
                    Extension service worker looks asleep (
                    <code>Receiving end does not exist</code>). Open the{" "}
                    <strong>Browser Agent</strong> toolbar popup, or go to{" "}
                    <code>chrome://extensions</code> → Browser Agent →{" "}
                    <strong>Reload</strong>. Keep this <code>/go</code> tab open —
                    do not run <code>session new</code> again.
                  </p>
                ) : null}
                <p style={{ margin: "0.5rem 0 0", fontSize: "0.9rem" }}>
                  Chrome 137+ ignores <code>--load-extension</code>. Load unpacked
                  once from the path above (chrome://extensions → Developer mode →
                  Load unpacked).
                </p>
                <p className="muted" style={{ fontSize: "0.85rem" }}>
                  Or run: <code>browser-agent install-chrome-extension</code>
                </p>
              </>
            )}
          </details>
        </>
      ) : null}
    </div>
  );
}

export default SessionPageApp;
