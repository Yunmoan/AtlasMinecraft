# 区块生成邻区读取失败修复

调查基线为 `ab744ab`。`run/logs/latest.log` 在种子 `-8981361456740854275` 的实际探索中记录了 52 次区块不可用：暖水海底废墟 20 次，大型滴水石 32 次。它们是可复现的生成失败，不是单纯生成缓慢。

| 请求区块 | 失败路径 | 缺失的读取区块 |
| --- | --- | --- |
| `(1,-32)` | `postprocess minecraft:ocean_ruin_warm in -2,-34` | `(-2,-34)` |
| `(11,-63)`、`(11,-62)`、`(11,-61)` | `feature minecraft:large_dripstone in 12,-62` | `(14,-62)` |

`TestOverworldRuntimeGenerationNeighborhood` 用日志中的种子调用生产 `OverworldGenerator.Generate`，修复前约 5.5 秒内同时复现这两条错误。

## 原因与修复

装饰回放预先准备目标周围的 5×5 基础地形。这能覆盖 3×3 装饰源的一般邻区，却不能保证涵盖大型滴水石的半径、风偏移和高度查询，也不能涵盖完整海底废墟簇的所有结构片段。请求同一失败区块会重新走到同一缺失读取，因此玩家移动时不断看到相同错误。

生成器拥有的 `OverworldFeatureRegion` 现在能按需补充缺失区块。普通地物读取通过已有的基础生成缓存获得噪声、地表和雕刻完成后的真实地形；这些快照继续保持不可变，写入进入本次回放的私有覆盖层。独立地物测试没有配置加载器时，仍然拒绝缺失邻区；加载器的真实错误继续向上传播，返回成功但没有安装区块也会被拒绝。

结构回放需要更早的装饰和结构写入，不能直接补一份裸基础地形。其加载器通过现有 `structureChunkBefore(x,z,step)` 缓存读取正确前置步骤，安装对应的不可变群系数据，并复制覆盖层和更新队列。这个加载器在前置回放完成后才绑定；前置回放自身的普通地物继续使用基础地形读取，避免递归请求同一整段装饰。Jigsaw、林地府邸、要塞等使用同一个前置区域入口，也继承了这项修复。

没有扩大所有区块的固定预生成范围，也没有把缺失地形替换为空气、跳过失败地物或忽略错误。此次不修改已有存档，不提高生成版本；已有 12 组完整区块摘要未变。

## 原版逐方块抽查

用已核验的官方 Java 1.21.1 包新建隔离世界，在冻结游戏后加载以下八个固定区块，确认 FULL、执行 `save-all flush`，然后正常停服。原版包 SHA-1 为 `59353fb40c36d304f2035d51e7d6e6baa98dc05c`，结构开启；Go 蜂巢盐固定为 0，比较 Y=-64～319 的完整方块状态。

| 区块 | 与原版完整状态一致率 |
| --- | ---: |
| `(1,-32)` | 100.000000% |
| `(0,-32)` | 99.692790% |
| `(2,-32)` | 100.000000% |
| `(-2,-34)` | 99.742635% |
| `(11,-63)` | 100.000000% |
| `(11,-62)` | 100.000000% |
| `(11,-61)` | 99.885050% |
| `(12,-62)` | 99.793498% |

合计比较 786,432 个位置，785,561 个完整状态相同，一致率 **99.889247%**；排除双方空气后为 271,626／272,497，**99.680363%**。四块完全相同，其余四块合计 871 个差异。每块双方 u16be 数组及 SHA-256 由现成比较工具保存，Python 独立逐项复算；原版数组再次与最终停止的存档逐项核对。

这是实际失败位置及其邻区的定点诊断集，不代表全世界平均拟合度，也不证明方块实体、实体、光照或更新队列与原版完全相同。旧版本无法成功生成的区块没有可用的旧一致率，不报告虚构的改善百分点。清单、逐块计数和数组哈希保存于 [`testdata/worldgen/runtime_neighborhood_2026-11-07/`](../../testdata/worldgen/runtime_neighborhood_2026-11-07/)。

## 检查与本地证据

- `TestFeatureRegionLazyChunkAccess` 检查负坐标向下换算、失败传播与重试、已加载邻区复用、加载器未安装区块时拒绝访问，以及缓存地形写入隔离。
- `TestStructureRegionLazyChunkRetainsPrecedingStage` 检查远处读取确实看见前置装饰覆盖层，后续写入不修改共享快照。
- `TestNormalWorldLoggedChunkFailuresGenerateAndReload` 经生产 World 队列并发请求四个日志失败区块，检查区块数据包、保存及重新打开后的加载和编辑保留，并检查远处临时读取邻区没有作为完整区块写入存档。
- 206 项基础测试、既有完整输出摘要、并行生成和前置快照并发隔离检查通过。
- 上述日志坐标复现、远邻区读取、前置快照隔离及真实 World 并发保存重开，在 worldgen／world 两个包的竞态检查中通过。海底废墟方块、矿井和 Jigsaw 焦点检查通过；首次扩展检查发现两份本地参考文件缺失，已用官方 JAR 探针补齐并提交。随后 288 个原版海底废墟选址案例（192 接受、36 成簇）和 13,824 个模板处理器／考古／NBT 案例全部通过。`go build ./...` 和差异格式检查通过。

复现／验证命令：

```powershell
go test ./internal/worldgen -run '^TestOverworldRuntimeGenerationNeighborhood$' -count=1 -timeout=5m
go test -race ./internal/worldgen ./internal/world -run '^(TestOverworldRuntimeGenerationNeighborhood|TestFeatureRegionLazyChunkAccess|TestStructureRegionLazyChunkRetainsPrecedingStage|TestStructureChunkSnapshotConcurrentIsolation|TestNormalWorldLoggedChunkFailuresGenerateAndReload)$' -count=1 -timeout=10m
python tools/run_tests.py --suite basic
go build ./...
./build.ps1
```

本次补齐的两份选址／处理器参考来自独立原版执行器，不是 Go 输出生成的期望值。可用以下命令重新核验（需完整 JDK 21，并通过 `JAVA_HOME` 指定）：

```powershell
python tools/worldgen/verify_structure_reference.py --probe OceanRuinStartsProbe
python tools/worldgen/verify_structure_reference.py --probe OceanRuinProcessorsProbe
```

本地证据位于 `artifacts/chunk-generation-fix/`：`runtime-errors.log` 仅保留实际种子和生成错误，`reproduction.txt`／`reproduction-fixed.txt` 为修复前后复现，`focused.txt`／`basic.txt` 为焦点及基础检查，`race-roundtrip.txt` 为两个包的竞态及保存重开检查；`structure-reference.txt` 保留缺失输入的首次失败，`structure-reference-completed.txt` 是补齐后两项检查的通过记录。`native/` 保留独立官方存档、源码哈希、双方数组及汇总，`native-comparison.txt` 为逐块生成日志。实际 `run/world` 未被测试工具修改。
