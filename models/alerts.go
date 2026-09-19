package models

import (
	"fmt"
	"math"
	"time"
)

// AlertHit 表示一次规则评估命中。
type AlertHit struct {
	Rule    *AlertRule
	Quote   *CoinQuote
	Value   float64
	Message string
}

// EvaluateAlert 判断当前行情是否触发规则，并尊重冷却时间。
func EvaluateAlert(rule *AlertRule, quote *CoinQuote, now time.Time) (*AlertHit, bool) {
	if rule == nil || quote == nil || !rule.Enabled || rule.CmcID != quote.CmcID {
		return nil, false
	}
	cooldown := time.Duration(rule.CooldownMinutes) * time.Minute
	if cooldown <= 0 {
		cooldown = 15 * time.Minute
	}
	if rule.LastTriggeredAt != nil && now.Sub(rule.LastTriggeredAt.UTC()) < cooldown {
		return nil, false
	}
	var value float64
	var ok bool
	var message string
	switch rule.Kind {
	case AlertPriceAbove:
		value = quote.PriceUSD
		ok = value >= rule.Threshold
		message = fmt.Sprintf("%s 价格 %.4f 向上突破 %.4f", quote.Symbol, value, rule.Threshold)
	case AlertPriceBelow:
		value = quote.PriceUSD
		ok = value <= rule.Threshold
		message = fmt.Sprintf("%s 价格 %.4f 向下跌破 %.4f", quote.Symbol, value, rule.Threshold)
	case AlertPct24hAbs:
		value = math.Abs(quote.PercentChange24h)
		ok = value >= rule.Threshold
		message = fmt.Sprintf("%s 24h 涨跌幅 %.2f%% 绝对值达到 %.2f%%", quote.Symbol, quote.PercentChange24h, rule.Threshold)
	default:
		return nil, false
	}
	if !ok {
		return nil, false
	}
	return &AlertHit{Rule: rule, Quote: quote, Value: value, Message: message}, true
}
