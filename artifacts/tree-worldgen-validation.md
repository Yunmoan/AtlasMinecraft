# 树木生成验证（2026-10-01）

本文件记录生成版本 2 的树木阶段。版本 3 的植被和基础交互结果见 [最新验证](vegetation-mechanics-validation.md)。

根目录 `atlas.exe` 已更新。运行时继续使用纯 Go 和标准库；Java 1.21.1 JAR 只用于对照测试。

## 已实现

- 普通橡树、大橡树、白桦、云杉和松形云杉的基础树干与树冠算法。
- 原版树高随机数、空间检查、障碍物裁剪、地面替换、树干轴向、树叶含水状态与距离更新。树叶距离保留原版 Java 集合遍历和重复待处理项的行为。
- 加权生成数量、地表水深过滤、动态高度图、树苗生存条件、嵌套随机选择器及粉雪中的环境搜索。
- 普通世界启用 7 个完整支持的树木放置配置：`trees_taiga`、`trees_grove`、`trees_snowy`、`trees_water`、`trees_windswept_forest`、`trees_windswept_hills`、`trees_badlands`。
- 新创建世界和新生成区块的 `AtlasWorldgen.Version=2`；新世界支持阶段包含 `supported_trees`。完整性标记仍为 `Complete=0`，结构生成仍关闭。已有完整区块保持原有方块。

## 验证证据

| 检查 | 结果 |
| --- | --- |
| 配置树木原版 JAR 对照 | 600/600，通过方块状态哈希、放置结果和后续随机数验证 |
| 完整放置与选择器原版 JAR 对照 | 1,008/1,008，覆盖全部 7 个运行时树木配置及 2 个嵌套粉雪放置配置 |
| 负坐标、跨区块、极端坐标、障碍、已有树木、源水/流水、世界高度边界 | 已包含在上述原版对照中 |
| 生成树木 → 编辑 → 生成邻区块 → 重载 | 种子 0、区块 (-96,-128)，27 个树木位置及其中一处编辑保留 |
| 官方 Java 1.21.1 服务端实际强制加载、保存 Atlas 世界 | 普通/超平坦均通过，27 个树木方块状态、编辑和全部 quart 生物群系保留 |
| 最终可执行文件从空目录创建普通世界并执行协议检查 | 22/22 |
| 最终可执行文件保存停止、重启、恢复位置/天气/时间并执行协议检查 | 22/22 |
| `go test ./... -timeout=10m` | 通过 |
| `go vet ./...` | 通过 |

本轮二进制验证存档与日志：

`E:\AtlasMinecraft\artifacts\worldgen-smoke\normal-smoke-206428759`

树木向量及生成方式：

- `docs/worldgen/TreeFeatureProbe.java`、`tree_feature_1.21.1.txt`
- `docs/worldgen/TreePlacementProbe.java`、`tree_placement_1.21.1.txt`
- `docs/worldgen/build_tree_tags.py` 从原版 JAR 提取并展开树木使用的方块标签。
- `docs/worldgen/verify_reference.py --probe TreeFeatureProbe` 和 `--probe TreePlacementProbe` 可重新核对向量。需要 JDK 21 或更新版本及 Python 3；本机无 Python，本轮使用 PowerShell 编译和执行相同 Java 探针。

## 尚未实现

带蜂巢等装饰器的树、巨型/特殊树、其余植被及装饰、结构、下界、末地和原版出生点搜索仍未完成。因此普通森林、白桦林、平原等依赖蜂巢配置的树木分布仍记录为跳过，不会被替换成其他分布。

跨区块装饰仍按固定顺序重放邻区块来源，只写入目标区块，不覆盖邻区块玩家编辑。该顺序尚未证明与原版完整生成执行顺序等价。上述对照验证不代表完整世界逐方块相同，也不属于真实 Minecraft 图形客户端的目视验收。

## 编译产物 SHA-256

```text
atlas.exe
D87145C5F352A84C9FE01CAD588C16C1320DB68E69EACDD4420A26486892F06A

atlas-test.exe
4E4403E38D2E1CEB1FD5ACD04C976DD6851713E9967AF400AFB3D1193074D54B
```

```powershell
.\atlas.exe -world .\new-world -level-type minecraft:normal -seed 0
```
