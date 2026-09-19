package private

import (
	"strings"

	"github.com/digitalwayhk/core/pkg/server/router"
	servertypes "github.com/digitalwayhk/core/pkg/server/types"
	"github.com/digitalwayhk/pulse/api/dto"
	"github.com/digitalwayhk/pulse/models"
)

// AddAlert 为当前登录用户创建一条价格或涨跌幅告警。
type AddAlert struct {
	CmcID           int     `json:"cmcID"`
	Kind            string  `json:"kind"`
	Threshold       float64 `json:"threshold"`
	CooldownMinutes int     `json:"cooldownMinutes"`
}

// Parse 绑定告警参数。
func (own *AddAlert) Parse(req servertypes.IRequest) error { return req.Bind(own) }

// Validation 校验登录身份与规则字段。
func (own *AddAlert) Validation(req servertypes.IRequest) error {
	userID, _ := req.GetUser()
	if strings.TrimSpace(userID) == "" {
		return models.NewBusinessError("用户身份无效")
	}
	if own.CmcID <= 0 {
		return models.NewValidationError("请提供 cmcID")
	}
	if !models.ValidAlertKind(own.Kind) {
		return models.NewValidationError("规则类型必须是 price_above、price_below 或 pct_24h_abs")
	}
	if own.Threshold <= 0 {
		return models.NewValidationError("阈值必须大于 0")
	}
	return nil
}

// Do 用本地行情补全名称后写入告警规则。
func (own *AddAlert) Do(req servertypes.IRequest) (interface{}, error) {
	quote, err := models.NewCoinQuote().FindByCmcID(own.CmcID)
	if err != nil {
		return nil, err
	}
	if quote == nil {
		return nil, models.NewBusinessError("本地还没有该币种行情，请先等待 CMC 采集完成")
	}
	rule := models.NewAlertRule()
	rule.SetID(req.NewID())
	rule.UserID, _ = req.GetUser()
	rule.CmcID = quote.CmcID
	rule.Symbol = quote.Symbol
	rule.Name = quote.Name
	rule.Kind = strings.TrimSpace(own.Kind)
	rule.Threshold = own.Threshold
	rule.Enabled = true
	if own.CooldownMinutes > 0 {
		rule.CooldownMinutes = own.CooldownMinutes
	}
	if err := rule.Insert(); err != nil {
		return nil, err
	}
	response := dto.NewAlertResponse(rule)
	NotifyAlertChange(response.WithAction("created"))
	return response, nil
}

// GetResponse 返回 OpenAPI 用的新增告警成功响应结构。
func (own *AddAlert) GetResponse() interface{} { return dto.AlertResponse{} }

// RouterInfo 将新增告警注册为需要认证的 POST 路由。
func (own *AddAlert) RouterInfo() *servertypes.RouterInfo {
	return router.DefaultRouterInfo(own)
}
