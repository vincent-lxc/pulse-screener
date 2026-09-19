package public

import (
	"net/http"

	"github.com/digitalwayhk/core/pkg/server/router"
	servertypes "github.com/digitalwayhk/core/pkg/server/types"
	"github.com/digitalwayhk/pulse/api/dto"
	"github.com/digitalwayhk/pulse/models"
)

// GetGlobal 返回最新全球市场指标。
type GetGlobal struct{}

// Parse 无查询参数。
func (own *GetGlobal) Parse(servertypes.IRequest) error { return nil }

// Validation 始终通过。
func (own *GetGlobal) Validation(servertypes.IRequest) error { return nil }

// Do 读取本地全球指标快照。
func (own *GetGlobal) Do(servertypes.IRequest) (interface{}, error) {
	item, err := models.NewGlobalMetrics().FindLatest()
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, models.NewBusinessError("还没有全球指标快照。请配置 CMC_API_KEY 后等待采集，或查看 /api/pulse/getstatus")
	}
	return dto.NewGlobalResponse(item), nil
}

// GetResponse 返回 OpenAPI 用的全球指标成功响应结构。
func (own *GetGlobal) GetResponse() interface{} { return dto.GlobalResponse{} }

// RouterInfo 将全球指标注册为公开 GET 路由。
func (own *GetGlobal) RouterInfo() *servertypes.RouterInfo {
	return router.DefaultRouterInfoWithOptions(own, router.WithMethod(http.MethodGet))
}

// GetCacheKey 实现 IRouterCacheKey。
func (own *GetGlobal) GetCacheKey() string { return "global" }
