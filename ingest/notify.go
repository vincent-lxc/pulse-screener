package ingest

import "github.com/digitalwayhk/pulse/api/dto"

// NotifyAlert 由服务启动时注入，避免 ingest ↔ api/private 循环依赖。
var NotifyAlert func(*dto.AlertResponse)

func notifyAlert(response *dto.AlertResponse) {
	if NotifyAlert != nil {
		NotifyAlert(response)
	}
}
