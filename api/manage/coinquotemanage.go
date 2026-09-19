package manage

import (
	"github.com/digitalwayhk/core/pkg/server/types"
	managepkg "github.com/digitalwayhk/core/service/manage"
	"github.com/digitalwayhk/core/service/manage/view"
	"github.com/digitalwayhk/pulse/models"
)

// CoinQuoteManage 只读展示 CMC 行情快照，供评委核对采集结果。
type CoinQuoteManage struct {
	*managepkg.ManageService[models.CoinQuote]
}

// NewCoinQuoteManage 创建行情快照管理服务并传入正确的 hook owner。
func NewCoinQuoteManage() *CoinQuoteManage {
	own := &CoinQuoteManage{}
	own.ManageService = managepkg.NewManageService[models.CoinQuote](own)
	return own
}

// GetList 通过 models 层取得当前服务统一的 Manage 模型列表。
func (*CoinQuoteManage) GetList() interface{} {
	return models.NewManageModelList[models.CoinQuote]()
}

// Routers 仅暴露 View 与 Search，避免后台手改 CMC 快照。
func (own *CoinQuoteManage) Routers() []types.IRouter {
	return []types.IRouter{own.View, own.Search}
}

// ViewModel 设置管理界面的行情标题和自动加载行为。
func (own *CoinQuoteManage) ViewModel(model *view.ViewModel) {
	model.Title = "CMC 行情快照"
	model.AutoLoad = true
}
