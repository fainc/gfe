package middleware

import (
	"github.com/gogf/gf/v2/net/ghttp"

	"github.com/fainc/gfe/response"
)

type resp struct {
	mime string
}

// Response 响应中间件
func Response(options ...string) *resp {
	if len(options) == 0 {
		return &resp{mime: response.MimeJSON}
	}
	return &resp{options[0]}
}

func (rec *resp) Register(r *ghttp.Request) {
	r.Middleware.Next()
	response.MiddlewareHandler(r, rec.mime)
}
