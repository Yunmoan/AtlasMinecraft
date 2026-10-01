"""Build exact 1.21.1 feature heightmap block classifiers from JAR vectors."""

import argparse
import base64
import gzip
import json
from pathlib import Path


ROOT = Path(__file__).resolve().parents[2]
REFERENCE = ROOT / "docs/worldgen/motion_blocking_1.21.1.txt"
STATES = ROOT / "internal/world/states.json.gz"
OUTPUT = ROOT / "internal/worldgen/motion_blocking_data.json"


def build() -> str:
    lines = REFERENCE.read_text(encoding="utf-8").splitlines()
    header = lines[0].split()
    if len(header) != 3 or header[:2] != ["motion_blocking_states", "26684"]:
        raise ValueError("invalid JAR motion blocking state header")
    flags = base64.b64decode(header[2], validate=True)
    states = json.loads(gzip.decompress(STATES.read_bytes()))
    if len(states) != 26684 or len(flags) != (len(states) + 7) // 8:
        raise ValueError("state registry does not match JAR motion blocking vectors")

    ocean_header = lines[1].split()
    if len(ocean_header) != 3 or ocean_header[:2] != ["ocean_floor_states", "26684"]:
        raise ValueError("invalid JAR ocean floor state header")
    ocean_flags = base64.b64decode(ocean_header[2], validate=True)
    if len(ocean_flags) != len(flags):
        raise ValueError("invalid JAR ocean floor flags")
    kinds = {}
    ocean_kinds = {}
    for line in lines[2:]:
        fields = line.split()
        if not fields:
            continue
        if fields[0] == "motion_blocking_name":
            if len(fields) != 3 or fields[1] in kinds or fields[2] not in {
                "always", "never", "waterlogged_default_true", "waterlogged_default_false"
            }:
                raise ValueError(f"invalid JAR motion blocking name row: {line}")
            kinds[fields[1]] = fields[2]
        elif fields[0] == "ocean_floor_name":
            if len(fields) != 3 or fields[1] in ocean_kinds or fields[2] not in {"always", "never"}:
                raise ValueError(f"invalid JAR ocean floor name row: {line}")
            ocean_kinds[fields[1]] = fields[2]
    if len(kinds) != 1060 or len(ocean_kinds) != 1060 or kinds.keys() != ocean_kinds.keys():
        raise ValueError(f"expected 1060 JAR block names, got {len(kinds)} and {len(ocean_kinds)}")
    for state_id, state in enumerate(states):
        kind = kinds.get(state["name"])
        if kind is None:
            raise ValueError(f"missing JAR classifier for {state['name']}")
        if kind == "always":
            blocking = True
        elif kind == "never":
            blocking = False
        else:
            if "waterlogged" not in state["properties"]:
                raise ValueError(f"missing waterlogged state property: {state['name']}")
            blocking = state["properties"]["waterlogged"] == "true"
        jar_blocking = bool(flags[state_id // 8] & (1 << (state_id % 8)))
        if blocking != jar_blocking:
            raise ValueError(f"block state {state_id} differs from the JAR classifier")
        jar_ocean = bool(ocean_flags[state_id // 8] & (1 << (state_id % 8)))
        if (ocean_kinds[state["name"]] == "always") != jar_ocean:
            raise ValueError(f"ocean floor state {state_id} differs from the JAR classifier")
    data = {name: {"motion": kind, "ocean_floor": ocean_kinds[name] == "always"}
            for name, kind in kinds.items()}
    return json.dumps(data, sort_keys=True, indent=2) + "\n"


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("--check", action="store_true")
    args = parser.parse_args()
    content = build()
    if args.check:
        if not OUTPUT.exists() or OUTPUT.read_text(encoding="utf-8") != content:
            raise SystemExit("motion blocking data is out of date")
        print(f"verified {OUTPUT}")
    else:
        OUTPUT.write_text(content, encoding="utf-8", newline="\n")
        print(f"updated {OUTPUT}")


if __name__ == "__main__":
    main()
