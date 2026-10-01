"""Build a compact, ordered 1.21.1 block-state table from blocks.json."""
import gzip
import itertools
import json
from pathlib import Path

ROOT = Path(__file__).resolve().parent
blocks = json.loads((ROOT / "blocks.json").read_text(encoding="utf-8"))
states = [None] * (max(b["maxStateId"] for b in blocks) + 1)

for block in blocks:
    choices = []
    for prop in block["states"]:
        if prop["type"] == "bool":
            values = ["true", "false"]
        else:
            values = [str(v) for v in prop.get("values", range(prop["num_values"]))]
        if len(values) != prop["num_values"]:
            raise ValueError((block["name"], prop["name"]))
        choices.append((prop["name"], values))
    combinations = itertools.product(*(values for _, values in choices))
    for state_id, combination in enumerate(combinations, block["minStateId"]):
        if state_id > block["maxStateId"]:
            raise ValueError(f"too many states for {block['name']}")
        states[state_id] = {
            "name": "minecraft:" + block["name"],
            "properties": dict(zip((p[0] for p in choices), combination)),
            "default": state_id == block["defaultState"],
        }
    if any(states[i] is None for i in range(block["minStateId"], block["maxStateId"] + 1)):
        raise ValueError(f"missing states for {block['name']}")

if any(state is None for state in states):
    raise ValueError("gaps in global block-state IDs")

output = ROOT.parent.parent / "internal" / "world" / "states.json.gz"
with output.open("wb") as out:
    with gzip.GzipFile(fileobj=out, mode="wb", mtime=0, filename="") as zipped:
        zipped.write(json.dumps(states, ensure_ascii=False, separators=(",", ":")).encode())
print(f"wrote {len(states)} block states to {output}")
