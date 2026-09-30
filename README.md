# Pulse / Pulse Screener

中文 | [English](#english)

Pulse 是给 **CoinMarketCap API Hackathon**（截止约 2026-09-30 UTC）准备的行情服务。产品名 **Pulse Screener**，赛道 **Markets and Trading Tools**。

它用 `github.com/digitalwayhk/core`（Apache-2.0，go-zero + GORM）组装 Public / Private / Manage 路由，把 CMC Pro/Basic 的 listings、quotes、global metrics 快照进 SQLite，并提供可点击的 `-view` 管理后台和 `/screener/` 薄 HTML。

- 作者：bitzoom · DoraHacks：Fearless · bitzoom2025@gmail.com
- Go module：`github.com/digitalwayhk/pulse`
- 标签：`#BuildwithCMC`

## 5 分钟演示

```bash
# 需要 Go 1.26+（go.mod 写的是 1.26.7；core v1.2.1 也要求 1.26）。
# go1.24.4 编不过，请先装 1.26.x：https://go.dev/dl/ 或 go install golang.org/dl/go1.26.7@latest && go1.26.7 download
export PATH="$HOME/sdk/go1.26.7/bin:$PATH"   # 若本机 go 版本不够新

CMC_API_KEY=... go run ./cmd/pulse -view 43123 -p 18091 -grpc 19091
```

等价构建：

```bash
go test ./...
go build -o pulse ./cmd/pulse
CMC_API_KEY=... ./pulse -view 43123 -p 18091 -grpc 19091
```

然后打开：

| 入口 | 地址 | 做什么 |
| --- | --- | --- |
| 筛选器（评委主路径） | http://127.0.0.1:43123/screener/ | 全球指标、涨跌/成交/市值筛选、点选 quotes/latest 详情、demo JWT 自选与告警 |
| Core 管理后台 | http://127.0.0.1:43123/ | 「菜单管理 → 更新菜单」后 CRUD Watchlist / AlertRule，只读查看 CMC 快照 |
| OpenAPI | http://127.0.0.1:43123/swagger/ | Public / Private 契约 |
| 同源 API 代理 | http://127.0.0.1:43123/api/pulse/… | 视图端口转发 Public / Private，curl 优先走这里 |

`-p 18091` 是框架 `servermanage` 端口；Pulse 业务进程会再占相邻端口（首次生成的 `etc/pulse.json`，本环境是 `18092`）。评委和脚本请走 `43123`，不必记业务端口。

没有真实密钥时，占位值（空、`placeholder`、`your-key-here`）会在打 CMC 之前失败，错误信息明确。默认 `PULSE_ALLOW_SAMPLE` 会加载 `cmc/testdata/` 脱敏样例，页面标记 `source=sample`，方便无密钥走完点击路径。设了真密钥后只走 live CMC。样例仅用于展示，不触发提醒；报价请求失败或未返回的资产不会使用旧快照触发提醒。密钥必须设置在运行进程的环境中，构建时设置不会传给可执行文件。

```bash
# 占位密钥：进程不崩，状态接口返回清晰错误
CMC_API_KEY=placeholder ./pulse -view 43123 -p 18091 -grpc 19091
curl -s http://127.0.0.1:43123/api/pulse/getstatus
```

## 环境变量

| 变量 | 必填 | 说明 |
| --- | --- | --- |
| `CMC_API_KEY` | 实时行情时必填 | 只从环境读取，禁止写入仓库或日志 |
| `PULSE_ALLOW_SAMPLE` | 否 | 默认允许。设 `false` 则缺密钥时不落样例 |

首次运行会在可执行文件目录自动生成 `etc/server.json`、`etc/pulse.json` 和 SQLite（`db/pulse/`）。不要提交这些文件。

## 使用的 CMC 端点

完整字段与脱敏响应见 [docs/cmc-endpoints.md](docs/cmc-endpoints.md)。

1. `GET /v1/cryptocurrency/listings/latest` — 市值榜，筛选用
2. `GET /v1/cryptocurrency/quotes/latest` — 点选详情 + 告警二次确认 + 自选/前 5 名补报价
3. `GET /v1/global-metrics/quotes/latest` — BTC/ETH 占比与总市值
4. `GET /v1/key/info` — 套餐额度（可选，失败不阻断）

缺密钥时客户端在设置 `X-CMC_PRO_API_KEY` 之前就返回明确错误，进程不退出。`go test ./cmc ./ingest ./api/public` 用 httptest 覆盖 Startup 端点与点选详情路径；未设置 `CMC_API_KEY` 时跳过 live 调用。

采集每 5 分钟一次，公开接口读 SQLite，并对 GetCoins / GetScreener / GetGlobal / GetQuote 启用 core `UseCache`（10–30 秒），避免评委刷新打穿 CMC 限速。

## Interesting use of the API

**中文：** 评委打分项「Interesting use of the API」对应的是三条 Startup 端点的组合，而不是只拉一张榜单。

1. **listings/latest → 筛选**：`/api/pulse/getscreener` 用市值榜做宇宙，涨跌/成交/市值分档都在这份快照上算。
2. **quotes/latest → 点选确认**：表格点一枚币，公开接口 `GET /api/pulse/getquote?id=` 当场打 `GET /v1/cryptocurrency/quotes/latest`。页面写明 **CMC quotes/latest**，JSON 里有 `endpoint`、`live`、价格、涨跌、成交额、`lastUpdated`。
3. **quotes/latest → 告警确认**：定时采集和 `POST /api/pulse/checkalerts`（按钮 Check alerts now）按自选 + 告警 ID 再拉 quotes，写入后再评估 `price_above` / `price_below`，不用过期 listings 快照当最终价。

`/api/pulse/getstatus` 的 `lastByEndpoint` 列出每个 CMC 路径最近一次成功/失败，方便截图当证据。完整字段见 [docs/cmc-endpoints.md](docs/cmc-endpoints.md)，提交步骤见 [docs/submission-checklist.md](docs/submission-checklist.md)。

**English:** The scoring item is a three-endpoint combo, not a single listing dump.

1. **listings/latest → screen** — `getscreener` ranks and filters the listing snapshot.
2. **quotes/latest → confirm** — clicking a row hits public `getquote`, which calls CMC `GET /v1/cryptocurrency/quotes/latest` for that id/symbol. The panel is labeled **CMC quotes/latest**.
3. **quotes/latest → alert** — ingest and `POST /api/pulse/checkalerts` re-fetch quotes for watched/alert IDs, then evaluate `price_above` / `price_below`. Alerts are not decided from a stale listing print alone.

`getstatus.lastByEndpoint` shows the last ingest per CMC path. Details: [docs/cmc-endpoints.md](docs/cmc-endpoints.md). Submit steps: [docs/submission-checklist.md](docs/submission-checklist.md).

## HTTP API

服务名 `pulse`。`public` / `private` 不出现在 URL 里。

| 类型 | 方法 | 路径 | 说明 |
| --- | --- | --- | --- |
| Public | GET | `/api/pulse/getcoins?limit=50` | 最新榜单 |
| Public | GET | `/api/pulse/getscreener?q=&minChange24h=&minVolume24h=&capBand=` | 筛选器。`capBand`=`micro\|small\|mid\|large\|mega` |
| Public | GET | `/api/pulse/getglobal` | 全球指标 |
| Public | GET | `/api/pulse/getquote?id=1` 或 `?symbol=BTC` | 点选详情，打 CMC quotes/latest |
| Public | GET | `/api/pulse/getstatus` | 密钥是否配置、最近采集、`lastByEndpoint` |
| Private | POST | `/api/pulse/addwatchlist` | `{ "cmcID": 1 }` |
| Private | GET | `/api/pulse/getwatchlist` | 本人自选 |
| Private | POST | `/api/pulse/deletewatchlist` | `{ "id": "..." }` |
| Private | POST | `/api/pulse/addalert` | `{ "cmcID":1,"kind":"price_above","threshold":70000 }` |
| Private | GET | `/api/pulse/getalerts` | 本人告警；可 WS 订阅 |
| Private | POST | `/api/pulse/checkalerts` | 先重拉 quotes/latest，再评估告警 |
| Private | POST | `/api/pulse/deletealert` | `{ "id": "..." }` |
| Manage | * | `/api/manage/pulse/{watchlistmanage\|alertrulemanage\|...}/{view,search,add,edit,remove}` | 后台 CRUD / 只读快照 |

```bash
BASE=http://127.0.0.1:43123
curl "$BASE/api/pulse/getcoins"
TOKEN=$(curl -s "$BASE/api/servermanage/testtoken?userid=demo" | jq -r .data.access_token)
curl -X POST "$BASE/api/pulse/addwatchlist" \
  -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" \
  -d '{"cmcID":1}'
```

`kind`：`price_above` / `price_below` / `pct_24h_abs`。采集后评估，默认冷却 15 分钟。WebSocket 连 Pulse 业务端口（日志里的 `service=pulse`，本环境 `ws://127.0.0.1:18092/ws`），先 `logon` 再订阅 `/api/pulse/getalerts`。

## 框架约定

本仓库 **不 vendor core 源码**，只通过 `go.mod` 依赖 `github.com/digitalwayhk/core v1.2.1`。消费模式对齐 `examples/01-simple-shop`：

- `contract` 无依赖服务名
- 模型按生命周期分基础资料（AlertRule）与业务事实（CoinQuote / Watchlist / IngestRun）
- Manage 只走 `models.NewManageModelList[T]()`
- Public/Private 走模型上的 `IDataAction` 方法，返回扁平 DTO
- 建表由 core 在首次访问时自动完成

安装消费方 skill（开发用，不要提交指向本机 module cache 的软链）：

```bash
CORE=$(go list -m -f '{{.Dir}}' github.com/digitalwayhk/core)
bash "$CORE/scripts/link-consumer-skill.sh" --target . --write-agents
```

## 黑客松提交清单

可复制字段、截图清单、演示视频要点和推文模板见 [docs/submission-checklist.md](docs/submission-checklist.md)。

- [ ] DoraHacks 项目名 **Pulse Screener**，赛道 Markets and Trading Tools
- [ ] 简介写明 listings → quotes（点选）→ quotes（告警）三端点组合，并附 `docs/cmc-endpoints.md`
- [ ] 演示：`/screener/` 筛选 + **点选 quotes/latest 详情** + Check alerts now；`-view` 更新菜单后展示 Watchlist / AlertRule / CMC 快照
- [ ] 录屏或截图：有 `CMC_API_KEY` 时的 live `source=cmc`，以及占位密钥的明确错误
- [ ] 文案带 `#BuildwithCMC`
- [ ] 联系人 Fearless / bitzoom2025@gmail.com
- [ ] **不要**把 API key 写进 README、issue 或提交

---

## English

Pulse is a CoinMarketCap-backed market data service for the **CMC API Hackathon** (deadline about 2026-09-30 UTC). The product name is **Pulse Screener** (Markets and Trading Tools).

It depends on `github.com/digitalwayhk/core` (Apache-2.0) and does **not** vendor that framework. Snapshots land in SQLite via core auto-migrate. Judges can click through `/screener/` and the built-in `-view` admin without a separate SPA.

```bash
CMC_API_KEY=... go run ./cmd/pulse -view 43123 -p 18091 -grpc 19091
```

Open http://127.0.0.1:43123/screener/ . Click a coin for a **CMC quotes/latest** confirm; use **Check alerts now** to re-fetch quotes before evaluating rules. A placeholder key fails before the network with a clear error; the process stays up. Endpoints, sanitized payloads, and credit use are listed in [docs/cmc-endpoints.md](docs/cmc-endpoints.md). Submission checklist: [docs/submission-checklist.md](docs/submission-checklist.md).

### Interesting use of the API

listings/latest screens the universe → quotes/latest confirms the selected coin (`GET /api/pulse/getquote`) → quotes/latest runs again for watch/alert IDs before `price_above` / `price_below` (`POST /api/pulse/checkalerts`). That is the judge-facing multi-endpoint story.

Module path: `github.com/digitalwayhk/pulse`. Owner: bitzoom / Fearless / bitzoom2025@gmail.com. Tag submissions `#BuildwithCMC`.
