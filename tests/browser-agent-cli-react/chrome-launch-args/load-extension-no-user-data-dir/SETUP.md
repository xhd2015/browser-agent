# SETUP

**Feature**: default-profile Chrome args — new window + URL, no load-extension / user-data-dir (F1)

**Flow**:
BuildChromeArgs(url, extractedPath)
  --new-window + session URL
  no --load-extension (ignored param; Load unpacked is operator Chrome)
  no --user-data-dir

**Leaf**: chrome-launch-args/load-extension-no-user-data-dir (name kept for history).
