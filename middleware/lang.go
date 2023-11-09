package middleware

import (
	"github.com/gogf/gf/v2/container/garray"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/i18n/gi18n"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/text/gstr"
)

// MultiLang 优先注册中间件，以免lang为空
// language:
// - en
// - xx
func MultiLang(r *ghttp.Request) {
	clientLang := gstr.Explode(",", r.Header.Get("Accept-Language"))
	serverLang := "en" // 默认
	supportLang := garray.NewArrayFrom(g.Cfg().MustGet(r.Context(), "language").Array(), true)
	if len(clientLang) != 0 && clientLang[0] != "" && supportLang.Contains(clientLang[0]) {
		serverLang = clientLang[0]
	}
	r.SetCtx(gi18n.WithLanguage(r.Context(), serverLang))
	r.Middleware.Next()
}
