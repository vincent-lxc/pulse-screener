package public

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/digitalwayhk/core/pkg/server/router"
	servertypes "github.com/digitalwayhk/core/pkg/server/types"
	"github.com/digitalwayhk/pulse/api/dto"
	"github.com/digitalwayhk/pulse/models"
)

// GetScreener 按涨跌幅、成交额和市值分档筛选行情。
type GetScreener struct {
	Query        string
	MinChange24h *float64
	MaxChange24h *float64
	MinVolume24h *float64
	CapBand      string
	Limit        int
}

// Parse 绑定筛选查询参数。
func (own *GetScreener) Parse(req servertypes.IRequest) error {
	own.Query = strings.TrimSpace(req.GetValue("q"))
	own.CapBand = strings.TrimSpace(req.GetValue("capBand"))
	own.Limit = parseLimit(req.GetValue("limit"), 50)
	own.MinChange24h = parseOptionalFloat(req.GetValue("minChange24h"))
	own.MaxChange24h = parseOptionalFloat(req.GetValue("maxChange24h"))
	own.MinVolume24h = parseOptionalFloat(req.GetValue("minVolume24h"))
	return nil
}

// Validation 接受空筛选，此时退化为榜单。
func (own *GetScreener) Validation(servertypes.IRequest) error { return nil }

// Do 在本地快照上应用筛选器。
func (own *GetScreener) Do(servertypes.IRequest) (interface{}, error) {
	items, err := models.NewCoinQuote().QueryLatest()
	if err != nil {
		return nil, err
	}
	filtered := models.ApplyScreener(items, models.ScreenerFilter{
		Query:        own.Query,
		MinChange24h: own.MinChange24h,
		MaxChange24h: own.MaxChange24h,
		MinVolume24h: own.MinVolume24h,
		CapBand:      own.CapBand,
		Limit:        own.Limit,
	})
	return dto.MarketListResponse{
		Source:  sourceOf(filtered),
		Updated: latestFetched(items),
		Items:   dto.CoinResponses(filtered),
	}, nil
}

// GetResponse 返回 OpenAPI 用的筛选成功响应结构。
func (own *GetScreener) GetResponse() interface{} { return dto.MarketListResponse{} }

// RouterInfo 将筛选器注册为公开 GET 路由。
func (own *GetScreener) RouterInfo() *servertypes.RouterInfo {
	return router.DefaultRouterInfoWithOptions(own, router.WithMethod(http.MethodGet))
}

// GetCacheKey 实现 IRouterCacheKey，按全部筛选条件隔离缓存。
func (own *GetScreener) GetCacheKey() string {
	return strings.Join([]string{
		own.Query,
		own.CapBand,
		strconv.Itoa(own.Limit),
		floatKey(own.MinChange24h),
		floatKey(own.MaxChange24h),
		floatKey(own.MinVolume24h),
	}, "|")
}

func parseOptionalFloat(raw string) *float64 {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	value, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return nil
	}
	return &value
}

func floatKey(value *float64) string {
	if value == nil {
		return ""
	}
	return strconv.FormatFloat(*value, 'f', 6, 64)
}
