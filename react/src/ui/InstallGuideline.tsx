import type { ProductConfig } from "../products/types";

export interface InstallGuidelineProps {
  product: ProductConfig;
  installPath?: string;
  /** When true, expand the guideline panel by default. */
  defaultOpen?: boolean;
}

/**
 * InstallGuideline — chrome://extensions / Load unpacked steps parameterized
 * by ProductConfig (browser-agent vs browser-trace ports and names).
 */
export function InstallGuideline({
  product,
  installPath,
  defaultOpen = true,
}: InstallGuidelineProps) {
  const port = product.controlPort;
  return (
    <details
      className="install-guideline"
      data-browser-agent-install={product.id === "browser-agent" ? "" : undefined}
      data-install-guideline
      open={defaultOpen}
    >
      <summary>Install {product.displayName} extension</summary>
      <div className="install-body">
        <p>
          Load the unpacked Chrome extension that connects to{" "}
          <code>
            127.0.0.1:{port}
          </code>
          .
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
      </div>
    </details>
  );
}

export default InstallGuideline;
