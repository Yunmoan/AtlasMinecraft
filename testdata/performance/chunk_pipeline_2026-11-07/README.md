# 区块生成／加载整体性能：上游与最终代码对照

这是整个性能 PR 的统一复测，不是把各轮局部优化百分比相乘。基线为上游 `Yunmoan/AtlasMinecraft` 的 `24e2b17f8511f67f73932861947fb4e4289f1644`；最终生产代码为 `fa9e2b9ab1f8b579eb6edd63b61dd30effcdc346`。后续 PR 提交只增加工具和证据，`internal/` 树保持该候选版本。

## 成绩

各项独立取三次中位数。每个操作为 81 个区块；GB／MB 是十进制累计分配字节，不是常驻内存。

| 指标 | 上游 | 最终代码 | 减少 |
| --- | ---: | ---: | ---: |
| 生成、光照及保存 | 30.821 s | 15.076 s | 51.1% |
| 生成累计分配 | 4.475 GB | 3.127 GB | 30.1% |
| 已保存区块加载 | 885.271 ms | 327.344 ms | 63.0% |
| 加载累计分配 | 124.045 MB | 92.235 MB | 25.6% |

三组均启动独立的旧／新测试进程，顺序交替。每个进程都测两条完整链路，不复用另一个进程的 Go 懒加载数据。

| 组／执行次序 | 上游生成 s | 最终生成 s | 上游加载 ms | 最终加载 ms |
| --- | ---: | ---: | ---: | ---: |
| 1（旧→新） | 30.848 | 15.548 | 909.726 | 325.813 |
| 2（新→旧） | 30.821 | 14.138 | 885.271 | 344.670 |
| 3（旧→新） | 30.327 | 15.076 | 870.010 | 327.344 |

## 测试口径

- Windows amd64，Go 1.27.0，Intel Core i7-11800H；GOMAXPROCS=16，生成工作协程为 8。
- 种子 1234，独立蜂巢盐 0，结构开启；以 `(40,-40)` 为中心的 9×9 区块。
- 生成经过生产 World 工作池、Region、光照及 Anvil 保存。加载包括 Anvil 解压／解码、实体读取及光照安装。计时排除存档创建、Open、Close；加载前先生成并关闭相同区域，因此操作系统文件缓存已预热。
- 两边使用同一份已提交的 `BenchmarkGenerateNormalArea`／`BenchmarkLoadSavedNormalArea`。基线通过 Go overlay 恢复全部有变化的上游 Go 文件，隐藏新增声明；只保留同一份基准入口。嵌入输入和 go.mod 无差异；工作区生产代码没有被临时覆盖。
- 桌面后台负载、CPU 频率与电源状态未锁定。成绩只覆盖该场景，不代表所有种子／维度、冷盘读取、网络交付或客户端帧率；没有与 Java 原版在同结束条件下比较吞吐。

主世界基础快照预算从 128 提高到 256。此前独立缓存诊断中，GC 后进程 Go 活堆约从 310.6 MiB 增到 434.6 MiB，约多 124 MiB；该诊断不是本表的 RSS 或峰值内存测量。可通过 `OverworldGeneratorOptions.TerrainCacheLimit` 降低预算；它不是 server.properties 配置项。下界／末地预算保持 128。

## 可核对的证据

- [environment.json](environment.json)：完整提交号、Go／平台、工作量、基准文件 SHA-256、基线与候选 internal 树及 overlay 清单。
- [results.json](results.json)：全部 12 条原始测量记录和中位数，没有删除较慢样本。
- 原始输出：[第 1 组旧](pair-1-before.txt)／[新](pair-1-after.txt)，[第 2 组旧](pair-2-before.txt)／[新](pair-2-after.txt)，[第 3 组旧](pair-3-before.txt)／[新](pair-3-after.txt)。
- 最新代码的已有验证输出：[206 项基础测试](validation-basic.txt)、[并发生成／保存重开竞态](validation-race-roundtrip.txt)、[288 个海底废墟选址及 13,824 个处理器参考案例](validation-ocean-reference.txt)。这些是修复提交阶段的记录，不是本次性能运行重新执行的测试。
- 批量求值阶段（`ab744ab`）的[12 块固定难例对照](../../worldgen/performance_fit_2026-11-07/audit.json)完整状态一致率 99.745178%，该阶段优化前后逐块 Go 数组相同；最终远邻修复的[8 块实际故障区块及邻区](../../worldgen/runtime_neighborhood_2026-11-07/audit.json)与原版一致率 99.889247%，排除双方空气后 99.680363%。两组选择和被测版本不同，不能横向比较百分比。最终代码的 12 组历史完整区块摘要也保持原期望值。
- [checksums.json](checksums.json) 给出本目录上述机器数据和原始输出的 SHA-256；文本统一 LF。发布前另行从六份原始输出复算各指标中位数并与 JSON 核对。

原版逐块测量不证明实体、方块实体 NBT、光照或一般运行刻完全等价；全量 reference／smoke 及 1,000 块随机审计未在本轮重跑。既有文档中的沉船最终熔岩请求顺序差异不属于本次性能验证，相关严格断言没有放宽。

## 复现

从包含本 PR 的仓库根目录运行，保持 `internal/` 和 go.mod 干净。若本地尚无上游合并提交，先获取：

```powershell
git fetch https://github.com/Yunmoan/AtlasMinecraft.git main
python tools/benchmark_chunk_pipeline.py --baseline 24e2b17f8511f67f73932861947fb4e4289f1644 --pairs 3 --output artifacts/chunk-pipeline-rerun
```

输出目录必须是新的；工具保留源码 overlay、全部输出和结构化结果。每组会额外生成加载场景的准备存档，这段准备时间不计入加载指标。该工具实际完成了本目录的三组测量。
