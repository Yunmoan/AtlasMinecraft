# 开发工具

所有示例命令从仓库根目录执行。普通服务端运行不需要这些工具；它们用于生成内嵌数据、验证原版兼容性或诊断问题。

| 路径 | 内容 |
| --- | --- |
| `protocol/*.py` | 协议、注册表、物品数据的生成和验证 |
| `protocol/probes/*.java` | 协议与物品的官方 JAR 参考探针 |
| `worldgen/build_*.py` | 世界生成数据、模板和参考结果的构建工具 |
| `worldgen/verify_*.py` | 对照原版参考结果进行验证 |
| `worldgen/diagnose_terrain_stages.py` | 指定种子、区块的地形阶段诊断 |
| `worldgen/probes/*.java` | 世界生成和游戏交互的参考探针及辅助类 |
| `worldgen/extract_native_bubble_columns.go` | 从原版世界提取气泡柱坐标 |

## 常用命令

```powershell
# 从已有协议源数据生成注册表、物品表
python tools/protocol/build_registries.py
python tools/protocol/build_items.py

# 验证一个随机数参考探针
python tools/worldgen/verify_reference.py --probe ReferenceProbe

# 验证结构选址参考结果
python tools/worldgen/verify_structure_reference.py --probe StructurePlacementProbe

# 比较抓包中的注册表
python tools/protocol/diff_registries.py capture-mine
```

Python 工具使用标准库。调用官方参考代码的工具还需要 JDK 21 的 `java`、`javac`，以及 `jar/server.jar`（Minecraft Java 1.21.1）。部分构建脚本使用事先解包的 `vanilla/server-inner.jar` 和 `vanilla/libraries`。Go 提取工具需要 Go。

参考输入和结果位于 [`testdata`](../testdata/README.md)，运行时内嵌资源位于 `internal`，临时日志和编译结果位于 `artifacts`。部分参考集需先运行对应 `build_*` 工具生成；目录整理不会补充尚未生成的测试数据。

验证脚本的 `--update` 会更新参考数据，仅在有意重新生成基准时使用。更详细的算法说明见[文档导航](../docs/README.md)。
