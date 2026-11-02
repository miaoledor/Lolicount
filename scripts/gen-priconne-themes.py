"""Export extracted PCR story models as static, independently selectable layers.

Requires Pillow, numpy, playwright and Chromium. Input model folders contain
dyn.json (Spine 3.6), dyn.atlas and atlas textures. No runtime is shipped to users.
"""
import argparse
import base64
import hashlib
import io
import json
import re
import threading
import time
from http.server import SimpleHTTPRequestHandler, ThreadingHTTPServer
from pathlib import Path

import numpy as np
from PIL import Image
from playwright.sync_api import sync_playwright


def read_json(path):
    return json.loads(path.read_text(encoding="utf-8"))


def write_json(path, value):
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(json.dumps(value, ensure_ascii=False, indent=2) + "\n", encoding="utf-8", newline="\n")


def decode(png):
    return Image.open(io.BytesIO(base64.b64decode(png))).convert("RGBA")


def export(result, dest, name, label):
    """Keep exact alpha, deduplicate complete pixels and verify reference blends."""
    size = (result["width"], result["height"])
    # Layer images are content-cropped, so placement plus size lives in ren.json
    # alone; recording it in config.json would go stale on re-export.
    manifest = [{"layer_id": 0}]
    ranges, seen, by_name = {}, {}, {}
    ren = dest / name / "ren"
    ren.mkdir(parents=True, exist_ok=True)
    for part in result["parts"]:
        image = Image.new("RGBA", size)
        image.paste(decode(part["png"]), (part.get("left", 0), part.get("top", 0)))
        box = image.getbbox()
        if box is None:
            box = (2, 2, 3, 3)
        if box[0] == 0 or box[1] == 0 or box[2] == size[0] or box[3] == size[1]:
            raise ValueError(f"{name}: attachment touches canvas edge")
        by_name[part["name"]] = image
        crop = image.crop(box)
        category = part["category"]
        # Layer images are content-cropped, so placement plus size lives in
        # ren.json alone and re-encoding an identical part changes nothing.
        key = (category, box, hashlib.sha256(crop.tobytes()).hexdigest())
        if key in seen:
            continue
        seen[key] = True
        layer_id = len(manifest)
        crop.save(ren / f"{layer_id}.webp", lossless=True, exact=True, method=4)
        manifest.append({"name": part["name"], "left": box[0], "top": box[1],
                         "visible": 1, "layer_id": layer_id, "group_layer_id": 0})
        if category not in ranges:
            ranges[category] = {"first": layer_id, "last": layer_id}
        ranges[category]["last"] = layer_id
    worst = 0
    for reference in result["references"]:
        composed = by_name[reference["body"]].copy()
        for category in ranges:
            for part_name in reference["names"]:
                if part_name.rsplit("_", 1)[-1].startswith(category):
                    composed.alpha_composite(by_name[part_name])
        expected = decode(reference["png"])
        # Compare on an opaque background so invisible RGB does not skew error.
        a = Image.new("RGBA", size, "white")
        b = a.copy()
        a.alpha_composite(composed)
        b.alpha_composite(expected)
        error = np.abs(np.asarray(a).astype(int) - np.asarray(b).astype(int))
        worst = max(worst, int(error.max()))
        if error.max() > 8 or error.mean() > 0.1:
            raise ValueError(f"{name}: reference mismatch {reference['skin']}: max={error.max()}, mean={error.mean():.5f}")
    if "body" not in ranges:
        raise ValueError(f"{name}: missing body")
    write_json(dest / name / "ren.json", manifest)
    write_json(dest / name / "config.json", {"canvasW": size[0], "canvasH": size[1], "ranges": ranges})
    write_json(dest / name / "display.json", {"size": 400})
    features = "independent body, eyes and mouth" if len(ranges) > 1 else "body only (no expression parts in source)"
    write_json(dest / name / "meta.json", {
        "name": name, "description": f"Princess Connect! Re:Dive / {label}; static story portrait, {features}.",
        "tags": ["princess-connect", "priconne", "公主连结", "プリコネ", label], "version": "1.0.0"})
    expected_images = {f"{index}.webp" for index in range(1, len(manifest))}
    for old_image in ren.glob("*.webp"):
        if old_image.stem.isdigit() and old_image.name not in expected_images:
            old_image.unlink()
    thumbnail = by_name.get("normal_body", by_name["default_body"]).copy()
    for category in (c for c in ranges if c != "body"):
        preferred = "normal_eyeOpen" if category == "eye" else "normal_mouthClose"
        candidate = by_name.get(preferred)
        if candidate is None:
            candidate = next(by_name[m["name"]] for m in manifest[1:] if m["name"].endswith("_eyeOpen" if category == "eye" else "_mouthClose"))
        thumbnail.alpha_composite(candidate)
    return {"name": name, "label": label, "layers": len(manifest) - 1,
            "counts": {k: v["last"] - v["first"] + 1 for k, v in ranges.items()},
            "reference_error": worst, "references": len(result["references"])}, thumbnail


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--models", type=Path, required=True)
    parser.add_argument("--runtime", type=Path, required=True)
    parser.add_argument("--manifest", type=Path, required=True)
    parser.add_argument("--names", type=Path, required=True)
    parser.add_argument("--exclude", type=Path, default=Path(__file__).with_name("priconne-exclusions.json"))
    parser.add_argument("--repo", type=Path, default=Path(__file__).resolve().parents[1])
    parser.add_argument("--only", nargs="*")
    parser.add_argument("--partition", type=int, nargs=2, metavar=("INDEX", "COUNT"))
    parser.add_argument("--resume", action="store_true")
    parser.add_argument("--max-edge", type=int, default=800)
    parser.add_argument("--wait-seconds", type=int, default=0)
    parser.add_argument("--report", type=Path, required=True)
    args = parser.parse_args()
    names = read_json(args.names)
    override_path = Path(__file__).with_name("priconne-source-overrides.json")
    body_only = read_json(override_path) if override_path.exists() else {}
    ids = sorted(set(re.findall(r"spine_full_(\d+)", args.manifest.read_text(encoding="utf-8"))))
    if args.exclude.exists():
        excluded = read_json(args.exclude)
        ids = [sid for sid in ids if sid not in excluded]
    if args.only:
        ids = [sid for sid in ids if sid in args.only]
    if args.partition:
        index, count = args.partition
        if not 0 <= index < count:
            parser.error("partition must satisfy 0 <= INDEX < COUNT")
        ids = ids[index::count]
    script = Path(__file__).with_name("priconne-render.js")

    class Handler(SimpleHTTPRequestHandler):
        def do_GET(self):
            if self.path == "/":
                content, mime = b'<canvas></canvas><script src="/runtime.js"></script><script src="/render.js"></script>', "text/html"
            elif self.path == "/runtime.js":
                content, mime = args.runtime.read_bytes(), "text/javascript"
            elif self.path == "/render.js":
                content, mime = script.read_bytes(), "text/javascript"
            elif self.path.startswith("/models/"):
                target = (args.models / self.path[len("/models/"):]).resolve()
                if not target.is_relative_to(args.models.resolve()):
                    self.send_error(403)
                    return
                content, mime = target.read_bytes(), self.guess_type(str(target))
                if target.suffix == ".atlas":
                    content = content.replace(b"\r", b"")
                    if target.parent.name.removeprefix("pcr-") in body_only:
                        content = content.strip().split(b"\n\n", 1)[0] + b"\n"
                if target.suffix == ".json" and target.parent.name.removeprefix("pcr-") in body_only:
                    data = json.loads(content)
                    for skin in data["skins"].values():
                        for slot in list(skin):
                            if slot.startswith(("eye", "mouth")):
                                del skin[slot]
                    for slot in data["slots"]:
                        if slot["name"].startswith(("eye", "mouth")):
                            slot.pop("attachment", None)
                    content = json.dumps(data).encode("utf-8")
                if target.suffix == ".png":
                    # Unity story atlases store premultiplied RGB. Normalize before
                    # filtering so exported PNG/WebP layers use straight alpha.
                    pixels = np.array(Image.open(io.BytesIO(content)).convert("RGBA"))
                    alpha = pixels[..., 3:4].astype(float)
                    pixels[..., :3] = np.clip(np.round(pixels[..., :3] * 255.0 / np.maximum(alpha, 1)), 0, 255).astype("uint8")
                    pixels[alpha[..., 0] == 0] = 0
                    output = io.BytesIO()
                    Image.fromarray(pixels).save(output, "PNG", compress_level=1)
                    content = output.getvalue()
            else:
                self.send_error(404)
                return
            self.send_response(200)
            self.send_header("Content-Type", mime)
            self.end_headers()
            self.wfile.write(content)

        def log_message(self, *args):
            pass

    server = ThreadingHTTPServer(("127.0.0.1", 0), Handler)
    threading.Thread(target=server.serve_forever, daemon=True).start()
    report = read_json(args.report) if args.resume and args.report.exists() else []
    complete = {entry["name"] for entry in report}
    ids = [sid for sid in ids if f"pcr-{sid}" not in complete]
    failures = []
    write_json(args.report.with_suffix(".failures.json"), failures)
    try:
        with sync_playwright() as p:
            browser = p.chromium.launch(headless=True)
            page = browser.new_page()
            for index, sid in enumerate(ids):
                name = f"pcr-{sid}"
                try:
                    deadline = time.monotonic() + args.wait_seconds
                    while not (args.models / name / "dyn.json").exists() and time.monotonic() < deadline:
                        time.sleep(0.5)
                    page.goto(f"http://127.0.0.1:{server.server_port}/")
                    result = page.evaluate("args => window.renderModel(args.model, args.maxEdge)", {"model": name, "maxEdge": args.max_edge})
                    label = names.get(sid[:4], f"Story {sid[:4]}")
                    info, thumbnail = export(result, args.repo / "assets/theme", name, label)
                    thumb_dir = args.repo / "web/public/images/theme-thumbs"
                    thumb_dir.mkdir(parents=True, exist_ok=True)
                    thumbnail.thumbnail((512, 256), Image.Resampling.LANCZOS)
                    thumbnail.save(thumb_dir / f"{name}.webp", quality=85, method=4)
                    report.append(info)
                    write_json(args.report, report)
                    print(f"{index + 1}/{len(ids)} {name}: {info['layers']} layers, reference error {info['reference_error']}", flush=True)
                except Exception as error:
                    failures.append({"name": name, "error": str(error)})
                    write_json(args.report.with_suffix(".failures.json"), failures)
                    print(f"FAIL {name}: {error}", flush=True)
            browser.close()
    finally:
        server.shutdown()
    if failures:
        raise SystemExit(f"{len(failures)} models failed; see failure report")


if __name__ == "__main__":
    main()
