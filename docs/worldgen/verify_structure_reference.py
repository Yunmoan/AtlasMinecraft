"""Compile the official-JAR structure oracle, then verify or update its vectors.

The separate script keeps independently maintained structure probes isolated
from the feature/noise reference runner. Requires JDK 21, Python 3.
"""

import argparse
import gzip
import json
import os
import subprocess
import tempfile
import time
import zipfile
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
HERE = Path(__file__).resolve().parent


def atomic_write(path: Path, encoded: bytes) -> None:
    """Publish a complete fixture without exposing a partial file to Go tests."""
    descriptor, name = tempfile.mkstemp(prefix=path.name + ".", suffix=".tmp", dir=path.parent)
    temporary = Path(name)
    published = False
    try:
        with os.fdopen(descriptor, "wb") as stream:
            stream.write(encoded)
            stream.flush()
            os.fsync(stream.fileno())
        for attempt in range(60):
            try:
                os.replace(temporary, path)
                published = True
                break
            except PermissionError:
                if attempt == 59:
                    raise PermissionError(f"fixture is still open; complete unpublished reference retained at {temporary}")
                time.sleep(0.2)
    finally:
        if published:
            temporary.unlink(missing_ok=True)


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("--update", action="store_true")
    parser.add_argument("--stderr-log", type=Path)
    parser.add_argument("--probe-arg", action="append", default=[])
    parser.add_argument("--output", type=Path)
    parser.add_argument("--probe", choices=("StructurePlacementProbe", "StructureBeardifierProbe", "StructureRingsProbe", "StructureStartsProbe", "StructureHutBlocksProbe", "StructurePyramidBlocksProbe", "StructureOrderProbe", "StructureJungleBlocksProbe", "StructureTransformProbe", "StructureEntityTransformProbe", "IglooStartsProbe", "IglooBlocksProbe", "RuinedPortalClassifiersProbe", "RuinedPortalProcessorsProbe", "RuinedPortalBlocksProbe", "RuinedPortalStartsProbe", "StructureSetSelectionProbe", "BuriedTreasureStartsProbe", "BuriedTreasureBlocksProbe", "NetherFossilStartsProbe", "NetherFossilBlocksProbe", "FortressStartsProbe", "FortressBlocksProbe"), default="StructurePlacementProbe")
    # The template probes share original-JAR loaders and defined input worlds.
    parser._option_string_actions["--probe"].choices += ("ShipwreckStartsProbe", "ShipwreckBlocksProbe", "BastionStubProbe", "JigsawStartsProbe", "JigsawBlocksProbe", "OceanRuinStartsProbe", "OceanRuinProcessorsProbe", "OceanRuinBlocksProbe")
    parser._option_string_actions["--probe"].choices += ("JigsawProcessorsProbe", "JigsawBlockEntitiesProbe", "JigsawOverworldBlocksProbe")
    parser._option_string_actions["--probe"].choices += ("JigsawOutpostBlocksProbe",)
    parser._option_string_actions["--probe"].choices += ("MineshaftStartsProbe", "MineshaftBlocksProbe")
    args = parser.parse_args()
    fixtures = {"StructurePlacementProbe": "structure_placement_1.21.1.txt",
                "StructureBeardifierProbe": "structure_beardifier_1.21.1.txt",
                "StructureRingsProbe": "structure_rings_1.21.1.txt",
                "StructureStartsProbe": "structure_starts_1.21.1.txt",
                "StructureHutBlocksProbe": "structure_hut_blocks_1.21.1.txt",
                "StructurePyramidBlocksProbe": "structure_pyramid_blocks_1.21.1.txt",
                "StructureOrderProbe": "structure_order_1.21.1.txt",
                "StructureJungleBlocksProbe": "structure_jungle_blocks_1.21.1.txt",
                "StructureTransformProbe": "structure_transforms_1.21.1.json.gz",
                "StructureEntityTransformProbe": "structure_entity_transforms_1.21.1.txt",
                "RuinedPortalClassifiersProbe": "ruined_portal_classifiers_1.21.1.json.gz",
                "RuinedPortalProcessorsProbe": "ruined_portal_processors_1.21.1.txt",
                "RuinedPortalBlocksProbe": "structure_ruined_portal_blocks_1.21.1.txt",
                "RuinedPortalStartsProbe": "structure_ruined_portal_starts_1.21.1.txt",
                "StructureSetSelectionProbe": "structure_set_selection_1.21.1.txt",
                "BuriedTreasureStartsProbe": "structure_buried_treasure_starts_1.21.1.txt",
                "BuriedTreasureBlocksProbe": "structure_buried_treasure_blocks_1.21.1.txt",
                "NetherFossilStartsProbe": "structure_nether_fossil_starts_1.21.1.txt",
                "NetherFossilBlocksProbe": "structure_nether_fossil_blocks_1.21.1.txt",
                "FortressStartsProbe": "structure_fortress_starts_1.21.1.txt",
                "FortressBlocksProbe": "structure_fortress_blocks_1.21.1.txt",
                "IglooStartsProbe": "structure_igloo_starts_1.21.1.txt",
                "IglooBlocksProbe": "structure_igloo_blocks_1.21.1.txt"}
    fixtures.update({"ShipwreckStartsProbe": "structure_shipwreck_starts_1.21.1.txt",
                     "ShipwreckBlocksProbe": "structure_shipwreck_blocks_1.21.1.txt",
                     "BastionStubProbe": "structure_bastion_stub_1.21.1.txt",
                     "JigsawStartsProbe": "structure_jigsaw_starts_1.21.1.txt",
                     "JigsawBlocksProbe": "structure_jigsaw_blocks_1.21.1.txt",
                     "OceanRuinStartsProbe": "structure_ocean_ruin_starts_1.21.1.txt",
                     "OceanRuinProcessorsProbe": "structure_ocean_ruin_processors_1.21.1.txt",
                     "OceanRuinBlocksProbe": "structure_ocean_ruin_blocks_1.21.1.txt"})
    fixtures["JigsawProcessorsProbe"] = "jigsaw_processors_1.21.1.txt"
    fixtures["JigsawBlockEntitiesProbe"] = "jigsaw_block_entities_1.21.1.txt"
    fixtures["JigsawOverworldBlocksProbe"] = "jigsaw_overworld_blocks_1.21.1.txt"
    fixtures["JigsawOutpostBlocksProbe"] = "jigsaw_outpost_blocks_1.21.1.txt"
    fixtures["MineshaftStartsProbe"] = "structure_mineshaft_starts_1.21.1.txt"
    fixtures["MineshaftBlocksProbe"] = "structure_mineshaft_blocks_1.21.1.txt"
    with zipfile.ZipFile(ROOT / "jar/server.jar") as bundle:
        version = json.loads(bundle.read("version.json"))
        if version["id"] != "1.21.1" or version["world_version"] != 3955:
            raise ValueError("expected Minecraft Java 1.21.1 server JAR")
        with tempfile.TemporaryDirectory(prefix="atlas-structure-ref-") as temp:
            target = Path(temp)
            inner = target / "server.jar"
            inner.write_bytes(bundle.read("META-INF/versions/1.21.1/server-1.21.1.jar"))
            jars = [inner]
            for name in bundle.namelist():
                if name.startswith("META-INF/libraries/") and name.endswith(".jar"):
                    jar = target / f"library-{len(jars)}.jar"
                    jar.write_bytes(bundle.read(name))
                    jars.append(jar)
            classpath = os.pathsep.join(map(str, jars))
            sources = [str(HERE / f"{args.probe}.java")]
            if args.probe in ("StructurePyramidBlocksProbe", "StructureJungleBlocksProbe"):
                sources.append(str(HERE / "StructureHutBlocksProbe.java"))
            if args.probe in ("IglooStartsProbe", "IglooBlocksProbe", "FortressBlocksProbe", "BastionStubProbe", "JigsawStartsProbe") or args.probe.startswith(("BuriedTreasure", "NetherFossil")):
                sources.append(str(HERE / "IglooProbeSupport.java"))
            if args.probe.startswith("RuinedPortal"):
                sources.extend(str(HERE / name) for name in ("IglooProbeSupport.java", "RuinedPortalProbeSupport.java"))
            if args.probe == "JigsawBlocksProbe":
                sources.extend(str(HERE / name) for name in ("IglooProbeSupport.java", "RuinedPortalProbeSupport.java", "JigsawProcessorsProbe.java"))
            if args.probe == "JigsawProcessorsProbe":
                sources.extend(str(HERE / name) for name in ("IglooProbeSupport.java", "RuinedPortalProbeSupport.java", "JigsawBlocksProbe.java"))
            if args.probe == "JigsawBlockEntitiesProbe":
                sources.extend(str(HERE / name) for name in ("IglooProbeSupport.java", "RuinedPortalProbeSupport.java", "JigsawBlocksProbe.java", "JigsawProcessorsProbe.java"))
            if args.probe in ("JigsawOverworldBlocksProbe", "JigsawOutpostBlocksProbe"):
                sources.extend(str(HERE / name) for name in ("IglooProbeSupport.java", "RuinedPortalProbeSupport.java", "JigsawBlocksProbe.java", "JigsawProcessorsProbe.java"))
            if args.probe == "StructureSetSelectionProbe":
                sources.extend(str(HERE / name) for name in ("IglooProbeSupport.java", "RuinedPortalProbeSupport.java", "RuinedPortalStartsProbe.java"))
            if args.probe.startswith("Mineshaft"):
                sources.extend(str(HERE / name) for name in ("IglooProbeSupport.java", "RuinedPortalProbeSupport.java", "RuinedPortalStartsProbe.java"))
            if args.probe == "MineshaftBlocksProbe":
                sources.extend(str(HERE / name) for name in ("RuinedPortalBlocksProbe.java", "OceanRuinProbeSupport.java", "OceanRuinProcessorsProbe.java", "OceanRuinBlocksProbe.java"))
            if args.probe.startswith("Shipwreck"):
                sources.extend(str(HERE / name) for name in ("IglooProbeSupport.java", "RuinedPortalProbeSupport.java", "RuinedPortalStartsProbe.java", "RuinedPortalBlocksProbe.java", "ShipwreckProbeSupport.java"))
            if args.probe.startswith("OceanRuin"):
                sources.extend(str(HERE / name) for name in ("IglooProbeSupport.java", "RuinedPortalProbeSupport.java", "RuinedPortalStartsProbe.java", "OceanRuinProbeSupport.java"))
            if args.probe == "OceanRuinBlocksProbe":
                sources.extend(str(HERE / name) for name in ("RuinedPortalBlocksProbe.java", "OceanRuinProcessorsProbe.java"))
            subprocess.run(["javac", "-proc:none", "-cp", classpath, "-d", temp, *sources], check=True)
            process = subprocess.run(["java", "-cp", os.pathsep.join([temp, classpath]),
                                      f"atlasprobe.{args.probe}", *args.probe_arg], cwd=target, capture_output=True)
            if args.stderr_log:
                atomic_write(args.stderr_log, process.stderr)
            if process.returncode:
                raise SystemExit(process.stderr.decode("utf-8", errors="replace") +
                                 process.stdout.decode("utf-8", errors="replace")[-3000:])
            actual = process.stdout.decode("utf-8").replace("\r\n", "\n")
    fixture = args.output or HERE / fixtures[args.probe]
    if args.probe == "RuinedPortalClassifiersProbe":
        embedded = ROOT / "internal/worldgen" / fixture.name
        value = json.loads(actual)
        if value["version"] != 3955 or len(value["states"]) != 26684:
            raise ValueError("incomplete original ruined portal classifiers")
        if args.update:
            encoded = gzip.compress(json.dumps(value, sort_keys=True, separators=(",", ":")).encode("utf-8"), mtime=0)
            atomic_write(fixture, encoded)
            atomic_write(embedded, encoded)
            print("updated 26684 original portal classifier/blackstone vectors")
        elif json.loads(gzip.decompress(fixture.read_bytes())) != value or embedded.read_bytes() != fixture.read_bytes():
            raise SystemExit(f"ruined portal classifier reference mismatch: {fixture}")
        else:
            print("verified 26684 original portal classifier/blackstone vectors")
        return
    if args.probe == "StructureTransformProbe":
        embedded = ROOT / "internal/worldgen" / fixture.name
        if args.update:
            value = json.loads(actual)
            if value["version"] != 3955 or len(value["states"]) != 26684 or len(value["positions"]) != 336:
                raise ValueError("incomplete original template transforms")
            encoded = gzip.compress(json.dumps(value, sort_keys=True, separators=(",", ":")).encode("utf-8"), mtime=0)
            atomic_write(fixture, encoded)
            atomic_write(embedded, encoded)
            print("updated 320208 original state transforms and 336 positional transforms")
        elif json.loads(gzip.decompress(fixture.read_bytes())) != json.loads(actual) or embedded.read_bytes() != fixture.read_bytes():
            raise SystemExit(f"structure transform reference mismatch: {fixture}")
        else:
            print("verified 320208 original state transforms and 336 positional transforms")
        return
    if args.probe == "JigsawBlockEntitiesProbe":
        embedded = ROOT / "internal/worldgen" / fixture.name
        if args.update:
            atomic_write(embedded, actual.encode("utf-8"))
        elif embedded.read_bytes() != fixture.read_bytes():
            raise SystemExit("embedded Jigsaw block entity normalization differs")
    if args.update:
        atomic_write(fixture, actual.encode("utf-8"))
        print(f"updated {fixture}: {len(actual.splitlines())} vectors")
    elif fixture.read_text(encoding="utf-8") != actual:
        raise SystemExit(f"structure reference mismatch: {fixture}")
    else:
        print(f"verified {fixture}: {len(actual.splitlines())} vectors")


if __name__ == "__main__":
    main()
