package manage

import (
	"github.com/digitalwayhk/core/pkg/server/types"
	managepkg "github.com/digitalwayhk/core/service/manage"
	"github.com/digitalwayhk/core/service/manage/view"
	"github.com/digitalwayhk/pulse/models"
)

// GlobalMetricsManage 只读展示全球市场指标快照。
type GlobalMetricsManage struct {
	*managepkg.ManageService[models.GlobalMetrics]
}

// NewGlobalMetricsManage 创建全球指标管理服务并传入正确的 hook owner。
func NewGlobalMetricsManage() *GlobalMetricsManage {
	own := &GlobalMetricsManage{}
	own.ManageService = managepkg.NewManageService[models.GlobalMetrics](own)
	return own
}

// GetList 通过 models 层取得当前服务统一的 Manage 模型列表。
func (*GlobalMetricsManage) GetList() interface{} {
	return models.NewManageModelList[models.GlobalMetrics]()
}

// Routers 仅暴露 View 与 Search。
func (own *GlobalMetricsManage) Routers() []types.IRouter {
	return []types.IRouter{own.View, own.Search}
}

// ViewModel 设置管理界面的全球指标标题和自动加载行为。
func (own *GlobalMetricsManage) ViewModel(model *view.ViewModel) {
	model.Title = "CMC 全球指标"
	model.AutoLoad = true
}
