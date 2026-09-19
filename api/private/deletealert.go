package private

import (
	"strconv"
	"strings"

	"github.com/digitalwayhk/core/pkg/server/router"
	servertypes "github.com/digitalwayhk/core/pkg/server/types"
	"github.com/digitalwayhk/pulse/api/dto"
	"github.com/digitalwayhk/pulse/models"
)

// DeleteAlert 按 ID 删除当前用户自己的告警规则。
type DeleteAlert struct {
	ID uint `json:"id,string"`
}

// Parse 绑定字符串或 JSON 形式的规则 ID。
func (own *DeleteAlert) Parse(req servertypes.IRequest) error {
	value := strings.TrimSpace(req.GetValue("id"))
	if value == "" {
		return req.Bind(own)
	}
	id, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		return models.NewBusinessError("告警不存在或无权操作")
	}
	own.ID = uint(id)
	return nil
}

// Validation 校验规则 ID 和可信登录身份。
func (own *DeleteAlert) Validation(req servertypes.IRequest) error {
	userID, _ := req.GetUser()
	if own.ID == 0 || strings.TrimSpace(userID) == "" {
		return models.NewBusinessError("告警不存在或无权操作")
	}
	return nil
}

// Do 使用 ID 与 UserID 组合条件删除。
func (own *DeleteAlert) Do(req servertypes.IRequest) (interface{}, error) {
	userID, _ := req.GetUser()
	item, err := models.NewAlertRule().FindOwned(own.ID, userID)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, models.NewBusinessError("告警不存在或无权操作")
	}
	if err := item.Delete(); err != nil {
		return nil, err
	}
	response := dto.NewAlertResponse(item)
	NotifyAlertChange(response.WithAction("deleted"))
	return response, nil
}

// GetResponse 返回 OpenAPI 用的删除告警成功响应结构。
func (own *DeleteAlert) GetResponse() interface{} { return dto.AlertResponse{} }

// RouterInfo 将告警删除注册为需要认证的 POST 路由。
func (own *DeleteAlert) RouterInfo() *servertypes.RouterInfo {
	return router.DefaultRouterInfo(own)
}
