// Package ingest 把 CoinMarketCap 行情写入 SQLite 快照，并评估告警规则。
package ingest

import (
	"context"
	"sync"
	"time"

	"github.com/digitalwayhk/pulse/api/dto"
	"github.com/digitalwayhk/pulse/cmc"
	"github.com/digitalwayhk/pulse/models"
	"github.com/zeromicro/go-zero/core/logx"
)

const (
	defaultInterval = 5 * time.Minute
	minInterval     = 60 * time.Second
)

var (
	mu         sync.Mutex
	started    bool
	cancelFn   context.CancelFunc
	lastStatus *dto.IngestStatus
)

// Start 在服务启动完成后开始定时采集。
func Start() {
	mu.Lock()
	if started {
		mu.Unlock()
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancelFn = cancel
	started = true
	mu.Unlock()
	go func() {
		RunOnce(ctx)
		ticker := time.NewTicker(defaultInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				RunOnce(ctx)
			}
		}
	}()
}

// Stop 停止后台采集。
func Stop() {
	mu.Lock()
	defer mu.Unlock()
	if cancelFn != nil {
		cancelFn()
		cancelFn = nil
	}
	started = false
}

// Status 返回最近一次采集的摘要。
func Status() *dto.IngestStatus {
	mu.Lock()
	defer mu.Unlock()
	if lastStatus == nil {
		return &dto.IngestStatus{Message: "ingest has not run yet"}
	}
	copyStatus := *lastStatus
	return &copyStatus
}

// RunOnce 执行一次 Startup 档三端点采集：listings、quotes、global-metrics。
func RunOnce(ctx context.Context) *dto.IngestStatus {
	mu.Lock()
	if lastStatus != nil && time.Since(lastStatus.FetchedAt) < minInterval && lastStatus.OK {
		status := *lastStatus
		mu.Unlock()
		return &status
	}
	mu.Unlock()

	return runOnceWithClient(ctx, cmc.NewFromEnv())
}

func runOnceWithClient(ctx context.Context, client *cmc.Client) *dto.IngestStatus {
	status := &dto.IngestStatus{KeyConfigured: client.KeyConfigured(), MaskedKey: client.MaskedKey(), FetchedAt: time.Now().UTC(), Source: "cmc", OK: true}
	startedAt := time.Now()
	record := func(endpoint string, count, credits int, err error) {
		status.Endpoints = append(status.Endpoints, endpoint)
		run := models.NewIngestRun()
		run.Endpoint, run.Source, run.Status = endpoint, "cmc", "ok"
		run.ItemCount, run.CreditCount = count, credits
		run.FetchedAt = time.Now().UTC()
		run.DurationMs = time.Since(startedAt).Milliseconds()
		if err != nil {
			run.Status = "error"
			run.Message = "CMC request or snapshot persistence failed"
			status.Message = run.Message
			if endpoint != cmc.KeyInfoPath {
				status.OK = false
			}
		}
		status.Credits += credits
		if err := run.Insert(); err != nil {
			logx.Errorw("ingest_run_insert_failed", logx.Field("error", err.Error()))
		}
	}
	if !client.KeyConfigured() {
		status.OK = false
		status.Message = cmc.ErrMissingAPIKey.Error()
		if cmc.AllowSample() {
			if err := persistSample(status); err == nil {
				status.Source = "sample"
				status.Message += " Loaded demo samples; alerts are disabled."
			}
		}
		recordRuns(status, time.Since(startedAt))
		setStatus(status)
		return status
	}
	now := time.Now().UTC()
	credits := 0
	listings, err := client.ListingsLatest(ctx, 1, 100)
	if err == nil {
		err = persistListings(listings, now, "cmc")
		status.Listings = len(listings.Data)
		credits = listings.Status.CreditCount
	}
	record(cmc.ListingsLatestPath, status.Listings, credits, err)
	credits = 0
	quotes, err := client.QuotesLatest(ctx, quoteIDs(listings, WatchedQuoteIDs()))
	if err == nil {
		coins := validCoins(quotes.Coins())
		err = persistCoins(coins, now, "cmc")
		if err == nil {
			status.Quotes = len(coins)
			status.AlertHits = evaluateAlerts(now, coins)
		}
		credits = quotes.Status.CreditCount
	}
	record(cmc.QuotesLatestPath, status.Quotes, credits, err)
	credits = 0
	global, err := client.GlobalMetrics(ctx)
	if err == nil {
		err = persistGlobal(global, now, "cmc")
		status.HasGlobal = err == nil
		credits = global.Status.CreditCount
	}
	record(cmc.GlobalMetricsPath, 0, credits, err)
	info, err := client.KeyInfo(ctx)
	if err == nil && info != nil {
		status.MonthlyCreditLimit = info.Data.Plan.CreditLimitMonthly
		status.RateLimitMinute = info.Data.Plan.RateLimitMinute
	}
	record(cmc.KeyInfoPath, 0, 0, err)
	if status.Message == "" {
		status.Message = "live CoinMarketCap snapshots stored"
	}
	setStatus(status)
	return status
}

func persistSample(status *dto.IngestStatus) error {
	listings, err := cmc.SampleListings()
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	if err := persistListings(listings, now, "sample"); err != nil {
		return err
	}
	status.Listings = len(listings.Data)
	status.Endpoints = append(status.Endpoints, "sample:"+cmc.ListingsLatestPath)
	if quotes, err := cmc.SampleQuotes(); err == nil {
		if coins := quotes.Coins(); len(coins) > 0 {
			if err := persistCoins(coins, now, "sample"); err == nil {
				status.Quotes = len(coins)
				status.Endpoints = append(status.Endpoints, "sample:"+cmc.QuotesLatestPath)
			}
		}
	}
	if global, err := cmc.SampleGlobal(); err == nil {
		if err := persistGlobal(global, now, "sample"); err == nil {
			status.HasGlobal = true
			status.Endpoints = append(status.Endpoints, "sample:"+cmc.GlobalMetricsPath)
		}
	}
	status.AlertHits = 0
	return nil
}

func persistListings(resp *cmc.ListingsResponse, now time.Time, source string) error {
	if resp == nil {
		return nil
	}
	return persistCoins(resp.Data, now, source)
}

func persistCoins(items []cmc.ListingCoin, now time.Time, source string) error {
	for _, item := range items {
		if source == "sample" {
			existing, err := models.NewCoinQuote().FindByCmcID(item.ID)
			if err != nil {
				return err
			}
			if existing != nil && existing.Source == "cmc" {
				continue
			}
		}
		quote := item.Quote["USD"]
		model := models.NewCoinQuote()
		model.CmcID = item.ID
		model.Name = item.Name
		model.Symbol = item.Symbol
		model.Slug = item.Slug
		model.Rank = item.CmcRank
		model.PriceUSD = quote.Price
		model.Volume24h = quote.Volume24h
		model.MarketCap = quote.MarketCap
		model.PercentChange1h = quote.PercentChange1h
		model.PercentChange24h = quote.PercentChange24h
		model.PercentChange7d = quote.PercentChange7d
		model.CirculatingSupply = item.CirculatingSupply
		model.TotalSupply = item.TotalSupply
		model.MaxSupply = item.MaxSupply
		model.MarketCapDominance = quote.MarketCapDominance
		model.LastUpdated = firstNonEmpty(quote.LastUpdated, item.LastUpdated)
		model.FetchedAt = now
		model.Source = source
		if err := model.Upsert(); err != nil {
			return err
		}
	}
	return nil
}

// quoteIDs 合并自选 ID 与榜单前 5 名，保证每次采集都打 quotes/latest。
func quoteIDs(listings *cmc.ListingsResponse, extra []int) []int {
	seen := make(map[int]struct{}, 8)
	ids := make([]int, 0, 8)
	add := func(id int) {
		if id <= 0 {
			return
		}
		if _, ok := seen[id]; ok {
			return
		}
		seen[id] = struct{}{}
		ids = append(ids, id)
	}
	for _, id := range extra {
		add(id)
	}
	if listings != nil {
		for i, item := range listings.Data {
			if i >= 5 {
				break
			}
			add(item.ID)
		}
	}
	return ids
}

func watchlistIDs() []int {
	ids, err := models.NewWatchlist().QueryCmcIDs()
	if err != nil {
		logx.Errorw("watchlist_ids_failed", logx.Field("error", err.Error()))
		return nil
	}
	return ids
}

func persistGlobal(resp *cmc.GlobalResponse, now time.Time, source string) error {
	usd := resp.Data.Quote["USD"]
	model := models.NewGlobalMetrics()
	model.ActiveCryptocurrencies = resp.Data.ActiveCryptocurrencies
	model.ActiveExchanges = resp.Data.ActiveExchanges
	model.BtcDominance = resp.Data.BtcDominance
	model.EthDominance = resp.Data.EthDominance
	model.TotalMarketCap = usd.TotalMarketCap
	model.TotalVolume24h = usd.TotalVolume24h
	model.LastUpdated = firstNonEmpty(usd.LastUpdated, resp.Data.LastUpdated)
	model.FetchedAt = now
	model.Source = source
	return model.Upsert()
}

func validCoins(items []cmc.ListingCoin) []cmc.ListingCoin {
	valid := make([]cmc.ListingCoin, 0, len(items))
	for _, item := range items {
		if usd, ok := item.Quote["USD"]; ok && item.ID > 0 && usd.Price > 0 {
			valid = append(valid, item)
		}
	}
	return valid
}

func evaluateAlerts(now time.Time, coins []cmc.ListingCoin) int {
	fresh := make(map[int]cmc.ListingCoin, len(coins))
	for _, coin := range coins {
		fresh[coin.ID] = coin
	}
	if len(fresh) == 0 {
		return 0
	}
	rules, err := models.NewAlertRule().QueryEnabled()
	if err != nil {
		logx.Errorw("alert_query_failed", logx.Field("error", err.Error()))
		return 0
	}
	hits := 0
	for _, rule := range rules {
		coin, returned := fresh[rule.CmcID]
		if !returned {
			continue
		}
		usd := coin.Quote["USD"]
		quote := &models.CoinQuote{CmcID: coin.ID, Symbol: coin.Symbol, PriceUSD: usd.Price, PercentChange24h: usd.PercentChange24h, FetchedAt: now, Source: "cmc"}
		hit, ok := models.EvaluateAlert(rule, quote, now)
		if !ok {
			continue
		}
		triggered := now
		rule.LastTriggeredAt = &triggered
		rule.LastValue = hit.Value
		rule.LastMessage = hit.Message
		if err := rule.Update(); err != nil {
			logx.Errorw("alert_update_failed", logx.Field("error", err.Error()))
			continue
		}
		hits++
		notifyAlert(dto.NewAlertResponse(rule).WithAction("triggered"))
	}
	return hits
}

func recordRuns(status *dto.IngestStatus, elapsed time.Duration) {
	for _, endpoint := range status.Endpoints {
		run := models.NewIngestRun()
		run.Endpoint = endpoint
		run.Status = "ok"
		if !status.OK {
			run.Status = "error"
		}
		if status.Source == "sample" {
			run.Status = "sample"
		}
		run.Message = status.Message
		run.ItemCount = status.Listings
		run.CreditCount = status.Credits
		run.DurationMs = elapsed.Milliseconds()
		run.Source = status.Source
		run.FetchedAt = status.FetchedAt
		if err := run.Insert(); err != nil {
			logx.Errorw("ingest_run_insert_failed", logx.Field("error", err.Error()))
		}
	}
	if len(status.Endpoints) == 0 {
		run := models.NewIngestRun()
		run.Endpoint = "startup"
		run.Status = "error"
		run.Message = status.Message
		run.DurationMs = elapsed.Milliseconds()
		run.FetchedAt = status.FetchedAt
		_ = run.Insert()
	}
}

func setStatus(status *dto.IngestStatus) {
	mu.Lock()
	lastStatus = status
	mu.Unlock()
	if status.OK {
		logx.Infow("cmc_ingest_ok",
			logx.Field("source", status.Source),
			logx.Field("listings", status.Listings),
			logx.Field("credits", status.Credits),
		)
		return
	}
	logx.Errorw("cmc_ingest_failed", logx.Field("error", status.Message))
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
