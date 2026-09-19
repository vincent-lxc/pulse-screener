package models

import (
	"strconv"
	"strings"

	"github.com/digitalwayhk/core/pkg/utils"
)

// Watchlist 是演示用户自选清单，按用户与 CMC ID 唯一。
type Watchlist struct {
	*BusinessModel
	UserID string `json:"userID" desc:"用户 ID"`
	CmcID  int    `json:"cmcID" desc:"CoinMarketCap 币种 ID"`
	Symbol string `json:"symbol" desc:"交易代码快照"`
	Name   string `json:"name" desc:"币种名称快照"`
	Note   string `json:"note" desc:"备注"`
}

// NewWatchlist 创建已初始化继承链的自选。
func NewWatchlist() *Watchlist {
	return &Watchlist{BusinessModel: newBusinessModel()}
}

// NewModel 供 ModelList 反射创建自选时初始化继承链。
func (own *Watchlist) NewModel() {
	if own.BusinessModel == nil || own.ServiceModel == nil || own.Model == nil {
		fresh := NewWatchlist()
		own.BusinessModel = fresh.BusinessModel
	}
}

// GetHash 以用户与 CMC ID 生成唯一哈希。
func (own *Watchlist) GetHash() string {
	userID := strings.TrimSpace(own.UserID)
	if userID == "" || own.CmcID <= 0 {
		if own.Model != nil {
			return own.Hashcode
		}
		return ""
	}
	return utils.HashCodes(userID + ":" + strconv.Itoa(own.CmcID))
}

// AddValid 校验新增自选。
func (own *Watchlist) AddValid() error {
	return own.validate()
}

// UpdateValid 校验修改自选。
func (own *Watchlist) UpdateValid(_ interface{}) error {
	return own.validate()
}

func (own *Watchlist) validate() error {
	own.UserID = strings.TrimSpace(own.UserID)
	own.Symbol = strings.ToUpper(strings.TrimSpace(own.Symbol))
	own.Name = strings.TrimSpace(own.Name)
	own.Note = strings.TrimSpace(own.Note)
	if own.UserID == "" {
		return NewValidationError("用户 ID 不能为空")
	}
	if own.CmcID <= 0 {
		return NewValidationError("CMC 币种 ID 必须大于 0")
	}
	return nil
}
