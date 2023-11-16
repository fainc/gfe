package middleware

import (
	"context"

	"github.com/gogf/gf/v2/container/gmap"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gcfg" //nolint:typecheck
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/gogf/gf/v2/os/genv"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/text/gstr"
	"github.com/gogf/gf/v2/util/gconv"

	"github.com/fainc/gfe/helper"
	"github.com/fainc/gfe/util"
)

// init 日志组件初始化覆盖部分默认配置
func init() {
	// 覆盖配置文件,关闭框架默认日志记录和终端打印，避免重复错误输出
	_ = g.Config().GetAdapter().(*gcfg.AdapterFile).Set("server.logPath", "")
	_ = g.Config().GetAdapter().(*gcfg.AdapterFile).Set("server.logStdout", false)

	// 通过环境变量屏蔽gview自带的错误打印，避免重复错误输出
	_ = genv.Set("GF_GVIEW_ERRORPRINT", "false")
}

type logger struct {
	LoggerOptions
}
type LoggerOptions struct {
	AccessLog       bool
	AccessStdout    bool
	AccessMaxLimit  int64
	AccessHeaderKey g.ArrayStr
}

// Logger 日志中间件，需要依赖logger.path配置才能写入文件，可选Access日志（错误日志是强制显示和记录的）
func Logger(options ...LoggerOptions) *logger {
	if len(options) == 0 {
		return &logger{
			LoggerOptions{
				AccessLog:      true,
				AccessStdout:   false,
				AccessMaxLimit: 1024,
			},
		}
	}
	return &logger{options[0]}
}
func (rec *logger) Register(r *ghttp.Request) {
	r.Middleware.Next()
	traceID := gctx.CtxId(r.Context())
	// 业务日志
	shouldLog := rec.AccessLog // 是否需要日志
	if shouldLog {
		// 移除不需要日志的内容
		if util.GetReqMetaStr(r, "x-logger-ignore") == "true" || gstr.Contains(r.RequestURI, "api.json") || gstr.Contains(r.RequestURI, "/debug/pprof/") || r.Response.Writer.Header().Get("Content-Type") == "application/force-download" {
			shouldLog = false
		}
	}

	if shouldLog {
		ct := r.GetHeader("Content-Type")
		referer := r.GetHeader("Referer")
		ua := r.GetHeader("User-Agent")
		bd := r.GetBodyString()
		cip := r.GetClientIp()
		ets := gtime.NewFromTimeStamp(r.EnterTime).String()
		rt := gtime.Now().TimestampMilli() - r.EnterTime
		buffer := r.Response.BufferString()
		if rec.AccessMaxLimit >= 1 && r.Request.ContentLength > rec.AccessMaxLimit {
			bd = "body bytes exceed the limit"
		}
		if rec.AccessMaxLimit >= 1 && r.Response.BufferLength() > int(rec.AccessMaxLimit) {
			buffer = "buffer bytes exceed the limit"
		}
		header := gmap.New()
		for _, key := range rec.AccessHeaderKey {
			header.Set(key, r.GetHeader(key))
		}
		jwtInfo := helper.CtxUser().Get(r.Context())
		logData := g.Map{"jwt": jwtInfo, "header": header, "remoteAddr": r.Request.RemoteAddr, "referer": referer, "traceId": traceID, "method": r.Response.Request.Method, "code": r.Response.Status, "uri": r.Request.RequestURI, "contentType": ct, "UA": ua, "body": bd, "ip": cip, "time": ets, "runTime": rt, "buffer": buffer, "respContent": r.Response.Writer.Header().Get("Content-Type")}
		rec.writeAccess(r.Context(), logData)
	}
	// 错误日志
	err := r.GetError()
	if err != nil {
		rec.Exception(r.Context(), err)
	}
}

// writeAccess access信息强制 Stdout = false,使用 Print 忽略 logger配置的 level，需要依赖配置 logger.path 才能记录
func (rec *logger) writeAccess(ctx context.Context, data interface{}) {
	g.Log().Async().Cat("access").Header(true).Stack(false).Stdout(rec.AccessStdout).Print(ctx, data)
}

// writeError 错误信息强制 Stdout = true，增强提示，需要依赖配置 logger.path 才能记录
func (rec *logger) writeError(ctx context.Context, err interface{}) {
	g.Log().Async().Cat("error").Header(true).Stack(false).Stdout(true).Error(ctx, err)
}
func (rec *logger) Exception(ctx context.Context, err error) {
	code := gerror.Code(err)
	// -1 未定义错误，一般直接抛出error产生，51，请求参数验证错误
	if (code.Code() < 400 && code.Code() != -1 && code.Code() != 51) || (code.Code() < 1000 && code.Code() >= 500) { // 需要捕获错误信息的code范围
		if gerror.HasStack(err) {
			rec.writeError(ctx, gerror.Stack(err)+gconv.String(code.Detail()))
		} else {
			rec.writeError(ctx, err)
		}
	}
}
func (rec *logger) Debug(ctx context.Context, data interface{}) {
	g.Log().Async().Cat("debug").Header(true).Stdout(true).Debug(ctx, data)
}
