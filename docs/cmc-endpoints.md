# CoinMarketCap endpoints used by Pulse

Pulse talks to the official CoinMarketCap **Pro API** host only:

`https://pro-api.coinmarketcap.com`

Authentication is the `X-CMC_PRO_API_KEY` header. The key is read from the
`CMC_API_KEY` environment variable and is never written to disk, logs, or git.

Current CoinMarketCap docs still route ranked lists to `listings/latest` and
known-asset prices to `quotes/latest`, with global aggregates under
`global-metrics`. Pulse uses the v1 Startup/Basic paths below. Authentication
is only `CMC_API_KEY` → header `X-CMC_PRO_API_KEY`.

A 5-minute poll stays under Startup rate limits (listings + a small quotes
batch + global metrics, typically three credits).

## Interesting combo: listings → quotes → alert

Pulse does **not** treat `listings/latest` as the only price source.

1. **Screen** — `GET /v1/cryptocurrency/listings/latest` fills the ranked
   universe for `/api/pulse/getscreener`.
2. **Confirm** — selecting a row calls public `GET /api/pulse/getquote`, which
   re-hits `GET /v1/cryptocurrency/quotes/latest` for that `id` or `symbol`.
   The UI labels the panel **CMC quotes/latest**. The JSON echoes
   `"endpoint":"/v1/cryptocurrency/quotes/latest"`.
3. **Alert** — scheduled ingest and private `POST /api/pulse/checkalerts`
   re-query `quotes/latest` for watchlist + alert IDs, persist the confirmed
   print, then evaluate `price_above` / `price_below` / `pct_24h_abs`. Alerts
   are not decided from a stale listings snapshot alone.

That three-call story is the hackathon “Interesting use of the API” evidence:
listings for discovery, quotes for confirmation, quotes again before an alert.

| Pulse use | Method | Path | Plan floor | Poll |
| --- | --- | --- | --- | --- |
| Ranked coin universe for the screener | `GET` | `/v1/cryptocurrency/listings/latest` | Basic / Startup | every 5 minutes, `start=1&limit=100&convert=USD` |
| Watchlist + top-5 quote refresh | `GET` | `/v1/cryptocurrency/quotes/latest` | Basic / Startup | every 5 minutes, `id=<watchlist,top5>&convert=USD` |
| Global market strip | `GET` | `/v1/global-metrics/quotes/latest` | Basic / Startup | every 5 minutes, `convert=USD` |
| Demo status (credits / rate limit) | `GET` | `/v1/key/info` | all keyed plans | best-effort after a successful ingest |

CMC also publishes v3 listings/quotes successors. Pulse stays on the v1 paths
requested for this hackathon slice. Switching the constants in `cmc/client.go`
is the only change needed if a judge key is restricted to v3.

## Sanitized response shapes

Secrets, account emails, and live prices from a personal key are not stored in
this repository. The shapes below match the official envelope (`status` +
`data`) and the fields Pulse persists.

### `GET /v1/cryptocurrency/listings/latest`

```json
{
  "status": {
    "timestamp": "2026-09-19T00:00:00.000Z",
    "error_code": 0,
    "error_message": "",
    "elapsed": 12,
    "credit_count": 1
  },
  "data": [
    {
      "id": 1,
      "name": "Bitcoin",
      "symbol": "BTC",
      "slug": "bitcoin",
      "cmc_rank": 1,
      "circulating_supply": 19800000,
      "total_supply": 19800000,
      "max_supply": 21000000,
      "last_updated": "2026-09-19T00:00:00.000Z",
      "quote": {
        "USD": {
          "price": 64210.55,
          "volume_24h": 28450000000,
          "percent_change_1h": 0.21,
          "percent_change_24h": 1.84,
          "percent_change_7d": -2.11,
          "market_cap": 1271400000000,
          "market_cap_dominance": 54.2,
          "fully_diluted_market_cap": 1348410000000,
          "last_updated": "2026-09-19T00:00:00.000Z"
        }
      }
    }
  ]
}
```

### `GET /v1/cryptocurrency/quotes/latest?id=1&convert=USD`

```json
{
  "status": {
    "timestamp": "2026-09-19T00:00:00.000Z",
    "error_code": 0,
    "error_message": "",
    "elapsed": 9,
    "credit_count": 1
  },
  "data": {
    "1": {
      "id": 1,
      "name": "Bitcoin",
      "symbol": "BTC",
      "quote": {
        "USD": {
          "price": 64210.55,
          "volume_24h": 28450000000,
          "percent_change_24h": 1.84,
          "market_cap": 1271400000000,
          "last_updated": "2026-09-19T00:00:00.000Z"
        }
      }
    }
  }
}
```

### `GET /v1/global-metrics/quotes/latest?convert=USD`

```json
{
  "status": {
    "timestamp": "2026-09-19T00:00:00.000Z",
    "error_code": 0,
    "error_message": "",
    "elapsed": 8,
    "credit_count": 1
  },
  "data": {
    "active_cryptocurrencies": 9842,
    "active_exchanges": 768,
    "btc_dominance": 54.2,
    "eth_dominance": 13.8,
    "last_updated": "2026-09-19T00:00:00.000Z",
    "quote": {
      "USD": {
        "total_market_cap": 2345000000000,
        "total_volume_24h": 84200000000,
        "last_updated": "2026-09-19T00:00:00.000Z"
      }
    }
  }
}
```

### `GET /v1/key/info`

Pulse only keeps plan counters. The API key itself is never stored.

```json
{
  "status": {
    "timestamp": "2026-09-19T00:00:00.000Z",
    "error_code": 0,
    "error_message": "",
    "elapsed": 4,
    "credit_count": 0
  },
  "data": {
    "plan": {
      "credit_limit_monthly": 10000,
      "credit_limit_monthly_reset": "In 3 days, 19 hours, 56 minutes",
      "rate_limit_minute": 30
    }
  }
}
```

### Placeholder / invalid key

A missing, `placeholder`, or `your-key-here` value fails **before** the network
call:

```text
CMC_API_KEY is missing or is a placeholder. Set a real CoinMarketCap Pro API key from https://pro.coinmarketcap.com/account
```

A real-looking but rejected key surfaces the CMC `status.error_message`, for
example `This API Key is invalid.` The process stays up. Public APIs then either
serve the last SQLite snapshot or, when `PULSE_ALLOW_SAMPLE` is not `false`,
load the sanitized fixtures in `cmc/testdata/` and mark `source=sample`.
