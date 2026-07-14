import type { ProductConfig } from "./types";

/** browser-trace product — control port 43759. */
export const browserTraceProduct: ProductConfig = {
  id: "browser-trace",
  displayName: "Browser Trace",
  cliName: "browser-trace",
  controlPort: 43759,
  features: ["browser-trace"],
  pageMarkerGlobal: "__BROWSER_TRACE_EXT__",
  extensionDirName: "Chrome-Ext-Capture-API",
};

export default browserTraceProduct;
