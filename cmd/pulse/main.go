package main

import (
	"io/fs"

	"github.com/digitalwayhk/core/pkg/server/run"
	"github.com/digitalwayhk/core/pkg/server/types"
	pulse "github.com/digitalwayhk/pulse"
	"github.com/digitalwayhk/pulse/web"
)

// main 启动框架内建管理服务和 Pulse 行情服务。
func main() {
	demoFS, err := fs.Sub(web.FS, ".")
	if err != nil {
		panic(err)
	}
	server := run.NewWebServer()
	server.AddIService(&pulse.PulseService{}, &types.ServerOption{
		IsWebSocket: true,
		Demo: &types.DemoOption{
			Pattern: "screener",
			File:    demoFS,
		},
	})
	server.Start()
}
