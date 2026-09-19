package models

import "testing"

func TestApplyScreenerFiltersRankAndBands(t *testing.T) {
	items := []*CoinQuote{
		{CmcID: 1, Name: "Bitcoin", Symbol: "BTC", Rank: 1, PercentChange24h: 2.5, Volume24h: 2e10, MarketCap: 8e11},
		{CmcID: 2, Name: "Ethereum", Symbol: "ETH", Rank: 2, PercentChange24h: -1.2, Volume24h: 1e10, MarketCap: 3e11},
		{CmcID: 1027, Name: "Tiny", Symbol: "TINY", Rank: 80, PercentChange24h: 12, Volume24h: 1e6, MarketCap: 5e7},
	}
	minChange := 0.0
	got := ApplyScreener(items, ScreenerFilter{Query: "bt", MinChange24h: &minChange, Limit: 10})
	if len(got) != 1 || got[0].Symbol != "BTC" {
		t.Fatalf("expected BTC only, got %#v", got)
	}
	got = ApplyScreener(items, ScreenerFilter{CapBand: "micro", Limit: 10})
	if len(got) != 1 || got[0].Symbol != "TINY" {
		t.Fatalf("expected TINY in micro band, got %#v", got)
	}
}
