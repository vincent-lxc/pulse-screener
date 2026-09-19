package models

import (
	"strconv"
	"strings"
	"time"

	"github.com/digitalwayhk/core/pkg/utils"
)

const (
	// AlertPriceAbove 表示价格向上突破阈值。
	AlertPriceAbove = "price_above"
	// AlertPriceBelow 表示价格向下跌破阈值。
	AlertPriceBelow = "price_below"
	// AlertPct24hAbs 表示 24 小时涨跌幅绝对值超过阈值。
	AlertPct24hAbs = "pct_24h_abs"
)

// AlertRule 是价格突破或涨跌幅告警规则，可供 Manage 与 Private API 维护。
type AlertRule struct {
	*BaseDataModel
	UserID           string     `json:"userID" desc:"用户 ID"`
	CmcID            int        `json:"cmcID" desc:"CoinMarketCap 币种 ID"`
	Symbol           string     `json:"symbol" desc:"交易代码快照"`
	Name             string     `json:"name" desc:"币种名称快照"`
	Kind             string     `json:"kind" desc:"规则类型 price_above / price_below / pct_24h_abs"`
	Threshold        float64    `json:"threshold" desc:"触发阈值"`
	Enabled          bool       `json:"enabled" desc:"是否启用"`
	LastTriggeredAt  *time.Time `json:"lastTriggeredAt" desc:"最近触发时间"`
	LastValue        float64    `json:"lastValue" desc:"最近一次比较值"`
	LastMessage      string     `json:"lastMessage" desc:"最近触发说明"`
	CooldownMinutes  int        `json:"cooldownMinutes" desc:"同一规则最小间隔分钟"`
}

// NewAlertRule 创建已初始化继承链的告警规则。
func NewAlertRule() *AlertRule {
	return &AlertRule{BaseDataModel: newBaseDataModel(), Enabled: true, CooldownMinutes: 15}
}

// NewModel 供 ModelList 反射创建告警规则时初始化继承链。
func (own *AlertRule) NewModel() {
	if own.BaseDataModel == nil || own.ServiceModel == nil || own.Model == nil {
		fresh := NewAlertRule()
		own.BaseDataModel = fresh.BaseDataModel
	}
}

// GetHash 以用户、币种、规则类型和阈值生成唯一哈希。
func (own *AlertRule) GetHash() string {
	userID := strings.TrimSpace(own.UserID)
	kind := strings.TrimSpace(own.Kind)
	if userID == "" || own.CmcID <= 0 || kind == "" {
		if own.Model != nil {
			return own.Hashcode
		}
		return ""
	}
	return utils.HashCodes(userID + ":" + strconv.Itoa(own.CmcID) + ":" + kind + ":" + strconv.FormatFloat(own.Threshold, 'f', 8, 64))
}

// AddValid 校验新增告警规则。
func (own *AlertRule) AddValid() error {
	return own.validate()
}

// UpdateValid 校验修改告警规则。
func (own *AlertRule) UpdateValid(_ interface{}) error {
	return own.validate()
}

func (own *AlertRule) validate() error {
	own.UserID = strings.TrimSpace(own.UserID)
	own.Symbol = strings.ToUpper(strings.TrimSpace(own.Symbol))
	own.Name = strings.TrimSpace(own.Name)
	own.Kind = strings.TrimSpace(own.Kind)
	if own.UserID == "" {
		return NewValidationError("用户 ID 不能为空")
	}
	if own.CmcID <= 0 {
		return NewValidationError("CMC 币种 ID 必须大于 0")
	}
	if !ValidAlertKind(own.Kind) {
		return NewValidationError("规则类型必须是 price_above、price_below 或 pct_24h_abs")
	}
	if own.Threshold <= 0 {
		return NewValidationError("阈值必须大于 0")
	}
	if own.CooldownMinutes <= 0 {
		own.CooldownMinutes = 15
	}
	return nil
}

// ValidAlertKind 判断告警类型是否受支持。
func ValidAlertKind(kind string) bool {
	switch strings.TrimSpace(kind) {
	case AlertPriceAbove, AlertPriceBelow, AlertPct24hAbs:
		return true
	default:
		return false
	}
}
