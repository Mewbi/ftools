#!/usr/bin/env python3
"""Build static/icons.svg, a sprite of the icons the templates use.

Usage (from the frontend directory): make icons

To add an icon, add its Lucide name to ICONS and run `make icons`.
Names: https://lucide.dev/icons
Lucide has no brand logos; those come from Simple Icons (BRAND_ICONS).
Names: https://simpleicons.org
"""

import os
import re
import sys
import urllib.request

LUCIDE_VERSION = "1.70.0"
SIMPLE_ICONS_VERSION = "13.21.0"

ICONS = [
    "arrow-up-right",
    "circle-alert",
    "circle-check",
    "download",
    "file",
    "file-text",
    "house",
    "image",
    "link",
    "loader-circle",
    "menu",
    "moon",
    "qr-code",
    "rotate-ccw",
    "shield-check",
    "shrink",
    "sun",
    "trash-2",
    "upload",
    "x",
]

BRAND_ICONS = [
    "github",
    "linkedin",
]

URL = "https://cdn.jsdelivr.net/npm/lucide-static@{v}/icons/{name}.svg"
BRAND_URL = "https://cdn.jsdelivr.net/npm/simple-icons@{v}/icons/{name}.svg"
OUT = os.path.join(os.path.dirname(__file__), "..", "static", "icons.svg")


def fetch_body(url: str) -> str:
    with urllib.request.urlopen(url) as r:
        svg = r.read().decode()
    svg = re.sub(r"<!--.*?-->", "", svg, flags=re.S)
    svg = re.sub(r"<title>.*?</title>", "", svg, flags=re.S)
    body = re.search(r"<svg[^>]*>(.*)</svg>", svg, re.S).group(1)
    return re.sub(r"\s*\n\s*", "", body.strip())


def main() -> None:
    symbols = []
    for name in sorted(ICONS):
        body = fetch_body(URL.format(v=LUCIDE_VERSION, name=name))
        symbols.append(
            f'<symbol id="{name}" viewBox="0 0 24 24" fill="none" stroke="currentColor" '
            f'stroke-width="2" stroke-linecap="round" stroke-linejoin="round">{body}</symbol>'
        )
    for name in sorted(BRAND_ICONS):
        body = fetch_body(BRAND_URL.format(v=SIMPLE_ICONS_VERSION, name=name))
        symbols.append(f'<symbol id="{name}" viewBox="0 0 24 24" fill="currentColor">{body}</symbol>')

    with open(OUT, "w") as f:
        f.write('<svg xmlns="http://www.w3.org/2000/svg">\n')
        f.write(
            f"<!-- Lucide icons (ISC license), lucide-static@{LUCIDE_VERSION}; brand icons from "
            f"Simple Icons (CC0), simple-icons@{SIMPLE_ICONS_VERSION}. Built by scripts/build-icons.py. -->\n"
        )
        f.write("\n".join(symbols))
        f.write("\n</svg>\n")
    print(f"wrote {len(symbols)} icons to {os.path.normpath(OUT)}", file=sys.stderr)


if __name__ == "__main__":
    main()
