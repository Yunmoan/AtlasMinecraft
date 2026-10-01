"""Regenerate and verify noise vectors against jar/server.jar (JDK 21 required).

The Java classes run only during verification. Atlas never loads the JAR at
runtime. Pass --update after deliberately changing reference probe inputs.
"""
import argparse
import gzip
import json
import os
import shutil
import subprocess
import tempfile
import zipfile
from pathlib import Path

HERE = Path(__file__).resolve().parent
ROOT = HERE.parents[1]
PROBES = {
    "ReferenceProbe": "xoroshiro_1.21.1.txt",
    "ImprovedNoiseProbe": "improved_noise_1.21.1.txt",
    "PerlinNoiseProbe": "perlin_noise_1.21.1.txt",
    "NormalNoiseProbe": "normal_noise_1.21.1.txt",
    "NamedNoiseProbe": "named_noise_1.21.1.txt",
    "LegacyRandomProbe": "legacy_random_1.21.1.txt",
    "DensityClimateProbe": "density_climate_1.21.1.txt",
    "BlendedNoiseProbe": "blended_noise_1.21.1.txt",
    "WeirdScaledProbe": "weird_scaled_1.21.1.txt",
    "SplineProbe": "spline_1.21.1.txt",
    "BiomePresetProbe": "../../internal/worldgen/biome_preset_1.21.1.txt",
    "NoiseChunkProbe": "noise_chunk_1.21.1.txt",
    "SurfaceRuleProbe": "surface_rules_1.21.1.txt",
    "BiomeTemperatureProbe": "biome_temperature_1.21.1.txt",
    "BiomeZoomProbe": "biome_zoom_1.21.1.txt",
    "SurfaceChunkProbe": "surface_chunk_1.21.1.txt",
    "SurfaceNoiseProbe": "surface_noise_1.21.1.txt",
    "CarverStartProbe": "carver_starts_1.21.1.txt",
    "CarverChunkProbe": "carver_chunk_1.21.1.txt",
    "FeatureOrderProbe": "../../internal/worldgen/feature_order_1.21.1.txt",
    "FeatureSeedProbe": "feature_seed_1.21.1.txt",
    "FeaturePlacementProbe": "feature_placement_1.21.1.txt",
    "FeatureStreamProbe": "feature_stream_1.21.1.txt",
    "OreFeatureProbe": "ore_feature_1.21.1.txt",
    "OreHeightProbe": "ore_height_1.21.1.txt",
    "ForestRockProbe": "forest_rock_1.21.1.txt",
    "MotionBlockingProbe": "motion_blocking_1.21.1.txt",
    "BlueIceProbe": "blue_ice_1.21.1.txt",
    "DiskFeatureProbe": "disk_feature_1.21.1.txt",
    "RuleDiskProbe": "rule_disk_1.21.1.txt",
    "LightEngineProbe": "light_engine_1.21.1.txt",
    "LightRulesProbe": "../../internal/light/rules_1.21.1.json.gz",
    "TreeFeatureProbe": "tree_feature_1.21.1.txt",
    "TreePlacementProbe": "tree_placement_1.21.1.txt",
    "VegetationProbe": "vegetation_1.21.1.txt",
    "BlockRulesProbe": "../../internal/blockrules/rules_1.21.1.json.gz",
    "FrozenSpringProbe": "frozen_spring_1.21.1.txt",
    "CaveFeatureProbe": "cave_features_1.21.1.txt",
    "CavePlacementProbe": "cave_placement_1.21.1.txt",
    "MultifaceSupportProbe": "../../internal/worldgen/multiface_support_1.21.1.txt",
}


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("--update", action="store_true")
    parser.add_argument("--probe", choices=PROBES)
    parser.add_argument("--vegetation-set", choices=("base", "columns"), default="base")
    args = parser.parse_args()
    probes = {args.probe: PROBES[args.probe]} if args.probe else PROBES
    javac, java = shutil.which("javac"), shutil.which("java")
    if not javac or not java:
        raise SystemExit("JDK 21 java and javac must be on PATH")

    with tempfile.TemporaryDirectory(prefix="atlas-worldgen-ref-") as temp:
        temporary = Path(temp)
        with zipfile.ZipFile(ROOT / "jar" / "server.jar") as bundle:
            version = json.loads(bundle.read("version.json"))
            if version["id"] != "1.21.1" or version["world_version"] != 3955:
                raise ValueError("expected Minecraft Java 1.21.1 server JAR")
            inner = temporary / "server-1.21.1.jar"
            inner.write_bytes(bundle.read("META-INF/versions/1.21.1/server-1.21.1.jar"))
            libraries = []
            for member in bundle.namelist():
                if member.startswith("META-INF/libraries/") and member.endswith(".jar"):
                    path = temporary / f"library-{len(libraries)}.jar"
                    path.write_bytes(bundle.read(member))
                    libraries.append(path)
        classpath = os.pathsep.join(map(str, [inner, *libraries]))
        sources = [str(HERE / f"{name}.java") for name in probes]
        if args.probe == "CarverChunkProbe":
            sources.append(str(HERE / "SurfaceChunkProbe.java"))
        if args.probe == "FeatureStreamProbe":
            sources.append(str(HERE / "FeaturePlacementProbe.java"))
        if args.probe in ("TreePlacementProbe", "VegetationProbe", "FrozenSpringProbe", "CaveFeatureProbe", "CavePlacementProbe"):
            sources.append(str(HERE / "TreeFeatureProbe.java"))
        if args.probe in ("CaveFeatureProbe", "CavePlacementProbe"):
            sources.append(str(HERE / "FrozenSpringProbe.java"))
        subprocess.run([javac, "-proc:none", "-cp", classpath, "-d", temp, *sources], check=True)
        for name, fixture_name in probes.items():
            command = [java, "-cp", os.pathsep.join([temp, classpath]), f"atlasprobe.{name}"]
            if name in ("NamedNoiseProbe", "DensityClimateProbe", "BlendedNoiseProbe", "WeirdScaledProbe"):
                command.append(str(inner))
            if name in ("SurfaceChunkProbe", "CarverChunkProbe"):
                command.append(str(HERE / "noise_chunk_1.21.1.txt"))
            if name == "CarverChunkProbe":
                command.append(str(inner))
            if name == "OreHeightProbe":
                command.append(str(HERE / "carver_chunk_1.21.1.txt"))
            if name == "MotionBlockingProbe":
                command.append(str(HERE / "carver_chunk_1.21.1.txt"))
            if name == "VegetationProbe":
                if args.vegetation_set == "columns":
                    command.extend([str(HERE / "column_vegetation_features_1.21.1.txt"), "only"])
                    fixture_name = "column_vegetation_1.21.1.txt"
                else:
                    command.append(str(HERE / "vegetation_features_1.21.1.txt"))
            process = subprocess.run(command, capture_output=True, cwd=temporary)
            if process.returncode:
                raise SystemExit(f"{name} failed:\n{process.stderr.decode('utf-8', errors='replace')}\n{process.stdout.decode('utf-8', errors='replace')[-4000:]}")
            actual = process.stdout.decode("utf-8").replace("\r\n", "\n")
            fixture = HERE / fixture_name
            if name in ("BlockRulesProbe", "LightRulesProbe"):
                if args.update:
                    fixture.write_bytes(gzip.compress(actual.encode("utf-8"), mtime=0))
                    print(f"updated {fixture_name}")
                elif json.loads(actual) != json.loads(gzip.decompress(fixture.read_bytes()).decode("utf-8")):
                    raise SystemExit(f"reference mismatch: {fixture_name}")
                else:
                    print(f"verified {fixture_name}")
                continue
            if args.update:
                fixture.write_text(actual, encoding="utf-8", newline="\n")
                print(f"updated {fixture_name}")
            elif actual != fixture.read_text(encoding="utf-8"):
                raise SystemExit(f"reference mismatch: {fixture_name}")
            else:
                print(f"verified {fixture_name}")


if __name__ == "__main__":
    main()
