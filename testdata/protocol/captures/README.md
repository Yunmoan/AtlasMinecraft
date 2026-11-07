# 协议抓包基准

| 目录 | 来源 |
| --- | --- |
| `atlas/` | 原根目录 `capture-mine`，Atlas 服务端样本 |
| `vanilla/` | 原根目录 `capture-vanilla`，原版服务端参考样本 |
| `legacy/` | 原默认输出目录 `capture`，保留其历史内容；具体来源未记录 |

`.bin` 文件按接收顺序、协议阶段和包编号命名，保存解帧后的服务端数据包负载。比较工具默认读取 `vanilla`，可传入 `atlas` 进行比较。

新的临时抓包默认写入 `artifacts/captures`。仅在确认需要作为基准保留后，才放入这里。
