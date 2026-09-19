package manage

import (
	"github.com/digitalwayhk/core/pkg/server/types"
	managepkg "github.com/digitalwayhk/core/service/manage"
	"github.com/digitalwayhk/core/service/manage/view"
	"github.com/digitalwayhk/pulse/models"
)

// AlertRuleManage 组装告警规则的管理 CRUD。
type AlertRuleManage struct {
	*managepkg.ManageService[models.AlertRule]
}

// NewAlertRuleManage 创建告警管理服务并传入正确的 hook owner。
func NewAlertRuleManage() *AlertRuleManage {
	own := &AlertRuleManage{}
	own.ManageService = managepkg.NewManageService[models.AlertRule](own)
	return own
}

// GetList 通过 models 层取得当前服务统一的 Manage 模型列表。
func (*AlertRuleManage) GetList() interface{} {
	return models.NewManageModelList[models.AlertRule]()
}

// Routers 暴露告警配置的标准 CRUD。
func (own *AlertRuleManage) Routers() []types.IRouter {
	return []types.IRouter{own.View, own.Search, own.Add, own.Edit, own.Remove}
}

// ValidationAfter 在通用唯一检查前执行告警模型校验。
func (own *AlertRuleManage) ValidationAfter(sender interface{}, _ types.IRequest) error {
	switch operation := sender.(type) {
	case *managepkg.Add[models.AlertRule]:
		if operation.Model != nil {
			return operation.Model.AddValid()
		}
	case *managepkg.Edit[models.AlertRule]:
		if operation.Model != nil {
			return operation.Model.UpdateValid(operation.OldItem)
		}
	}
	return nil
}

// ViewModel 设置管理界面的告警标题和自动加载行为。
func (own *AlertRuleManage) ViewModel(model *view.ViewModel) {
	model.Title = "AlertRule 告警配置"
	model.AutoLoad = true
}
