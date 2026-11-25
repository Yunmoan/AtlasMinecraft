# 以原版拟合度验收的区块生成优化

本轮以 `9195292` 为基线。验收目标调整为原版 Java 1.21.1 的生成行为与性能：允许改变 Atlas 旧结果，但需要用固定的原版逐方块对照说明变化，不能仅因更快而接受拟合退步。旧 Atlas 摘要仍用于发现无意变化；将来有原版证据支持的语义修复可以更新这些摘要。本轮没有改动已有世界、重生成存档或提高生成版本。

## 实现

- **整格批量密度求值**：`NoiseChunk` 在消费方块前填充整格数组。将插值器外的表达式编译为共享数组操作，保持原来的 Y 倒序、X、Z 消费顺序。每个插值器每格读取八个角点，并按填格路径的 X→Y→Z 顺序复用中间量；含水层及普通点采样仍走原来的上下文。`range_choice` 只对每个位置求被选中的分支，不预先填充未选分支。数组仅属于这次噪声生成，不进入基础地形常驻缓存。
- **原版零乘短路**：已绑定图中的一般二元乘法，在左值为零时不求右值，并返回正零。常量操作数对应原版 `MulOrAdd`，继续求输入并执行普通乘法，保留负零与错误语义；手写／未绑定探针继续走原路径。批量执行器不将一般二元乘法的右侧提前填满，以免破坏短路。

原版证据沿用[上一轮官方 JAR 机制对照](vanilla_generation_performance_reference_2026-11-07.md)。另检查了官方 `DensityFunctions$TwoArgumentSimpleFunction.create`（`dys$y`）字节码：362–469 对含常量的加／乘建立 `MulOrAdd`，470–484 才建立一般 `Ap2`，因此不能把“左值为零就返回正零”套到所有乘法。反汇编保存在 `artifacts/vanilla-fitting/vanilla-binary-factory.javap.txt`。

## 固定小样本的原版逐方块对照

复用已冻结、保存并独立验证过的官方 1.21.1 存档 `artifacts/random-world-comparison-2026-11-07/`，没有重新运行原版世界来挑选更有利结果。官方包 SHA-1 为 `59353fb40c36d304f2035d51e7d6e6baa98dc05c`，结构开启、Go 蜂巢盐为 0，比较主世界完整 Y=-64～319，包含方块属性。

在运行候选前固定种子 0、1234，各选六块：旧 100 块清单按 X/Z 排序后的索引 0、25、50、75，以及旧结果差异最多的两块。这个诊断集有意加入难例，其一致率不能与全范围随机 500 块的 99.98% 当成同一分布比较。坐标和逐块计数／哈希见 [`cases.json`](../../testdata/worldgen/performance_fit_2026-11-07/cases.json) 与 [`audit.json`](../../testdata/worldgen/performance_fit_2026-11-07/audit.json)。

| 指标 | 基线 | 仅批量求值 | 批量求值＋零乘短路 |
| --- | ---: | ---: | ---: |
| 完整方块状态一致率 | 99.745178% | 99.745178% | 99.745178% |
| 排除双方空气后一致率 | 99.243050% | 99.243050% | 99.243050% |
| 差异位置 | 3,006 | 3,006 | 3,006 |
| 完全一致区块 | 7/12 | 7/12 | 7/12 |
| 最差区块一致率 | 99.046834% | 99.046834% | 99.046834% |

共比较 1,179,648 个位置。非空气指标分母为至少一端非空气的位置，共 397,120 个；其中 394,114 个完整状态相同。三份 Go 数组逐块 SHA-256 相同，使用的原版数组也与原 500 块实验保存的数组逐块相同。Python 汇总工具独立读取双方 u16be 数组，复算完整／非空气计数、差异类型和 SHA-256；每个阶段另外用 `-verify-vanilla` 逐项核对最终停止的原版存档。

因此本轮的抽查结论是**保持原版拟合度**，没有消除已有树木、滴水石等差异。没有重新跑三维 1,000 块随机审计，也没有建立与原版同条件的吞吐倍数结论。若后续改为共享装饰阶段，需要增加相邻连续区块、不同进入顺序及结构交界的原版样本，逐块查看新增和消除的差异，不能只看汇总百分比。

## 性能测量

环境仍为 Windows、i7-11800H、GOMAXPROCS=16。完整链路使用既有 `BenchmarkGenerateNormalArea/workers=8`：种子 1234、蜂巢盐 0、结构开启、中心 40,-40 的 81 个目标，计入生成、光照及保存，排除创建／打开／关闭世界。

首轮顺序各跑三次：基线 13.472、15.871、14.061 秒，候选 16.310、14.588、13.221 秒；中位数为 14.061→14.588 秒（增加 3.7%）。这组数据没有确认完整链路收益，不能引用局部组件的改善来代替整体吞吐。随后以独立编译的旧／新测试程序交替运行，另测包含含水层与矿脉的完整噪声填充阶段，以区分局部收益和整条链路的波动。

交替组每次启动新测试进程，顺序为旧→新、新→旧、旧→新；与首组单进程内复用懒加载数据的 `-count=3` 不能混合统计：

| 次序 | 基线（秒） | 候选（秒） | 候选耗时变化 |
| --- | ---: | ---: | ---: |
| 第一对：旧→新 | 19.888 | 16.549 | −16.8% |
| 第二对：新→旧 | 18.473 | 19.035 | +3.0% |
| 第三对：旧→新 | 17.133 | 15.773 | −7.9% |
| 两边独立取中位数 | 18.473 | 16.549 | −10.4% |

交替组的分配字节中位数基本不变（3.130→3.128 GB），分配次数减少约 0.6%。这一组观察到完整链路改善，但第二对仍变慢，且首组中位数反向；硬件频率及桌面负载未锁定，不能承诺稳定减少 10.4%。保留两组完整记录，不将最快一次当成结果。

完整噪声阶段的三轮中位数为 **41.728→32.325 ms，减少 22.5%**；基线范围为 37.859–43.275 ms，候选为 31.853–36.727 ms。此基准实际填充一个主世界区块，包含最终密度、含水层、矿脉、方块数组及后处理记录；不含地表、雕刻、装饰、光照或持久化。分配字节中位数为 580.243→576.663 KB，分配次数为 956→817。与密度子组件一起，证明本轮在噪声阶段减少了工作，尚不能把同一百分比外推到所有世界或完整服务器吞吐。

最终组件对照中，同样启用零乘短路，仅比较标量和整格执行：一份完整主世界最终密度遍历的三轮中位数为 16.631→11.760 ms，减少 29.3%；累计分配 155.156→163.770 KB，增加约 8.6 KB，约 5.6%。这是密度组件，不包含含水层、地表、装饰、光照或保存。

## 验证与复现

新测试覆盖三维、两个种子、负坐标、上下边界及梯度区的批量／标量浮点位对照；覆盖未选分支不报错、错误后恢复上下文、一般乘法的正零短路、常量乘法保留负零以及未绑定探针的错误路径。12 组既有完整输出摘要及主世界原版噪声／密度／地表参考检查通过。基础组 204 项通过。

新执行器及并行主世界／维度生成的焦点竞态检查通过。普通世界生成→编辑→保存→重开、数据包缓存／所有权、出生点持久化检查通过，`go build ./...`、Python 语法检查和 `git diff --check` 通过。最后核对了生产文件及嵌入数据的 SHA-256，和 `final/sources.json` 中冻结的运行时源码一致；后续只添加测试基准及文档，没有改变被抽查的生成实现。

固定样本使用已有 `compare-chunks` 与独立汇总工具。例如先从本地原版审计数据中复用种子 0 的存档，输出到新的实验目录：

```powershell
go build -o artifacts/fit-compare.exe ./tools/worldgen/compare-chunks
$sample = Get-Content testdata/worldgen/performance_fit_2026-11-07/cases.json -Raw | ConvertFrom-Json
$sample | Where-Object seed -eq 0 | ConvertTo-Json | Set-Content artifacts/fit-cases-0.json -Encoding utf8
./artifacts/fit-compare.exe -cases artifacts/fit-cases-0.json -world artifacts/random-world-comparison-2026-11-07/0/vanilla/world -output artifacts/fit-new-0/chunks -workers 2
./artifacts/fit-compare.exe -verify-vanilla -cases artifacts/fit-cases-0.json -world artifacts/random-world-comparison-2026-11-07/0/vanilla/world -output artifacts/fit-new-0/chunks

go test ./internal/worldgen -run '^Test(DensityCellFill|VanillaDensityMultiply)' -count=1
go test ./internal/worldgen -run '^$' -bench '^(BenchmarkDensityCellFill|BenchmarkNoiseChunkGeneration)$' -benchtime=10x -count=3
go test ./internal/world -run '^$' -bench 'BenchmarkGenerateNormalArea/workers=8$' -benchtime=1x -count=3 -timeout=15m
```

种子 1234 使用同样命令和对应原版存档。每次更改生成代码都使用新结果目录，不能复用旧 Go 数组；复用前还应以旧实验的 `chunks` 输出运行 `-verify-vanilla`，确认参考世界未改变。完整实验目录的 `manifest.json` 可设置 `selection` 描述固定选样，汇总工具据此生成定点诊断报告；随机抽样模式保持原来的随机种子和坐标范围说明。`source_tree_sha256` 可标明包含工作区改动的源码，避免仅以 HEAD 标记候选。缺少本地原版存档时，仍使用 `compare_random_worlds.py --count 6 --seed-limit 2 --parallel 1 --output <新目录>` 重新采样；这会得到另一组坐标，不能直接当成上述固定难例集。

证据目录为 `artifacts/vanilla-fitting/`：`before/`、`after/`、`final/` 分别保存基线、批量执行和最终候选的源码哈希、原版及 Go 数组、逐块统计和独立汇总。`baseline-overlay.json` 使用 `.source` 格式的旧噪声循环与旧密度执行器，不引入额外 Go 包；新增数组执行器在基线生产路径中不被调用。`zero-before.txt` 保存修复前对零乘语义的失败，`final-validation.txt` 保存焦点参考检查，`basic.txt` 保存基础组结果，`race.txt`／`roundtrip.txt` 保存竞态及存档回归。`pipeline-before.txt`／`pipeline-after.txt` 保留首组测量，`paired-1..3-before/after.txt` 保留交替组，`noise-before.txt`／`noise-after.txt` 是完整噪声阶段，`density-cell-final.txt` 是密度组件对照。
