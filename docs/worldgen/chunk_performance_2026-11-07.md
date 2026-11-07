# 区块生成与加载性能优化（2026-11-07）

第一轮（提交 `6f40b0e`）在 `d5b2aa2` 的生成版本 19 上优化缓存和数据展开。第二轮的进一步优化见文末。两轮均保持噪声公式、浮点运算顺序、随机流、地物／结构顺序、生成版本及存档格式。

## 第一轮完整链路测量

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
New-Item -ItemType Directory -Force artifacts/chunk-performance/.tmp | Out-Null
$env:TEMP = (Resolve-Path artifacts/chunk-performance/.tmp).Path
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

## 第二轮：密度图执行、群系角点与光照种子

以已提交的 `6f40b0e` 为本轮基线，继续使用同一种子、独立盐、结构开关、81 区块区域和 8 个工作协程。两边重新顺序执行三轮，独立取中位数。机器状态未锁定，两轮绝对耗时存在波动，因此本轮只和重新测量的 `6f40b0e` 比较，不将上一轮的绝对值当作本轮基线。

| 指标（每次 81 区块） | 本轮基线 `6f40b0e` | 最终代码 | 变化 |
| --- | ---: | ---: | ---: |
| 新区块生成、照明和保存 | 27.895 s | 22.260 s | 耗时减少 20.2% |
| 生成累计分配字节 | 3.815 GB | 3.899 GB | 增加 2.2% |
| 已保存区块加载 | 605.331 ms | 356.075 ms | 耗时减少 41.2% |
| 加载累计分配字节 | 91.994 MB | 92.242 MB | 增加 0.3% |

最终代码生成耗时范围为 21.663–23.192 s，加载为 352.218–370.870 ms；基线分别为 27.272–30.996 s 和 582.629–631.928 ms。局部缓存以略增分配量换取速度，常驻内存未另行测量。修复无效输入处理后重新完成了三轮最终测量；较早候选的加载中位数为 307.014 ms，不能将其中最快值替代上表。并发缓存淘汰和首次懒初始化也会影响累计分配量。

本轮将原本每个方块都要做的重复查找和扫描移至解析或状态变化时：

- 解析密度函数时生成整数操作码，保持各分支的求值顺序和数学表达式。每份设置图有独立的槽位命名空间，采样器的私有数组保存噪声对象、当前插值格和角点缓存。来自另一设置图的节点及手写探针继续走通用映射，避免槽位碰撞。正的二次幂格宽使用位运算向下对齐，保留负坐标的整除行为。
- 地表处理每区块只计算一次群系缩放种子哈希，按固定的 6×6×高度窗口复用角点偏移。距离平方、角点顺序和严格小于的平局判定保持原式；范围外继续直接计算。主世界每次地表处理增加 94,208 字节、两次分配的短期缓存，随该次处理释放，不进入地形常驻缓存。
- 光照列懒计算直射光边界及发光方块索引。改变方块规则的编辑使种子元数据失效；相同光照规则的状态替换继续复用。求解器按直射光的相邻区间生成边缘队列，并直接访问稀疏发光源，省掉每次 3×3 重算的全高度扫描。天空上下填充层和六向遮挡判定保持原规则。代价是每个已加载列增加约 2 KiB 的直射光边界及按发光源数量增长的索引。

组件测量（三轮中位数）显示：密度槽位相对同一执行器的映射路径从 22.434 ms 降至 17.710 ms；地表群系缩放遍历从 26.175 ms 降至 7.691 ms。81 列平坦光照安装从 1,176.000 ms 降至 627.250 ms，携带可信导入光照时从 605.149 ms 降至 365.019 ms。这些是组件场景，不是客户端帧率或完整区块吞吐量。

结果保持检查继续使用第一轮的 12 组完整区块摘要，未更新任何期望值。新增设置图槽位／映射的浮点位对照、跨图及手写节点隔离、极值和负坐标格点对齐、三种子群系角点窗口／边界对照；现有群系缩放、密度和地表原版参考检查通过。光照新增四组基于 `6f40b0e` 的固定重放摘要，覆盖 116 个逐步状态，每一步记录所有已加载列的光照、完整性标志和修订编号；同时覆盖不同高度、天空开关、导入光照、阶梯／流体遮挡、光源编辑和卸载重装。原版静态／动态光照检查也通过。

第二轮基础组 198 项及工具检查通过，另补测缺失插值参数时保持返回 `nil density node`，共验证 199 项基础用例。焦点竞态检查在 worldgen／world／light 三个包通过，包含新缓存、并行生成、光照重放、加载与卸载。实际普通世界生成→编辑→保存→重开、区块数据包缓存／所有权及跨区域方块／光照包顺序检查通过；最终 `go build ./...` 通过。临时目录采用 `.tmp` 名称，使 Go 的递归包扫描忽略并行检查产生的编译源码。

本轮证据目录为 `artifacts/chunk-performance-aggressive/`：`pipeline-before.txt`／`pipeline-after.txt` 是最终对比，`pipeline-after-pre-guard.txt` 保留之前的候选测量；组件记录为 `density-component.txt`、`zoom-component.txt`、`light-component-before.txt`／`light-component-after.txt`；正确性记录为 `density-validation.txt`、`zoom-terrain-validation.txt`、`light-validation.txt`、`light-replay-before.txt`、`basic.txt`、`race.txt` 和 `roundtrip-packets.txt`。原代码对照仍使用 `.source` 文件形式的 Go overlay；不会将测量快照引入 Go 包扫描。
