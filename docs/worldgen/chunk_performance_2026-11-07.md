# 区块生成与加载性能优化（2026-11-07）

本次在 `d5b2aa2` 的生成版本 19 上优化缓存和数据展开。噪声公式、浮点运算顺序、随机流、地物／结构顺序、生成版本及存档格式保持原样。

## 完整链路测量

环境：Windows amd64、Go 1.27.0、Intel Core i7-11800H、16 个逻辑处理器，生成工作协程为 8。原代码与候选代码依次各运行三轮，表中各指标独立取中位数。桌面后台负载和电源状态没有锁定，这些数字是本机基准结果。

场景为种子 `1234`、蜂巢独立盐 `0`、结构开启，以区块 `(40,-40)` 为中心的 9×9 相邻区块。每轮生成使用全新临时存档，经过生产工作池、Region、光照和 Anvil 保存。加载场景先生成并关闭相同区域，再用新 World 实例加载；包含 Anvil 解压／解码、实体加载及光照安装。计时排除存档创建、Open 和 Close。加载时操作系统文件缓存已被预热，不能解释为冷盘读取延迟，也不包含网络交付或客户端帧率。

| 指标（每次 81 区块） | 原代码 | 优化后 | 减少 |
| --- | ---: | ---: | ---: |
| 新区块生成、照明和保存 | 36.385 s | 19.575 s | 46.2% |
| 生成累计分配字节 | 4.189 GB | 3.740 GB | 10.7% |
| 已保存区块加载 | 941.021 ms | 565.406 ms | 39.9% |
| 加载累计分配字节 | 124.025 MB | 91.965 MB | 25.8% |

GB/MB 为十进制累计分配量，并非常驻内存。原代码生成耗时范围为 35.582–38.768 s，候选为 19.370–23.560 s；加载分别为 905.829–960.286 ms 和 535.219–570.890 ms。首次进程内的懒初始化和并发缓存淘汰会影响分配量。

原代码通过 Go `-overlay` 读取上述提交的五个生产源文件，不临时覆盖工作区。两边使用同一份基准文件；原有未跟踪的性能基准得以保留，并补充了固定蜂巢盐和加载区块数指标。

## 改动与定位依据

初始 CPU 剖析中，基础地形准备占约 75%，密度采样占约 41%，群系气候采样占约 18%。这些百分比存在调用包含关系，不能相加。保存区域的独立加载剖析还显示了光照求解和方块展开的开销。

- 群系采样器复用内置路由中 `flat_cache`／`cache_2d` 的二维密度结果。缓存属于单个采样器，最多 4096 项；不缓存群系选择结果，保留每次调用的上一群系平局偏好。通用直接密度采样器仍保留探针表达式的原有语义。
- 噪声插值角点在小型规则格点中直接索引，避免以“节点＋三维坐标”反复扩展哈希表。超出区块、非规则格点和过大的探针配置走原坐标映射路径；只缓存成功求值的结果。
- 加载区块的光照快照按节／打包字展开，空节和均匀节使用快速填充。方块 SHA-256 按节流式计算，保留相同的小端字节序，减少两个完整区块大小的临时数组。
- 若求解区域中所有已加载列都携带可信光照、且各自仍缺完整邻域，原求解器的发布阶段必然没有输出。此时提前返回，保留修订编号递增；其他情况继续运行原求解器。

组件基准为同一进程内各三轮、每轮十次。垂直群系遍历的中位数从 33.989 ms 降至 13.060 ms；角点缓存从 19.415 ms 降至 16.860 ms，分配字节从 749,584 降至 154,783。组件数字不能单独解释完整链路的全部收益。

## 结果保持与验证范围

- `TestChunkOptimizationOutput` 的 12 个完整输出摘要来自原始 `d5b2aa2`：三个维度、种子 `1234`／`-987654321`、结构开／关。候选全部匹配，覆盖方块／属性／群系数组、有序后处理和计划请求、方块实体、结构及生成实体请求。
- 二维气候缓存与不缓存的直接路径比较原始浮点位、量化气候和群系选择，包含不同高度、负坐标、边界坐标和缓存淘汰；角点缓存与原坐标映射在三个维度、三个种子上比较原始浮点位。
- 光照快速路径与原求解器比较完整更新集合、快照及修订编号，覆盖可信导入、不完整邻域变完整、混合未计算地形、天空／无天空及编辑。打包展开与逐方块读取比较所有值和 SHA-256，覆盖空／均匀节、4–8 位局部调色板及 15 位直接编码。
- 已有主世界噪声、密度和地表原版参考检查通过。顺序／并行完整区块的旧摘要断言在原提交上已过期；先分别运行原代码与候选确认结果相同，再将断言校准到版本 19。
- 基础组 194 项测试、公共工具的 6 项 Python 测试和测试计划工具检查通过。焦点竞态检查覆盖并行主世界／维度生成、可信光照安装、并行预加载、光照恢复、保留不完整光照及卸载重载；三个包均通过。
- 普通世界生成→编辑→保存→重开、区块数据包缓存／返回值所有权、出生区域保存与编辑保留，以及静态／动态原版光照参考检查通过；`go build ./...` 全仓编译通过。

当前缺少 `nether_noise_1.21.1.txt`、`end_noise_1.21.1.txt`、`nether_surface_1.21.1.txt` 和 `end_surface_1.21.1.txt`，相应原版参考测试无法完成。三个维度的旧／新结果对比证明抽样输出保持，不替代独立原版兼容性证明；没有重跑全部 1,000 个随机区块或真人客户端验收。

## 复现

从仓库根目录执行，先将临时目录放在允许原子重命名的工作区内：

```powershell
New-Item -ItemType Directory -Force artifacts/chunk-performance/tmp | Out-Null
$env:TEMP = (Resolve-Path artifacts/chunk-performance/tmp).Path
$env:TMP = $env:TEMP

go test ./internal/world -run '^$' `
  -bench 'Benchmark(GenerateNormalArea|LoadSavedNormalArea)/workers=8$' `
  -benchtime=1x -count=3 -timeout=15m

go test ./internal/worldgen -run '^TestChunkOptimizationOutput$' -count=1
go test ./internal/worldgen -run '^(TestBiome.*Cache.*|TestBiomeVerticalTraversalMatchesDirect|TestDenseCornerCacheMatchesCoordinateMap|TestCornerCacheFallbackRetainsProbeSemantics)$' `
  -bench '^(BenchmarkNoiseCornerCache|BenchmarkBiomeVerticalTraversal)$' -benchtime=10x -count=3

python tools/run_tests.py --suite basic
go build ./...

go test -race ./internal/worldgen ./internal/world ./internal/light `
  -run '^(TestParallelOverworldAreaMatchesPostprocessedOutput|TestDimensionGeneratorConcurrentReplay|TestImportedLightFastPath.*|TestLightSectionExpansionAndHashMatchBlockReads|TestPrepareSpawnUsesParallelWorkersAndCoalesces|TestLightingRestartAndExternalBlockChange|TestIncompleteLightingIsPreservedOnRoundTrip|TestUnloadIdleReleasesLightingAndPacketsAndReloadsEdits)$' `
  -count=1 -timeout=10m
```

本地证据目录：`artifacts/chunk-performance/`。其中 `pipeline-before.txt`／`pipeline-after.txt` 是最终三轮测量，`component-final.txt` 是组件测量，`output-original.txt`／`output-candidate.txt` 是完整输出验证，`original-digests.txt` 记录原提交上的旧断言失败；`basic.txt`、`race.txt` 和 `roundtrip-light-reference.txt` 记录基础组、竞态及保存重载／原版光照检查；`baseline_cpu.pprof`／`baseline_mem.pprof` 和 `load-before.pprof` 用于定位。初始剖析的蜂巢盐尚未固定，正式对比采用后续固定盐的三轮结果。
