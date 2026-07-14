import type { ProductConfig } from "./types";

/** browser-agent product — control port 43761. */
export const browserAgentProduct: ProductConfig = {
  id: "browser-agent",
  displayName: "Browser Agent",
  cliName: "browser-agent",
  controlPort: 43761,
  features: ["browser-agent"],
  pageMarkerGlobal: "__BROWSER_AGENT_EXT__",
  extensionDirName: "Chrome-Ext-Browser-Agent",
};

export default browserAgentProduct;
