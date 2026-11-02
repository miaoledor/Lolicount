#!/usr/bin/env python3
"""Verify generated Unity UI nodes against their source RectTransforms."""

from __future__ import annotations

import argparse
import json
import math
from pathlib import Path

import UnityPy


EXCLUDED_NAVICHARA = {"17", "23"}


def serialized_file(bundle_path: Path):
    """Load the first serialized file stored in a Unity AssetBundle."""
    env = UnityPy.load(str(bundle_path))
    bundle = next(iter(env.files.values()))
    return next(iter(bundle.files.values()))


def close(left: float, right: float, tolerance: float = 1e-4) -> bool:
    """Return whether two serialized float values are equivalent."""
    return math.isclose(float(left), float(right), abs_tol=tolerance)


def script_name(obj) -> str:
    """Read a MonoBehaviour's declaring class."""
    try:
        return obj.read().m_Script.read().m_ClassName
    except Exception:
        return ""


def verify_prefab(prefab: Path, output_root: Path) -> list[str]:
    """Compare one generated node list with the prefab's local transforms."""
    suffix = prefab.stem.removeprefix("ui_navichara_")
    manifest = output_root / f"maimai-navichara-{suffix}" / "model.json"
    if not manifest.is_file():
        return [f"{prefab.name}: missing {manifest}"]

    sf = serialized_file(prefab)
    transforms = {
        path_id: obj.read_typetree()
        for path_id, obj in sf.objects.items()
        if obj.type.name == "RectTransform"
    }
    game_objects = {
        path_id: obj.read()
        for path_id, obj in sf.objects.items()
        if obj.type.name == "GameObject"
    }
    go_to_transform = {
        tree["m_GameObject"]["m_PathID"]: path_id
        for path_id, tree in transforms.items()
    }
    roots = [path_id for path_id, tree in transforms.items() if not tree["m_Father"]["m_PathID"]]
    ordered: list[int] = []

    def visit(path_id: int) -> None:
        ordered.append(path_id)
        for child in transforms[path_id]["m_Children"]:
            if child["m_PathID"] in transforms:
                visit(child["m_PathID"])

    for root in roots:
        visit(root)

    nodes = json.loads(manifest.read_text(encoding="utf-8"))["nodes"]
    if len(nodes) != len(ordered):
        return [f"{prefab.name}: node count {len(nodes)} != {len(ordered)}"]

    errors = []
    indices = {path_id: index for index, path_id in enumerate(ordered)}
    expected_constraints = {}
    disabled_renderers = {}
    for obj in sf.objects.values():
        if obj.type.name != "MonoBehaviour":
            continue
        class_name = script_name(obj)
        tree = obj.read_typetree()
        if class_name in {"Image", "MultipleImage"} and not tree.get("m_Enabled", 1):
            self_path = go_to_transform[tree["m_GameObject"]["m_PathID"]]
            disabled_renderers[indices[self_path]] = class_name
        if class_name != "UIConstraintParentLocal":
            continue
        self_path = go_to_transform[tree["m_GameObject"]["m_PathID"]]
        target_path = tree["target"]["m_PathID"]
        expected_constraints[indices[self_path]] = {
            "target": indices[target_path],
            "mode": "offset" if tree.get("mode") == 1 else "driven",
            "syncPosition": bool(tree.get("syncPosition")),
            "syncRotation": bool(tree.get("syncRotation")),
            "syncScale": bool(tree.get("syncScale")),
            "positionOffset": [tree["positionOffset"][key] for key in ("x", "y", "z")],
            "rotationOffset": [tree["rotationOffset"][key] for key in ("x", "y", "z")],
            "scaleOffset": [tree["scaleOffset"][key] for key in ("x", "y", "z")],
        }
    for index, expected in expected_constraints.items():
        actual = nodes[index].get("constraint")
        if actual != expected:
            errors.append(
                f"{prefab.stem}/{nodes[index]['name']}: constraint {actual} != {expected}"
            )
    for index, class_name in disabled_renderers.items():
        if nodes[index].get("rendererEnabled", True):
            errors.append(
                f"{prefab.stem}/{nodes[index]['name']}: disabled {class_name} renderer was exported as visible"
            )
    for index, path_id in enumerate(ordered):
        tree = transforms[path_id]
        source = tree["m_LocalPosition"]
        actual = nodes[index]["position"]
        if all(close(actual[axis], source[key]) for axis, key in enumerate(("x", "y", "z"))):
            continue
        game_object = game_objects[tree["m_GameObject"]["m_PathID"]]
        errors.append(
            f"{prefab.stem}/{game_object.m_Name}: position {actual} != "
            f"local [{source['x']}, {source['y']}, {source['z']}]"
        )
    return errors


def main() -> int:
    """Verify every generated navigation-character model."""
    parser = argparse.ArgumentParser()
    parser.add_argument("source", type=Path, help="AssetBundleImages directory")
    parser.add_argument("--output", type=Path, default=Path("assets/unity"))
    args = parser.parse_args()

    prefab_root = args.source / "navichara" / "prefab"
    if not prefab_root.is_dir():
        parser.error("source does not contain navichara/prefab")
    prefabs = [
        prefab
        for prefab in sorted(prefab_root.glob("ui_navichara_*.ab"))
        if prefab.stem.removeprefix("ui_navichara_") not in EXCLUDED_NAVICHARA
        if (args.source / "navichara" / "animation" / prefab.stem).is_dir()
    ]
    if not prefabs:
        parser.error("source contains no animated navigation-character prefabs")
    errors = [
        error
        for prefab in prefabs
        for error in verify_prefab(prefab, args.output)
    ]
    if errors:
        for error in errors[:20]:
            print(error)
        print(f"FAILED: {len(errors)} Unity UI manifest mismatches")
        return 1
    print(f"OK: {len(prefabs)} generated models match Unity transforms and constraints")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
