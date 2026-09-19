package models

import persistencetypes "github.com/digitalwayhk/core/pkg/persistence/types"

// Upsert 按 CMC ID 写入或更新最新行情快照。
func (own *CoinQuote) Upsert() error {
	if err := own.validate(); err != nil {
		return err
	}
	own.SetHashcode(own.GetHash())
	existing, err := NewCoinQuote().FindByCmcID(own.CmcID)
	if err != nil {
		return err
	}
	if existing != nil {
		own.SetID(existing.ID)
		if existing.CreatedAt != nil {
			own.SetCreatedAt(*existing.CreatedAt)
		}
		own.SetUpdatedAt(own.FetchedAt)
		return getDataAction().Update(own)
	}
	own.SetCreatedAt(own.FetchedAt)
	own.SetUpdatedAt(own.FetchedAt)
	return getDataAction().Insert(own)
}

// FindByCmcID 按 CoinMarketCap ID 查找最新行情。
func (own *CoinQuote) FindByCmcID(cmcID int) (*CoinQuote, error) {
	if cmcID <= 0 {
		return nil, nil
	}
	search := newCoinSearch(own, 1)
	search.AddWhereN("CmcID", cmcID)
	var items []*CoinQuote
	if err := getDataAction().Load(search, &items); err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, nil
	}
	return items[0], nil
}

// FindBySymbol 按交易代码查找最新行情。
func (own *CoinQuote) FindBySymbol(symbol string) (*CoinQuote, error) {
	symbol = stringsToUpper(symbol)
	if symbol == "" {
		return nil, nil
	}
	search := newCoinSearch(own, 1)
	search.AddWhereN("Symbol", symbol)
	var items []*CoinQuote
	if err := getDataAction().Load(search, &items); err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, nil
	}
	return items[0], nil
}

// QueryLatest 返回当前库中的全部最新行情，最多 500 条。
func (own *CoinQuote) QueryLatest() ([]*CoinQuote, error) {
	search := newCoinSearch(own, 500)
	var items []*CoinQuote
	err := getDataAction().Load(search, &items)
	return items, err
}

func newCoinSearch(model *CoinQuote, size int) *persistencetypes.SearchItem {
	return &persistencetypes.SearchItem{Page: 1, Size: size, Model: model}
}

func stringsToUpper(value string) string {
	out := make([]rune, 0, len(value))
	for _, r := range value {
		if r >= 'a' && r <= 'z' {
			out = append(out, r-32)
			continue
		}
		if r != ' ' && r != '\t' {
			out = append(out, r)
		}
	}
	return string(out)
}
