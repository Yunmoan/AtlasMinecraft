# Atlas 文档导航

Atlas 是面向 Minecraft Java 1.21.1 的 Go 服务端。项目的启动方式、功能现状和限制见[项目 README](../README.md)。

## 按问题查阅

| 想了解什么 | 从这里开始 |
| --- | --- |
| 模块职责、依赖与常见调用流程 | [架构导航](architecture.md) |
| 世界、区块、区域、实体、地物与结构等术语 | [术语表](../GLOSSARY.md) |
| 网络包编号和字段 | [协议说明](protocol/README.md) |
| 世界生成现状与验证方法 | [世界生成概览](worldgen/README.md) |
| 结构如何选址、拼装和保存 | [结构生成流水线](worldgen/structure_pipeline.md) |
| 下界、末地与传送门 | [维度流水线](worldgen/dimension_pipeline.md) |
| 初始动物生成 | [动物生成说明](worldgen/initial_animals_1.21.1.md) |
| 调试地形或性能问题 | [河流海洋回归](worldgen/river_ocean_regression.md)、[Jigsaw 性能](worldgen/jigsaw_performance.md)、[区块生成与加载性能](worldgen/chunk_performance_2026-11-07.md)、[原版机制对照](worldgen/vanilla_generation_performance_reference_2026-11-07.md)、[原版拟合与性能验收](worldgen/vanilla_fit_performance_2026-11-07.md) |
| 本次 PR 的整体性能成绩和原始输出 | [上游与最终代码统一复测](../testdata/performance/chunk_pipeline_2026-11-07/README.md) |
| 选择测试组、检查测试前置条件 | [开发工具与测试入口](../tools/README.md) |

## 文档、工具和数据分别放在哪里

```text
docs/                       阅读文档
  protocol/                 网络协议说明
  worldgen/                 世界生成、维度、结构和验证说明
tools/                      开发工具；命令从仓库根目录执行
  protocol/                 协议数据生成和验证脚本
    probes/                 Java 参考探针
  worldgen/                 世界数据生成、验证和诊断脚本
    probes/                 Java 参考探针及辅助类
testdata/                   参考数据和测试输入
  protocol/                 协议 JSON、注册表和 mcmeta 数据
  worldgen/                 原版参考结果及压缩数据
artifacts/                  测试日志、临时世界、性能报告和编译产物
configs/                    可提交的配置模板
bin/                        本地编译出的程序
run/                        本地配置、实际世界、玩家数据与日志
```

工具用法见[开发工具导航](../tools/README.md)，数据用途见[参考数据说明](../testdata/README.md)。运行时需要内嵌的数据仍保留在 `internal` 对应模块中。
