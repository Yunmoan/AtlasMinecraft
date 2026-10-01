"""Build the protocol-767 item registry and component framing schema."""
import json
from pathlib import Path

HERE = Path(__file__).resolve().parent
DEST = HERE.parents[1] / "internal" / "item"
DEST.mkdir(exist_ok=True)
protocol = json.loads((HERE / "protocol.json").read_text(encoding="utf-8"))
types = protocol["types"]
# The upstream table includes potion custom_name from a later release and
# misframes FoodProperties. Verified against the 1.21.1 server STREAM_CODECs
# (PotionContents/cwu, FoodProperties/cpr and PossibleEffect/cpr$b).
fields = types["SlotComponent"][1][1]["type"][1]["fields"]
fields["potion_contents"][1] = [
    field for field in fields["potion_contents"][1] if field["name"] != "customName"
]
for field in fields["food"][1]:
    if field["name"] == "usingConvertsTo":
        # Optional.of(ItemStack.STREAM_CODEC), including its presence boolean.
        field["type"] = ["option", "Slot"]
    elif field["name"] == "effects":
        field["type"][1]["type"][1][0]["type"] = "ItemPotionEffect"
# Keep only dependencies of Slot; this is framing, not gameplay/registry data.
needed = {}

def visit(value):
    if isinstance(value, str):
        if value in types and value not in needed and types[value] != "native":
            needed[value] = types[value]
            visit(types[value])
    elif isinstance(value, list):
        for child in value:
            visit(child)
    elif isinstance(value, dict):
        for child in value.values():
            visit(child)

visit("Slot")
(DEST / "schema_767.json").write_text(json.dumps(needed, separators=(",", ":")), encoding="utf-8")
items = json.loads((HERE / "items.json").read_text(encoding="utf-8"))
(DEST / "items_767.json").write_text(json.dumps([
    {"id": i["id"], "name": "minecraft:" + i["name"], "stackSize": i["stackSize"]}
    for i in items
], separators=(",", ":")), encoding="utf-8")
print("built", len(items), "items and", len(needed), "component schema types")
