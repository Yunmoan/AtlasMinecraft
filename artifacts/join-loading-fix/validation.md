# 首次入服加载等待修复

最新根目录程序与背包/基础交互验证见 [背包与交互修复](../inventory-interaction-fix/validation.md)。下文哈希及计时对应本阶段历史构建。

日期：2026-10-01。协议：Minecraft Java 1.21.1 / 767。运行时保持纯 Go、标准库。

## 原因与修改

入服流程漏发 Game Event 13（`level_chunks_load_start`，float 值 0）。官方 1.21.1 客户端的 `LevelLoadStatusManager` 初始状态是 `WAITING_FOR_SERVER`；`ClientPacketListener.handleGameEvent` 收到这个事件才调用 `loadingPacketsReceived()`，转入 `WAITING_FOR_PLAYER_CHUNK`。随后等待玩家所在位置的 section 编译完成等条件。单纯收到位置和区块不会触发这个状态转换。

`ReceivingLevelScreen.tick()` 的退出条件是 level-ready 或创建时间超过 30,000 毫秒。因此缺少事件会使正常加载退化为约 30 秒的超时退出。这里依据官方客户端字节码确定机制；没有把 TCP 到达时间当成图形界面退出时间。

- `internal/server/play.go`：在入服世界状态之后、位置及区块之前发送该事件，仅发送一次。
- `internal/proto/ids.go`：增加命名事件常量。
- `internal/server/join_protocol_test.go`：压缩 TCP 回归覆盖新玩家、远处存档位置、负小数坐标、当前块优先、完整 batch、天气/时间同步不重复加载事件。
- `cmd/atlas-test/main.go`：可执行文件 smoke 增加入服事件的值、次数及顺序检查，并输出 TCP 加载计时。

修复前新增 TCP 回归在两个场景均失败：`initial position arrived without exactly one level chunks load start`；修复后通过。

## TCP 计时

机器：Windows amd64 / AMD Ryzen 9 8945HX；seed=1234，8 个生成线程，视距 3，共 45 块。使用独立临时普通世界，开服前预生成 7×7。计时从连接前开始，包含登录、配置、包编码/压缩以及真实本机 TCP 读取。预生成不计入入服时间。

| 场景 | 重复 | 首块 | 近邻 3×3 | 全部 45 块及 batch 完成 |
| --- | --- | --- | --- | --- |
| 已预生成并在内存中 | 1 | 3.451 ms | 6.296 ms | 14.179 ms |
| 已预生成并在内存中 | 2 | 4.325 ms | 9.707 ms | 16.691 ms |
| 重启读取磁盘，未调用启动预加载 | 1 | 7.696 ms | 143.011 ms | 273.317 ms |
| 重启读取磁盘，未调用启动预加载 | 2 | 8.336 ms | 208.181 ms | 361.906 ms |

这些结果检验首次入服的包到达速度；远方新地形生成、较大视距、真实网络及客户端渲染仍会增加时间，没有进行 Paper 同机对比。

复现：

```powershell
$env:GOCACHE = Join-Path $env:TEMP 'atlas-go-cache'
$env:ATLAS_JOIN_STREAM_BENCH = '1'
& (Join-Path $env:TEMP 'atlas-go/go/bin/go.exe') test ./internal/server -run '^Test(TCPJoinStartsLevelLoadingBeforePositionAndChunks|NormalWorldTCPJoinLoadingTimings)$' -v -count=2 -timeout=5m
```

## 验证与产物

- `go test ./... -count=1 -timeout=10m` 通过；最终 smoke 客户端修改后再运行 `go test ./cmd/atlas-test ./internal/server -count=1 -timeout=2m` 通过。
- 最终 `go vet ./...` 通过。
- 最终二进制创建普通世界 **41/41**，重启恢复存档位置 **45/45**；启动半径和 worker CLI 覆盖检查通过。
- 二进制首次入服首块 4.376 ms，25 块 11.705 ms；重启恢复边缘存档位置首块 4.368 ms，25 块 1.741 s（包含预生成范围外新块）。
- smoke 日志与存档：`smoke/normal-smoke-3109548347`；测试输出：`binary-smoke.txt`。
- 根目录 `atlas.exe`、`atlas-test.exe` 已替换，替换前没有运行中的 Atlas；旧文件保留在本目录 `atlas-previous.exe`、`atlas-test-previous.exe`。
- `atlas.exe` 构建时间：`2026-10-01T08:23:14Z`（北京时间 16:23:14）。
- Windows race detector 本轮未执行。图形 Minecraft 客户端加载界面尚未实测。

| 文件（根目录与本目录一致） | SHA-256 |
| --- | --- |
| `atlas.exe` | `047C40110098D0DA90150F923A15633957C83877ED1A45079E0F5C2476F8A2D1` |
| `atlas-test.exe` | `FAB647DF6E30C0389CC845AF82D641F8CB2E8A506116CAFFD464E6977BBE72D2` |

参考依据保存在 `../runtime-fix/join-reference/`：官方 SHA-1 校验后的客户端 mappings、`client-loading-bytecode.txt`、`client-packet-listener-bytecode.txt`、`vanilla-player-list-bytecode.txt`。服务端参考也在世界状态发送中显式发送该事件及值 0。

完整原版世界生成仍有结构、部分植被/树木、下界/末地及运行时维度等缺项；本次修复不代表全部世界生成已完成。
