"""Embed the separately verified original structure registry order."""
import argparse
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]

def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("--check", action="store_true")
    args = parser.parse_args()
    fixture = ROOT / "docs/worldgen/structure_order_1.21.1.txt"
    target = ROOT / "internal/worldgen/structure_order_1.21.1.txt"
    lines = [line for line in fixture.read_text(encoding="utf-8").splitlines() if line.startswith("order ")]
    if len(lines) != 34:
        raise ValueError("expected all 34 original registry entries")
    actual = "\n".join(lines) + "\n"
    if args.check:
        if target.read_text(encoding="utf-8") != actual:
            raise SystemExit("embedded structure registry order differs from verified fixture")
        print("verified 34 embedded original registry entries")
    else:
        target.write_text(actual, encoding="utf-8", newline="\n")
        print(f"generated {target}: 34 original registry entries")

if __name__ == "__main__":
    main()
