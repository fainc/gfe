package response

import (
	"context"
	"os"

	"github.com/gogf/gf/v2/frame/g"

	"github.com/gogf/gf/v2/crypto/gmd5"
	"github.com/gogf/gf/v2/i18n/gi18n"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/util/gconv"
)

var agent = g.Cfg().MustGet(context.Background(), "server.serverAgent").String()

func SetDefaultResponseHeader(r *ghttp.Request) {
	serverName, _ := os.Hostname()
	serverID, _ := gmd5.Encrypt(serverName)
	if agent == "" {
		// Keep server fingerprints safety.
		// Overwrite the default server agent "GoFrame HTTP Server".
		r.Server.SetServerAgent("Unknown")
	}
	r.Response.Header().Set("Server-Id", serverID)
	r.Response.Header().Set("Server-Lang", gi18n.LanguageFromCtx(r.Context()))
	r.Response.Header().Set("Server-RunTime", gconv.String(gtime.Now().TimestampMilli()-r.EnterTime))
}
