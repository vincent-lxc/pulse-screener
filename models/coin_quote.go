package models

import (
	"strconv"
	"strings"
	"time"

	"github.com/digitalwayhk/core/pkg/utils"
)

// CoinQuote 保存一枚币种的最新 CMC 行情快照，按 CmcID 唯一。
type CoinQuote struct {
	*BusinessModel
	CmcID              int       `json:"cmcID" desc:"CoinMarketCap 稳定币种 ID"`
	Name               string    `json:"name" desc:"币种名称"`
	Symbol             string    `json:"symbol" desc:"交易代码"`
	Slug               string    `json:"slug" desc:"CMC slug"`
	Rank               int       `json:"rank" desc:"市值排名"`
	PriceUSD           float64   `json:"priceUSD" desc:"USD 价格"`
	Volume24h          float64   `json:"volume24h" desc:"24 小时成交额 USD"`
	MarketCap          float64   `json:"marketCap" desc:"流通市值 USD"`
	PercentChange1h    float64   `json:"percentChange1h" desc:"1 小时涨跌幅 %"`
	PercentChange24h   float64   `json:"percentChange24h" desc:"24 小时涨跌幅 %"`
	PercentChange7d    float64   `json:"percentChange7d" desc:"7 日涨跌幅 %"`
	CirculatingSupply  float64   `json:"circulatingSupply" desc:"流通量"`
	TotalSupply        float64   `json:"totalSupply" desc:"总量"`
	MaxSupply          float64   `json:"maxSupply" desc:"最大供给"`
	MarketCapDominance float64   `json:"marketCapDominance" desc:"市值占比 %"`
	LastUpdated        string    `json:"lastUpdated" desc:"CMC 行情更新时间"`
	FetchedAt          time.Time `json:"fetchedAt" desc:"本地采集时间"`
	Source             string    `json:"source" desc:"数据来源 cmc 或 sample"`
}

// NewCoinQuote 创建已初始化继承链的行情快照。
func NewCoinQuote() *CoinQuote {
	return &CoinQuote{BusinessModel: newBusinessModel()}
}

// NewModel 供 ModelList 在反射创建行情快照时初始化继承链。
func (own *CoinQuote) NewModel() {
	if own.BusinessModel == nil || own.ServiceModel == nil || own.Model == nil {
		fresh := NewCoinQuote()
		own.BusinessModel = fresh.BusinessModel
	}
}

// GetHash 以 CMC 稳定 ID 作为唯一哈希，保证每币一行。
func (own *CoinQuote) GetHash() string {
	if own.CmcID <= 0 {
		if own.Model != nil {
			return own.Hashcode
		}
		return ""
	}
	return utils.HashCodes("cmc:" + strconv.Itoa(own.CmcID))
}

// AddValid 校验行情快照写入约束。
func (own *CoinQuote) AddValid() error {
	return own.validate()
}

// UpdateValid 校验行情快照更新约束。
func (own *CoinQuote) UpdateValid(_ interface{}) error {
	return own.validate()
}

func (own *CoinQuote) validate() error {
	own.Symbol = strings.ToUpper(strings.TrimSpace(own.Symbol))
	own.Name = strings.TrimSpace(own.Name)
	if own.CmcID <= 0 {
		return NewValidationError("CMC 币种 ID 必须大于 0")
	}
	if own.Symbol == "" || own.Name == "" {
		return NewValidationError("币种名称和交易代码不能为空")
	}
	return nil
}
