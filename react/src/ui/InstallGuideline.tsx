import type { ProductConfig } from "../products/types";

/** Operator browser for install steps. Default chrome preserves existing UX. */
export type InstallBrowser = "chrome" | "firefox";

export interface InstallGuidelineProps {
  product: ProductConfig;
  installPath?: string;
  /** Absolute filesystem path to signed .xpi when available. */
  xpiPath?: string;
  /** file:///… URL for the signed .xpi (copy/paste; click from http may be blocked). */
  xpiFileURL?: string;
  /** Control-plane download URL (reliable click-to-install from session page). */
  xpiHttpURL?: string;
  /** When true, expand the guideline panel by default. */
  defaultOpen?: boolean;
  /**
   * Browser path for install steps.
   * - chrome (default): chrome://extensions → Load unpacked
   * - firefox: signed .xpi (preferred) + about:debugging temporary add-on
   */
  browser?: InstallBrowser;
}

/**
 * InstallGuideline — Chrome chrome://extensions / Load unpacked, or Firefox
 * permanent .xpi + about:debugging / Load Temporary Add-on, parameterized by ProductConfig.
 */
export function InstallGuideline({
  product,
  installPath,
  xpiPath,
  xpiFileURL,
  xpiHttpURL,
  defaultOpen = true,
  browser = "chrome",
}: InstallGuidelineProps) {
  const port = product.controlPort;
  const isFirefox = browser === "firefox";
  const httpXpi = xpiHttpURL || "/v1/firefox-xpi";
  const fileURL =
    xpiFileURL ||
    (xpiPath
      ? xpiPath.startsWith("file:")
        ? xpiPath
        : `file://${xpiPath.startsWith("/") ? "" : "/"}${xpiPath}`
      : "");
  return (
    <details
      className="install-guideline"
      data-browser-agent-install={product.id === "browser-agent" ? "" : undefined}
      data-install-guideline
      data-install-browser={browser}
      open={defaultOpen}
    >
      <summary>Install {product.displayName} extension</summary>
      <div className="install-body">
        {isFirefox ? (
          <>
            <p>
              <strong>Permanent install (recommended)</strong> — open the
              signed package that connects to{" "}
              <code>127.0.0.1:{port}</code>:
            </p>
            <p data-firefox-xpi-http>
              <a href={httpXpi}>Download / open browser-agent.xpi</a>{" "}
              (<code>{httpXpi}</code>) — Firefox should prompt to install.
            </p>
            {fileURL ? (
              <p data-firefox-xpi-file style={{ wordBreak: "break-all" }}>
                Local file URL (paste into the address bar if the link is
                blocked from this page):{" "}
                <a href={fileURL}>{fileURL}</a>
              </p>
            ) : xpiPath ? (
              <p data-firefox-xpi-path style={{ wordBreak: "break-all" }}>
                Local path: <code>{xpiPath}</code>
              </p>
            ) : (
              <p className="muted">
                After <code>browser-agent install-firefox-extension</code>, a{" "}
                <code>file://</code> path appears here when the signed{" "}
                <code>.xpi</code> is embedded.
              </p>
            )}
            <p className="muted">
              Note: browsers often block navigating from <code>http://</code>{" "}
              pages to <code>file://</code> — prefer the download link above, or
              paste the file URL into the address bar.
            </p>

            <p>
              Or load a temporary add-on (unloads when Firefox restarts;
              path segment <code>browser-agent-firefox</code>):
            </p>
            <ol>
              <li>
                Open{" "}
                <strong>about:debugging#/runtime/this-firefox</strong>{" "}
                (or type <strong>about:debugging</strong> → This Firefox)
              </li>
              <li>
                Click <strong>Load Temporary Add-on…</strong>
              </li>
              <li>
                In the file picker, open the extension folder
                {installPath ? (
                  <>
                    {" "}
                    at <code>{installPath}</code>
                  </>
                ) : (
                  <>
                    {" "}
                    under <code>…/browser-agent-firefox/&lt;version&gt;/</code>
                  </>
                )}
              </li>
              <li>
                Select the file <strong>manifest.json</strong> (not the folder),
                then Open
              </li>
              <li>Keep the session page open so the extension can attach</li>
            </ol>
            <p className="muted">
              Or run: <code>browser-agent install-firefox-extension</code>
            </p>
          </>
        ) : (
          <>
            <p>
              Load the unpacked Chrome extension that connects to{" "}
              <code>127.0.0.1:{port}</code>.
            </p>
            <ol>
              <li>
                Open <strong>chrome://extensions</strong>
              </li>
              <li>
                Enable <strong>Developer mode</strong>
              </li>
              <li>
                Click <strong>Load unpacked</strong>
                {installPath ? (
                  <>
                    {" "}
                    and select <code>{installPath}</code>
                  </>
                ) : null}
              </li>
              <li>Keep the session page open so the extension can attach</li>
            </ol>
          </>
        )}
      </div>
    </details>
  );
}

export default InstallGuideline;
