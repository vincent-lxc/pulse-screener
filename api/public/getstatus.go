package public

import (
	"net/http"
	"sort"

	"github.com/digitalwayhk/core/pkg/server/router"
	servertypes "github.com/digitalwayhk/core/pkg/server/types"
	"github.com/digitalwayhk/pulse/api/dto"
	"github.com/digitalwayhk/pulse/contract"
	"github.com/digitalwayhk/pulse/ingest"
	"github.com/digitalwayhk/pulse/models"
)

// GetStatus 返回采集状态、密钥是否配置以及本地快照是否可用。
type GetStatus struct{}

// Parse 无查询参数。
func (own *GetStatus) Parse(servertypes.IRequest) error { return nil }

// Validation 始终通过。
func (own *GetStatus) Validation(servertypes.IRequest) error { return nil }

// Do 组合采集器状态与本地行情计数。
func (own *GetStatus) Do(servertypes.IRequest) (interface{}, error) {
	items, err := models.NewCoinQuote().QueryLatest()
	if err != nil {
		return nil, err
	}
	runs, _ := models.NewIngestRun().QueryRecent(80)
	return dto.StatusResponse{
		Product:        contract.ProductName,
		Service:        contract.ServiceName,
		Ingest:         ingest.Status(),
		HasQuotes:      len(items) > 0,
		QuoteCount:     len(items),
		LastByEndpoint: endpointSnapshots(runs),
	}, nil
}

func endpointSnapshots(runs []*models.IngestRun) []*dto.EndpointSnapshot {
	latest := models.LatestByEndpoint(runs)
	out := make([]*dto.EndpointSnapshot, 0, len(latest))
	for _, run := range latest {
		if run == nil {
			continue
		}
		out = append(out, &dto.EndpointSnapshot{
			Endpoint:  run.Endpoint,
			Status:    run.Status,
			Source:    run.Source,
			ItemCount: run.ItemCount,
			FetchedAt: run.FetchedAt.UTC().Format("2006-01-02T15:04:05Z"),
			Message:   run.Message,
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Endpoint < out[j].Endpoint })
	return out
}

// GetResponse 返回 OpenAPI 用的状态成功响应结构。
func (own *GetStatus) GetResponse() interface{} { return dto.StatusResponse{} }

// RouterInfo 将状态查询注册为公开 GET 路由。
func (own *GetStatus) RouterInfo() *servertypes.RouterInfo {
	return router.DefaultRouterInfoWithOptions(own, router.WithMethod(http.MethodGet))
}
