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

# 验证结构选址参考结果
python tools/worldgen/verify_structure_reference.py --probe StructurePlacementProbe

# 比较抓包中的注册表
python tools/protocol/diff_registries.py testdata/protocol/captures/atlas
```

Python 工具使用标准库。调用官方参考代码的工具还需要 JDK 21+ 的 `java`、`javac`，以及 `jar/server.jar`（Minecraft Java 1.21.1）。优先使用 `JAVA_HOME` 指向的完整 JDK，否则查找 `PATH`。公共工具会验证官方包的 SHA-1、版本和 DataVersion，并将探针使用的 JAR 准备到 `artifacts/reference-java`。Go 提取工具需要 Go。

`atlas_tools.py` 集中管理环境查找、参考包解压、classpath 和原子文件写入；两个主题目录的 `_bootstrap.py` 只负责直接执行脚本时的导入路径。公共工具测试：`python -m unittest discover -s tools -p "test_*.py"`。

参考输入和结果位于 [`testdata`](../testdata/README.md)，运行时内嵌资源位于 `internal`，临时日志和编译结果位于 `artifacts`。部分参考集需先运行对应 `build_*` 工具生成；目录整理不会补充尚未生成的测试数据。

验证脚本的 `--update` 会更新参考数据，仅在有意重新生成基准时使用。更详细的算法说明见[文档导航](../docs/README.md)。
