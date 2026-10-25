#!/usr/bin/env python3
"""Capture deterministic full-character thumbnails from the Unity web player."""

from __future__ import annotations

import argparse
import io
import json
import urllib.parse
import urllib.request
from pathlib import Path

from PIL import Image
from playwright.sync_api import sync_playwright


def model_names(base_url: str) -> list[str]:
    """Return the maimai models exposed by the running server."""
    with urllib.request.urlopen(f"{base_url}/api/unity/models") as response:
        payload = json.load(response)
    return [name for name in payload["models"] if name.startswith("maimai-navichara-")]


def encode_thumbnail(png: bytes) -> bytes:
    """Downsample a browser capture into the gallery's transparent WebP format."""
    with Image.open(io.BytesIO(png)) as source:
        image = source.convert("RGBA").resize((256, 256), Image.Resampling.LANCZOS)
    output = io.BytesIO()
    image.save(output, "WEBP", quality=90, method=4)
    return output.getvalue()


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--base-url", default="http://127.0.0.1:9721")
    parser.add_argument("--output", type=Path, default=Path("web/public/images/emote-thumbs"))
    parser.add_argument("--check", action="store_true", help="fail if committed thumbnails differ")
    args = parser.parse_args()
    base_url = args.base_url.rstrip("/")
    names = model_names(base_url)
    args.output.mkdir(parents=True, exist_ok=True)
    mismatches: list[str] = []

    with sync_playwright() as playwright:
        browser = playwright.chromium.launch()
        page = browser.new_page(viewport={"width": 512, "height": 512})
        for name in names:
            query = urllib.parse.urlencode({
                "model": name,
                "motion": "Navi_Default",
                "name": "thumbnail",
                "thumbnail": "1",
                "time": "0",
                "player": "2",
            })
            page.goto(f"{base_url}/unity-player.html?{query}", wait_until="networkidle")
            page.wait_for_function(
                "document.documentElement.dataset.unityReady === 'true'"
            )
            encoded = encode_thumbnail(page.screenshot(omit_background=True))
            target = args.output / f"{name}.webp"
            if args.check:
                if not target.is_file() or target.read_bytes() != encoded:
                    mismatches.append(name)
            else:
                target.write_bytes(encoded)
                print(f"captured {name}")
        browser.close()

    if mismatches:
        print(f"FAILED: {len(mismatches)} stale Unity thumbnails: {', '.join(mismatches)}")
        return 1
    action = "match" if args.check else "captured"
    print(f"OK: {len(names)} Unity thumbnails {action} the rendered models")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
