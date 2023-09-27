package middleware

import (
	"context"

	"github.com/gogf/gf/v2/container/garray"
	"github.com/gogf/gf/v2/container/gmap"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/util/gconv"

	"github.com/fainc/gfe/token"
)

var limit int64
var logStdout bool
var logHeaderExtKey g.Array
var logWhiteURI *garray.Array
var serviceLogPath string

// config sample:
// server:
// 	logPath: "logs/server"
// 	serviceLogPath: "logs/service"
// 	logStdout: false
// 	logDataLimit: 100
// 	logHeaderExtKey:
// 		- "AppId"
// 		- "Lang"
// 	logWhiteUri:
// 		- "/hello"

func init() {
	limit = g.Cfg().MustGet(context.Background(), "server.logDataLimit").Int64()
	if limit >= 1024 { // 最高限制1M，防止日志数据过于膨胀
		limit = 1024
	}
	limit *= 1024
	logStdout = g.Cfg().MustGet(context.Background(), "server.logStdout").Bool()
	logHeaderExtKey = g.Cfg().MustGet(context.Background(), "server.logHeaderExtKey").Array()
	logWhiteURI = garray.NewArrayFrom(g.Cfg().MustGet(context.Background(), "server.logWhiteUri").Array())
	serviceLogPath = g.Cfg().MustGet(context.Background(), "server.serviceLogPath").String()
}

func Logger(r *ghttp.Request) {
	r.Middleware.Next()
	if !logWhiteURI.Contains(r.Request.RequestURI) {
		traceID := gctx.CtxId(r.Context())
		ct := r.GetHeader("Content-Type")
		referer := r.GetHeader("Referer")
		ua := r.GetHeader("User-Agent")
		bd := r.GetBodyString()
		cip := r.GetClientIp()
		ets := gtime.NewFromTimeStamp(r.EnterTime).String()
		rt := gtime.Now().TimestampMilli() - r.EnterTime
		buffer := r.Response.BufferString()
		if limit >= 1 && r.Request.ContentLength >= limit {
			bd = "data bytes exceed the limit"
		}
		if limit >= 1 && r.Response.BufferLength() >= int(limit) {
			buffer = "data bytes exceed the limit"
		}
		ext := gmap.New()
		for _, key := range logHeaderExtKey {
			k := gconv.String(key)
			ext.Set(k, r.GetHeader(k))
		}
		jwtUser := token.Helper().GetCtxUser(r.Context())
		logData := g.Map{"jwtUser": jwtUser, "headerExt": ext, "remoteAddr": r.Request.RemoteAddr, "referer": referer, "traceId": traceID, "method": r.Response.Request.Method, "code": r.Response.Status, "uri": r.Request.RequestURI, "contentType": ct, "UA": ua, "body": bd, "ip": cip, "time": ets, "runTime": rt, "buffer": buffer}
		g.Log().Async().Header(false).Path(serviceLogPath).Stdout(logStdout).Print(context.Background(), logData)
	}
}
