package public

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/digitalwayhk/core/pkg/server/router"
	servertypes "github.com/digitalwayhk/core/pkg/server/types"
	"github.com/digitalwayhk/pulse/api/dto"
	"github.com/digitalwayhk/pulse/ingest"
)

// GetQuote 对点选币种打 CMC GET /v1/cryptocurrency/quotes/latest。
type GetQuote struct {
	CmcID  int
	Symbol string
}

// Parse 读取 id / cmcID / symbol。
func (own *GetQuote) Parse(req servertypes.IRequest) error {
	own.Symbol = strings.ToUpper(strings.TrimSpace(req.GetValue("symbol")))
	raw := strings.TrimSpace(req.GetValue("id"))
	if raw == "" {
		raw = strings.TrimSpace(req.GetValue("cmcID"))
	}
	if raw != "" {
		if value, err := strconv.Atoi(raw); err == nil {
			own.CmcID = value
		}
	}
	return nil
}

// Validation 要求至少提供 CMC id 或交易代码。
func (own *GetQuote) Validation(servertypes.IRequest) error {
	if own.CmcID <= 0 && own.Symbol == "" {
		return dtoQuoteIDRequired()
	}
	return nil
}

// Do 走 on-demand quotes/latest，失败时回退本地快照并标明来源。
func (own *GetQuote) Do(req servertypes.IRequest) (interface{}, error) {
	lookup, err := ingest.LookupQuote(reqContext(req), own.CmcID, own.Symbol)
	if err != nil {
		return nil, err
	}
	return dto.NewQuoteDetailResponse(lookup.Endpoint, lookup.Source, lookup.Message, lookup.Live, lookup.Coin), nil
}

// GetResponse 返回 OpenAPI 用的点选详情结构。
func (own *GetQuote) GetResponse() interface{} { return dto.QuoteDetailResponse{} }

// RouterInfo 将点选详情注册为公开 GET /api/pulse/getquote。
func (own *GetQuote) RouterInfo() *servertypes.RouterInfo {
	return router.DefaultRouterInfoWithOptions(own, router.WithMethod(http.MethodGet))
}

// GetCacheKey 按币种隔离短缓存，避免点选打穿 CMC。
func (own *GetQuote) GetCacheKey() string {
	return strconv.Itoa(own.CmcID) + "|" + own.Symbol
}
