package private

import (
	"context"
	"strings"

	"github.com/digitalwayhk/core/pkg/server/router"
	servertypes "github.com/digitalwayhk/core/pkg/server/types"
	"github.com/digitalwayhk/pulse/api/dto"
	"github.com/digitalwayhk/pulse/ingest"
	"github.com/digitalwayhk/pulse/models"
)

// CheckAlerts 先按自选/告警 ID 重拉 CMC quotes/latest，再评估规则。
type CheckAlerts struct{}

// Parse 无请求体字段。
func (own *CheckAlerts) Parse(servertypes.IRequest) error { return nil }

// Validation 要求已登录。
func (own *CheckAlerts) Validation(req servertypes.IRequest) error {
	userID, _ := req.GetUser()
	if strings.TrimSpace(userID) == "" {
		return models.NewBusinessError("用户身份无效")
	}
	return nil
}

// Do 对关注币种打 quotes/latest，再用确认价评估 price_above / price_below。
func (own *CheckAlerts) Do(req servertypes.IRequest) (interface{}, error) {
	userID, _ := req.GetUser()
	ids := ingest.WatchedQuoteIDs()
	result, err := ingest.RefreshQuotesAndEvaluate(reqContext(req), ids)
	if err != nil {
		return nil, err
	}
	items, err := models.NewAlertRule().QueryByUser(userID)
	if err != nil {
		return nil, err
	}
	return &dto.AlertCheckResponse{
		Endpoint: result.Endpoint,
		Source:   result.Source,
		Live:     result.Live,
		Quotes:   result.Quotes,
		Hits:     result.Hits,
		Message:  result.Message,
		Alerts:   dto.AlertResponses(items),
	}, nil
}

// GetResponse 返回 OpenAPI 用的告警确认结构。
func (own *CheckAlerts) GetResponse() interface{} { return dto.AlertCheckResponse{} }

// RouterInfo 将手动评估注册为需要认证的 POST /api/pulse/checkalerts。
func (own *CheckAlerts) RouterInfo() *servertypes.RouterInfo {
	return router.DefaultRouterInfo(own)
}

func reqContext(req servertypes.IRequest) context.Context {
	if req != nil {
		if httpReq, ok := req.(servertypes.IRequestHttp); ok {
			if raw := httpReq.GetHttpRequest(); raw != nil {
				return raw.Context()
			}
		}
	}
	return context.Background()
}
