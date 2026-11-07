# 对照原版 1.21.1 的区块生成性能机制

本笔记以 Atlas `44be121` 为调查起点，针对“继续提速，同时保持当前生成结果”梳理原版可借鉴的实现。结论来自固定的 Mojang/Microsoft 官方服务端 JAR 的字节码，官方映射只用于恢复名称。没有把后续版本 Wiki 的现行算法直接套到 1.21.1；也没有将第三方反编译仓库当作实现证据。

一个需要优先量化的结构性差距是工作量组织：原版维护共享区块及逐阶段 future，Atlas 为保持目标区块的确定性而重放邻区装饰。插值、缓存和短路可以减少单次工作的成本，但不能消除这项结构性重复。另一项差距是保存口径：Atlas 的完整链路基准包括同步持久化，原版的常规保存提交异步写入，不能直接用二者的可见区块生成速度推算倍数。

## 版本、来源与可复核证据

- [官方版本清单](https://piston-meta.mojang.com/mc/game/version_manifest_v2.json) 中的 `1.21.1` 指向[该版本元数据](https://piston-meta.mojang.com/v1/packages/82ba6501da174a665c1787775f946487422f2432/1.21.1.json)。下载 URL、长度和 SHA-1 均从此元数据读取。
- [官方 1.21.1 服务端包](https://piston-data.mojang.com/v1/objects/59353fb40c36d304f2035d51e7d6e6baa98dc05c/server.jar)：51,627,615 字节，SHA-1 `59353fb40c36d304f2035d51e7d6e6baa98dc05c`。本地 `jar/server.jar` 的原始字节 SHA-1 实测匹配。
- [官方服务端名称映射](https://piston-data.mojang.com/v1/objects/03f8985492bda0afc0898465341eb0acef35f570/server.txt)：SHA-1 `03f8985492bda0afc0898465341eb0acef35f570`。本地 `artifacts/worldgen-probes/server-mappings-1.21.1.txt` 使用 CRLF；还原 LF 后的 SHA-1 与官方文件匹配。
- 内层服务端为 `artifacts/reference-java/server-1.21.1.jar`，由项目的 [`atlas_tools.py`](../../tools/atlas_tools.py) 从核验后的官方 bundle 提取。本次反汇编使用 Microsoft OpenJDK 21.0.9。
- [NoiseChunk 的版本锁定名称导航](https://mappings.dev/1.21.1/net/minecraft/world/level/levelgen/NoiseChunk.html) 方便交叉查看不同命名空间；以下控制流、运算顺序和缓存行为均另行检查了本地官方字节码。

反汇编证据保留于 `artifacts/vanilla-perf-research/`，不提交原版代码或整份映射。下文的数字偏移均指 `javap -c` 的方法内字节码偏移，不是推测的源码行号。

## 1. 逐轴插值与两张切片

官方名称 `NoiseChunk$NoiseInterpolator` 对应 `dzc$i`。证据为 `noise-interpolator.javap.txt`：

- 持有 `slice0`、`slice1` 两张二维切片；`swapSlices`（混淆名 `l()`）交换两个引用，而不是重新建立整份三维数据。
- `updateForY`（`a(double)`，偏移 0–64）由八个角点求四个 Y 中间量；`updateForX`（`b(double)`，0–32）由这四个量求两个 X 中间量；`updateForZ`（`c(double)`，0–16）最后求一个值。
- `compute`（`a(dyr$b)`，136–140）在普通插值模式直接返回已更新的 `value`，不会在每次读取时重新做七次插值。
- 同一个 `compute` 在 `fillingCell=true` 时（49–135）调用完整的 `Mth.lerp3`。这条路径按 X→Y→Z 组合角点；普通预更新路径按 Y→X→Z 组合。两者不能在要求浮点位一致时任意互换。[官方 JAR](https://piston-data.mojang.com/v1/objects/59353fb40c36d304f2035d51e7d6e6baa98dc05c/server.jar)，类 `dzc$i`。

Atlas 起点的 [`DensitySampler.interpolate`](../../internal/worldgen/density_eval.go) 已缓存八个角点，但每次采样仍计算格点坐标、比例、拷贝角点并重算完整插值。可以借鉴“复用中间量”，同时分别保留现有 `fillingCell` 两种运算次序；不能把两种数学等价但舍入不同的公式统一成一个公式。缓存键还要区分格子、坐标和上下文模式。

原版双切片减少存储与切片重算，并不意味着 Atlas 应删除所有角点缓存。Atlas 还支持非顺序探针、地表高度查询和含水层的范围外采样；这类调用需要保留通用坐标回退。

## 2. CacheOnce 与整格批量填充

`NoiseChunk$CacheOnce` 对应 `dzc$f`，`NoiseChunk$CacheAllInCell` 对应 `dzc$e`，证据见 `noise-caches.javap.txt` 与 `noise-chunk.javap.txt`。

`CacheOnce.compute` 的步骤为：

1. 若调用上下文不是持有它的 `NoiseChunk`，直接求底层函数（0–18）。
2. 若存在与当前 `arrayInterpolationCounter` 相同的批量结果，按 `arrayIndex` 读取（19–53）。
3. 否则，若 `interpolationCounter` 未变，返回最后一个标量值（54–73）。
4. 最后才求底层函数并记录值（74–102）。其 `fillArray` 同样复用匹配的数组计数代次，否则计算后复制／克隆数组。[官方 JAR](https://piston-data.mojang.com/v1/objects/59353fb40c36d304f2035d51e7d6e6baa98dc05c/server.jar)，类 `dzc$f`。

这不是一个按全世界 XYZ 永久记忆的缓存。它绑定一次插值遍历和批量填充的计数代次。Atlas 在起点将 `cache_once` 当作直接透传，因此相同位置的共享密度子图可被重复计算。本次试过仅对绑定的不可变内置图，用“最后 XYZ＋`singlePoint`／`fillingCell` 模式＋有效位”复用成功值，未测到有意义的收益，已撤回，详见下文。手写节点、另一份设置图、失败求值和上下文变化必须保持直接路径，避免更改错误处理或跨图误命中。

`CacheAllInCell` 构造器分配 `cellWidth * cellWidth * cellHeight` 个 double，并注册到 `NoiseChunk.cellCaches`。`compute` 在合法格内按 `((cellHeight-1-inCellY)*cellWidth+inCellX)*cellWidth+inCellZ` 读取数组，否则求底层函数（`dzc$e.a(dyr$b)`，64–159）。`NoiseChunk.selectCellYZ`（`dzc.b(int,int)`）设置 `fillingCell=true`，逐一对这些格缓存调用底层 `fillArray`，再恢复为 false（16–122）。[官方 JAR](https://piston-data.mojang.com/v1/objects/59353fb40c36d304f2035d51e7d6e6baa98dc05c/server.jar)，类 `dzc`、`dzc$e`。

原版是“先批量填格，再在方块循环读值”。若 Atlas 只是增加一个 128 点数组，仍对每点递归求图一次，并没有减少首次运算；真正可借鉴的是批量图执行、共享子数组和 CacheOnce 代次复用。应通过每节点求值计数证明减少了工作，而不是只依据类名存在缓存就认定有收益。

## 3. 二元函数的边界短路

`DensityFunctions$Ap2` 对应 `dys$a`，证据见 `density-binary.javap.txt` 的 `a(dyr$b)`：先求左值，然后按运算类型分派。

| 运算 | 原版跳过右值的条件 | 字节码偏移 |
| --- | --- | --- |
| add | 不跳过 | 58–70 |
| mul | 左值等于零，直接返回正零 | 73–95 |
| min | 左值严格小于右侧函数的 `minValue` | 98–130 |
| max | 左值严格大于右侧函数的 `maxValue` | 133–162 |

批量 `fillArray` 也使用相同的条件逐元素避免求右侧值。[官方 JAR](https://piston-data.mojang.com/v1/objects/59353fb40c36d304f2035d51e7d6e6baa98dc05c/server.jar)，类 `dys$a`。

Atlas 起点的二元分支无条件求两侧。这里有潜力，但不能机械照搬：`0 * negative` 的负零、`0 * NaN`、右侧报错和手写节点的行为都可能变化。安全实现需要可靠的有限区间证明、内置纯图限定，以及现有符号零／错误语义的保留；缺少可靠边界的节点应继续求值。只凭少量地形摘要一致，不能证明任意表达式短路正确。

## 4. 含水层采用状态数组，不只缓存随机中心

`Aquifer$NoiseBasedAquifer` 对应 `dyk$c`，证据见 `aquifer.javap.txt`：

- 构造器 197–221 分配相同长度的 `FluidStatus[] aquiferCache` 和 `long[] aquiferLocationCache`，位置数组以 `Long.MAX_VALUE` 作无效哨兵。
- `getIndex`（`a(int,int,int)`，0–42）先减去最小格坐标，索引为 `((dy*gridSizeZ)+dz)*gridSizeX+dx`。
- `getAquiferStatus`（`a(long)`，0–50）拆出随机中心的块坐标，再分别按 16、12、16 向下整除还原格坐标。52–68 直接读取状态数组；69–91 缺失时调用 `computeFluid` 并写入数组。[官方 JAR](https://piston-data.mojang.com/v1/objects/59353fb40c36d304f2035d51e7d6e6baa98dc05c/server.jar)，类 `dyk$c`。

Atlas 起点的 [`noiseAquifer`](../../internal/worldgen/aquifer.go) 已有 `centerGrid`，但 `fluids` 仍是三维坐标映射。随机中心的抖动范围小于 16×12×16 格宽，可从中心还原原格并访问对应状态槽。借鉴原版的数组索引是清晰且局部的优化：保留负数 floor 除法、范围外坐标映射、失败不缓存及随机数消费顺序即可，不必改变含水层判定公式。

## 5. 共享区块阶段的结构性差距

原版 `GenerationChunkHolder`（`aqn`）持有按状态索引的 `AtomicReferenceArray<CompletableFuture<ChunkResult<ChunkAccess>>>`、`startedWork` 和生成引用计数。`getOrCreateFuture` 使用 `compareAndExchange` 安装共享 future；`applyStep`（`a(dwc,aqm,azi)`）只有获取新状态推进权的一方才调用工作函数，其他调用者复用该状态的 future（15–54）。已完成的生成请求直接返回现有 future（`a(dvz,aqb)`，12–26）。证据见 `chunk-holder.javap.txt`。[官方 JAR](https://piston-data.mojang.com/v1/objects/59353fb40c36d304f2035d51e7d6e6baa98dc05c/server.jar)，类 `aqn`。

`ChunkGenerationTask`（`apy`）按阶段调度邻区并等待该层 future；`ChunkStep`（`dwc`）与 `ChunkPyramid`（`dvy`）区分直接依赖、累积依赖和写入半径。它们没有要求每个依赖邻区都先成为 FULL。例如生成金字塔中，NOISE 的显式要求是 STRUCTURE_STARTS 半径 8、BIOMES 半径 1；FEATURES 为 STRUCTURE_STARTS 半径 8、CARVERS 半径 1，并有方块写入半径 1。完整依赖还包含前序中心阶段及累积关系，不能只按这两个数建立新的调度器。证据见 `chunk-stages.javap.txt` 中 `dvy.t(dwc$a)`、`dvy.q(dwc$a)`、`apy.a(dvz,boolean)` 与 `dwc`。[官方 JAR](https://piston-data.mojang.com/v1/objects/59353fb40c36d304f2035d51e7d6e6baa98dc05c/server.jar)，类 `dvy`、`dwc`、`apy`。

Atlas 的 [`decorationBases`／`decorateRegion`](../../internal/worldgen/overworld.go) 为每个目标先取得 5×5 份完成噪声、地表和雕刻的基础，再以固定次序重放 3×3 个源区块的装饰，最后只物化目标区块。它以私有覆盖层保护邻区及玩家编辑，并保证请求顺序不改目标结果。

对 9×9 的 81 个目标作几何计数，可得到：

| 项目 | 次数／坐标并集 |
| --- | ---: |
| 目标进入 5×5 基础区的逻辑查询 | 81×25 = 2,025 |
| 这些基础坐标的理论并集（不含结构额外查询） | 13×13 = 169 |
| 目标重放的源装饰上下文 | 81×9 = 729 |
| 源区块坐标并集 | 11×11 = 121 |
| 平均每个源坐标参与的重放上下文 | 约 6.0 |

这不是实际噪声生成次数或原版 FEATURE 任务次数：Atlas 已有基础 single-flight 缓存和结构前置阶段缓存，原版也有累积邻区依赖。尤其 `169 > overworldCacheLimit(128)` 不能单独证明 LRU 抖动，滑动工作集可能小于整个并集。必须计数 `generateBase` 实际调用、唯一坐标数、命中、等待、淘汰及前置结构重放，才能判断扩容、活跃任务 pin 或改变准入策略的收益。

原版共享 ProtoChunk 的邻区写入与阶段完成关系，会改变相邻地物看见的方块、装饰先后和后处理请求。Atlas 本来就把 `neighbor_decoration_order` 列为生成缺口。直接将“9 源重放”改成“全局每源装饰一次”，可能改善与原版的接近度，却会改变当前输出；在本次硬约束下不能作为普通缓存优化提交。若缓存重放结果，键必须涵盖确切读视图、前置步骤及顺序，且缓存快照要不可变，不能仅以源坐标代替上下文。

## 6. 同步保存与异步保存的速度不能混算

`ChunkMap`（`aqb`）的 `saveChunkIfNeeded`（`d(apz)`）检查下次允许保存时间，在成功提交后设置 `now+10000ms`（89–100）；`save(ChunkAccess)`（`a(duy)`）序列化 NBT 后调用返回 `CompletableFuture` 的写接口，只挂接错误处理、不在该方法等待完成（116–133）。这是常规保存路径的证据，不代表 flush、停服或所有存档调用都不等待。见 `chunk-map.javap.txt`。[官方 JAR](https://piston-data.mojang.com/v1/objects/59353fb40c36d304f2035d51e7d6e6baa98dc05c/server.jar)，类 `aqb`。

Atlas 的 [`loadBlockChunk`](../../internal/world/vanilla.go) 生成缺失区块后先 `persistChunk`；[`genWorker`](../../internal/world/world.go) 安装光照后若 dirty，又持久化一次。[Anvil 写入](../../internal/anvil/region.go) 原子替换整个区域文件。因此已有“生成、照明并保存”基准具有比“区块已经出现在原版客户端”更强的结束条件。

应先拆开纯世界生成、光照、NBT／压缩、区域文件替换与保存次数，保证与原版在同一结束条件比较。把保存移出计时能改变数字，但不能据此宣称计算效率提高；取消中间保存或批量提交还需保留已有持久化、错误传播和实体实现契约。

## 后续优化方向与验证边界

1. 在当前浮点位和完整区块摘要约束下，评估批量密度图执行、共享子数组和逐轴循环推进的配合。下面的试验说明，孤立增加点级缓存和插值中间量缓存没有收益，不能只移植类名对应的局部机制。
2. 利用本次只在诊断基准启用的工作量计数，继续评估更大探索范围和不同调度下的活跃工作集、缓存准入及任务保护策略。当前 256 上限解决了此场景的大部分基础重算；继续扩容需同时量化内存，不能从一个区域外推全部负载。
3. 对短路先证明范围、有限性和错误／符号零语义，再逐种节点开放；无证明的路径继续原样执行。
4. 把真正的共享阶段 DAG 作为单独架构任务，先决定是否允许改变现有邻区装饰语义，再用独立原版世界以及当前样本分别说明变化。它不是可以无条件替换的微优化。

本次没有运行同硬件、同种子、同坐标、同结构配置、相同 FULL／光照／保存结束条件的原版吞吐基准，因此不报告“Atlas 慢原版几倍”，也不将推导出的逻辑重放次数换算成速度倍数。未启动长时间 Java 世界生成，以免干扰当前 Go 性能测量。

## 本次实现与诊断

采用两个局部改动，保留生成版本 19 和既有完整区块摘要：

- 含水层为已有随机中心数组增添流体状态槽，中心坐标反推网格后校验确实是该槽的随机中心，再复用状态。任意非中心、范围外坐标继续使用原映射；失败求值不写入状态槽。原始坐标映射作为独立测试对照。三轮 300 ms 热查询基准的中位数为映射 31.59 ns、数组 10.93 ns，均为零分配。此组件差值不能直接换算成完整区块吞吐提升。
- 主世界基础地形保留窗口从 128 调到 256；通过 `OverworldGeneratorOptions.TerrainCacheLimit` 可指定较小预算，例如 128。它是生成器的运行时资源选项，不是 `level.dat` 的生成设置。下界、末地仍使用原来的 128 窗口。缓存仍有上限；正在生成的条目与原实现一样不会中途驱逐。

最终工作量探针（种子 1234、结构开启、中心 40,-40、9×9 目标、8 工作协程）观察到：128 窗口有 725 次基础快照生成，只涉及 298 个不同坐标，出现 427 次重算；256 窗口为 301 次生成／298 个不同坐标，仅重算 3 次。此前重复诊断的 128 窗口为 710–724 次，说明具体淘汰次数也受并发调度影响。两边均有 1,188 次源装饰上下文访问，包括目标的 729 次访问和结构前置重放。探针的计数在 Go overlay 中注入，生产生成器没有增加计数回调。

这证实了基础缓存抖动；不表示已经消除 9 源装饰重放。保持生成器存活后主动 GC 测得的进程 Go 活堆约从 310.6 MiB 增到 434.6 MiB，约增加 124 MiB。此数包含全局表、结构缓存及生成器，既不是缓存独占字节，也不是峰值 RSS；地形、结构和调度会影响具体数值。

本次还试做了私有图 `cache_once` 点级复用和区分浮点次序的插值轴中间量复用，差分及完整输出检查均通过，但组件测量未显示收益：三轮中位数依次为原路径 24.36 ms、仅 CacheOnce 24.20 ms、仅轴缓存 25.31 ms、两者同时 27.22 ms。它们缺少原版批量填格、数组共享和逐轴循环推进的完整配合；不能因原版存在该机制便保留一个额外分派成本更高的实现。这两个候选已撤回，未把无收益实现带入生产代码。候选源码 overlay 和测量保留在 `artifacts/vanilla-performance/rejected/`、`cache-components.txt`。

工作量与内存探针可复现为：

```powershell
python tools/worldgen/measure_stage_reuse.py --workers 8 --structures --cache-limit 128
python tools/worldgen/measure_stage_reuse.py --workers 8 --structures --cache-limit 256
```

`base_calls/op` 是实际 `generateBase` 调用数，`unique_bases/op` 是生成过的基础坐标数，`repeated_bases/op` 为两者之差；`source_visits/op` 还包含结构前置装饰上下文。`live_heap_MiB/op` 是生成器仍存活时整个进程的 GC 后 Go 活堆。探针增加了计数锁和 GC，其 `ns/op` 不作为完整链路性能结论。

### 完整链路复测

Windows、i7-11800H、Go 默认 GOMAXPROCS=16；基准仍为种子 1234、独立蜂巢盐 0、结构开启、中心 40,-40 的 81 个新区块、8 工作协程。使用生产 World 工作池，计入生成、光照和持久化，排除创建／打开／关闭世界。基线 `44be121` 通过原源码 overlay 执行，随后执行最终候选，两边各三轮、独立取中位数。机器负载未锁定，不用上一轮报告的绝对时间代替这次基线。

| 指标（每次 81 区块） | 基线 `44be121` | 最终候选 | 变化 |
| --- | ---: | ---: | ---: |
| 生成、光照及保存耗时 | 22.098 s | 15.196 s | 减少 31.2% |
| 累计分配字节 | 3.925 GB | 2.827 GB | 减少 28.0% |
| 累计分配次数 | 14.994 M | 12.754 M | 减少 14.9% |

基线三轮耗时为 22.524、22.098、20.140 s；候选为 15.426、13.534、15.196 s。累计分配量降低和 GC 后保留堆增大并不矛盾：本次用更长的基础快照保留时间，换取更少的重新计算与临时分配。未重新测量存档区块加载吞吐；这两个改动位于新区块生成路径。

吞吐基准可复现为：

```powershell
go test ./internal/world -run '^$' -bench 'BenchmarkGenerateNormalArea/workers=8$' -benchtime=1x -count=3 -timeout=15m
```

本轮证据目录为 `artifacts/vanilla-performance/`：`pipeline-before.txt`／`pipeline-after.txt` 是以上完整链路对比；`original.json` 及 `original/*.source` 为基线 overlay；`final-stage-128.txt`／`final-stage-256.txt` 是最终工作量及活堆诊断，`aquifer-status.txt` 是含水层查询组件结果，`validation.txt` 包含原有完整输出摘要、原版密度／噪声／地表参考、不同缓存预算完整输出及并行生成检查。基础组 201 项通过。既有 12 组完整区块摘要保持原期望值，覆盖两个种子、主世界／下界／末地与结构开关；这能证明现有样本保持结果，不代表已经补齐原版邻区装饰语义。

worldgen／world 焦点竞态检查通过，覆盖含水层数组、不同缓存预算、并行主世界及维度生成、光照重载与空闲卸载；普通世界生成→编辑→保存→重开、区块数据包缓存及所有权、出生点持久化检查通过。记录见 `race.txt`、`roundtrip.txt`。最终 `go build ./...`、诊断工具 Python 语法检查及单目标 overlay 冒烟检查通过；冒烟观察到 25 个基础坐标各生成一次、9 次源上下文访问，记录见 `probe-smoke.txt`。

## 复现反汇编

以下命令从仓库根目录运行；`JAVA_HOME` 应指向可运行的完整 JDK 21+，本机可用 `C:/Users/32506/.jdks/ms-21.0.9`。不必使用 PATH 中已超时的 Oracle Java 别名。

```powershell
$env:JAVA_HOME = 'C:/Users/32506/.jdks/ms-21.0.9'
Get-FileHash jar/server.jar -Algorithm SHA1
$env:PYTHONPATH = 'tools'
python -c "import atlas_tools; print(atlas_tools.reference_server())"
$javap = Join-Path $env:JAVA_HOME 'bin/javap.exe'
$reference = 'artifacts/reference-java/server-1.21.1.jar'
& $javap -c -p -classpath $reference 'dzc$i'
& $javap -c -p -classpath $reference 'dzc$e' 'dzc$f'
& $javap -c -p -classpath $reference 'dys$a'
& $javap -c -p -classpath $reference 'dyk$c'
& $javap -c -p -classpath $reference dzc dvy dwc apy aqn aqb
```

`$` 内部类名使用单引号，避免 PowerShell 展开。官方元数据记录在 `artifacts/vanilla-perf-research/official-downloads.json`，反汇编文本名为 `noise-interpolator.javap.txt`、`noise-caches.javap.txt`、`noise-chunk.javap.txt`、`density-binary.javap.txt`、`aquifer.javap.txt`、`chunk-stages.javap.txt`、`chunk-holder.javap.txt` 和 `chunk-map.javap.txt`。
