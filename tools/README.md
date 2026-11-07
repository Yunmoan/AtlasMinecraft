# 开发工具

所有示例命令从仓库根目录执行。普通服务端运行不需要这些工具；它们用于生成内嵌数据、验证原版兼容性或诊断问题。

| 路径 | 内容 |
| --- | --- |
| `protocol/*.py` | 协议、注册表、物品数据的生成和验证 |
| `protocol/probes/*.java` | 协议与物品的官方 JAR 参考探针 |
| `worldgen/build_*.py` | 世界生成数据、模板和参考结果的构建工具 |
| `worldgen/verify_*.py` | 对照原版参考结果进行验证 |
| `worldgen/diagnose_terrain_stages.py` | 指定种子、区块的地形阶段诊断 |
| `worldgen/compare_random_worlds.py` | 使用 5 个种子，在完整主世界坐标范围内随机抽样，运行冻结的官方服务端并逐方块比较 Go 生成结果；保留独立存档、坐标和完整数组 |
| `worldgen/summarize_random_comparison.py` | 独立复算双方数组，输出完整状态／非空气一致率、差异分布、每区块 CSV 和 Markdown 报告 |
| `worldgen/probes/*.java` | 世界生成和游戏交互的参考探针及辅助类 |
| `worldgen/extract_native_bubble_columns.go` | 从原版世界提取气泡柱坐标 |
| `worldgen/build_scheduled_tick_probe.py` | 对照原版计划更新队列与 NBT 加载／保存；`--check` 校验固定参考文件 |
| `worldgen/build_fluid_runtime_fixture.py` | 在原版服务端冻结／逐刻运行 28 个流体、植物、树叶、气泡柱及下落方块场景；提取完整方块、方块实体、下落实体和两套有序计划队列，支持 `--check` 独立重采样 |
| `worldgen/build_falling_block_probe.py` | 对照原版下落方块回调和完整实体 NBT；`--motion` 校验逐刻运动、碰撞和落地；`--drops` 校验掉落判定、物品 NBT 与两条随机流；`--brushable` 切换到可疑沙／沙砾及刷取重置参考；`--kelp` 捕获海带延迟生存检查和原版物品生成，`--check` 独立重采样 |
| `worldgen/build_fluid_container_probe.py` | 穷举原版容器状态／流体组合，校验接受结果、更新顺序、随机数和熄火音效包 |
| `worldgen/build_block_shape_probe.py` | 穷举所有原版方块状态的六向形状回调；`--plants` 增补植物支撑、上下半部和亮度上下文；`--check` 独立重采样 |
| `worldgen/build_block_tick_probe.py` | 提取原版全部方块的计划更新方法来源、全状态叶距查询和树叶延迟更新参考结果 |
| `worldgen/build_tree_tick_probe.py` | 对照 2,824 个原版树生成案例，检查边界更新请求顺序和 ProtoChunkTicks 打包结果；支持 `--check` |
| `worldgen/build_template_tick_probe.py` | 对照原版模板边缘及最终六向更新，覆盖 386 状态（含考古方块、海带、海草）、12 周边环境和跨区块／空腔形状；逐项比较方块、回调、随机数后续值和原版 ProtoChunkTicks 打包，支持 `--check` |
| `worldgen/build_lake_tick_probe.py` | 对照 512 个原版熔岩湖案例及洞穴空气计划请求，包含越过建筑高度的请求；支持 `--check` |
| `worldgen/build_bubble_tick_probe.py` | 对照气泡柱、岩浆块、灵魂沙的延迟更新／放置回调和全部状态的生成后处理标记；支持 `--check` |
| `worldgen/build_postprocessing_probe.py` | 原版噪声填充、雕刻和 21 类洞穴地物的后处理列表，逐项保留顺序和重复位置；支持 `--check` |
| `worldgen/build_final_ocean_fixture.py` | 冻结原版世界并推进 100 刻，保存前后完整区块和初始方块／流体队列；`--preset ocean-ruins` 提取冷／暖海遗迹，`--preset shipwrecks` 提取海底／搁浅沉船，均包含完整方块实体；`--check` 独立重采样，`--trace-decoration` 观察原版装饰、激活及保存入口的内存队列，并在结构模式发布完整有序调度记录 |
| `inspect-save/`、`terrain-scan/` | 独立 Go 存档检查和地形扫描工具；以命令参数为准 |

## 常用命令

测试统一入口为仓库根目录的 `test.ps1`，也可直接执行 `python tools/run_tests.py`：

| 测试组 | 命令 | 要求 |
| --- | --- | --- |
| 基础 | `./test.ps1 -Suite basic` | Go、Python；运行公共工具测试和无需额外参考集的 Go 测试 |
| 世界集成 | `./test.ps1 -Suite integration` | Go；实际创建、加载或生成世界，耗时较长 |
| 参考 | `./test.ps1 -Suite reference` | Go、对应参考文件；部分互操作场景要求额外环境变量 |
| 协议冒烟 | `./test.ps1 -Suite smoke` | 自动构建服务端与模拟客户端，使用独立临时世界 |
| 性能场景 | `./test.ps1 -Suite performance` | 有界的独立进程性能/协议场景；不自动执行所有微基准 |
| 全部 | `./test.ps1 -Suite all` | 上述全部输入和环境 |

使用 `-List` 查看测试与缺失输入，`-Match 名称片段` 缩小范围。例如 `./test.ps1 -Suite smoke -Match SpawnRadius`。`-Timeout 45m` 可指定每个 Go 包的时间限制；基础组默认为 10 分钟，其余为 30 分钟。Go AST 计划工具会沿同包辅助函数调用收集静态文件路径，并识别世界创建、加载和生成调用；动态计算的路径仍由实际测试验证。测试分组来自当前源码，无需维护易失效的测试名称清单。

所有测试产物集中到 `artifacts/test-*`，临时目录位于 `artifacts/.tmp/tests`。完整参考集没有随仓库全部提供，缺失时参考组会在测试启动前报告所需文件，退出码为 2；这与实现测试失败分开处理。

```powershell
# 检查 Go；参考工具还会检查 JAVA_HOME/PATH 中的 JDK 和官方包 SHA-1
python tools/check_env.py
python tools/check_env.py --reference

# 从已有协议源数据生成注册表、物品表
python tools/protocol/build_registries.py
python tools/protocol/build_items.py

# 验证一个随机数参考探针
python tools/worldgen/verify_reference.py --probe ReferenceProbe

# 5 个种子，每种子 100 个完整区块（较耗时）；中断后原命令可继续
python tools/worldgen/compare_random_worlds.py --count 100
python tools/worldgen/summarize_random_comparison.py artifacts/random-world-comparison-2026-11-07

# 下界／末地：沿用 5 个种子，每种子每维度 50 个区块，完整高度 0..255
python tools/worldgen/compare_random_worlds.py --dimension nether --count 50
python tools/worldgen/compare_random_worlds.py --dimension end --count 50
python tools/worldgen/summarize_random_comparison.py artifacts/random-nether-comparison-2026-11-07
python tools/worldgen/summarize_random_comparison.py artifacts/random-end-comparison-2026-11-07

# 验证结构选址参考结果
python tools/worldgen/verify_structure_reference.py --probe StructurePlacementProbe

# 比较抓包中的注册表
python tools/protocol/diff_registries.py testdata/protocol/captures/atlas
```

Python 工具使用标准库。调用官方参考代码的工具还需要 JDK 21+ 的 `java`、`javac`，以及 `jar/server.jar`（Minecraft Java 1.21.1）。优先使用 `JAVA_HOME` 指向的完整 JDK，否则查找 `PATH`。公共工具会验证官方包的 SHA-1、版本和 DataVersion，并将探针使用的 JAR 准备到 `artifacts/reference-java`。Go 提取工具需要 Go。

`atlas_tools.py` 集中管理环境查找、参考包解压、classpath 和原子文件写入；两个主题目录的 `_bootstrap.py` 只负责直接执行脚本时的导入路径。公共工具测试：`python -m unittest discover -s tools -p "test_*.py"`。

参考输入和结果位于 [`testdata`](../testdata/README.md)，运行时内嵌资源位于 `internal`，临时日志和编译结果位于 `artifacts`。部分参考集需先运行对应 `build_*` 工具生成；目录整理不会补充尚未生成的测试数据。

验证脚本的 `--update` 会更新参考数据，仅在有意重新生成基准时使用。更详细的算法说明见[文档导航](../docs/README.md)。

随机区块比较默认写入当天的 `artifacts/random-{world|nether|end}-comparison-YYYY-MM-DD`，可用 `--output` 固定目录。继续运行会检查生成代码与记录提交是否一致，并核对当前原版数组及双方已存数组的 SHA-256；参考变化或数组损坏会重新比较对应区块。代码基线不符时应提交代码并指定新目录。汇总工具再次逐项复算；双方全空区块的排空指标为空值。

本次 5 种子、三个维度的结果和验证限制见[随机区块对比简报](../docs/worldgen/random_chunk_comparison_2026-11-07.md)，可公开的坐标、计数和哈希位于 `testdata/worldgen/random_comparison_2026-11-07/`。
