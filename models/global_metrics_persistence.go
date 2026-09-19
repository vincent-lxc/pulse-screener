package models

import persistencetypes "github.com/digitalwayhk/core/pkg/persistence/types"

// Upsert 写入或覆盖唯一的全球指标行。
func (own *GlobalMetrics) Upsert() error {
	own.SetHashcode(own.GetHash())
	existing, err := NewGlobalMetrics().FindLatest()
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

// FindLatest 读取当前全球指标。
func (own *GlobalMetrics) FindLatest() (*GlobalMetrics, error) {
	search := &persistencetypes.SearchItem{Page: 1, Size: 1, Model: own}
	var items []*GlobalMetrics
	if err := getDataAction().Load(search, &items); err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, nil
	}
	return items[0], nil
}
