package dto

import (
	"time"

	"github.com/digitalwayhk/pulse/models"
)

// CoinResponse 是公开行情的最小 DTO。
type CoinResponse struct {
	CmcID            int     `json:"cmcID" desc:"CoinMarketCap 币种 ID"`
	Name             string  `json:"name" desc:"币种名称"`
	Symbol           string  `json:"symbol" desc:"交易代码"`
	Slug             string  `json:"slug" desc:"CMC slug"`
	Rank             int     `json:"rank" desc:"市值排名"`
	PriceUSD         float64 `json:"priceUSD" desc:"USD 价格"`
	Volume24h        float64 `json:"volume24h" desc:"24 小时成交额"`
	MarketCap        float64 `json:"marketCap" desc:"流通市值"`
	PercentChange1h  float64 `json:"percentChange1h" desc:"1 小时涨跌幅"`
	PercentChange24h float64 `json:"percentChange24h" desc:"24 小时涨跌幅"`
	PercentChange7d  float64 `json:"percentChange7d" desc:"7 日涨跌幅"`
	LastUpdated      string  `json:"lastUpdated" desc:"CMC 更新时间"`
	Source           string  `json:"source" desc:"cmc 或 sample"`
}

// NewCoinResponse 从行情快照创建公开 DTO。
func NewCoinResponse(model *models.CoinQuote) *CoinResponse {
	if model == nil {
		return nil
	}
	return &CoinResponse{
		CmcID:            model.CmcID,
		Name:             model.Name,
		Symbol:           model.Symbol,
		Slug:             model.Slug,
		Rank:             model.Rank,
		PriceUSD:         model.PriceUSD,
		Volume24h:        model.Volume24h,
		MarketCap:        model.MarketCap,
		PercentChange1h:  model.PercentChange1h,
		PercentChange24h: model.PercentChange24h,
		PercentChange7d:  model.PercentChange7d,
		LastUpdated:      model.LastUpdated,
		Source:           model.Source,
	}
}

// CoinResponses 将行情列表转换为公开 DTO 列表。
func CoinResponses(items []*models.CoinQuote) []*CoinResponse {
	result := make([]*CoinResponse, 0, len(items))
	for _, item := range items {
		if response := NewCoinResponse(item); response != nil {
			result = append(result, response)
		}
	}
	return result
}

// MarketListResponse 包装公开列表与数据来源。
type MarketListResponse struct {
	Source  string          `json:"source" desc:"cmc 或 sample"`
	Updated string          `json:"updated" desc:"本地最近采集时间"`
	Items   []*CoinResponse `json:"items" desc:"行情列表"`
}

// GlobalResponse 是全球指标公开 DTO。
type GlobalResponse struct {
	ActiveCryptocurrencies int     `json:"activeCryptocurrencies" desc:"活跃币种数"`
	ActiveExchanges        int     `json:"activeExchanges" desc:"活跃交易所数"`
	BtcDominance           float64 `json:"btcDominance" desc:"BTC 市值占比"`
	EthDominance           float64 `json:"ethDominance" desc:"ETH 市值占比"`
	TotalMarketCap         float64 `json:"totalMarketCap" desc:"总市值"`
	TotalVolume24h         float64 `json:"totalVolume24h" desc:"24 小时总成交额"`
	LastUpdated            string  `json:"lastUpdated" desc:"CMC 更新时间"`
	Source                 string  `json:"source" desc:"cmc 或 sample"`
}

// NewGlobalResponse 从全球指标快照创建公开 DTO。
func NewGlobalResponse(model *models.GlobalMetrics) *GlobalResponse {
	if model == nil {
		return nil
	}
	return &GlobalResponse{
		ActiveCryptocurrencies: model.ActiveCryptocurrencies,
		ActiveExchanges:        model.ActiveExchanges,
		BtcDominance:           model.BtcDominance,
		EthDominance:           model.EthDominance,
		TotalMarketCap:         model.TotalMarketCap,
		TotalVolume24h:         model.TotalVolume24h,
		LastUpdated:            model.LastUpdated,
		Source:                 model.Source,
	}
}

// IngestStatus 描述最近一次 CMC 采集。
type IngestStatus struct {
	OK                 bool      `json:"ok" desc:"最近一次采集是否可用"`
	Source             string    `json:"source" desc:"cmc 或 sample"`
	Message            string    `json:"message" desc:"可读说明"`
	KeyConfigured      bool      `json:"keyConfigured" desc:"是否配置了非占位密钥"`
	MaskedKey          string    `json:"maskedKey" desc:"打码后的密钥"`
	Listings           int       `json:"listings" desc:"写入币种数"`
	Quotes             int       `json:"quotes" desc:"quotes/latest 刷新条数"`
	HasGlobal          bool      `json:"hasGlobal" desc:"是否写入全球指标"`
	Credits            int       `json:"credits" desc:"本次声明的 CMC credit"`
	AlertHits          int       `json:"alertHits" desc:"本次触发告警数"`
	Endpoints          []string  `json:"endpoints" desc:"实际访问的 CMC 路径"`
	MonthlyCreditLimit int       `json:"monthlyCreditLimit,omitempty" desc:"套餐月度 credit"`
	RateLimitMinute    int       `json:"rateLimitMinute,omitempty" desc:"每分钟限速"`
	FetchedAt          time.Time `json:"fetchedAt" desc:"采集时间"`
}

// StatusResponse 是公开状态页 DTO。
type StatusResponse struct {
	Product        string              `json:"product" desc:"产品名"`
	Service        string              `json:"service" desc:"服务名"`
	Ingest         *IngestStatus       `json:"ingest" desc:"最近采集"`
	HasQuotes      bool                `json:"hasQuotes" desc:"库中是否已有行情"`
	QuoteCount     int                 `json:"quoteCount" desc:"行情行数"`
	LastByEndpoint []*EndpointSnapshot `json:"lastByEndpoint" desc:"各 CMC 端点最近一次采集"`
}
