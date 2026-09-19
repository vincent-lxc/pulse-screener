package models

import (
	"strings"
	"time"

	persistencetypes "github.com/digitalwayhk/core/pkg/persistence/types"
)

// Insert 规范化哈希后写入告警规则。
func (own *AlertRule) Insert() error {
	if err := own.validate(); err != nil {
		return err
	}
	own.SetHashcode(own.GetHash())
	return getDataAction().Insert(own)
}

// Update 保存告警规则的触发状态或配置变更。
func (own *AlertRule) Update() error {
	if err := own.validate(); err != nil {
		return err
	}
	own.SetHashcode(own.GetHash())
	own.SetUpdatedAt(time.Now().UTC())
	return getDataAction().Update(own)
}

// Delete 物理删除已查询的告警规则。
func (own *AlertRule) Delete() error {
	return getDataAction().Delete(own)
}

// QueryByUser 查询指定用户的告警规则。
func (own *AlertRule) QueryByUser(userID string) ([]*AlertRule, error) {
	search := newAlertSearch(own, 200)
	search.AddWhereN("UserID", strings.TrimSpace(userID))
	var items []*AlertRule
	err := getDataAction().Load(search, &items)
	return items, err
}

// QueryEnabled 查询全部启用中的告警规则，供采集后评估。
func (own *AlertRule) QueryEnabled() ([]*AlertRule, error) {
	search := newAlertSearch(own, 500)
	search.AddWhereN("Enabled", true)
	var items []*AlertRule
	err := getDataAction().Load(search, &items)
	return items, err
}

// FindOwned 使用 ID 与用户 ID 组合条件查找当前用户的告警规则。
func (own *AlertRule) FindOwned(id uint, userID string) (*AlertRule, error) {
	search := newAlertSearch(own, 1)
	search.AddWhereN("ID", id)
	search.AddWhereN("UserID", strings.TrimSpace(userID))
	var items []*AlertRule
	if err := getDataAction().Load(search, &items); err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, nil
	}
	return items[0], nil
}

// QueryCmcIDs 返回启用中告警规则的 CMC ID，供 quotes/latest 二次确认。
func (own *AlertRule) QueryCmcIDs() ([]int, error) {
	search := newAlertSearch(own, 500)
	search.AddWhereN("Enabled", true)
	var items []*AlertRule
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

func newAlertSearch(model *AlertRule, size int) *persistencetypes.SearchItem {
	return &persistencetypes.SearchItem{Page: 1, Size: size, Model: model}
}
