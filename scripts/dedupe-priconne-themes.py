"""Ship each Princess Connect portrait once, however many costume entries repeat it.

A character's real-world portrait (`31`) is drawn once and reused by every costume
unit, the story happens inside a VR game, and generic story NPCs are reused across
batches. Those repeats are not always byte-identical: the same art is re-exported on
a different canvas, or it reappears as one of the mutually exclusive sets inside a
merged theme. So the comparison renders every portrait on a transparent canvas,
aligns it by its opaque bounding box and scales it to a common height, then groups
the portraits whose alpha mask and luma match within a strict tolerance.

One portrait per group survives: a named theme beats an unnamed one, a character
name without a costume beats a costume entry, the smallest model id breaks the rest.
The others are dropped - a whole theme when that portrait was all it shipped, or a
single set when the theme keeps other portraits - and every dropped title stays
searchable through the aliases of the theme that keeps the artwork. The plan lands in
`scripts/priconne-deduped-themes.json`, and dropped model ids are recorded in
`scripts/priconne-excluded-themes.json`, which the asset test enforces.
"""

import argparse
import json
import pathlib
import shutil
from collections import defaultdict

from PIL import Image


HEIGHT = 192
MASK_IOU = 0.995
LUMA_TOLERANCE = 1.5


def read_json(path):
    return json.loads(path.read_text(encoding="utf-8"))


def write_json(path, value):
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(json.dumps(value, ensure_ascii=False, indent=2) + "\n", encoding="utf-8", newline="\n")


def read_sets(directory):
    """Return the theme's selectable portraits as (model, canvas, ranges) triples."""
    config = read_json(directory / "config.json")
    if config.get("sets"):
        return [
            (configured["name"], (configured["canvasW"], configured["canvasH"]), configured["ranges"])
            for configured in config["sets"]
        ]
    return [(directory.name.removeprefix("pcr-"), (config["canvasW"], config["canvasH"]), config["ranges"])]


def render_portrait(directory, canvas, ranges, manifest):
    image = Image.new("RGBA", canvas, (0, 0, 0, 0))
    for category in ("body", "eye", "mouth"):
        if category not in ranges:
            continue
        entry = manifest[ranges[category]["first"]]
        layer = Image.open(directory / "ren" / f"{entry['layer_id']}.webp").convert("RGBA")
        image.alpha_composite(layer, (entry.get("left", 0), entry.get("top", 0)))
    box = image.split()[3].getbbox()
    if box is None:
        return None
    cropped = image.crop(box)
    scale = HEIGHT / cropped.height
    return cropped.resize((max(1, round(cropped.width * scale)), HEIGHT), Image.Resampling.LANCZOS)


def signature(image):
    alpha = image.split()[3].resize((32, 64), Image.Resampling.BILINEAR)
    flat = Image.new("RGB", image.size, (255, 255, 255))
    flat.paste(image, mask=image.split()[3])
    return alpha, flat.convert("L").resize((32, 64), Image.Resampling.BILINEAR)


def look_alike(first, second):
    alpha_a, luma_a = first
    alpha_b, luma_b = second
    pixels_a, pixels_b = list(alpha_a.getdata()), list(alpha_b.getdata())
    intersection = union = 0
    for a, b in zip(pixels_a, pixels_b):
        if a > 128 or b > 128:
            union += 1
        if a > 128 and b > 128:
            intersection += 1
    if union == 0:
        return 0.0, None
    iou = intersection / union
    if iou < MASK_IOU:
        return iou, None
    total = count = 0
    for a, b, alpha in zip(luma_a.getdata(), luma_b.getdata(), pixels_a):
        if alpha > 128:
            total += abs(a - b)
            count += 1
    luma = total / count if count else None
    if luma is None:
        return iou, None
    return iou, (luma if luma <= LUMA_TOLERANCE else None)


def expression_count(ranges):
    value = 1
    for category in ("eye", "mouth"):
        span = ranges.get(category)
        if span:
            value *= span["last"] - span["first"] + 1
    return value


def survivor_rank(theme, model, ranges, names):
    """Rank one portrait for keeping the artwork: name first, then the richer set."""
    stem = theme.removeprefix("pcr-")[:4]
    character = names.get(stem, f"Story {stem}")
    unnamed = character.startswith("Story ")
    costume = "（" in character or "(" in character
    return (unnamed, costume, -expression_count(ranges), model)


def character_label(model, names, overrides):
    stem = model[:4]
    override = overrides.get(model)
    if override == "":
        return f"Story {stem}"
    if override:
        return override
    return names.get(stem) or f"Story {stem}"


def refresh_meta(directory, theme_name, models, names, overrides):
    """Keep the metadata in step with the portraits a theme still holds."""
    meta_path = directory / "meta.json"
    if not meta_path.exists():
        return
    meta = read_json(meta_path)
    label = character_label(models[0], names, overrides)
    meta["name"] = theme_name
    if len(models) > 1:
        meta["description"] = (
            f"Princess Connect! Re:Dive / {label}; {len(models)} mutually exclusive story "
            "portraits, with expressions selected inside the chosen portrait."
        )
    else:
        meta["description"] = (
            f"Princess Connect! Re:Dive / {label}; static story portrait, "
            "independent body, eyes and mouth."
        )
    meta["tags"] = ["princess-connect", "priconne", "公主连结", "プリコネ", label]
    meta["version"] = "1.1.0" if len(models) > 1 else meta.get("version", "1.0.0")
    write_json(meta_path, meta)


def compact(directory, keep_models):
    """Rewrite a merged theme so it only holds the given portraits."""
    config = read_json(directory / "config.json")
    manifest = read_json(directory / "ren.json")
    kept = [configured for configured in config["sets"] if configured["name"] in keep_models]
    if not kept or len(kept) == len(config["sets"]):
        return False
    rebuilt = [{"layer_id": 0}]
    sets = []
    temporary = directory / "ren-compacted"
    temporary.mkdir()
    for configured in kept:
        ranges = {}
        for category in sorted(configured["ranges"], key=lambda key: configured["ranges"][key]["first"]):
            span = configured["ranges"][category]
            first = len(rebuilt)
            for index in range(span["first"], span["last"] + 1):
                layer = dict(manifest[index])
                old_id = layer["layer_id"]
                layer["layer_id"] = len(rebuilt)
                shutil.copy2(directory / "ren" / f"{old_id}.webp", temporary / f"{layer['layer_id']}.webp")
                rebuilt.append(layer)
            ranges[category] = {"first": first, "last": len(rebuilt) - 1}
        sets.append({**configured, "ranges": ranges})
    shutil.rmtree(directory / "ren")
    temporary.rename(directory / "ren")
    write_json(directory / "ren.json", rebuilt)
    if len(sets) == 1:
        single = sets[0]
        write_json(directory / "config.json", {
            "canvasW": single["canvasW"], "canvasH": single["canvasH"], "ranges": single["ranges"],
        })
    else:
        write_json(directory / "config.json", {"sets": sets})
    return True


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--repo", type=pathlib.Path, default=pathlib.Path(__file__).resolve().parents[1])
    parser.add_argument("--dry-run", action="store_true", help="report the duplicates without changing anything")
    parser.add_argument("--report", action="store_true", help="print the match quality of every linked pair")
    args = parser.parse_args()
    report = [] if args.report else None
    repo = args.repo.resolve()
    theme_root = repo / "assets/theme"
    thumb_root = repo / "web/public/images/theme-thumbs"
    names = read_json(repo / "scripts/priconne-names.json")
    overrides = read_json(repo / "scripts/priconne-model-names.json")
    merged_path = repo / "scripts/priconne-merged-themes.json"
    plan = read_json(merged_path)

    portraits = {}
    buckets = defaultdict(list)
    ranges_by_node = {}
    models_by_theme = {}
    for directory in sorted(theme_root.iterdir()):
        if not directory.is_dir() or len(directory.name) != 10:
            continue
        manifest = read_json(directory / "ren.json")
        sets = read_sets(directory)
        models_by_theme[directory.name] = [model for model, _, _ in sets]
        for model, canvas, ranges in sets:
            image = render_portrait(directory, canvas, ranges, manifest)
            if image is None:
                continue
            node = (directory.name, model)
            portraits[node] = signature(image)
            ranges_by_node[node] = ranges
            buckets[round(image.width / 3)].append(node)

    parent = {node: node for node in portraits}

    def find(node):
        while parent[node] != node:
            parent[node] = parent[parent[node]]
            node = parent[node]
        return node

    for key, nodes in buckets.items():
        for other_key in (key, key + 1):
            for index, first in enumerate(buckets.get(key, [])):
                for second in buckets.get(other_key, []):
                    if second <= first:
                        continue
                    iou, luma = look_alike(portraits[first], portraits[second])
                    if luma is not None and report is not None:
                        report.append((round(iou, 4), round(luma, 3), first, second))
                    if luma is not None:
                        root_a, root_b = find(first), find(second)
                        if root_a != root_b:
                            parent[root_b] = root_a

    groups = defaultdict(list)
    for node in portraits:
        groups[find(node)].append(node)
    duplicates = [sorted(members) for members in groups.values() if len(members) > 1]

    dropped_themes = {}
    dropped_sets = defaultdict(list)
    alias_fold = {}
    for group in duplicates:
        ranked = sorted(
            group,
            key=lambda node: (survivor_rank(node[0], node[1], ranges_by_node[node], names), node[0], node[1]),
        )
        survivor = ranked[0]
        for theme, model in ranked[1:]:
            alias_fold.setdefault(survivor[0], []).append(model)
            if len(models_by_theme[theme]) == 1:
                dropped_themes.setdefault(survivor[0], []).append(model)
            else:
                dropped_sets[theme].append(model)

    print(f"portraits: {len(portraits)}")
    print(f"look-alike groups: {len(duplicates)} covering {sum(len(g) for g in duplicates)} portraits")
    print(f"themes dropped entirely: {sum(len(v) for v in dropped_themes.values())}")
    print(f"survivors that absorb a dropped theme: {len(dropped_themes)}")
    print(f"merged themes losing a duplicate set: {len(dropped_sets)} ({sum(len(v) for v in dropped_sets.values())} sets)")
    if args.report:
        for iou, luma, first, second in sorted(report, key=lambda item: -item[1])[:20]:
            print(f"  loose match iou={iou} luma={luma} {first} ~ {second}")
        print(f"  linked pairs: {len(report)}")
    if args.dry_run:
        for survivor, dropped in sorted(dropped_themes.items())[:12]:
            print(f"  keep {survivor} <- {dropped}")
        for theme, dropped in sorted(dropped_sets.items())[:12]:
            print(f"  {theme}: drop sets {dropped}")
        return

    for survivor, dropped in dropped_themes.items():
        for model in dropped:
            shutil.rmtree(theme_root / f"pcr-{model}")
            thumb = thumb_root / f"pcr-{model}.webp"
            if thumb.exists():
                thumb.unlink()
    renames = {}
    emptied = []
    for theme, dropped in dropped_sets.items():
        keep = [model for model in models_by_theme[theme] if model not in dropped]
        if not keep:
            # Every portrait of this theme repeats one kept elsewhere.
            shutil.rmtree(theme_root / theme)
            thumb = thumb_root / f"{theme}.webp"
            if thumb.exists():
                thumb.unlink()
            emptied.append(theme)
            plan.pop(theme, None)
            continue
        compact(theme_root / theme, set(keep))
        new_name = f"pcr-{keep[0]}"
        if new_name != theme:
            # The surviving portrait is not the one that named the directory, so the
            # theme moves to the id it now holds.
            target = theme_root / new_name
            if target.exists():
                raise SystemExit(f"cannot move {theme} to {new_name}: target exists")
            shutil.move(str(theme_root / theme), str(target))
            old_thumb = thumb_root / f"{theme}.webp"
            new_thumb = thumb_root / f"{new_name}.webp"
            if old_thumb.exists():
                shutil.move(str(old_thumb), str(new_thumb))
            renames[theme] = new_name
        plan.pop(theme, None)
        refresh_meta(theme_root / new_name, new_name, keep, names, overrides)
        if len(keep) > 1:
            plan[new_name] = keep

    for theme, members in list(plan.items()):
        if len(members) < 2:
            del plan[theme]
    write_json(merged_path, dict(sorted(plan.items())))

    deduped_path = repo / "scripts/priconne-deduped-themes.json"
    deduped = read_json(deduped_path) if deduped_path.exists() else {}
    for survivor, dropped in alias_fold.items():
        key = renames.get(survivor, survivor)
        deduped[key] = sorted(set(deduped.get(key, [])) | set(dropped))
    write_json(deduped_path, dict(sorted(deduped.items())))

    excluded_path = repo / "scripts/priconne-excluded-themes.json"
    excluded = set(read_json(excluded_path))
    excluded.update(f"pcr-{model}" for model in (*sum(dropped_themes.values(), []), *sum(dropped_sets.values(), [])))
    write_json(excluded_path, sorted(excluded))
    print(f"themes emptied by duplicate sets: {len(emptied)}")
    print(f"recorded {len(excluded)} excluded themes")


if __name__ == "__main__":
    main()
