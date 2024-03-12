package response

import (
	"os"

	"github.com/gogf/gf/v2/crypto/gmd5"
	"github.com/gogf/gf/v2/i18n/gi18n"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/util/gconv"
)

// SetDefaultResponseHeader 设置公共响应头
func SetDefaultResponseHeader(r *ghttp.Request) {
	serverName, _ := os.Hostname()
	serverID, _ := gmd5.Encrypt(serverName)
	r.Response.Header().Set("Server-Id", serverID)
	r.Response.Header().Set("Server-Lang", gi18n.LanguageFromCtx(r.Context()))
	r.Response.Header().Set("Server-RunTime", gconv.String(gtime.Now().TimestampMilli()-r.EnterTime))
}
