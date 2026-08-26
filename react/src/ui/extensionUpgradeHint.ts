/** Match statuses from GET /v1/session extension_match. */
export type ExtensionMatch =
  | "ok"
  | "not_connected"
  | "version_mismatch"
  | "md5_mismatch"
  | "md5_unknown"
  | string;

export type InstallBrowserKind = "chrome" | "firefox";

export type ExtensionUpgradeKind =
  | "extension"
  | "browser_agent"
  | "extension_rebuild";

export interface ExtensionUpgradeHint {
  kind: ExtensionUpgradeKind;
  /** Human title, e.g. "Extension upgradeable". */
  title: string;
  from: string;
  to: string;
  /** Shell command the user should run. */
  command: string;
}

export const BROWSER_AGENT_INSTALL_SH =
  "curl -fsSL https://raw.githubusercontent.com/xhd2015/browser-agent/master/install.sh | bash";

/** Loose semver compare aligned with browseragent.CompareVersion (-1/0/+1). */
export function compareVersion(a: string, b: string): number {
  const av = parseVersionTuple(a);
  const bv = parseVersionTuple(b);
  for (let i = 0; i < 3; i++) {
    if (av[i] < bv[i]) return -1;
    if (av[i] > bv[i]) return 1;
  }
  const apre = hasPrereleaseSuffix(a);
  const bpre = hasPrereleaseSuffix(b);
  if (apre && !bpre) return -1;
  if (!apre && bpre) return 1;
  return 0;
}

function hasPrereleaseSuffix(v: string): boolean {
  const s = (v || "").trim();
  return /[-+]/.test(s);
}

function parseVersionTuple(v: string): [number, number, number] {
  let s = (v || "").trim();
  const cut = s.search(/[-+]/);
  if (cut >= 0) s = s.slice(0, cut);
  const parts = s.split(".");
  const out: [number, number, number] = [0, 0, 0];
  for (let i = 0; i < 3; i++) {
    const n = parseInt(parts[i] || "0", 10);
    out[i] = Number.isFinite(n) ? n : 0;
  }
  return out;
}

function installExtensionCommand(browser: InstallBrowserKind): string {
  return browser === "firefox"
    ? "browser-agent install-firefox-extension"
    : "browser-agent install-chrome-extension";
}

/**
 * Derive an actionable upgrade callout from session snap identity fields.
 * Returns null when no upgrade guidance should be shown.
 */
export function resolveExtensionUpgradeHint(opts: {
  match: ExtensionMatch;
  bundledVersion?: string;
  loadedVersion?: string;
  browser?: InstallBrowserKind;
}): ExtensionUpgradeHint | null {
  const match = (opts.match || "").trim() || "not_connected";
  if (match === "ok" || match === "not_connected") {
    return null;
  }

  const bundled = (opts.bundledVersion || "").trim();
  const loaded = (opts.loadedVersion || "").trim();
  const browser: InstallBrowserKind =
    opts.browser === "firefox" ? "firefox" : "chrome";
  const extCmd = installExtensionCommand(browser);

  if (match === "md5_mismatch" || match === "md5_unknown") {
    const ver = bundled || loaded || "—";
    return {
      kind: "extension_rebuild",
      title: "Extension rebuild needed",
      from: ver,
      to: ver,
      command: extCmd,
    };
  }

  // version_mismatch (and any other warn status with both versions)
  if (!bundled || !loaded) {
    return null;
  }

  const cmp = compareVersion(loaded, bundled);
  if (cmp < 0) {
    return {
      kind: "extension",
      title: "Extension upgradeable",
      from: loaded,
      to: bundled,
      command: extCmd,
    };
  }
  if (cmp > 0) {
    return {
      kind: "browser_agent",
      title: "browser-agent upgradeable",
      from: bundled,
      to: loaded,
      command: BROWSER_AGENT_INSTALL_SH,
    };
  }

  // Same version string but match still warns (unexpected) → rebuild.
  return {
    kind: "extension_rebuild",
    title: "Extension rebuild needed",
    from: loaded,
    to: bundled,
    command: extCmd,
  };
}

/** One-line summary: "Extension upgradeable: 1.0.12 → 1.0.14". */
export function formatUpgradeHintHeadline(hint: ExtensionUpgradeHint): string {
  if (hint.kind === "extension_rebuild") {
    return `${hint.title}: version ${hint.from} matches, md5 differs`;
  }
  return `${hint.title}: ${hint.from} → ${hint.to}`;
}
