package models

import (
	"time"

	"github.com/digitalwayhk/core/pkg/utils"
)

// GlobalMetrics 保存全市场最新汇总指标，全表仅保留一行。
type GlobalMetrics struct {
	*BusinessModel
	ActiveCryptocurrencies int       `json:"activeCryptocurrencies" desc:"活跃币种数"`
	ActiveExchanges        int       `json:"activeExchanges" desc:"活跃交易所数"`
	BtcDominance           float64   `json:"btcDominance" desc:"BTC 市值占比 %"`
	EthDominance           float64   `json:"ethDominance" desc:"ETH 市值占比 %"`
	TotalMarketCap         float64   `json:"totalMarketCap" desc:"总市值 USD"`
	TotalVolume24h         float64   `json:"totalVolume24h" desc:"24 小时总成交额 USD"`
	LastUpdated            string    `json:"lastUpdated" desc:"CMC 更新时间"`
	FetchedAt              time.Time `json:"fetchedAt" desc:"本地采集时间"`
	Source                 string    `json:"source" desc:"数据来源 cmc 或 sample"`
}

// NewGlobalMetrics 创建已初始化继承链的全球指标。
func NewGlobalMetrics() *GlobalMetrics {
	return &GlobalMetrics{BusinessModel: newBusinessModel()}
}

// NewModel 供 ModelList 反射创建全球指标时初始化继承链。
func (own *GlobalMetrics) NewModel() {
	if own.BusinessModel == nil || own.ServiceModel == nil || own.Model == nil {
		fresh := NewGlobalMetrics()
		own.BusinessModel = fresh.BusinessModel
	}
}

// GetHash 全球指标使用固定哈希，保证仅一行。
func (own *GlobalMetrics) GetHash() string {
	return utils.HashCodes("global-metrics")
}
