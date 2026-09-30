// Package cmc 是 Pulse 对 CoinMarketCap Pro API 的唯一集成点。
// 密钥只从环境变量 CMC_API_KEY 读取，响应形状与官方文档对齐。
package cmc

import "encoding/json"

// Status 是 CMC 统一响应信封中的 status 对象。
type Status struct {
	Timestamp    string `json:"timestamp"`
	ErrorCode    int    `json:"error_code"`
	ErrorMessage string `json:"error_message"`
	Elapsed      int    `json:"elapsed"`
	CreditCount  int    `json:"credit_count"`
}

// ListingsResponse 对应 GET /v1/cryptocurrency/listings/latest。
type ListingsResponse struct {
	Status Status          `json:"status"`
	Data   []ListingCoin   `json:"data"`
}

// QuotesResponse 对应 GET /v1/cryptocurrency/quotes/latest。
type QuotesResponse struct {
	Status Status                     `json:"status"`
	Data   map[string]json.RawMessage `json:"data"`
}

// Coins 把 quotes/latest 按 ID 索引的对象展开为与 listings 相同的币种结构。
func (own *QuotesResponse) Coins() []ListingCoin {
	if own == nil || len(own.Data) == 0 {
		return nil
	}
	result := make([]ListingCoin, 0, len(own.Data))
	for _, raw := range own.Data {
		var coin ListingCoin
		if err := json.Unmarshal(raw, &coin); err != nil || coin.ID <= 0 {
			continue
		}
		result = append(result, coin)
	}
	return result
}

// GlobalResponse 对应 GET /v1/global-metrics/quotes/latest。
type GlobalResponse struct {
	Status Status      `json:"status"`
	Data   GlobalData  `json:"data"`
}

// KeyInfoResponse 对应 GET /v1/key/info。
type KeyInfoResponse struct {
	Status Status  `json:"status"`
	Data   KeyInfo `json:"data"`
}

// ListingCoin 是 listings/latest 中的一枚币。
type ListingCoin struct {
	ID                int                    `json:"id"`
	Name              string                 `json:"name"`
	Symbol            string                 `json:"symbol"`
	Slug              string                 `json:"slug"`
	CmcRank           int                    `json:"cmc_rank"`
	CirculatingSupply float64                `json:"circulating_supply"`
	TotalSupply       float64                `json:"total_supply"`
	MaxSupply         float64                `json:"max_supply"`
	LastUpdated       string                 `json:"last_updated"`
	Quote             map[string]QuoteUSD    `json:"quote"`
}

// QuoteUSD 是 quote.USD 对象。
type QuoteUSD struct {
	Price                float64 `json:"price"`
	Volume24h            float64 `json:"volume_24h"`
	PercentChange1h      float64 `json:"percent_change_1h"`
	PercentChange24h     float64 `json:"percent_change_24h"`
	PercentChange7d      float64 `json:"percent_change_7d"`
	MarketCap            float64 `json:"market_cap"`
	MarketCapDominance   float64 `json:"market_cap_dominance"`
	FullyDilutedMarketCap float64 `json:"fully_diluted_market_cap"`
	LastUpdated          string  `json:"last_updated"`
}

// GlobalData 是全球指标主体。
type GlobalData struct {
	ActiveCryptocurrencies int                  `json:"active_cryptocurrencies"`
	ActiveExchanges        int                  `json:"active_exchanges"`
	BtcDominance           float64              `json:"btc_dominance"`
	EthDominance           float64              `json:"eth_dominance"`
	Quote                  map[string]GlobalUSD `json:"quote"`
	LastUpdated            string               `json:"last_updated"`
}

// GlobalUSD 是全球指标的 USD 报价。
type GlobalUSD struct {
	TotalMarketCap float64 `json:"total_market_cap"`
	TotalVolume24h float64 `json:"total_volume_24h"`
	LastUpdated    string  `json:"last_updated"`
}

// KeyInfo 是 /v1/key/info 的 plan 摘要（不含密钥）。
type KeyInfo struct {
	Plan KeyPlan `json:"plan"`
}

// KeyPlan 描述当前套餐的信用与限速，供演示状态页展示。
type KeyPlan struct {
	CreditLimitMonthly      int    `json:"credit_limit_monthly"`
	CreditLimitMonthlyReset string `json:"credit_limit_monthly_reset"`
	RateLimitMinute         int    `json:"rate_limit_minute"`
}
