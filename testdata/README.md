# 参考数据与测试输入

| 目录 | 内容 | 使用方 |
| --- | --- | --- |
| `protocol/` | 协议 JSON、方块/物品表、注册表、mcmeta 元素及采集资料 | `tools/protocol` |
| `worldgen/` | 原版探针输出、地形测试向量、结构和实体数据，含压缩参考集 | Go 测试和 `tools/worldgen` |
| `protocol/captures/` | 保留的 Atlas、原版和历史抓包基准 | `tools/protocol/diff_registries.py` |

这些文件用于开发验证和资源生成。服务端运行时直接使用的内嵌资源放在 `internal` 各模块中。

部分完整参考集需要通过对应生成工具获得，详见[工具导航](../tools/README.md)。生成日志、临时世界和编译结果应写入 `artifacts`，不要混入参考数据目录。

`worldgen/random_comparison_2026-11-07/` 保存三个维度随机抽样的清单、逐区块计数、双方数组 SHA-256 和汇总 JSON。它是一次测量的公开记录，含已知差异；不是要求全部吻合的固定测试向量。原版 JAR、世界和完整方块数组保留在本地 `artifacts`。方法与限制见[对比简报](../docs/worldgen/random_chunk_comparison_2026-11-07.md)。
