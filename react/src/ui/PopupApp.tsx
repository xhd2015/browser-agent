import type { ProductConfig } from "../products/types";
import { browserAgentProduct } from "../products/browser-agent";

export interface PopupAppProps {
  product?: ProductConfig;
  connected?: boolean;
  /** Package identity from bundle-sum (extension load). */
  version?: string;
  md5?: string;
}

/** Extension popup shell — package identity + connection-oriented UI. */
export function PopupApp({
  product = browserAgentProduct,
  connected = false,
  version = "—",
  md5 = "—",
}: PopupAppProps) {
  return (
    <div className="popup-app" data-product={product.id}>
      <h1>{product.displayName}</h1>
      <p>
        Control port <strong>{product.controlPort}</strong>
      </p>
      <p>
        version <code>{version || "—"}</code>
      </p>
      <p>
        md5 <code style={{ wordBreak: "break-all" }}>{md5 || "—"}</code>
      </p>
      <p data-connection-status>
        {connected ? "Connected" : "Not connected"}
      </p>
      <p className="muted">
        HTTP health is not the same as session WebSocket attach. Open a{" "}
        {product.cliName} /go?session= page so this extension can register.
      </p>
    </div>
  );
}

export default PopupApp;
