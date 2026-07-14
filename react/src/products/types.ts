/** Shared product configuration for browser-agent / browser-trace SPAs. */
export type ProductId = "browser-agent" | "browser-trace";

export interface ProductConfig {
  id: ProductId;
  displayName: string;
  cliName: string;
  controlPort: number;
  features: string[];
  pageMarkerGlobal: string;
  extensionDirName: string;
}
