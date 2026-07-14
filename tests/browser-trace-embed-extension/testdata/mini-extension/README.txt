Mini-extension fixture (implementer contract)
=============================================

This directory is the **canonical mini MV3** used so CI/doctests never need
webpack or npm. The product package should either:

  A) Stage this tree (or an equivalent) into browsertrace/embedded/extension/
     under //go:embed for development and CI, OR
  B) Accept an injectable fs.FS of the same shape for tests.

Production release path: ./script/browser-trace/bundle builds
Chrome-Ext-Capture-API and stages the real unpacked extension into
browsertrace/embedded/extension/ (overwriting the mini fixture).

Required files for extract tests:
  - manifest.json  (must include "version" string)
  - at least one other file referenced by the manifest (e.g. background.js)

Extract layout at runtime (product):
  {BaseDir}/extension/{version}/manifest.json
  {BaseDir}/extension/{version}/...

Doctests assert path layout + readable version; they do not execute the
extension in Chrome.
