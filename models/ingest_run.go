package models

import (
	"strconv"
	"time"

	persistencetypes "github.com/digitalwayhk/core/pkg/persistence/types"
	"github.com/digitalwayhk/core/pkg/utils"
)

// IngestRun 记录一次 CMC 采集的审计结果，便于评委核对端点与失败原因。
type IngestRun struct {
	*BusinessModel
	Endpoint    string    `json:"endpoint" desc:"CMC 路径"`
	Status      string    `json:"status" desc:"ok / error / sample"`
	Message     string    `json:"message" desc:"可读结果说明"`
	ItemCount   int       `json:"itemCount" desc:"写入条数"`
	CreditCount int       `json:"creditCount" desc:"CMC 声明的 credit 消耗"`
	DurationMs  int64     `json:"durationMs" desc:"耗时毫秒"`
	Source      string    `json:"source" desc:"cmc 或 sample"`
	FetchedAt   time.Time `json:"fetchedAt" desc:"采集时间"`
}

// NewIngestRun 创建已初始化继承链的采集记录。
func NewIngestRun() *IngestRun {
	return &IngestRun{BusinessModel: newBusinessModel()}
}

// NewModel 供 ModelList 反射创建采集记录时初始化继承链。
func (own *IngestRun) NewModel() {
	if own.BusinessModel == nil || own.ServiceModel == nil || own.Model == nil {
		fresh := NewIngestRun()
		own.BusinessModel = fresh.BusinessModel
	}
}

// GetHash 以端点、时间和状态生成追加型哈希。
func (own *IngestRun) GetHash() string {
	stamp := own.FetchedAt.UTC().Format(time.RFC3339Nano)
	return utils.HashCodes(own.Endpoint + ":" + stamp + ":" + strconv.FormatInt(own.DurationMs, 10))
}

// Insert 写入一条采集审计。
func (own *IngestRun) Insert() error {
	if own.FetchedAt.IsZero() {
		own.FetchedAt = time.Now().UTC()
	}
	own.SetHashcode(own.GetHash())
	own.SetCreatedAt(own.FetchedAt)
	own.SetUpdatedAt(own.FetchedAt)
	return getDataAction().Insert(own)
}

// QueryRecent 返回最近的采集记录。
func (own *IngestRun) QueryRecent(limit int) ([]*IngestRun, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	search := newIngestSearch(own, limit)
	var items []*IngestRun
	err := getDataAction().Load(search, &items)
	return items, err
}

// LatestByEndpoint 按端点取最近一条采集记录（成功或失败都保留）。
func LatestByEndpoint(items []*IngestRun) map[string]*IngestRun {
	latest := make(map[string]*IngestRun, len(items))
	for _, item := range items {
		if item == nil || item.Endpoint == "" {
			continue
		}
		prev := latest[item.Endpoint]
		if prev == nil || item.FetchedAt.After(prev.FetchedAt) {
			latest[item.Endpoint] = item
		}
	}
	return latest
}

func newIngestSearch(model *IngestRun, size int) *persistencetypes.SearchItem {
	return &persistencetypes.SearchItem{Page: 1, Size: size, Model: model}
}
