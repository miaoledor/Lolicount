"""Consolidate Princess Connect portraits that show the same skin of one character.

Every retained story model ships as its own theme, because the leading four digits
of a model ID are a story batch rather than a character. Portraits that carry the
same character name *and* the same tens digit of the story suffix are the same
character in the same skin line, so they are consolidated into a single theme:
`config.json` keeps one mutually exclusive `sets` entry per portrait, and rendering
therefore picks a portrait first and an expression inside it. Variants add up
instead of multiplying, and a card can never shuffle between different characters,
because only same-name same-tens portraits merge.

Story-only batches keep one theme per portrait: their four-digit ids are story-local
NPC slots, and one tens digit there regularly holds unrelated characters (a fox
girl, a wizard and a raccoon boy share `Story 0056` or `Story 0004`).

Portraits the reviewer flagged as unusable are deleted outright, and a portrait can
be kept out of merging through the exceptions file. Both lists are data in
`scripts/priconne-merge-exceptions.json`, so the checked-in theme tree, the catalog
and the asset test all derive from the same declaration.
"""

import argparse
import json
import shutil
import tempfile
import time
from pathlib import Path


def read_json(path):
    return json.loads(path.read_text(encoding="utf-8"))


def write_json(path, value):
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(json.dumps(value, ensure_ascii=False, indent=2) + "\n", encoding="utf-8", newline="\n")


def group_name(model, names, overrides):
    """Return the character label that decides which portraits belong together."""
    stem = model[:4]
    override = overrides.get(model)
    if override == "":
        return f"Story {stem}"
    if override:
        return override
    return names.get(stem) or f"Story {stem}"


def group_identity(model, names, chinese, overrides, romaji):
    """Return the identity two portraits must share to be consolidated.

    The same character's nameplate appears in either language depending on the story
    file (``ラジラジ`` and ``拉基拉基``), so a portrait whose nameplate is the stem's
    Japanese or Chinese name groups under the stem's romanized name. A portrait whose
    nameplate is something else names another character inside the same batch and
    keeps its own identity, which is what stops a mixed batch from merging.
    """
    stem = model[:4]
    stem_name = names.get(stem) or f"Story {stem}"
    override = overrides.get(model)
    if override == "" or stem_name.startswith("Story "):
        return f"Story {stem}"
    resolved = override or stem_name
    if resolved in {stem_name, chinese.get(stem)}:
        return romaji.get(stem) or stem_name
    return resolved


def replace_file(source, target):
    """Copy one thumbnail over another, tolerating a transient Windows lock."""
    if source == target:
        return
    for attempt in range(5):
        try:
            shutil.copy2(source, target)
            return
        except PermissionError:
            if attempt == 4:
                raise
            time.sleep(0.5)


def merge_group(repo, stage, destination_name, source_names, label):
    theme_root = repo / "assets/theme"
    destination = stage / destination_name
    ren_dir = destination / "ren"
    ren_dir.mkdir(parents=True)
    manifest = [{"layer_id": 0}]
    sets = []
    for source_name in sorted(source_names):
        source = theme_root / source_name
        source_manifest = read_json(source / "ren.json")
        source_config = read_json(source / "config.json")
        offset = len(manifest) - 1
        for entry in source_manifest[1:]:
            copied = dict(entry)
            copied["layer_id"] = len(manifest)
            shutil.copy2(source / "ren" / f"{entry['layer_id']}.webp", ren_dir / f"{copied['layer_id']}.webp")
            manifest.append(copied)
        sets.append({
            "name": source_name.removeprefix("pcr-"),
            "canvasW": source_config["canvasW"],
            "canvasH": source_config["canvasH"],
            "ranges": {
                category: {"first": span["first"] + offset, "last": span["last"] + offset}
                for category, span in source_config["ranges"].items()
            },
        })
    write_json(destination / "ren.json", manifest)
    write_json(destination / "config.json", {"sets": sets})
    write_json(destination / "display.json", {"size": 400})
    write_json(destination / "meta.json", {
        "name": destination_name,
        "description": (
            f"Princess Connect! Re:Dive / {label}; {len(sets)} mutually exclusive story "
            "portraits, with expressions selected inside the chosen portrait."
        ),
        "tags": ["princess-connect", "priconne", "公主连结", "プリコネ", label],
        "version": "1.1.0",
    })
    return destination_name


def build_plan(repo, theme_names, names, chinese, overrides, romaji, deleted, standalone, standalone_batches):
    """Return the merge plan: representative theme name -> sorted member model ids.

    Themes that carry a real character name are grouped, and so are story-only
    batches whose portraits were each checked to show one character. Story-only
    batches reuse the same slot index for unrelated NPCs -- one tens digit there
    can hold a fox girl, a wizard and a raccoon boy -- so the batches where a
    visual audit found that mix stay out of merging entirely.
    """
    groups = {}
    for theme_name in theme_names:
        model = theme_name.removeprefix("pcr-")
        if theme_name in deleted or theme_name in standalone:
            continue
        identity = group_identity(model, names, chinese, overrides, romaji)
        if identity.startswith("Story ") and model[:4] in standalone_batches:
            continue
        key = (identity, model[4])
        groups.setdefault(key, []).append(model)
    plan = {}
    for key, members in groups.items():
        members = sorted(members)
        if len(members) < 2:
            continue
        plan[f"pcr-{members[0]}"] = members
    return plan


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--repo", type=Path, default=Path(__file__).resolve().parents[1])
    parser.add_argument("--dry-run", action="store_true", help="print the plan without touching the theme tree")
    args = parser.parse_args()
    repo = args.repo.resolve()
    theme_root = repo / "assets/theme"
    thumb_root = repo / "web/public/images/theme-thumbs"

    names = read_json(repo / "scripts/priconne-names.json")
    chinese = read_json(repo / "scripts/priconne-zh-names.json")
    romaji = read_json(repo / "scripts/priconne-romaji.json")
    overrides = read_json(repo / "scripts/priconne-model-names.json")
    exceptions = read_json(repo / "scripts/priconne-merge-exceptions.json")
    deleted = {f"pcr-{model}" for model in exceptions["deleted"]}
    standalone = {f"pcr-{model}" for model in exceptions["standalone"]}
    standalone_batches = set(exceptions.get("standalone_batches", []))
    dusk_tens = set(exceptions.get("dusk_tens", []))

    theme_names = sorted(
        directory.name
        for directory in theme_root.iterdir()
        if directory.is_dir() and directory.name.startswith("pcr-") and len(directory.name) == 10
    )
    # A suffix tens digit listed in ``dusk_tens`` marks the evening rendering of the
    # same portrait: same pose and outfit, darker and often without the weapon, so
    # it ships no theme of its own.
    dusk = {name for name in theme_names if name.removeprefix("pcr-")[4] in dusk_tens}
    deleted |= dusk
    missing = sorted(name for name in deleted if name not in theme_names)
    if missing:
        raise SystemExit(f"deletion list names themes that are not shipped: {missing}")
    plan = build_plan(repo, theme_names, names, chinese, overrides, romaji, deleted, standalone, standalone_batches)
    members = {member for group in plan.values() for member in group}
    print(f"shipped themes: {len(theme_names)}")
    print(f"deleted by request: {len(deleted) - len(dusk)} {sorted(deleted - dusk)}")
    print(f"deleted as dusk ({sorted(dusk_tens)}): {len(dusk)}")
    print(f"kept out of merging: {sorted(standalone)}")
    print(f"merge groups: {len(plan)} covering {len(members)} portraits")
    print(f"resulting themes: {len(theme_names) - len(deleted) - len(members) + len(plan)}")
    if args.dry_run:
        for destination, group in sorted(plan.items(), key=lambda item: -len(item[1]))[:15]:
            print(f"  {destination} <- {group}")
        return

    labels = {
        destination: group_name(group[0], names, overrides) for destination, group in plan.items()
    }
    with tempfile.TemporaryDirectory(prefix="priconne-merge-", dir=repo / "assets") as temporary:
        stage = Path(temporary)
        for destination, group in sorted(plan.items()):
            merge_group(repo, stage, destination, [f"pcr-{model}" for model in group], labels[destination])
        for destination in sorted(plan):
            target = theme_root / destination
            if target.exists():
                shutil.rmtree(target)
            shutil.move(stage / destination, target)
            source_thumb = thumb_root / f"pcr-{plan[destination][0]}.webp"
            if not source_thumb.exists():
                raise SystemExit(f"missing thumbnail for merged representative {source_thumb.name}")
            replace_file(source_thumb, thumb_root / f"{destination}.webp")

    representatives = {name.removeprefix("pcr-") for name in plan}
    removed = sorted((members - representatives) | {name.removeprefix("pcr-") for name in deleted})
    for model in removed:
        shutil.rmtree(theme_root / f"pcr-{model}")
        thumb = thumb_root / f"pcr-{model}.webp"
        if thumb.exists():
            thumb.unlink()

    write_json(repo / "scripts/priconne-merged-themes.json", plan)

    excluded_path = repo / "scripts/priconne-excluded-themes.json"
    excluded = set(read_json(excluded_path)) if excluded_path.exists() else set()
    excluded.update(deleted)
    write_json(excluded_path, sorted(excluded))
    print(f"merged {len(members)} portraits into {len(plan)} themes")
    print(f"deleted {len(removed) - (len(members) - len(plan))} flagged themes; "
          f"{len(excluded)} themes recorded as excluded")


if __name__ == "__main__":
    main()
