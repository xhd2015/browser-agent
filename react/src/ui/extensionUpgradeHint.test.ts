import assert from "node:assert/strict";
import { describe, it } from "node:test";
import {
  BROWSER_AGENT_INSTALL_SH,
  compareVersion,
  formatUpgradeHintHeadline,
  resolveExtensionUpgradeHint,
} from "./extensionUpgradeHint.ts";

describe("compareVersion", () => {
  it("orders major.minor.patch", () => {
    assert.equal(compareVersion("1.0.12", "1.0.14"), -1);
    assert.equal(compareVersion("1.0.14", "1.0.12"), 1);
    assert.equal(compareVersion("1.0.14", "1.0.14"), 0);
  });
});

describe("resolveExtensionUpgradeHint", () => {
  it("returns null for ok and not_connected", () => {
    assert.equal(
      resolveExtensionUpgradeHint({
        match: "ok",
        bundledVersion: "1.0.14",
        loadedVersion: "1.0.14",
      }),
      null,
    );
    assert.equal(
      resolveExtensionUpgradeHint({
        match: "not_connected",
        bundledVersion: "1.0.14",
      }),
      null,
    );
  });

  it("extension upgradeable when loaded < bundled", () => {
    const hint = resolveExtensionUpgradeHint({
      match: "version_mismatch",
      bundledVersion: "1.0.14",
      loadedVersion: "1.0.12",
      browser: "chrome",
    });
    assert.ok(hint);
    assert.equal(hint.kind, "extension");
    assert.equal(hint.from, "1.0.12");
    assert.equal(hint.to, "1.0.14");
    assert.equal(hint.command, "browser-agent install-chrome-extension");
    assert.equal(
      formatUpgradeHintHeadline(hint),
      "Extension upgradeable: 1.0.12 → 1.0.14",
    );
  });

  it("uses firefox install command when browser is firefox", () => {
    const hint = resolveExtensionUpgradeHint({
      match: "version_mismatch",
      bundledVersion: "1.0.14",
      loadedVersion: "1.0.12",
      browser: "firefox",
    });
    assert.ok(hint);
    assert.equal(hint.command, "browser-agent install-firefox-extension");
  });

  it("browser-agent upgradeable when loaded > bundled", () => {
    const hint = resolveExtensionUpgradeHint({
      match: "version_mismatch",
      bundledVersion: "1.0.12",
      loadedVersion: "1.0.14",
      browser: "chrome",
    });
    assert.ok(hint);
    assert.equal(hint.kind, "browser_agent");
    assert.equal(hint.from, "1.0.12");
    assert.equal(hint.to, "1.0.14");
    assert.equal(hint.command, BROWSER_AGENT_INSTALL_SH);
    assert.equal(
      formatUpgradeHintHeadline(hint),
      "browser-agent upgradeable: 1.0.12 → 1.0.14",
    );
  });

  it("rebuild callout for md5_mismatch", () => {
    const hint = resolveExtensionUpgradeHint({
      match: "md5_mismatch",
      bundledVersion: "1.0.14",
      loadedVersion: "1.0.14",
      browser: "chrome",
    });
    assert.ok(hint);
    assert.equal(hint.kind, "extension_rebuild");
    assert.equal(hint.command, "browser-agent install-chrome-extension");
    assert.equal(
      formatUpgradeHintHeadline(hint),
      "Extension rebuild needed: version 1.0.14 matches, md5 differs",
    );
  });
});
