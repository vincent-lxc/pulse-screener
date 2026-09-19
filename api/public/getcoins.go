package public

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/digitalwayhk/core/pkg/server/router"
	servertypes "github.com/digitalwayhk/core/pkg/server/types"
	"github.com/digitalwayhk/pulse/api/dto"
	"github.com/digitalwayhk/pulse/models"
)

// GetCoins 返回按市值排名的最新币种列表。
type GetCoins struct {
	Limit int
}

// Parse 读取可选 limit。
func (own *GetCoins) Parse(req servertypes.IRequest) error {
	own.Limit = parseLimit(req.GetValue("limit"), 50)
	return nil
}

// Validation 接受默认榜单查询。
func (own *GetCoins) Validation(servertypes.IRequest) error { return nil }

// Do 读取本地快照，不直接调用 CMC，以遵守限速。
func (own *GetCoins) Do(servertypes.IRequest) (interface{}, error) {
	items, err := models.NewCoinQuote().QueryLatest()
	if err != nil {
		return nil, err
	}
	filtered := models.ApplyScreener(items, models.ScreenerFilter{Limit: own.Limit})
	return dto.MarketListResponse{
		Source:  sourceOf(filtered),
		Updated: latestFetched(items),
		Items:   dto.CoinResponses(filtered),
	}, nil
}

// GetResponse 返回 OpenAPI 用的榜单成功响应结构。
func (own *GetCoins) GetResponse() interface{} { return dto.MarketListResponse{} }

// RouterInfo 将榜单查询注册为公开 GET 路由。
func (own *GetCoins) RouterInfo() *servertypes.RouterInfo {
	return router.DefaultRouterInfoWithOptions(own, router.WithMethod(http.MethodGet))
}

// GetCacheKey 实现 IRouterCacheKey，按 limit 隔离缓存。
func (own *GetCoins) GetCacheKey() string {
	return "coins:" + strconv.Itoa(own.Limit)
}

func parseLimit(raw string, fallback int) int {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return fallback
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	return value
}

func sourceOf(items []*models.CoinQuote) string {
	if len(items) == 0 {
		return ""
	}
	return items[0].Source
}

func latestFetched(items []*models.CoinQuote) string {
	var latest time.Time
	for _, item := range items {
		if item != nil && item.FetchedAt.After(latest) {
			latest = item.FetchedAt
		}
	}
	if latest.IsZero() {
		return ""
	}
	return latest.UTC().Format(time.RFC3339)
}
