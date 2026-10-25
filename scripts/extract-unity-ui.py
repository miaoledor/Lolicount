#!/usr/bin/env python3
"""Convert maimai DX Unity UGUI bundles into Lolicount web models.

The source bundle tree is expected to contain the navichara/prefab,
navichara/animation and navichara/sprite directories from the game. UnityPy
is intentionally a build-time dependency only; generated JSON and WebP files
are embedded into the Go binary and need no Unity runtime in production.
"""

from __future__ import annotations

import argparse
import json
import math
import re
import struct
import sys
import zlib
from collections import defaultdict
from pathlib import Path
from typing import Any

import UnityPy
from PIL import Image


TRANSFORM = 4
GAME_OBJECT = 1
MONO_BEHAVIOUR = 114
RECT_TRANSFORM = 224
EXCLUDED_NAVICHARA = {"17", "23"}  # Static partner characters, not standalone themes.

ATTR = {
    "active": zlib.crc32(b"m_IsActive") & 0xFFFFFFFF,
    "anchored_x": zlib.crc32(b"m_AnchoredPosition.x") & 0xFFFFFFFF,
    "anchored_y": zlib.crc32(b"m_AnchoredPosition.y") & 0xFFFFFFFF,
    "local_x": zlib.crc32(b"m_LocalPosition.x") & 0xFFFFFFFF,
    "local_y": zlib.crc32(b"m_LocalPosition.y") & 0xFFFFFFFF,
    "local_z": zlib.crc32(b"m_LocalPosition.z") & 0xFFFFFFFF,
    "pivot_x": zlib.crc32(b"m_Pivot.x") & 0xFFFFFFFF,
    "pivot_y": zlib.crc32(b"m_Pivot.y") & 0xFFFFFFFF,
    "size_x": zlib.crc32(b"m_SizeDelta.x") & 0xFFFFFFFF,
    "size_y": zlib.crc32(b"m_SizeDelta.y") & 0xFFFFFFFF,
    "sprite_index": zlib.crc32(b"_selectSpriteIndex") & 0xFFFFFFFF,
    "sprite": zlib.crc32(b"m_Sprite") & 0xFFFFFFFF,
    "color_r": zlib.crc32(b"m_Color.r") & 0xFFFFFFFF,
    "color_g": zlib.crc32(b"m_Color.g") & 0xFFFFFFFF,
    "color_b": zlib.crc32(b"m_Color.b") & 0xFFFFFFFF,
    "color_a": zlib.crc32(b"m_Color.a") & 0xFFFFFFFF,
}


def serialized_file(bundle_path: Path):
    """Load the first serialized file stored in a Unity AssetBundle."""
    env = UnityPy.load(str(bundle_path))
    bundle = next(iter(env.files.values()))
    return next(iter(bundle.files.values()))


def build_cab_index(root: Path) -> dict[str, Path]:
    """Map Unity's archive CAB names back to the individual bundle paths."""
    index: dict[str, Path] = {}
    for path in (root / "navichara").rglob("*.ab"):
        sf = serialized_file(path)
        index[sf.name.lower()] = path
    return index


class SpriteStore:
    """Resolve cross-bundle sprite pointers and emit de-duplicated WebPs."""

    def __init__(self, cab_index: dict[str, Path], output: Path):
        self.cab_index = cab_index
        self.output = output
        self.files: dict[str, Any] = {}
        self.emitted: dict[tuple[str, int], str] = {}
        self.additive_files: dict[str, str] = {}
        self.used_names: set[str] = set()

    def _file(self, cab: str):
        if cab not in self.files:
            path = self.cab_index.get(cab)
            if path is None:
                raise KeyError(f"missing dependency CAB: {cab}")
            self.files[cab] = serialized_file(path)
        return self.files[cab]

    def emit(self, owner, pointer: dict[str, int] | Any) -> str | None:
        """Resolve a PPtr relative to owner and save its Sprite as lossless WebP."""
        file_id = pointer.get("m_FileID", 0) if isinstance(pointer, dict) else pointer.m_FileID
        path_id = pointer.get("m_PathID", 0) if isinstance(pointer, dict) else pointer.m_PathID
        if not path_id:
            return None
        target = owner
        cab = owner.name.lower()
        if file_id:
            external = owner.externals[file_id - 1]
            cab = external.name.lower()
            target = self._file(cab)
        key = (cab, path_id)
        if key in self.emitted:
            return self.emitted[key]
        obj = target.objects.get(path_id)
        if obj is None or obj.type.name != "Sprite":
            return None
        sprite = obj.read()
        base = re.sub(r"[^a-z0-9_-]+", "-", sprite.m_Name.lower()).strip("-") or "sprite"
        name = f"{base}.webp"
        serial = 2
        while name in self.used_names:
            name = f"{base}-{serial}.webp"
            serial += 1
        self.used_names.add(name)
        sprite.image.save(self.output / name, "WEBP", lossless=True, quality=100, method=0)
        self.emitted[key] = name
        return name

    def additive(self, name: str | None) -> str | None:
        """Convert a black-background additive sprite to straight alpha."""
        if not name:
            return None
        if name in self.additive_files:
            return self.additive_files[name]
        path = self.output / name
        with Image.open(path) as source:
            image = source.convert("RGBA")
        converted = Image.new("RGBA", image.size)
        source_pixels = image.load()
        converted_pixels = converted.load()
        for y in range(image.height):
            for x in range(image.width):
                red, green, blue, alpha = source_pixels[x, y]
                intensity = max(red, green, blue)
                if intensity == 0 or alpha == 0:
                    converted_pixels[x, y] = (0, 0, 0, 0)
                    continue
                converted_pixels[x, y] = (
                    min(255, round(red * 255 / intensity)),
                    min(255, round(green * 255 / intensity)),
                    min(255, round(blue * 255 / intensity)),
                    round(alpha * intensity / 255),
                )
        additive_name = f"{Path(name).stem}-add.webp"
        converted.save(self.output / additive_name, "WEBP", lossless=True, quality=100, method=0)
        self.used_names.add(additive_name)
        self.additive_files[name] = additive_name
        return additive_name


def pptr(tree: dict[str, Any], key: str) -> dict[str, int] | None:
    """Return a serialized pointer field when it contains an object."""
    value = tree.get(key)
    return value if isinstance(value, dict) and value.get("m_PathID") else None


def script_name(obj) -> str:
    """Read a MonoBehaviour's declaring class without relying on filenames."""
    try:
        return obj.read().m_Script.read().m_ClassName
    except Exception:
        return ""


def quaternion_z(q: dict[str, float]) -> float:
    """Convert a Unity quaternion to its two-dimensional Z rotation."""
    sin_z = 2 * (q["w"] * q["z"] + q["x"] * q["y"])
    cos_z = 1 - 2 * (q["y"] * q["y"] + q["z"] * q["z"])
    return math.degrees(math.atan2(sin_z, cos_z))


def transform_width(binding) -> int:
    """Return the number of scalar curves represented by a generic binding."""
    if binding.typeID != TRANSFORM:
        return 1
    return {1: 3, 2: 4, 3: 3, 4: 3}.get(binding.attribute, 1)


def curve_slots(bindings) -> list[tuple[Any, int]]:
    """Expand Unity's vector bindings into scalar curve slots."""
    slots: list[tuple[Any, int]] = []
    for binding in bindings:
        slots.extend((binding, component) for component in range(transform_width(binding)))
    return slots


def streamed_keys(clip) -> dict[int, list[list[float]]]:
    """Decode StreamedClip polynomial keys using Unity's little-endian layout."""
    words = clip.m_StreamedClip.data
    raw = struct.pack(f"<{len(words)}I", *words)
    curves: dict[int, list[list[float]]] = defaultdict(list)
    pos = 0
    sentinel: dict[int, list[float]] = {}
    while pos < len(raw):
        time, count = struct.unpack_from("<fi", raw, pos)
        pos += 8
        for _ in range(count):
            index, a, b, c, value = struct.unpack_from("<i4f", raw, pos)
            pos += 20
            key = [time, value, a, b, c]
            if time < -1e20:
                sentinel[index] = [0.0, value, a, b, c]
            elif math.isfinite(time):
                curves[index].append(key)
    for index, key in sentinel.items():
        if index not in curves or curves[index][0][0] > 0:
            curves[index].insert(0, key)
    return curves


def all_clip_curves(animation) -> dict[int, dict[str, Any]]:
    """Combine streamed, dense and constant Unity animation storage."""
    clip = animation.m_MuscleClip.m_Clip.data
    curves = {
        index: {"cubic": True, "keys": keys}
        for index, keys in streamed_keys(clip).items()
    }
    stream_count = clip.m_StreamedClip.curveCount
    dense = clip.m_DenseClip
    for curve_index in range(dense.m_CurveCount):
        keys = []
        for frame in range(dense.m_FrameCount):
            time = dense.m_BeginTime + frame / dense.m_SampleRate
            value = dense.m_SampleArray[frame * dense.m_CurveCount + curve_index]
            keys.append([time, value])
        curves[stream_count + curve_index] = {"cubic": False, "keys": keys}
    constant = clip.m_ConstantClip.data if clip.m_ConstantClip else []
    start = stream_count + dense.m_CurveCount
    duration = animation.m_MuscleClip.m_StopTime
    for index, value in enumerate(constant):
        curves[start + index] = {
            "cubic": False,
            "keys": [[0.0, value], [duration, value]],
        }
    return curves


def track_property(binding, component: int) -> str | None:
    """Translate the bindings used by these UGUI models to player properties."""
    if binding.typeID == TRANSFORM:
        prefix = {1: "position", 2: "quaternion", 3: "scale", 4: "euler"}.get(binding.attribute)
        axes = "xyzw" if binding.attribute == 2 else "xyz"
        return f"{prefix}.{axes[component]}" if prefix else None
    if binding.typeID == GAME_OBJECT and binding.attribute == ATTR["active"]:
        return "active"
    if binding.typeID == RECT_TRANSFORM:
        return {
            ATTR["anchored_x"]: "position.x",
            ATTR["anchored_y"]: "position.y",
            ATTR["local_x"]: "position.x",
            ATTR["local_y"]: "position.y",
            ATTR["local_z"]: "position.z",
            ATTR["pivot_x"]: "pivot.x",
            ATTR["pivot_y"]: "pivot.y",
            ATTR["size_x"]: "size.x",
            ATTR["size_y"]: "size.y",
        }.get(binding.attribute)
    if binding.typeID == MONO_BEHAVIOUR:
        return {
            ATTR["sprite_index"]: "spriteIndex",
            ATTR["sprite"]: "sprite",
            ATTR["color_r"]: "color.r",
            ATTR["color_g"]: "color.g",
            ATTR["color_b"]: "color.b",
            ATTR["color_a"]: "color.a",
        }.get(binding.attribute)
    return None


def localize_curve(
    binding,
    component: int,
    node: int,
    curve: dict[str, Any],
    offsets: dict[int, list[float]],
) -> dict[str, Any]:
    """Convert animated anchored positions to the parent's local coordinate space."""
    if binding.typeID != RECT_TRANSFORM:
        return curve
    axes = {
        ATTR["anchored_x"]: 0,
        ATTR["anchored_y"]: 1,
    }
    axis = axes.get(binding.attribute)
    if axis is None or component != 0:
        return curve
    offset = offsets[node][axis]
    if abs(offset) < 1e-8:
        return curve
    return {
        **curve,
        "keys": [[key[0], key[1] + offset, *key[2:]] for key in curve["keys"]],
    }


def export_animation(
    path: Path,
    path_nodes: dict[int, int],
    position_offsets: dict[int, list[float]],
    sprites: SpriteStore,
) -> dict[str, Any] | None:
    """Convert one AnimationClip bundle to compact scalar tracks."""
    sf = serialized_file(path)
    obj = next((o for o in sf.objects.values() if o.type.name == "AnimationClip"), None)
    if obj is None:
        return None
    animation = obj.read()
    slots = curve_slots(animation.m_ClipBindingConstant.genericBindings)
    tracks = []
    for index, curve in all_clip_curves(animation).items():
        if index >= len(slots):
            continue
        binding, component = slots[index]
        node = path_nodes.get(binding.path)
        prop = track_property(binding, component)
        if node is None or prop is None:
            continue
        curve = localize_curve(binding, component, node, curve, position_offsets)
        if prop == "sprite":
            mapping = animation.m_ClipBindingConstant.pptrCurveMapping
            converted = []
            for key in curve["keys"]:
                pointer_index = int(round(key[1]))
                image = sprites.emit(sf, mapping[pointer_index]) if 0 <= pointer_index < len(mapping) else None
                if image:
                    converted.append([key[0], image])
            if not converted:
                continue
            curve = {"cubic": False, "keys": converted}
        elif prop in {"active", "spriteIndex"}:
            curve = {**curve, "discrete": True}
        tracks.append({"node": node, "property": prop, **curve})
    if not tracks:
        return None
    return {
        "name": animation.m_Name,
        "duration": round(float(animation.m_MuscleClip.m_StopTime), 6),
        "loop": bool(animation.m_MuscleClip.m_LoopTime),
        "tracks": tracks,
    }


def multiply(a: list[float], b: list[float]) -> list[float]:
    """Multiply two compact 2D affine matrices."""
    return [
        a[0] * b[0] + a[2] * b[1],
        a[1] * b[0] + a[3] * b[1],
        a[0] * b[2] + a[2] * b[3],
        a[1] * b[2] + a[3] * b[3],
        a[0] * b[4] + a[2] * b[5] + a[4],
        a[1] * b[4] + a[3] * b[5] + a[5],
    ]


def model_bounds(nodes: list[dict[str, Any]]) -> list[float]:
    """Calculate an initial-pose content box used by the browser's contain fit."""
    poses: list[dict[str, Any] | None] = [None] * len(nodes)
    resolving: set[int] = set()

    def matrix_pose(matrix: list[float]) -> dict[str, Any]:
        scale_x = math.hypot(matrix[0], matrix[1])
        determinant = matrix[0] * matrix[3] - matrix[1] * matrix[2]
        scale_y = determinant / scale_x if scale_x else math.hypot(matrix[2], matrix[3])
        return {
            "matrix": matrix,
            "position": matrix[4:6],
            "rotation": math.degrees(math.atan2(matrix[1], matrix[0])),
            "scale": [scale_x, scale_y],
        }

    def pose_matrix(position: list[float], rotation: float, scale: list[float]) -> list[float]:
        angle = math.radians(rotation)
        return [
            math.cos(angle) * scale[0], math.sin(angle) * scale[0],
            -math.sin(angle) * scale[1], math.cos(angle) * scale[1],
            position[0], position[1],
        ]

    def resolve(index: int) -> dict[str, Any]:
        if poses[index] is not None:
            return poses[index]  # type: ignore[return-value]
        node = nodes[index]
        if index in resolving:
            return matrix_pose(pose_matrix(node["position"][:2], node["rotation"], node["scale"]))
        resolving.add(index)
        local = pose_matrix(node["position"][:2], node["rotation"], node["scale"])
        parent = node["parent"]
        normal = matrix_pose(multiply(resolve(parent)["matrix"], local) if parent is not None else local)
        constraint = node.get("constraint")
        if constraint and constraint["target"] != index:
            target = resolve(constraint["target"])
            if constraint["mode"] == "offset":
                position = [
                    target["matrix"][0] * constraint["positionOffset"][0] + target["matrix"][2] * constraint["positionOffset"][1] + target["matrix"][4],
                    target["matrix"][1] * constraint["positionOffset"][0] + target["matrix"][3] * constraint["positionOffset"][1] + target["matrix"][5],
                ]
                rotation = target["rotation"] + constraint["rotationOffset"][2]
                scale = [target["scale"][0] * constraint["scaleOffset"][0], target["scale"][1] * constraint["scaleOffset"][1]]
            else:
                position, rotation, scale = target["position"][:], target["rotation"], target["scale"][:]
            position = position if constraint["syncPosition"] else normal["position"]
            rotation = rotation if constraint["syncRotation"] else normal["rotation"]
            scale = scale if constraint["syncScale"] else normal["scale"]
            normal = matrix_pose(pose_matrix(position, rotation, scale))
        resolving.remove(index)
        poses[index] = normal
        return normal

    points: list[tuple[float, float]] = []
    for index, node in enumerate(nodes):
        world = resolve(index)["matrix"]
        if not node.get("sprite") or not node["active"] or node.get("rendererEnabled") is False:
            continue
        width, height = node["size"]
        pivot_x, pivot_y = node["pivot"]
        for x, y in (
            (-pivot_x * width, -pivot_y * height),
            ((1 - pivot_x) * width, -pivot_y * height),
            ((1 - pivot_x) * width, (1 - pivot_y) * height),
            (-pivot_x * width, (1 - pivot_y) * height),
        ):
            points.append((world[0] * x + world[2] * y + world[4], world[1] * x + world[3] * y + world[5]))
    if not points:
        return [-300, -400, 300, 400]
    xs, ys = zip(*points)
    return [round(min(xs), 3), round(min(ys), 3), round(max(xs), 3), round(max(ys), 3)]


def export_model(source: Path, prefab: Path, output_root: Path, cab_index: dict[str, Path]) -> str:
    """Export one navigation-character prefab and all of its animation clips."""
    suffix = prefab.stem.removeprefix("ui_navichara_")
    model_name = f"maimai-navichara-{suffix}"
    output = output_root / model_name
    if output.is_dir():
        for existing in output.iterdir():
            if existing.is_file() and existing.suffix in {".json", ".webp"}:
                existing.unlink()
    output.mkdir(parents=True, exist_ok=True)
    sprite_store = SpriteStore(cab_index, output)
    sf = serialized_file(prefab)
    objects = sf.objects
    game_objects = {
        path_id: obj.read()
        for path_id, obj in objects.items()
        if obj.type.name == "GameObject"
    }
    names = {path_id: obj.m_Name for path_id, obj in game_objects.items()}
    transforms = {
        path_id: obj.read_typetree()
        for path_id, obj in objects.items()
        if obj.type.name == "RectTransform"
    }
    go_to_transform = {tree["m_GameObject"]["m_PathID"]: path_id for path_id, tree in transforms.items()}
    images: dict[int, dict[str, Any]] = {}
    constraints: dict[int, dict[str, Any]] = {}
    additive_objects: set[int] = set()
    for obj in objects.values():
        if obj.type.name != "MonoBehaviour" or script_name(obj) != "UIConstraintParentLocal":
            continue
        tree = obj.read_typetree()
        constraints[tree["m_GameObject"]["m_PathID"]] = {
            "targetPath": tree["target"]["m_PathID"],
            "mode": "offset" if tree.get("mode") == 1 else "driven",
            "syncPosition": bool(tree.get("syncPosition")),
            "syncRotation": bool(tree.get("syncRotation")),
            "syncScale": bool(tree.get("syncScale")),
            "positionOffset": [tree["positionOffset"][key] for key in ("x", "y", "z")],
            "rotationOffset": [tree["rotationOffset"][key] for key in ("x", "y", "z")],
            "scaleOffset": [tree["scaleOffset"][key] for key in ("x", "y", "z")],
        }
    for obj in objects.values():
        if obj.type.name != "MonoBehaviour" or script_name(obj) != "BlendImage":
            continue
        tree = obj.read_typetree()
        if tree.get("_blendType") == 1:
            additive_objects.add(tree["m_GameObject"]["m_PathID"])
    for obj in objects.values():
        if obj.type.name != "MonoBehaviour" or script_name(obj) not in {"Image", "MultipleImage"}:
            continue
        tree = obj.read_typetree()
        go_id = tree["m_GameObject"]["m_PathID"]
        image = {
            "sprite": sprite_store.emit(sf, pptr(tree, "m_Sprite")) if pptr(tree, "m_Sprite") else None,
            "sprites": [],
            "color": [tree["m_Color"][key] for key in ("r", "g", "b", "a")],
            "rendererEnabled": bool(tree.get("m_Enabled", 1)),
            "additive": go_id in additive_objects,
            "preserveAspect": bool(tree.get("m_PreserveAspect", False)),
        }
        for pointer in tree.get("MultiSprites", []):
            emitted = sprite_store.emit(sf, pointer)
            if emitted:
                image["sprites"].append(emitted)
        images[go_id] = image
    for image in images.values():
        if not image["additive"]:
            continue
        image["sprite"] = sprite_store.additive(image["sprite"])
        image["sprites"] = [sprite_store.additive(name) for name in image["sprites"]]

    roots = [path_id for path_id, tree in transforms.items() if not tree["m_Father"]["m_PathID"]]
    ordered: list[int] = []

    def visit(path_id: int) -> None:
        ordered.append(path_id)
        for child in transforms[path_id]["m_Children"]:
            if child["m_PathID"] in transforms:
                visit(child["m_PathID"])

    for root in roots:
        visit(root)
    indices = {path_id: index for index, path_id in enumerate(ordered)}
    nodes = []
    paths: dict[int, str] = {}
    for path_id in ordered:
        tree = transforms[path_id]
        go_id = tree["m_GameObject"]["m_PathID"]
        parent_id = tree["m_Father"]["m_PathID"]
        parent = indices.get(parent_id)
        relative = names[go_id] if parent is None else f"{paths[parent_id]}/{names[go_id]}".lstrip("/")
        paths[path_id] = relative if parent is not None else ""
        game_object = game_objects[go_id]
        image = images.get(go_id, {})
        node = {
            "name": names[go_id],
            "parent": parent,
            "position": [tree["m_LocalPosition"][key] for key in ("x", "y", "z")],
            "rotation": quaternion_z(tree["m_LocalRotation"]),
            "scale": [tree["m_LocalScale"]["x"], tree["m_LocalScale"]["y"]],
            "size": [tree["m_SizeDelta"]["x"], tree["m_SizeDelta"]["y"]],
            "pivot": [tree["m_Pivot"]["x"], tree["m_Pivot"]["y"]],
            "active": bool(game_object.m_IsActive),
            "sprite": image.get("sprite"),
            "sprites": image.get("sprites", []),
            "color": image.get("color", [1, 1, 1, 1]),
            "additive": image.get("additive", False),
            "preserveAspect": image.get("preserveAspect", False),
        }
        constraint = constraints.get(go_id)
        if constraint:
            node["constraint"] = {
                **{key: value for key, value in constraint.items() if key != "targetPath"},
                "target": indices[constraint["targetPath"]],
            }
        if image and not image.get("rendererEnabled", True):
            node["rendererEnabled"] = False
        nodes.append(node)
    path_nodes = {
        zlib.crc32(relative.encode()) & 0xFFFFFFFF: indices[path_id]
        for path_id, relative in paths.items()
    }
    position_offsets = {
        indices[path_id]: [
            tree["m_LocalPosition"]["x"] - tree["m_AnchoredPosition"]["x"],
            tree["m_LocalPosition"]["y"] - tree["m_AnchoredPosition"]["y"],
        ]
        for path_id, tree in transforms.items()
    }
    animations = []
    animation_dir = source / "navichara" / "animation" / prefab.stem
    for path in sorted(animation_dir.glob("*.ab")):
        animation = export_animation(path, path_nodes, position_offsets, sprite_store)
        if animation:
            animations.append(animation)
    model = {
        "version": 1,
        "name": model_name,
        "bounds": model_bounds(nodes),
        "nodes": nodes,
        "animations": animations,
    }
    (output / "model.json").write_text(json.dumps(model, ensure_ascii=False, separators=(",", ":")), encoding="utf-8")
    return model_name


def main() -> int:
    """Parse arguments and export every animated navigation character."""
    parser = argparse.ArgumentParser()
    parser.add_argument("source", type=Path, help="AssetBundleImages directory")
    parser.add_argument("--output", type=Path, default=Path("assets/unity"))
    args = parser.parse_args()
    if not (args.source / "navichara" / "prefab").is_dir():
        parser.error("source does not contain navichara/prefab")
    args.output.mkdir(parents=True, exist_ok=True)
    cab_index = build_cab_index(args.source)
    prefabs = sorted((args.source / "navichara" / "prefab").glob("ui_navichara_*.ab"))
    prefabs = [
        prefab for prefab in prefabs
        if prefab.stem.removeprefix("ui_navichara_") not in EXCLUDED_NAVICHARA
    ]
    exported = []
    for prefab in prefabs:
        animation_dir = args.source / "navichara" / "animation" / prefab.stem
        if not animation_dir.is_dir():
            continue
        name = export_model(args.source, prefab, args.output, cab_index)
        exported.append(name)
        print(f"exported {name}", file=sys.stderr)
    (args.output / "README.md").write_text(
        "# Unity UI models\n\nGenerated by `scripts/extract-unity-ui.py`; do not edit model files manually.\n",
        encoding="utf-8",
    )
    print(json.dumps(exported))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
