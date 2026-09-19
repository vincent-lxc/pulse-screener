package private

import (
	"net/http"
	"strings"

	"github.com/digitalwayhk/core/pkg/server/router"
	servertypes "github.com/digitalwayhk/core/pkg/server/types"
	"github.com/digitalwayhk/pulse/api/dto"
	"github.com/digitalwayhk/pulse/models"
)

// GetWatchlist 查询当前登录用户的自选清单。
type GetWatchlist struct{}

// Parse 不接受客户端身份参数。
func (own *GetWatchlist) Parse(servertypes.IRequest) error { return nil }

// Validation 验证登录身份。
func (own *GetWatchlist) Validation(req servertypes.IRequest) error {
	userID, _ := req.GetUser()
	if strings.TrimSpace(userID) == "" {
		return models.NewBusinessError("用户身份无效")
	}
	return nil
}

// Do 只按令牌身份查询自选。
func (own *GetWatchlist) Do(req servertypes.IRequest) (interface{}, error) {
	userID, _ := req.GetUser()
	items, err := models.NewWatchlist().QueryByUser(userID)
	if err != nil {
		return nil, err
	}
	return dto.WatchResponses(items), nil
}

// GetResponse 返回 OpenAPI 用的自选列表成功响应结构。
func (own *GetWatchlist) GetResponse() interface{} { return []*dto.WatchResponse{} }

// RouterInfo 将自选查询注册为需要认证的 GET 路由。
func (own *GetWatchlist) RouterInfo() *servertypes.RouterInfo {
	return router.DefaultRouterInfoWithOptions(own, router.WithMethod(http.MethodGet))
}
