# Pulse Screener — DoraHacks / CMC submission checklist

No secrets. Do not paste `CMC_API_KEY` into DoraHacks, GitHub, tweets, or screenshots.

## Screenshots to capture

1. `/screener/` with the global strip and a filtered table (`source=cmc` if the key is live).
2. **Detail-on-select**: a clicked coin and the panel that says **CMC quotes/latest**, plus price / % / volume / last updated.
3. Status chip or `/api/pulse/getstatus` JSON showing `lastByEndpoint` for
   `listings/latest`, `quotes/latest`, `global-metrics/quotes/latest`.
4. Logged-in demo user: watchlist + one alert + **Check alerts now** result
   that mentions `quotes/latest`.
5. Placeholder-key run: clear missing-key message, `source=sample` still clickable.

## Demo video (keep under 5 minutes)

Must appear on camera:

- Open `/screener/` (not only the admin home).
- Apply a filter (24h change or cap band).
- Click a coin → live/cached `quotes/latest` detail panel.
- Login as demo → add watch + alert → **Check alerts now**.
- Flash `docs/cmc-endpoints.md` or the status `endpoints` / `lastByEndpoint` list.
- Say the combo out loud: listings for screen → quotes to confirm → quotes again before alert.
- End card: **#BuildwithCMC** · Pulse Screener · Markets and Trading Tools.

## #BuildwithCMC tweet template (English)

```
Built Pulse Screener for the CoinMarketCap API Hackathon.

Listings/latest to screen the market, quotes/latest when you click a coin and again before a price alert, plus global-metrics for the tape.

#BuildwithCMC
```

## DoraHacks BUIDL fields

| Field | Paste |
| --- | --- |
| Project name | Pulse Screener |
| Track | Markets and Trading Tools |
| One-liner | Screener + click-to-confirm quotes + alerts on CoinMarketCap Startup APIs |
| Description | Use the README section 「Interesting use of the API」 plus `docs/cmc-endpoints.md`. Mention the three Startup endpoints and the listings→quotes→alert combo. |
| Demo URL | `/screener/` on the view port (`-view 43123`) |
| Repo | Origin / GitHub URL of this project (no keys in the tree) |
| Video | The 5-minute walkthrough above |
| Tags | `#BuildwithCMC`, CoinMarketCap, Markets, Go |
| Contact | Fearless / bitzoom2025@gmail.com |

## Run for judges (Go 1.26+ required)

```bash
# This repo's go.mod is `go 1.26.7` because github.com/digitalwayhk/core v1.2.1
# needs Go 1.26. go1.24.4 will not compile.
export PATH="$HOME/sdk/go1.26.7/bin:$PATH"   # or any go1.26+
CMC_API_KEY=... go run ./cmd/pulse -view 43123 -p 18091 -grpc 19091
```

Open http://127.0.0.1:43123/screener/
