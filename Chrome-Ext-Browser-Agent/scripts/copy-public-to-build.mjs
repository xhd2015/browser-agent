#!/usr/bin/env node
/**
 * Stage MV3 package: public/ → build/ (no bundler required for this shell).
 */
import { cpSync, existsSync, mkdirSync, rmSync } from "node:fs";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";

const root = join(dirname(fileURLToPath(import.meta.url)), "..");
const src = join(root, "public");
const dest = join(root, "build");

if (!existsSync(src)) {
  console.error("Chrome-Ext-Browser-Agent/public missing");
  process.exit(1);
}

rmSync(dest, { recursive: true, force: true });
mkdirSync(dest, { recursive: true });
cpSync(src, dest, { recursive: true });
console.log("Staged public/ → build/");
