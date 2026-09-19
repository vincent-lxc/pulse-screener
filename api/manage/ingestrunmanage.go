package manage

import (
	"github.com/digitalwayhk/core/pkg/server/types"
	managepkg "github.com/digitalwayhk/core/service/manage"
	"github.com/digitalwayhk/core/service/manage/view"
	"github.com/digitalwayhk/pulse/models"
)

// IngestRunManage 只读展示 CMC 采集审计，作为黑客松演示证据。
type IngestRunManage struct {
	*managepkg.ManageService[models.IngestRun]
}

// NewIngestRunManage 创建采集审计管理服务并传入正确的 hook owner。
func NewIngestRunManage() *IngestRunManage {
	own := &IngestRunManage{}
	own.ManageService = managepkg.NewManageService[models.IngestRun](own)
	return own
}

// GetList 通过 models 层取得当前服务统一的 Manage 模型列表。
func (*IngestRunManage) GetList() interface{} {
	return models.NewManageModelList[models.IngestRun]()
}

// Routers 仅暴露 View 与 Search。
func (own *IngestRunManage) Routers() []types.IRouter {
	return []types.IRouter{own.View, own.Search}
}

// ViewModel 设置管理界面的采集审计标题和自动加载行为。
func (own *IngestRunManage) ViewModel(model *view.ViewModel) {
	model.Title = "CMC 采集审计"
	model.AutoLoad = true
}
