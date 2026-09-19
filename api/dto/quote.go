package dto

import "github.com/digitalwayhk/pulse/models"

// QuoteDetailResponse 是点选详情对 CMC quotes/latest 的确认结果。
type QuoteDetailResponse struct {
	Endpoint string        `json:"endpoint" desc:"CMC 路径"`
	Source   string        `json:"source" desc:"cmc、sample 或 snapshot"`
	Live     bool          `json:"live" desc:"是否刚打过 CMC"`
	Message  string        `json:"message" desc:"评委可读说明"`
	Coin     *CoinResponse `json:"coin" desc:"确认后的行情"`
}

// NewQuoteDetailResponse 从查找结果创建详情 DTO。
func NewQuoteDetailResponse(endpoint, source, message string, live bool, coin *models.CoinQuote) *QuoteDetailResponse {
	return &QuoteDetailResponse{
		Endpoint: endpoint,
		Source:   source,
		Live:     live,
		Message:  message,
		Coin:     NewCoinResponse(coin),
	}
}

// EndpointSnapshot 记录某一 CMC 路径最近一次成功或失败。
type EndpointSnapshot struct {
	Endpoint  string `json:"endpoint" desc:"CMC 路径"`
	Status    string `json:"status" desc:"ok / error / sample"`
	Source    string `json:"source" desc:"cmc 或 sample"`
	ItemCount int    `json:"itemCount" desc:"写入条数"`
	FetchedAt string `json:"fetchedAt" desc:"最近时间"`
	Message   string `json:"message,omitempty" desc:"可读说明"`
}

// AlertCheckResponse 是手动评估告警的返回。
type AlertCheckResponse struct {
	Endpoint string           `json:"endpoint" desc:"CMC 路径"`
	Source   string           `json:"source" desc:"cmc、sample 或 snapshot"`
	Live     bool             `json:"live" desc:"是否刚打过 CMC"`
	Quotes   int              `json:"quotes" desc:"确认条数"`
	Hits     int              `json:"hits" desc:"本次触发数"`
	Message  string           `json:"message" desc:"评委可读说明"`
	Alerts   []*AlertResponse `json:"alerts" desc:"当前用户告警"`
}
