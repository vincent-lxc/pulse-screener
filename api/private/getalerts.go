package private

import (
	"net/http"
	"strings"

	"github.com/digitalwayhk/core/pkg/server/router"
	servertypes "github.com/digitalwayhk/core/pkg/server/types"
	"github.com/digitalwayhk/core/pkg/utils"
	"github.com/digitalwayhk/pulse/api/dto"
	"github.com/digitalwayhk/pulse/models"
)

// NotifyAlertChange 通过已注册的告警 RouterInfo 发布用户通知。
func NotifyAlertChange(response *dto.AlertResponse) {
	info := (&GetAlerts{}).RouterInfo()
	if info != nil {
		info.NoticeWebSocket(response)
	}
}

// GetAlerts 查询当前登录用户的告警规则，并作为该用户的 WebSocket 订阅路由。
type GetAlerts struct {
	subscriptionUserID string
}

// Parse 不接受客户端身份参数。
func (own *GetAlerts) Parse(servertypes.IRequest) error { return nil }

// Validation 验证 HTTP 认证上下文或 WebSocket 订阅会话中的用户身份。
func (own *GetAlerts) Validation(req servertypes.IRequest) error {
	if own.resolveUserID(req) == "" {
		return models.NewBusinessError("用户身份无效")
	}
	return nil
}

// Do 只按可信请求或订阅身份查询告警。
func (own *GetAlerts) Do(req servertypes.IRequest) (interface{}, error) {
	items, err := models.NewAlertRule().QueryByUser(own.resolveUserID(req))
	if err != nil {
		return nil, err
	}
	return dto.AlertResponses(items), nil
}

// GetResponse 返回 OpenAPI 用的告警列表成功响应结构。
func (own *GetAlerts) GetResponse() interface{} { return []*dto.AlertResponse{} }

// RouterInfo 将告警查询注册为需要认证的 GET 路由。
func (own *GetAlerts) RouterInfo() *servertypes.RouterInfo {
	return router.DefaultRouterInfoWithOptions(own, router.WithMethod(http.MethodGet))
}

func (own *GetAlerts) resolveUserID(req servertypes.IRequest) string {
	if req != nil {
		if userID, _ := req.GetUser(); strings.TrimSpace(userID) != "" {
			return strings.TrimSpace(userID)
		}
	}
	return strings.TrimSpace(own.subscriptionUserID)
}

// SetUserID 实现 IWebSocketUserIdentity，接收 WebSocket 会话解析出的可信身份。
func (own *GetAlerts) SetUserID(userID, _ string) {
	own.subscriptionUserID = strings.TrimSpace(userID)
}

// GetUserID 实现 IWebSocketUserIdentity，返回当前订阅绑定的可信用户 ID。
func (own *GetAlerts) GetUserID() string { return own.subscriptionUserID }

// GetHashKey 实现 IRouterHashKey，以用户 ID 隔离不同用户的订阅组。
func (own *GetAlerts) GetHashKey() uint64 {
	return utils.HashCode64(own.subscriptionUserID)
}

// NoticeFiltersRouter 实现 IWebSocketRouterNotice，只向规则所属用户投递事件。
func (own *GetAlerts) NoticeFiltersRouter(message interface{}, api servertypes.IRouter) (bool, interface{}) {
	response, ok := message.(*dto.AlertResponse)
	if !ok || response == nil {
		return false, nil
	}
	subscription, ok := api.(*GetAlerts)
	if !ok || subscription.subscriptionUserID == "" {
		return false, nil
	}
	if response.UserID != subscription.subscriptionUserID {
		return false, nil
	}
	return true, response
}
