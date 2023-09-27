package middleware

import (
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gbuild"
)

func BuildVersion(r *ghttp.Request) {
	// 需要在build varMap 定义
	// buildVersion: xxx
	r.SetCtxVar("BUILD_VERSION", gbuild.Get("buildVersion", "UNKNOWN"))
	r.Middleware.Next()
}
