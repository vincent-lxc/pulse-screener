package private

import (
	"strings"

	"github.com/digitalwayhk/core/pkg/server/router"
	servertypes "github.com/digitalwayhk/core/pkg/server/types"
	"github.com/digitalwayhk/pulse/api/dto"
	"github.com/digitalwayhk/pulse/models"
)

// AddWatchlist 将一枚币加入当前登录用户的自选清单。
type AddWatchlist struct {
	CmcID  int    `json:"cmcID"`
	Symbol string `json:"symbol"`
	Note   string `json:"note"`
}

// Parse 绑定自选参数，UserID 不属于客户端可提交字段。
func (own *AddWatchlist) Parse(req servertypes.IRequest) error {
	return req.Bind(own)
}

// Validation 校验登录身份与币种标识。
func (own *AddWatchlist) Validation(req servertypes.IRequest) error {
	userID, _ := req.GetUser()
	if strings.TrimSpace(userID) == "" {
		return models.NewBusinessError("用户身份无效")
	}
	if own.CmcID <= 0 && strings.TrimSpace(own.Symbol) == "" {
		return models.NewValidationError("请提供 cmcID 或 symbol")
	}
	return nil
}

// Do 用本地行情快照补全名称后写入自选。
func (own *AddWatchlist) Do(req servertypes.IRequest) (interface{}, error) {
	quote, err := resolveQuote(own.CmcID, own.Symbol)
	if err != nil {
		return nil, err
	}
	if quote == nil {
		return nil, models.NewBusinessError("本地还没有该币种行情，请先等待 CMC 采集完成")
	}
	item := models.NewWatchlist()
	item.SetID(req.NewID())
	item.UserID, _ = req.GetUser()
	item.CmcID = quote.CmcID
	item.Symbol = quote.Symbol
	item.Name = quote.Name
	item.Note = strings.TrimSpace(own.Note)
	if err := item.Insert(); err != nil {
		return nil, err
	}
	return dto.NewWatchResponse(item, quote), nil
}

// GetResponse 返回 OpenAPI 用的新增自选成功响应结构。
func (own *AddWatchlist) GetResponse() interface{} { return dto.WatchResponse{} }

// RouterInfo 将新增自选注册为需要认证的 POST 路由。
func (own *AddWatchlist) RouterInfo() *servertypes.RouterInfo {
	return router.DefaultRouterInfo(own)
}

func resolveQuote(cmcID int, symbol string) (*models.CoinQuote, error) {
	if cmcID > 0 {
		return models.NewCoinQuote().FindByCmcID(cmcID)
	}
	return models.NewCoinQuote().FindBySymbol(symbol)
}
