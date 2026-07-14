#!/usr/bin/env python3
"""Summarize API-relevant entries from a Chrome-Ext-Capture-API HAR file."""

from __future__ import annotations

import argparse
import json
import re
import sys
from pathlib import Path
from typing import Any
from urllib.parse import urlparse

NOISE_HOST_SUFFIXES = (
    "dem.some-x.com",
    "google-analytics.com",
    "googletagmanager.com",
)
NOISE_PATH_PATTERNS = (
    re.compile(r"/tags/web-(performance|custom)/"),
    re.compile(r"/dem/entrance/"),
    re.compile(r"\.(js|css|png|jpg|jpeg|gif|webp|svg|ico|woff2?|ttf|map)(\?|$)", re.I),
)


def is_noise(url: str) -> bool:
    host = urlparse(url).hostname or ""
    path = urlparse(url).path
    if any(host.endswith(s) for s in NOISE_HOST_SUFFIXES):
        return True
    return any(p.search(path) for p in NOISE_PATH_PATTERNS)


def path_only(url: str) -> str:
    p = urlparse(url)
    return p.path + (f"?{p.query}" if p.query else "")


def truncate(text: str, max_len: int) -> str:
    text = text.strip()
    if not text:
        return ""
    if len(text) <= max_len:
        return text
    return text[:max_len] + "..."


def request_body(entry: dict) -> str:
    return ((entry.get("request") or {}).get("postData") or {}).get("text") or ""


def response_body(entry: dict) -> tuple[str, str | None]:
    content = (entry.get("response") or {}).get("content") or {}
    text = content.get("text") or ""
    encoding = content.get("encoding")
    return text, encoding


def summarize_json_body(text: str) -> dict[str, Any] | None:
    if not text:
        return None
    try:
        data = json.loads(text)
    except json.JSONDecodeError:
        return None
    if not isinstance(data, dict):
        return {"type": type(data).__name__}
    summary: dict[str, Any] = {}
    for key in ("code", "msg", "message", "reqId"):
        if key in data:
            summary[key] = data[key]
    result = data.get("result")
    if isinstance(result, dict):
        result_summary: dict[str, Any] = {}
        for key in ("id", "jiraKey", "jiraId", "status", "parentId", "parentJiraKey"):
            if key in result:
                result_summary[key] = result[key]
        if "data" in result and isinstance(result["data"], list):
            result_summary["data_count"] = len(result["data"])
            if result["data"]:
                first = result["data"][0]
                if isinstance(first, dict):
                    result_summary["first_item"] = {
                        key: first.get(key)
                        for key in ("id", "jiraKey", "status", "summary")
                        if key in first
                    }
        if result_summary:
            summary["result"] = result_summary
    elif result is not None:
        summary["result_type"] = type(result).__name__
    return summary or None


def body_coverage(rows: list[dict[str, Any]]) -> dict[str, int]:
    return {
        "entries": len(rows),
        "with_request_body": sum(1 for r in rows if r["request_body_len"] > 0),
        "with_response_body": sum(1 for r in rows if r["response_body_len"] > 0),
    }


def build_row(entry: dict, index: int, max_len: int) -> dict[str, Any]:
    req = entry.get("request") or {}
    resp = entry.get("response") or {}
    req_text = request_body(entry)
    res_text, res_encoding = response_body(entry)
    req_summary = summarize_json_body(req_text)
    res_summary = summarize_json_body(res_text) if not res_encoding else None

    return {
        "index": index,
        "started": entry.get("startedDateTime", ""),
        "method": req.get("method", "?"),
        "url": req.get("url", ""),
        "path": path_only(req.get("url", "")),
        "status": resp.get("status", 0),
        "time_ms": entry.get("time"),
        "request_body_len": len(req_text),
        "request_body_preview": truncate(req_text, max_len),
        "request_body_summary": req_summary,
        "response_body_len": len(res_text),
        "response_body_encoding": res_encoding,
        "response_body_preview": truncate(res_text, max_len) if not res_encoding else "<base64>",
        "response_body_summary": res_summary,
    }


def creator_info(har: dict) -> dict[str, str]:
    creator = har.get("log", {}).get("creator") or {}
    return {
        "name": creator.get("name", "unknown"),
        "version": creator.get("version", "unknown"),
    }


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("har", type=Path, help="Path to .har file")
    parser.add_argument(
        "--host",
        help="Only include URLs whose host contains this substring (e.g. app.example.com)",
    )
    parser.add_argument(
        "--json",
        action="store_true",
        help="Emit machine-readable JSON instead of text",
    )
    parser.add_argument(
        "--max-len",
        type=int,
        default=200,
        help="Max characters for body preview fields (default: 200)",
    )
    args = parser.parse_args()

    with args.har.open(encoding="utf-8") as f:
        har = json.load(f)

    creator = creator_info(har)
    entries = har.get("log", {}).get("entries", [])

    rows = []
    for i, e in enumerate(entries):
        req = e.get("request") or {}
        url = req.get("url", "")
        if not url or is_noise(url):
            continue
        if args.host and args.host not in url:
            continue
        rows.append(build_row(e, i, args.max_len))

    coverage = body_coverage(rows)
    payload = {
        "creator": creator,
        "body_coverage": coverage,
        "entries": rows,
    }

    if args.json:
        print(json.dumps(payload, indent=2))
        return 0

    print(f"HAR creator: {creator['name']} v{creator['version']}")
    print(f"API entries: {coverage['entries']} / {len(entries)} total")
    print(
        "Body coverage: "
        f"request {coverage['with_request_body']}/{coverage['entries']}, "
        f"response {coverage['with_response_body']}/{coverage['entries']}"
    )
    if (
        creator["name"] == "API Capture - HAR Recorder"
        and creator["version"] < "1.1.0"
        and coverage["with_response_body"] == 0
    ):
        print(
            "Warning: extension v1.0.0 does not capture response bodies. "
            "Rebuild with: go run ./script/chrome-ext-capture-api/build"
        )
    print()

    for r in rows:
        print(
            f"[{r['index']:>3}] {r['started']}  {r['method']:4}  {r['status']}  "
            f"{r['time_ms']}ms  {r['path']}"
        )
        if r["request_body_len"] > 0:
            print(f"       request: {r['request_body_preview']}")
            if r["request_body_summary"]:
                print(f"       request_json: {json.dumps(r['request_body_summary'], ensure_ascii=False)}")
        else:
            print("       request: <empty>")

        if r["response_body_len"] > 0:
            if r["response_body_encoding"]:
                print(
                    f"       response: <base64 {r['response_body_len']} chars, "
                    f"encoding={r['response_body_encoding']}>"
                )
            else:
                print(f"       response: {r['response_body_preview']}")
                if r["response_body_summary"]:
                    print(
                        f"       response_json: {json.dumps(r['response_body_summary'], ensure_ascii=False)}"
                    )
        else:
            size = ((entries[r["index"]].get("response") or {}).get("content") or {}).get("size")
            if size:
                print(f"       response: <empty text; content.size={size}B>")
            else:
                print("       response: <empty>")
    return 0


if __name__ == "__main__":
    sys.exit(main())