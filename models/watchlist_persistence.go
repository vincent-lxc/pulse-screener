package models

import (
	"strings"

	persistencetypes "github.com/digitalwayhk/core/pkg/persistence/types"
)

// Insert 规范化哈希后写入自选。
func (own *Watchlist) Insert() error {
	if err := own.validate(); err != nil {
		return err
	}
	own.SetHashcode(own.GetHash())
	exists, err := own.ExistsForUser(own.UserID, own.CmcID, 0)
	if err != nil {
		return err
	}
	if exists {
		return NewBusinessError("该币种已在自选清单中")
	}
	return getDataAction().Insert(own)
}

// Delete 物理删除已查询的自选。
func (own *Watchlist) Delete() error {
	return getDataAction().Delete(own)
}

// QueryByUser 查询指定用户的自选清单。
func (own *Watchlist) QueryByUser(userID string) ([]*Watchlist, error) {
	search := newWatchSearch(own, 200)
	search.AddWhereN("UserID", strings.TrimSpace(userID))
	var items []*Watchlist
	err := getDataAction().Load(search, &items)
	return items, err
}

// FindOwned 使用 ID 与用户 ID 组合条件查找当前用户的自选。
func (own *Watchlist) FindOwned(id uint, userID string) (*Watchlist, error) {
	search := newWatchSearch(own, 1)
	search.AddWhereN("ID", id)
	search.AddWhereN("UserID", strings.TrimSpace(userID))
	var items []*Watchlist
	if err := getDataAction().Load(search, &items); err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, nil
	}
	return items[0], nil
}

// ExistsForUser 检查用户是否已关注该币种。
func (own *Watchlist) ExistsForUser(userID string, cmcID int, excludeID uint) (bool, error) {
	probe := NewWatchlist()
	probe.UserID = strings.TrimSpace(userID)
	probe.CmcID = cmcID
	search := newWatchSearch(own, 2)
	search.AddWhereN("Hashcode", probe.GetHash())
	var items []*Watchlist
	if err := getDataAction().Load(search, &items); err != nil {
		return false, err
	}
	for _, item := range items {
		if item != nil && item.ID != excludeID {
			return true, nil
		}
	}
	return false, nil
}

// QueryCmcIDs 返回当前库中全部自选的 CMC ID，供 quotes/latest 批量刷新。
func (own *Watchlist) QueryCmcIDs() ([]int, error) {
	search := newWatchSearch(own, 200)
	var items []*Watchlist
	if err := getDataAction().Load(search, &items); err != nil {
		return nil, err
	}
	seen := make(map[int]struct{}, len(items))
	ids := make([]int, 0, len(items))
	for _, item := range items {
		if item == nil || item.CmcID <= 0 {
			continue
		}
		if _, ok := seen[item.CmcID]; ok {
			continue
		}
		seen[item.CmcID] = struct{}{}
		ids = append(ids, item.CmcID)
	}
	return ids, nil
}

func newWatchSearch(model *Watchlist, size int) *persistencetypes.SearchItem {
	return &persistencetypes.SearchItem{Page: 1, Size: size, Model: model}
}
