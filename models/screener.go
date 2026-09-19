package models

import (
	"sort"
	"strings"
)

// ScreenerFilter 描述公开筛选器的查询条件。
type ScreenerFilter struct {
	Query         string
	MinChange24h  *float64
	MaxChange24h  *float64
	MinVolume24h  *float64
	MinMarketCap  *float64
	MaxMarketCap  *float64
	CapBand       string
	Limit         int
}

// ApplyScreener 在内存中对最新行情做筛选、按排名排序并截断。
func ApplyScreener(items []*CoinQuote, filter ScreenerFilter) []*CoinQuote {
	query := strings.ToUpper(strings.TrimSpace(filter.Query))
	minCap, maxCap := capBandRange(filter.CapBand)
	if filter.MinMarketCap != nil {
		minCap = filter.MinMarketCap
	}
	if filter.MaxMarketCap != nil {
		maxCap = filter.MaxMarketCap
	}
	result := make([]*CoinQuote, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		if query != "" && !strings.Contains(strings.ToUpper(item.Symbol), query) && !strings.Contains(strings.ToUpper(item.Name), query) {
			continue
		}
		if filter.MinChange24h != nil && item.PercentChange24h < *filter.MinChange24h {
			continue
		}
		if filter.MaxChange24h != nil && item.PercentChange24h > *filter.MaxChange24h {
			continue
		}
		if filter.MinVolume24h != nil && item.Volume24h < *filter.MinVolume24h {
			continue
		}
		if minCap != nil && item.MarketCap < *minCap {
			continue
		}
		if maxCap != nil && item.MarketCap >= *maxCap {
			continue
		}
		result = append(result, item)
	}
	sort.SliceStable(result, func(i, j int) bool {
		if result[i].Rank == result[j].Rank {
			return result[i].CmcID < result[j].CmcID
		}
		if result[i].Rank == 0 {
			return false
		}
		if result[j].Rank == 0 {
			return true
		}
		return result[i].Rank < result[j].Rank
	})
	limit := filter.Limit
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	if len(result) > limit {
		result = result[:limit]
	}
	return result
}

func capBandRange(band string) (min *float64, max *float64) {
	switch strings.ToLower(strings.TrimSpace(band)) {
	case "micro":
		return nil, floatPtr(100_000_000)
	case "small":
		return floatPtr(100_000_000), floatPtr(1_000_000_000)
	case "mid":
		return floatPtr(1_000_000_000), floatPtr(10_000_000_000)
	case "large":
		return floatPtr(10_000_000_000), floatPtr(100_000_000_000)
	case "mega":
		return floatPtr(100_000_000_000), nil
	default:
		return nil, nil
	}
}

func floatPtr(value float64) *float64 {
	return &value
}
