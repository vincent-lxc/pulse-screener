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

	client := cmc.NewFromEnv()
	status := &dto.IngestStatus{
		KeyConfigured: client.KeyConfigured(),
		MaskedKey:     client.MaskedKey(),
		FetchedAt:     time.Now().UTC(),
	}
	startedAt := time.Now()
	if !client.KeyConfigured() {
		status.Message = cmc.ErrMissingAPIKey.Error()
		if cmc.AllowSample() {
			if err := persistSample(status); err != nil {
				status.OK = false
				status.Message = err.Error()
			} else {
				status.OK = true
				status.Source = "sample"
				status.Message = cmc.ErrMissingAPIKey.Error() + " Loaded offline sample quotes for the demo UI."
			}
		}
		recordRuns(status, time.Since(startedAt))
		setStatus(status)
		return status
	}

	listings, err := client.ListingsLatest(ctx, 1, 100)
	if err != nil {
		status.OK = false
		status.Message = err.Error()
		if cmc.AllowSample() {
			if sampleErr := persistSample(status); sampleErr == nil {
				status.OK = true
				status.Source = "sample"
				status.Message = err.Error() + " Fell back to offline sample quotes."
			}
		}
		recordRuns(status, time.Since(startedAt))
		setStatus(status)
		return status
	}
	now := time.Now().UTC()
	if err := persistListings(listings, now, "cmc"); err != nil {
		status.OK = false
		status.Message = err.Error()
		recordRuns(status, time.Since(startedAt))
		setStatus(status)
		return status
	}
	status.Listings = len(listings.Data)
	status.Credits += listings.Status.CreditCount
	status.Endpoints = append(status.Endpoints, cmc.ListingsLatestPath)

	if quotes, err := client.QuotesLatest(ctx, quoteIDs(listings, WatchedQuoteIDs())); err != nil {
		logx.Errorw("cmc_quotes_latest_failed", logx.Field("error", err.Error()))
		if status.Message == "" {
			status.Message = err.Error()
		}
	} else if coins := quotes.Coins(); len(coins) > 0 {
		if err := persistCoins(coins, now, "cmc"); err != nil {
			status.OK = false
			status.Message = err.Error()
			recordRuns(status, time.Since(startedAt))
			setStatus(status)
			return status
		}
		status.Quotes = len(coins)
		status.Credits += quotes.Status.CreditCount
		status.Endpoints = append(status.Endpoints, cmc.QuotesLatestPath)
	}

	if global, err := client.GlobalMetrics(ctx); err != nil {
		logx.Errorw("cmc_global_metrics_failed", logx.Field("error", err.Error()))
		status.Message = err.Error()
	} else if err := persistGlobal(global, now, "cmc"); err != nil {
		status.Message = err.Error()
	} else {
		status.HasGlobal = true
		status.Credits += global.Status.CreditCount
		status.Endpoints = append(status.Endpoints, cmc.GlobalMetricsPath)
	}

	if info, err := client.KeyInfo(ctx); err == nil && info != nil {
		status.MonthlyCreditLimit = info.Data.Plan.CreditLimitMonthly
		status.RateLimitMinute = info.Data.Plan.RateLimitMinute
		status.Endpoints = append(status.Endpoints, cmc.KeyInfoPath)
	}

	hits := evaluateAlerts(now)
	status.AlertHits = hits
	status.OK = true
	status.Source = "cmc"
	if status.Message == "" {
		status.Message = "live CoinMarketCap snapshots stored"
	}
	recordRuns(status, time.Since(startedAt))
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
	status.AlertHits = evaluateAlerts(now)
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

func evaluateAlerts(now time.Time) int {
	rules, err := models.NewAlertRule().QueryEnabled()
	if err != nil {
		logx.Errorw("alert_query_failed", logx.Field("error", err.Error()))
		return 0
	}
	hits := 0
	for _, rule := range rules {
		quote, err := models.NewCoinQuote().FindByCmcID(rule.CmcID)
		if err != nil || quote == nil {
			continue
		}
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
