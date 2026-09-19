package pulse

import (
	"context"
	"errors"
	"time"

	"github.com/digitalwayhk/core/pkg/server/types"
	"github.com/digitalwayhk/pulse/api/dto"
	"github.com/digitalwayhk/pulse/api/manage"
	privateapi "github.com/digitalwayhk/pulse/api/private"
	publicapi "github.com/digitalwayhk/pulse/api/public"
	"github.com/digitalwayhk/pulse/contract"
	"github.com/digitalwayhk/pulse/ingest"
)

// PulseService 组装行情采集服务的管理、公开和用户私有路由。
type PulseService struct{}

// OnAuth 在框架签发 Access Token 前注入服务名 Claim。
func (*PulseService) OnAuth(_ context.Context, args *types.AuthHookArgs) error {
	if args == nil || args.Claims == nil {
		return errors.New("认证钩子参数不完整")
	}
	args.Claims.AddData("example_service", contract.ServiceName)
	return nil
}

// ServiceName 返回配置和路由共同使用的稳定服务名。
func (*PulseService) ServiceName() string { return contract.ServiceName }

// Routers 返回 Pulse 全部路由。
func (*PulseService) Routers() []types.IRouter {
	routers := make([]types.IRouter, 0, 32)
	routers = append(routers, manage.NewWatchlistManage().Routers()...)
	routers = append(routers, manage.NewAlertRuleManage().Routers()...)
	routers = append(routers, manage.NewCoinQuoteManage().Routers()...)
	routers = append(routers, manage.NewGlobalMetricsManage().Routers()...)
	routers = append(routers, manage.NewIngestRunManage().Routers()...)
	routers = append(routers,
		&publicapi.GetCoins{},
		&publicapi.GetScreener{},
		&publicapi.GetGlobal{},
		&publicapi.GetQuote{},
		&publicapi.GetStatus{},
		&privateapi.AddWatchlist{},
		&privateapi.GetWatchlist{},
		&privateapi.DeleteWatchlist{},
		&privateapi.AddAlert{},
		&privateapi.GetAlerts{},
		&privateapi.CheckAlerts{},
		&privateapi.DeleteAlert{},
	)
	return routers
}

// Start 在服务启动完成后启用公开接口缓存并开始 CMC 采集。
func (*PulseService) Start() {
	if info := (&publicapi.GetCoins{}).RouterInfo(); info != nil {
		info.UseCache(30 * time.Second)
	}
	if info := (&publicapi.GetScreener{}).RouterInfo(); info != nil {
		info.UseCache(15 * time.Second)
	}
	if info := (&publicapi.GetGlobal{}).RouterInfo(); info != nil {
		info.UseCache(30 * time.Second)
	}
	if info := (&publicapi.GetQuote{}).RouterInfo(); info != nil {
		info.UseCache(10 * time.Second)
	}
	ingest.NotifyAlert = func(response *dto.AlertResponse) {
		privateapi.NotifyAlertChange(response)
	}
	ingest.Start()
}

// Stop 停止后台采集。
func (*PulseService) Stop() {
	ingest.Stop()
}
