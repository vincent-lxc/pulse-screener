package ingest

import (
	"context"
	"strings"
	"time"

	"github.com/digitalwayhk/pulse/api/dto"
	"github.com/digitalwayhk/pulse/cmc"
	"github.com/digitalwayhk/pulse/models"
	"github.com/zeromicro/go-zero/core/logx"
)

// QuoteLookup 是点选详情对 CMC quotes/latest 的一次确认结果。
type QuoteLookup struct {
	Coin     *models.CoinQuote
	Endpoint string
	Source   string
	Live     bool
	Message  string
}

// AlertCheckResult 是告警二次确认的结果，必须先打 quotes/latest。
type AlertCheckResult struct {
	Endpoint string
	Source   string
	Live     bool
	Quotes   int
	Hits     int
	Alerts   []*dto.AlertResponse
	Message  string
}

// LookupQuote 为点选详情打 CMC quotes/latest；失败时回退本地快照并标明来源。
func LookupQuote(ctx context.Context, cmcID int, symbol string) (*QuoteLookup, error) {
	return LookupQuoteWithClient(ctx, cmc.NewFromEnv(), cmcID, symbol)
}

// LookupQuoteWithClient 允许测试注入 httptest 客户端。
func LookupQuoteWithClient(ctx context.Context, client *cmc.Client, cmcID int, symbol string) (*QuoteLookup, error) {
	now := time.Now().UTC()
	resp, err := fetchQuotes(ctx, client, cmcID, symbol)
	if err == nil && resp != nil {
		coins := resp.Coins()
		if err := persistCoins(coins, now, "cmc"); err != nil {
			return nil, err
		}
		recordQuoteCall(cmc.QuotesLatestPath, "ok", "cmc", len(coins), now)
		coin := pickCoin(coins, cmcID, symbol)
		if coin == nil && cmcID > 0 {
			coin, _ = models.NewCoinQuote().FindByCmcID(cmcID)
		}
		if coin == nil && symbol != "" {
			coin, _ = models.NewCoinQuote().FindBySymbol(symbol)
		}
		if coin == nil {
			return nil, models.NewBusinessError("CMC quotes/latest 没有返回该币种")
		}
		return &QuoteLookup{
			Coin:     coin,
			Endpoint: cmc.QuotesLatestPath,
			Source:   "cmc",
			Live:     true,
			Message:  "confirmed with CMC GET /v1/cryptocurrency/quotes/latest",
		}, nil
	}

	fallback, ferr := localOrSampleQuote(cmcID, symbol)
	if ferr != nil {
		return nil, ferr
	}
	if fallback == nil {
		if err != nil {
			return nil, err
		}
		return nil, models.NewBusinessError("还没有该币种的 quotes/latest 数据")
	}
	message := "CMC quotes/latest unavailable; showing local snapshot"
	if err != nil {
		message = err.Error() + " Showing local or sample snapshot."
	}
	return &QuoteLookup{
		Coin:     fallback,
		Endpoint: cmc.QuotesLatestPath,
		Source:   fallback.Source,
		Live:     false,
		Message:  message,
	}, nil
}

// RefreshQuotesAndEvaluate 先按 ID 重拉 quotes/latest，再评估告警。
func RefreshQuotesAndEvaluate(ctx context.Context, ids []int) (*AlertCheckResult, error) {
	return RefreshQuotesAndEvaluateWithClient(ctx, cmc.NewFromEnv(), ids)
}

// RefreshQuotesAndEvaluateWithClient 允许测试注入 httptest 客户端。
func RefreshQuotesAndEvaluateWithClient(ctx context.Context, client *cmc.Client, ids []int) (*AlertCheckResult, error) {
	result := &AlertCheckResult{Endpoint: cmc.QuotesLatestPath, Alerts: []*dto.AlertResponse{}}
	now := time.Now().UTC()
	if len(ids) == 0 {
		result.Message = "no CMC ids to confirm via quotes/latest"
		return result, nil
	}
	resp, err := client.QuotesLatest(ctx, ids)
	if err != nil {
		result.Message = err.Error()
		if cmc.AllowSample() {
			if sample, sampleErr := cmc.SampleQuotes(); sampleErr == nil {
				_ = persistCoins(sample.Coins(), now, "sample")
				result.Source = "sample"
				result.Quotes = len(sample.Coins())
				result.Hits = evaluateAlerts(now)
				result.Message = err.Error() + " Evaluated against sample/local quotes."
				return result, nil
			}
		}
		result.Hits = evaluateAlerts(now)
		result.Source = "snapshot"
		result.Message = err.Error() + " Evaluated against last SQLite snapshot."
		return result, nil
	}
	coins := resp.Coins()
	if err := persistCoins(coins, now, "cmc"); err != nil {
		return nil, err
	}
	recordQuoteCall(cmc.QuotesLatestPath, "ok", "cmc", len(coins), now)
	result.Live = true
	result.Source = "cmc"
	result.Quotes = len(coins)
	result.Hits = evaluateAlerts(now)
	result.Message = "alerts confirmed with CMC GET /v1/cryptocurrency/quotes/latest"
	return result, nil
}

func fetchQuotes(ctx context.Context, client *cmc.Client, cmcID int, symbol string) (*cmc.QuotesResponse, error) {
	if client == nil {
		return nil, cmc.ErrMissingAPIKey
	}
	if cmcID > 0 {
		return client.QuotesLatest(ctx, []int{cmcID})
	}
	return client.QuotesLatestBySymbol(ctx, symbol)
}

func pickCoin(coins []cmc.ListingCoin, cmcID int, symbol string) *models.CoinQuote {
	symbol = strings.ToUpper(strings.TrimSpace(symbol))
	for _, item := range coins {
		if cmcID > 0 && item.ID != cmcID {
			continue
		}
		if symbol != "" && strings.ToUpper(item.Symbol) != symbol && cmcID <= 0 {
			continue
		}
		if cmcID > 0 || symbol == "" || strings.ToUpper(item.Symbol) == symbol {
			quote, _ := models.NewCoinQuote().FindByCmcID(item.ID)
			return quote
		}
	}
	return nil
}

func localOrSampleQuote(cmcID int, symbol string) (*models.CoinQuote, error) {
	if cmcID > 0 {
		if item, err := models.NewCoinQuote().FindByCmcID(cmcID); err != nil || item != nil {
			return item, err
		}
	}
	if symbol != "" {
		if item, err := models.NewCoinQuote().FindBySymbol(symbol); err != nil || item != nil {
			return item, err
		}
	}
	if !cmc.AllowSample() {
		return nil, nil
	}
	sample, err := cmc.SampleQuotes()
	if err != nil {
		return nil, err
	}
	if err := persistCoins(sample.Coins(), time.Now().UTC(), "sample"); err != nil {
		return nil, err
	}
	if cmcID > 0 {
		return models.NewCoinQuote().FindByCmcID(cmcID)
	}
	return models.NewCoinQuote().FindBySymbol(symbol)
}

// WatchedQuoteIDs 合并自选与告警规则中的 CMC ID，供 quotes/latest 二次确认。
func WatchedQuoteIDs() []int {
	seen := make(map[int]struct{}, 16)
	ids := make([]int, 0, 16)
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
	for _, id := range watchlistIDs() {
		add(id)
	}
	for _, id := range alertRuleIDs() {
		add(id)
	}
	return ids
}

func alertRuleIDs() []int {
	ids, err := models.NewAlertRule().QueryCmcIDs()
	if err != nil {
		logx.Errorw("alert_ids_failed", logx.Field("error", err.Error()))
		return nil
	}
	return ids
}

func recordQuoteCall(endpoint, status, source string, count int, now time.Time) {
	run := models.NewIngestRun()
	run.Endpoint = endpoint
	run.Status = status
	run.Source = source
	run.ItemCount = count
	run.FetchedAt = now
	run.Message = "on-demand CMC quotes/latest"
	if err := run.Insert(); err != nil {
		logx.Errorw("quote_run_insert_failed", logx.Field("error", err.Error()))
	}
}
