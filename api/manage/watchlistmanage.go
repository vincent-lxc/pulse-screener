package manage

import (
	"github.com/digitalwayhk/core/pkg/server/types"
	managepkg "github.com/digitalwayhk/core/service/manage"
	"github.com/digitalwayhk/core/service/manage/view"
	"github.com/digitalwayhk/pulse/models"
)

// WatchlistManage 组装自选清单的管理 CRUD，便于评委在 -view 后台查看演示数据。
type WatchlistManage struct {
	*managepkg.ManageService[models.Watchlist]
}

// NewWatchlistManage 创建自选管理服务并传入正确的 hook owner。
func NewWatchlistManage() *WatchlistManage {
	own := &WatchlistManage{}
	own.ManageService = managepkg.NewManageService[models.Watchlist](own)
	return own
}

// GetList 通过 models 层取得当前服务统一的 Manage 模型列表。
func (*WatchlistManage) GetList() interface{} {
	return models.NewManageModelList[models.Watchlist]()
}

// Routers 暴露自选配置的标准 CRUD。
func (own *WatchlistManage) Routers() []types.IRouter {
	return []types.IRouter{own.View, own.Search, own.Add, own.Edit, own.Remove}
}

// ValidationAfter 在通用唯一检查前执行自选模型校验。
func (own *WatchlistManage) ValidationAfter(sender interface{}, _ types.IRequest) error {
	switch operation := sender.(type) {
	case *managepkg.Add[models.Watchlist]:
		if operation.Model != nil {
			return operation.Model.AddValid()
		}
	case *managepkg.Edit[models.Watchlist]:
		if operation.Model != nil {
			return operation.Model.UpdateValid(operation.OldItem)
		}
	}
	return nil
}

// ViewModel 设置管理界面的自选标题和自动加载行为。
func (own *WatchlistManage) ViewModel(model *view.ViewModel) {
	model.Title = "Watchlist 自选配置"
	model.AutoLoad = true
}
