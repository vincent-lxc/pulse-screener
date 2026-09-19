// Package web 内嵌评委可点击的薄 HTML 筛选器页面。
package web

import "embed"

// FS 是 /screener/ 静态演示页。
//
//go:embed index.html
var FS embed.FS
