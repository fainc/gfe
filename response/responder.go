package response

import (
	"net/http"

	"github.com/gogf/gf/v2/errors/gcode"

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
		mime = gmeta.Get(res, "mime").String()
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
		// 业务级错误 ：-1 未定义的错误；>=1000 自定义错误码；51 框架参数验证错误；401 授权错误
		isBizError := code.Code() == -1 || code.Code() >= 1000 || code.Code() == gcode.CodeValidationFailed.Code() || code.Code() == 401
		if isBizError {
			f.StandardError(ctx, err)
			return
		}
		// 其它非业务级别错误，屏蔽错误细节后输出 500 InternalError
		f.InternalError(ctx, InternalError(ctx))
		return
	}
	// 无错误但有响应状态码
	if r.Response.Status > 0 && r.Response.Status != http.StatusOK {
		responseError := CodeErrorTranslate(ctx, r.Response.Status, http.StatusText(r.Response.Status))
		f.SyncHTTPCodeError(ctx, responseError)
		r.SetError(responseError) // Set an error for the next middleware, e.g. logs.
		return
	}
	f.Success(ctx, res)
}
