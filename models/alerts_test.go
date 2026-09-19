package models

import (
	"testing"
	"time"
)

func TestEvaluateAlertRespectsKindAndCooldown(t *testing.T) {
	quote := &CoinQuote{CmcID: 1, Symbol: "BTC", PriceUSD: 100, PercentChange24h: -8}
	rule := &AlertRule{CmcID: 1, Kind: AlertPriceAbove, Threshold: 90, Enabled: true, CooldownMinutes: 15}
	if hit, ok := EvaluateAlert(rule, quote, time.Now().UTC()); !ok || hit == nil {
		t.Fatal("expected price_above hit")
	}
	rule.Kind = AlertPriceBelow
	rule.Threshold = 90
	if _, ok := EvaluateAlert(rule, quote, time.Now().UTC()); ok {
		t.Fatal("did not expect price_below hit")
	}
	rule.Kind = AlertPct24hAbs
	rule.Threshold = 5
	if hit, ok := EvaluateAlert(rule, quote, time.Now().UTC()); !ok || hit == nil {
		t.Fatal("expected pct_24h_abs hit")
	}
	triggered := time.Now().UTC()
	rule.LastTriggeredAt = &triggered
	if _, ok := EvaluateAlert(rule, quote, triggered.Add(time.Minute)); ok {
		t.Fatal("expected cooldown to suppress alert")
	}
}
