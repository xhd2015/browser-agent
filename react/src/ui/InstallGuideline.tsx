import type { ProductConfig } from "../products/types";

/** Operator browser for install steps. Default chrome preserves existing UX. */
export type InstallBrowser = "chrome" | "firefox";

export interface InstallGuidelineProps {
  product: ProductConfig;
  installPath?: string;
  /** When true, expand the guideline panel by default. */
  defaultOpen?: boolean;
  /**
   * Browser path for install steps.
   * - chrome (default): chrome://extensions → Load unpacked
   * - firefox: about:debugging → Load Temporary Add-on → browser-agent-firefox
   */
  browser?: InstallBrowser;
}

/**
 * InstallGuideline — Chrome chrome://extensions / Load unpacked, or Firefox
 * about:debugging / Load Temporary Add-on, parameterized by ProductConfig.
 */
export function InstallGuideline({
  product,
  installPath,
  defaultOpen = true,
  browser = "chrome",
}: InstallGuidelineProps) {
  const port = product.controlPort;
  const isFirefox = browser === "firefox";

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
              Load the temporary Firefox add-on that connects to{" "}
              <code>127.0.0.1:{port}</code>. Canonical extract path segment:{" "}
              <code>browser-agent-firefox</code>.
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
              Temporary add-ons unload when Firefox restarts — re-run{" "}
              <code>browser-agent install-firefox-extension</code> and Load
              Temporary Add-on after a restart.
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
