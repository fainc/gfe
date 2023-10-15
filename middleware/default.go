package middleware

import (
	"github.com/gogf/gf/v2/net/ghttp"

	"github.com/fainc/gfe/response"
)

// DefaultMiddleware 常用全局默认中间件统一注册
func DefaultMiddleware(s *ghttp.Server, defaultMime string) {
	s.BindMiddlewareDefault(CORSDefault, MultiLang, Logger().Register, Response(defaultMime).Register)
}

// GroupDefaultMiddleware  常用组默认中间件统一注册
func GroupDefaultMiddleware(group *ghttp.RouterGroup, options ...string) {
	defaultMime := response.MimeJSON
	if len(options) >= 1 {
		defaultMime = options[0]
	}
	group.Middleware(CORSDefault, MultiLang, Logger().Register, Response(defaultMime).Register)
}
