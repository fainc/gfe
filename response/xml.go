package response

import (
	"context"
	"os"

	"github.com/gogf/gf/v2/crypto/gmd5"
	"github.com/gogf/gf/v2/frame/g"
)

type xml struct{}

var xmlVar = xml{}

func Xml() *xml {
	return &xmlVar
}

func (rec *xml) CustomWriter(ctx context.Context, data interface{}, status ...int) {
	statusCode := 200
	if len(status) >= 1 && status[0] != 200 {
		statusCode = status[0]
	}
	r := g.RequestFromCtx(ctx)
	r.Response.WriteStatus(statusCode)
	r.Response.ClearBuffer()
	r.Response.WriteXml(data, "xml")
	r.Response.Header().Set("Content-Type", "application/xml;charset=utf-8")
	serverName, _ := os.Hostname()
	serverId, _ := gmd5.Encrypt(serverName)
	r.Response.Header().Set("Server-Id", serverId)
	r.ExitAll()
}
