package dto

import "github.com/digitalwayhk/pulse/models"

// WatchResponse 是自选清单对外 DTO。
type WatchResponse struct {
	Action string  `json:"action,omitempty" desc:"变更动作"`
	ID     uint    `json:"id,string" desc:"自选 ID"`
	CmcID  int     `json:"cmcID" desc:"CoinMarketCap 币种 ID"`
	Symbol string  `json:"symbol" desc:"交易代码"`
	Name   string  `json:"name" desc:"币种名称"`
	Note   string  `json:"note" desc:"备注"`
	Quote  *CoinResponse `json:"quote,omitempty" desc:"当前行情"`
}

// NewWatchResponse 从自选记录创建 DTO。
func NewWatchResponse(model *models.Watchlist, quote *models.CoinQuote) *WatchResponse {
	if model == nil {
		return nil
	}
	return &WatchResponse{
		ID:     model.ID,
		CmcID:  model.CmcID,
		Symbol: model.Symbol,
		Name:   model.Name,
		Note:   model.Note,
		Quote:  NewCoinResponse(quote),
	}
}

// WatchResponses 转换自选列表。
func WatchResponses(items []*models.Watchlist) []*WatchResponse {
	result := make([]*WatchResponse, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		quote, _ := models.NewCoinQuote().FindByCmcID(item.CmcID)
		if response := NewWatchResponse(item, quote); response != nil {
			result = append(result, response)
		}
	}
	return result
}

// AlertResponse 是告警规则对外 DTO。
type AlertResponse struct {
	Action          string  `json:"action,omitempty" desc:"变更动作"`
	ID              uint    `json:"id,string" desc:"规则 ID"`
	UserID          string  `json:"userID" desc:"用户 ID"`
	CmcID           int     `json:"cmcID" desc:"CoinMarketCap 币种 ID"`
	Symbol          string  `json:"symbol" desc:"交易代码"`
	Name            string  `json:"name" desc:"币种名称"`
	Kind            string  `json:"kind" desc:"规则类型"`
	Threshold       float64 `json:"threshold" desc:"阈值"`
	Enabled         bool    `json:"enabled" desc:"是否启用"`
	LastTriggeredAt string  `json:"lastTriggeredAt,omitempty" desc:"最近触发时间"`
	LastValue       float64 `json:"lastValue" desc:"最近比较值"`
	LastMessage     string  `json:"lastMessage" desc:"最近触发说明"`
	CooldownMinutes int     `json:"cooldownMinutes" desc:"冷却分钟"`
}

// WithAction 创建用于 WebSocket 通知的独立副本。
func (own *AlertResponse) WithAction(action string) *AlertResponse {
	if own == nil {
		return nil
	}
	result := *own
	result.Action = action
	return &result
}

// NewAlertResponse 从告警规则创建 DTO。
func NewAlertResponse(model *models.AlertRule) *AlertResponse {
	if model == nil {
		return nil
	}
	triggered := ""
	if model.LastTriggeredAt != nil {
		triggered = model.LastTriggeredAt.UTC().Format("2006-01-02T15:04:05Z")
	}
	return &AlertResponse{
		ID:              model.ID,
		UserID:          model.UserID,
		CmcID:           model.CmcID,
		Symbol:          model.Symbol,
		Name:            model.Name,
		Kind:            model.Kind,
		Threshold:       model.Threshold,
		Enabled:         model.Enabled,
		LastTriggeredAt: triggered,
		LastValue:       model.LastValue,
		LastMessage:     model.LastMessage,
		CooldownMinutes: model.CooldownMinutes,
	}
}

// AlertResponses 转换告警规则列表。
func AlertResponses(items []*models.AlertRule) []*AlertResponse {
	result := make([]*AlertResponse, 0, len(items))
	for _, item := range items {
		if response := NewAlertResponse(item); response != nil {
			result = append(result, response)
		}
	}
	return result
}
