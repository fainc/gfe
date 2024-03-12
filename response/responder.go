package response

import (
	"net/http"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/text/gstr"
	"github.com/gogf/gf/v2/util/gmeta"
)

type responder struct {
	defaultFormat string
}

func NewResponder(formatOptions ...string) *responder {
	if len(formatOptions) == 0 {
		return &responder{defaultFormat: FormatJSON}
	}
	return &responder{formatOptions[0]}
}

// Middleware 后置中间件守卫
func (rec *responder) Middleware(r *ghttp.Request) {
	r.Middleware.Next() // after middleware
	var (
		ctx  = r.Context()
		err  = r.GetError()
		res  = r.GetHandlerResponse()
		code = gerror.Code(err)
		mime = gmeta.Get(res, "format").String()
	)
	if mime == "" {
		mime = rec.defaultFormat // set default format
	}
	if mime == FormatCustom {
		SetDefaultResponseHeader(r)
		return
	}
	// special
	if r.IsExited() ||
		gstr.Contains(r.RequestURI, "api.json") ||
		gstr.Contains(r.RequestURI, "/debug/pprof/") ||
		r.Response.Writer.Header().Get("Content-Type") == "application/force-download" {
		return
	}
	f := FormatWriter(mime)
	if err != nil {
		// -1:未定义错误code；>=1000 自定义错误码；51:参数验证错误
		if code.Code() == -1 || code.Code() >= 1000 || code.Code() == 51 {
			f.StandardError(ctx, err)
			return
		}
		// 4XX  常用同步 HTTP 状态码
		if code.Code() >= 400 && code.Code() < 500 {
			f.SyncHTTPCodeError(ctx, err)
			return
		}
		// 其它错误，屏蔽错误细节再输出
		f.SyncHTTPCodeError(ctx, InternalError(ctx))
		return
	}
	// 无错误但有响应状态码
	if r.Response.Status > 0 && r.Response.Status != http.StatusOK {
		f.SyncHTTPCodeError(ctx, CodeErrorTranslate(ctx, r.Response.Status, http.StatusText(r.Response.Status)))
		return
	}
	f.Success(ctx, res)
}
