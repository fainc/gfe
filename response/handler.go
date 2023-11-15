package response

import (
	"net/http"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/text/gstr"
	"github.com/gogf/gf/v2/util/gmeta"
)

func MiddlewareHandler(r *ghttp.Request, defaultMime string) {
	var (
		ctx  = r.Context()
		err  = r.GetError()
		res  = r.GetHandlerResponse()
		code = gerror.Code(err)
		mime = gmeta.Get(res, "mime").String()
	)
	// openapi/pprof/已退出程序流程/下载任务
	if r.IsExited() || gstr.Contains(r.RequestURI, "api.json") || gstr.Contains(r.RequestURI, "/debug/pprof/") || r.Response.Writer.Header().Get("Content-Type") == "application/force-download" {
		return
	}
	// 声明meta为自定义输出时，不走当前中间件
	if mime == "custom" {
		setServerHeader(r)
		return
	}
	if mime == "" { // 无指定MIME，使用默认MIME
		mime = defaultMime
	}
	f := formatWriter(mime)
	// 已有err错误
	if err != nil {
		// -1:未定义错误code的error；>=1000 自定义错误码error；51:参数验证错误
		if code.Code() == -1 || code.Code() >= 1000 || code.Code() == 51 {
			f.StandardError(ctx, err)
			return
		}
		// 4XX 常用error，同步http code；
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
		var e error
		switch r.Response.Status {
		case http.StatusForbidden: // 403
			e = ForbiddenError(ctx)
		case http.StatusNotFound: // 404
			e = NotFoundError(ctx)
		case http.StatusMethodNotAllowed: // 405
			e = MethodNotAllowedError(ctx)
		case http.StatusTooManyRequests: // 429
			e = TooManyRequestsError(ctx)
		default: // 未知的错误状态码
			e = InternalError(ctx)
		}
		f.SyncHTTPCodeError(ctx, e)
		return
	}
	var tpl string
	if mime == MimeHTML {
		tpl = gmeta.Get(res, "x-tpl").String()
	}
	f.Success(ctx, res, tpl)
}
