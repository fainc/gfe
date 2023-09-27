package middleware

import (
	"github.com/gogf/gf/v2/net/ghttp"

	"github.com/fainc/gfe/response"
)

func JsonResponse(r *ghttp.Request) {
	r.Middleware.Next()
	response.Handler(r)
}
