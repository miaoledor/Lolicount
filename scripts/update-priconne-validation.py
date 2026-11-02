"""Rebuild the Princess Connect validation summary from the theme tree.

Reference counts are the whole-image renders the exporter compared against the
split layers. They are reported per outfit, and consolidating outfits into NPC
themes needs the total for outfits that no longer exist on their own, so every
outfit keeps its own entry here: NPC themes list one entry per set. Nothing is
derived from a previous summary, so any snapshot reproduces the same numbers
from the theme tree, the manifest, the exclusion lists and the exporter
reports written into the theme catalogs.
"""

import hashlib
import json
from pathlib import Path


def read_json(path):
    return json.loads(path.read_text(encoding="utf-8"))


def write_json(path, value):
    path.write_text(json.dumps(value, ensure_ascii=False, indent=2) + "\n", encoding="utf-8", newline="\n")


def range_counts(ranges):
    return {key: span["last"] - span["first"] + 1 for key, span in ranges.items()}


def combinations(ranges):
    value = 1
    for count in range_counts(ranges).values():
        value *= count
    return value


def manifest_models(repo):
    """Return the story model ids in manifest order."""
    models = []
    for line in (repo / "scripts/priconne-manifest.txt").read_text(encoding="utf-8").splitlines():
        if "spine_full_" in line:
            models.append(line.split("spine_full_", 1)[1].split(".", 1)[0].split(",", 1)[0])
    return models


def read_report(path):
    """Return exporter results keyed by theme name.

    Exports recorded before outfit merging ran under the outfit's own theme
    name, which no longer exists for merged NPC outfits; those records are
    resolved by the outfit id stored in the set name instead.
    """
    if not path.exists():
        return {}
    return {entry["name"]: entry for entry in read_json(path)}


def read_existing_validation(path):
    """Keep exporter measurements already recorded in the checked-in file."""
    if not path.exists():
        return {}
    reports = {}
    for model in read_json(path).get("models", []):
        if model.get("sets"):
            for configured in model["sets"]:
                reports[f"pcr-{configured['name']}"] = {
                    "references": configured.get("references", 0),
                    "reference_error": configured.get("maximumReferenceError", 0),
                }
        else:
            reports[model["name"]] = {
                "references": model.get("references", 0),
                "reference_error": model.get("maximumReferenceError", model.get("reference_error", 0)),
            }
    return reports


def outfit_report(reports, outfit):
    """Look up the exporter result of one outfit id."""
    return reports.get(f"pcr-{outfit}", {})


def main():
    repo = Path(__file__).resolve().parents[1]
    excluded = read_json(repo / "scripts/priconne-exclusions.json")
    silhouettes = read_json(repo / "scripts/priconne-silhouettes.json")
    low_variant = set(read_json(repo / "scripts/priconne-excluded-themes.json"))

    # Every retained story model is its own theme; the leading four digits of a
    # model ID are a story batch, not a character, so they never name a theme.
    source_models = [
        model for model in manifest_models(repo)
        if model not in excluded and model not in silhouettes and f"pcr-{model}" not in low_variant
    ]
    validation_path = repo / "docs/priconne-validation.json"
    reports = read_existing_validation(validation_path)
    reports.update(read_report(repo / "scripts/priconne-render-report.json"))
    theme_root = repo / "assets/theme"

    models = []
    asset_bytes = 0
    body_only = 0
    for directory in sorted(theme_root.glob("pcr-*")):
        config = read_json(directory / "config.json")
        layers = len(read_json(directory / "ren.json")) - 1
        asset_bytes += sum(path.stat().st_size for path in (directory / "ren").glob("*.webp"))
        if config.get("sets"):
            sets = []
            for configured in config["sets"]:
                source = outfit_report(reports, configured["name"])
                sets.append({
                    "name": configured["name"],
                    "counts": range_counts(configured["ranges"]),
                    "combinations": combinations(configured["ranges"]),
                    "references": source.get("references", 0),
                    "maximumReferenceError": source.get("reference_error", 0),
                })
            models.append({
                "name": directory.name,
                "outfits": len(sets),
                "layers": layers,
                "combinations": sum(item["combinations"] for item in sets),
                "references": sum(item["references"] for item in sets),
                "maximumReferenceError": max(item["maximumReferenceError"] for item in sets),
                "sets": sets,
            })
            if all(set(item["counts"]) == {"body"} for item in sets):
                body_only += 1
        else:
            source = outfit_report(reports, directory.name.removeprefix("pcr-"))
            counts = range_counts(config["ranges"])
            models.append({
                "name": directory.name,
                "layers": layers,
                "counts": counts,
                "combinations": combinations(config["ranges"]),
                "references": source.get("references", 0),
                "maximumReferenceError": source.get("reference_error", 0),
            })
            if set(counts) == {"body"}:
                body_only += 1

    manifest = (repo / "scripts/priconne-manifest.txt").read_text(encoding="utf-8").replace("\r\n", "\n").encode()
    summary = {
        "themes": len(models),
        "sourceModels": sum(model.get("outfits", 1) for model in models),
        "sourceManifestModels": len(source_models),
        "layers": sum(model["layers"] for model in models),
        "combinations": sum(model["combinations"] for model in models),
        "references": sum(model["references"] for model in models),
        "maximumReferenceError": max(model["maximumReferenceError"] for model in models),
        "bodyOnlyThemes": body_only,
        "assetBytes": asset_bytes,
        "excludedNonPortraitModels": len(excluded),
        "excludedSilhouetteModels": len(silhouettes),
        "excludedThemes": len(low_variant),
        "manifestSHA256_UTF8_LF": hashlib.sha256(manifest).hexdigest(),
    }
    write_json(validation_path, {"summary": summary, "models": models})
    print(f"updated validation for {len(models)} themes from {summary['sourceModels']} source models")


if __name__ == "__main__":
    main()
