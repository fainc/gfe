package middleware

import (
	"github.com/gogf/gf/v2/net/ghttp"

	"github.com/fainc/gfe/response"
)

// DefaultMiddlewareRegister 常用全局默认中间件统一注册
func DefaultMiddlewareRegister(s *ghttp.Server, defaultMime string) {
	s.BindMiddlewareDefault(CORSDefaultRegister, MultiLangRegister, Logger().Register, response.NewResponder(defaultMime).Middleware)
}

// GroupDefaultMiddlewareRegister  常用组默认中间件统一注册
func GroupDefaultMiddlewareRegister(group *ghttp.RouterGroup, options ...string) {
	defaultMime := response.FormatJSON
	if len(options) >= 1 {
		defaultMime = options[0]
	}
	group.Middleware(CORSDefaultRegister, MultiLangRegister, Logger().Register, response.NewResponder(defaultMime).Middleware)
}
