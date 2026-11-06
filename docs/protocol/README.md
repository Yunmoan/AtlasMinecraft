# Minecraft Java 1.21.1 协议

本项目使用协议版本 **767**。服务端处理握手、状态查询、登录、配置和游戏阶段。

| 内容 | 位置 |
| --- | --- |
| 数据包编号 | [packet_ids.md](packet_ids.md) |
| 数据包字段结构 | [packet_structures.md](packet_structures.md) |
| 协议和物品 JSON、注册表及 mcmeta | [testdata/protocol](../../testdata/protocol/) |
| 数据生成与验证工具 | [tools/protocol](../../tools/protocol/) |
| Java 参考探针 | [tools/protocol/probes](../../tools/protocol/probes/) |

从仓库根目录生成注册表：

```powershell
python tools/protocol/build_registries.py
```

该命令读取 `testdata/protocol/mcmeta`，生成 `internal/registry/registries.bin`。其他工具及前置要求见[工具导航](../../tools/README.md)。

返回[文档导航](../README.md)。
