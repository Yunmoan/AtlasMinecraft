"""Independently recount saved u16be arrays and report full-chunk similarity."""

import argparse
import array
from collections import Counter
import csv
import gzip
import hashlib
import json
from pathlib import Path
import statistics
import sys

from _bootstrap import common


def require(condition, message):
    if not condition:
        raise ValueError(message)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("directory", type=Path)
    args = parser.parse_args()
    directory = args.directory.resolve()
    config = json.loads((directory / "manifest.json").read_text())
    min_y, max_y = config["y"]
    blocks_per_chunk = (max_y - min_y + 1) * 256
    dimension = config["dimension"].removeprefix("minecraft:").removeprefix("the_")
    dimension_label = {"overworld": "主世界", "nether": "下界", "end": "末地"}[
        dimension
    ]

    def pct(value):
        return "—" if value is None else f"{value:.6f}%"

    validation_path = directory / "final-vanilla-validation.log"
    validated_seeds = set()
    if validation_path.exists():
        for line in validation_path.read_text(encoding="utf-8-sig").splitlines():
            words = line.split()
            if (
                len(words) == 4
                and words[0] == "VERIFIED_FINAL_VANILLA"
                and int(words[2]) == config["count_per_seed"]
            ):
                validated_seeds.add(int(words[1]))
    table = json.loads(
        gzip.decompress(
            (common.ROOT / "internal/blockstate/states.json.gz").read_bytes()
        )
    )
    names = [s["name"] for s in table]
    air = {
        i
        for i, n in enumerate(names)
        if n in ("minecraft:air", "minecraft:cave_air", "minecraft:void_air")
    }
    fields = [
        "Total",
        "Equal",
        "NonAirUnion",
        "NonAirEqual",
        "SameBlock",
        "SameOccupancy",
        "SurfaceColumns",
        "SameSurface",
        "SurfaceAbsoluteError",
    ]
    rows = []
    pairs = Counter()
    biomes = Counter()
    skipped = Counter()
    stages = Counter()
    y_diffs = Counter()
    all_examples = []
    for case in config["cases"]:
        base = directory / str(case["seed"]) / "chunks" / f'{case["x"]}_{case["z"]}'
        if not base.with_suffix(".json").exists():
            continue
        row = json.loads(base.with_suffix(".json").read_text())
        require(all(row[k] == case[k] for k in case), f"Case metadata mismatch: {base}")
        if row.get("Dimension"):
            require(
                row["Dimension"] == dimension
                and row["MinY"] == min_y
                and row["Height"] == max_y - min_y + 1,
                f"Dimension bounds mismatch: {base}",
            )
        arrays = []
        for label, digest in [("vanilla", "VanillaSHA256"), ("go", "GoSHA256")]:
            data = base.with_suffix(f".{label}.u16be").read_bytes()
            require(
                len(data) == blocks_per_chunk * 2, f"Truncated array: {base} ({label})"
            )
            actual = hashlib.sha256(data).hexdigest()
            require(actual == row.get(digest), f"SHA-256 mismatch: {base} ({label})")
            a = array.array("H")
            a.frombytes(data)
            if sys.byteorder == "little":
                a.byteswap()
            arrays.append(a)
        # Recompute independently from the persisted arrays, including air union.
        equal = union = nonair_equal = sameblock = occupancy = 0
        vanilla_nonair = go_nonair = 0
        local_pairs = Counter()
        for index, (a, b) in enumerate(zip(*arrays)):
            equal += a == b
            vanilla_nonair += a not in air
            go_nonair += b not in air
            if a not in air or b not in air:
                union += 1
                nonair_equal += a == b
            sameblock += names[a] == names[b]
            occupancy += (a in air) == (b in air)
            if a != b:
                local_pairs[f"{names[a]} -> {names[b]}"] += 1
                y_diffs[index // 256 + min_y] += 1
        require(
            [blocks_per_chunk, equal, union, nonair_equal, sameblock, occupancy]
            == [row[k] for k in fields[:6]],
            f"Saved counts differ from array recount: {base}",
        )
        require(
            dict(local_pairs) == row["Pairs"], f"Saved difference pairs differ: {base}"
        )
        pairs.update(local_pairs)
        biomes.update(row["Biomes"])
        skipped.update(row.get("SkippedFeatures") or [])
        stages.update(row.get("SkippedStages") or [])
        if len(all_examples) < 30:
            all_examples.extend(
                dict(seed=case["seed"], chunk=[case["x"], case["z"]], **e)
                for e in row.get("Examples") or []
            )
        row["exact_percent"] = equal / blocks_per_chunk * 100
        row["nonair_percent"] = nonair_equal / union * 100 if union else None
        row["vanilla_nonair"] = vanilla_nonair
        row["go_nonair"] = go_nonair
        rows.append(row)

    def summary(group):
        totals = {k: sum(row[k] for row in group) for k in fields}
        if not group:
            return dict(chunks=0)
        totals.update(
            chunks=len(group),
            exact_percent=100 * totals["Equal"] / totals["Total"],
            nonair_percent=(
                100 * totals["NonAirEqual"] / totals["NonAirUnion"]
                if totals["NonAirUnion"]
                else None
            ),
            block_type_percent=100 * totals["SameBlock"] / totals["Total"],
            occupancy_percent=100 * totals["SameOccupancy"] / totals["Total"],
            surface_height_equal_percent=100
            * totals["SameSurface"]
            / totals["SurfaceColumns"],
            surface_height_mae=totals["SurfaceAbsoluteError"]
            / totals["SurfaceColumns"],
            perfect_chunks=sum(row["Equal"] == row["Total"] for row in group),
            both_empty_chunks=sum(row["NonAirUnion"] == 0 for row in group),
            nonempty_chunks=sum(row["NonAirUnion"] > 0 for row in group),
            vanilla_nonair=sum(row["vanilla_nonair"] for row in group),
            go_nonair=sum(row["go_nonair"] for row in group),
            chunk_min_percent=min(row["exact_percent"] for row in group),
            chunk_median_percent=statistics.median(
                row["exact_percent"] for row in group
            ),
            chunk_max_percent=max(row["exact_percent"] for row in group),
        )
        return totals

    shutdowns = {}
    for seed in config["seeds"]:
        parent = directory / str(seed) / "vanilla"
        for filename in ("shutdown-audit.json", "complete.json"):
            if (parent / filename).exists():
                shutdowns[str(seed)] = json.loads((parent / filename).read_text())
                break
    validation_complete = set(config["seeds"]) <= validated_seeds
    result = {
        "dimension": dimension,
        "planned_chunks": len(config["cases"]),
        "complete": len(rows) == len(config["cases"]) and validation_complete,
        "final_vanilla_validation_complete": validation_complete,
        "overall": summary(rows),
        "seeds": {
            str(s): summary([r for r in rows if r["seed"] == s])
            for s in config["seeds"]
        },
        "top_differences_vanilla_to_go": pairs.most_common(30),
        "differences_by_y": sorted(y_diffs.items()),
        "go_quart_biome_counts": biomes,
        "skipped_feature_occurrences": skipped,
        "skipped_stage_occurrences": stages,
        "sample_quadrants": dict(
            Counter(
                ("+" if c["x"] >= 0 else "-") + ("+" if c["z"] >= 0 else "-")
                for c in config["cases"]
            )
        ),
        "official_shutdown_audit": shutdowns,
        "worst_chunks": [
            {k: r[k] for k in ("seed", "x", "z", "exact_percent", "nonair_percent")}
            for r in sorted(rows, key=lambda r: r["Equal"])[:15]
        ],
        "examples": all_examples[:30],
    }
    common.atomic_write(
        directory / "summary.json", json.dumps(result, indent=2).encode()
    )
    with (directory / "chunks.csv").open("w", newline="", encoding="utf-8-sig") as f:
        keys = [
            "seed",
            "x",
            "z",
            "exact_percent",
            "nonair_percent",
            "vanilla_nonair",
            "go_nonair",
            *fields,
            "Seconds",
        ]
        writer = csv.DictWriter(f, fieldnames=keys, extrasaction="ignore")
        writer.writeheader()
        writer.writerows(rows)
    out = [
        "# 随机区块逐方块对比实测",
        "",
        f'- 基线：Go `{config["head"]}`，生成版本 19；官方 Minecraft Java 1.21.1。',
        f"- 范围：{dimension_label}，结构开启；Y={min_y}～{max_y}，每块 {blocks_per_chunk:,} 个位置。",
        f'- 抽样：每种子 {config["count_per_seed"]} 块，随机抽样器种子 {config["sampling_seed"]}；区块 X/Z 均匀抽取自 {config["chunk_coordinate_bounds"]}。',
        f'- 已完成 {len(rows)}/{len(config["cases"])} 块；所有已计入数组均经过独立逐项复算及 SHA-256 检查。',
        "",
        "| 世界种子 | 区块数 | 完整状态一致率 | 排除双方空气后一致率 | 全部吻合区块 |",
        "|---|---:|---:|---:|---:|",
    ]
    for seed, s in result["seeds"].items():
        if s["chunks"]:
            out.append(
                f'| {seed} | {s["chunks"]} | {s["exact_percent"]:.6f}% | {pct(s["nonair_percent"])} | {s["perfect_chunks"]} |'
            )
    total = result["overall"]
    if rows:
        out += [
            f'| 总计 | {total["chunks"]} | {total["exact_percent"]:.6f}% | {pct(total["nonair_percent"])} | {total["perfect_chunks"]} |',
            "",
            f'完整状态一致 {total["Equal"]:,}/{total["Total"]:,}；差异 {total["Total"]-total["Equal"]:,}。',
            f'双方全空区块 {total["both_empty_chunks"]} 个；至少一端含非空气方块的区块 {total["nonempty_chunks"]} 个。全空样本的非空气一致率记为空值，不计成 100%。',
            f'方块种类一致率 {total["block_type_percent"]:.6f}%；空气/非空气占据一致率 {total["occupancy_percent"]:.6f}%。',
            f'最高非空气方块 Y 相同的列占 {total["surface_height_equal_percent"]:.6f}%，平均绝对高度误差 {total["surface_height_mae"]:.6f} 格（包含水面、树冠、建筑）。',
            f'单区块一致率最小/中位/最大：{total["chunk_min_percent"]:.6f}% / {total["chunk_median_percent"]:.6f}% / {total["chunk_max_percent"]:.6f}%。',
            "",
            "## 主要差异（原版 → Go）",
            "",
            "| 方块种类 | 差异位置数 |",
            "|---|---:|",
        ]
        out.extend(f"| {pair} | {count:,} |" for pair, count in pairs.most_common(15))
        out += ["", "同名 → 同名表示方块种类相同、属性不同，仍计入完整方块状态差异。"]
    out += [
        "",
        "## 统计口径与限制",
        "",
        "完整状态比较包含方块属性；排除双方空气的分母为至少一端是非空气的位置，空气包括 air/cave_air/void_air。种类一致率仅比较方块名称。",
        f'原版服务端完成 FULL 区块后保存，抽样加载前关闭随机刻并冻结游戏；Go 使用生产 {"OverworldGenerator" if dimension=="overworld" else "DimensionGenerator"} 的完整装饰和结构路径，无额外运行刻。',
        (
            "Go 固定 BeehiveShuffleSalt=0；原版部分蜂巢方向存在独立随机性，未为提高吻合率调整输出。"
            if dimension == "overworld"
            else "使用维度的完整建造高度进行比较；非空气指标可避免大量上层空气或末地虚空抬高一致率。"
        ),
        f"此处衡量新生成{dimension_label}方块状态，不比较方块实体 NBT、实体、光照、生物群系等价性或运行若干刻后的状态；不代表整个维度的穷举结果或其他维度。",
        "均匀随机空间样本的结果不等于每种群系/结构分别均衡抽样的结果。manifest.json 保存全部坐标，chunks.csv 保存每区块统计，种子目录保留原版存档和双方 u16be 方块数组。",
    ]
    if dimension == "end":
        out += [
            "末地采用全范围均匀抽样，未针对中央主岛或末地城定点加样；非空气一致率仅统计至少一端含方块的位置。"
        ]
    if (directory / "repeat-oracle-audit.json").exists():
        audit = json.loads((directory / "repeat-oracle-audit.json").read_text())
        changes = "、".join(
            f'{case["vanilla"]["different_positions"]:,}' for case in audit["cases"]
        )
        unchanged = all(
            case["go"]["different_positions"] == 0 for case in audit["cases"]
        )
        out += [
            f'重复采样记录：{len(audit["cases"])} 个下界区块在预跑与正式跑中，原版数组分别相差 {changes} 个位置；Go 数组逐项相同：{unchanged}。本报告以正式跑存档为准；尚未定位原版两次输出变化的原因。具体坐标、哈希与差异类型见 repeat-oracle-audit.json。'
        ]
    forced = [
        seed
        for seed, audit in shutdowns.items()
        if audit.get("shutdown", "clean") != "clean"
    ]
    if forced:
        verification = (
            "最终存档也已与统计使用的原版数组逐项核对一致"
            if validation_complete
            else "最终存档还需与统计使用的原版数组逐项核对"
        )
        out += [
            "",
            f'原版停服记录：种子 {", ".join(forced)} 在确认 save-all flush 成功后，stop 收尾超时，由测试工具结束进程。每种子保存的 {config["count_per_seed"]} 个抽样区块均独立验证为 FULL，并完整解码；{verification}，验证记录见 final-vanilla-validation.log。该问题影响停服收尾，不能据此证明正常停服行为。',
        ]
    common.atomic_write(directory / "report.md", "\n".join(out).encode("utf-8"))
    print(
        json.dumps(
            {
                "complete": result["complete"],
                "overall": total,
                "seeds": result["seeds"],
            },
            indent=2,
        )
    )


if __name__ == "__main__":
    main()
