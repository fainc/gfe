package response

import (
	"net/http"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/text/gstr"
	"github.com/gogf/gf/v2/util/gmeta"
)

func Handler(r *ghttp.Request, defaultMime string) {
	var (
		ctx  = r.Context()
		err  = r.GetError()
		res  = r.GetHandlerResponse()
		code = gerror.Code(err)
		mime = gmeta.Get(res, "mime").String()
	)
	// openapi/已退出程序流程/下载任务
	if r.IsExited() || gstr.Contains(r.RequestURI, "api.json") || r.Response.Writer.Header().Get("Content-Type") == "application/force-download" {
		return
	}
	// 声明meta为自定义输出时，不走当前中间件
	if mime == "custom" {
		setServerHeader(r)
		return
	}
	if mime == "" { // 无指定，使用默认mime
		mime = defaultMime
	}
	f := Format(FormatOptions{Mime: mime})
	// 已有err错误
	if err != nil {
		// -1:未定义错误code的error；>=1000 自定义错误码error；51:参数验证错误
		if code.Code() == -1 || code.Code() >= 1000 || code.Code() == 51 {
			f.Error(ctx, code.Code(), err.Error(), code.Detail())
			return
		}
		// 4XX 常用error，同步http code；
		if code.Code() >= 400 && code.Code() < 500 {
			f.errorSyncHTTPStatus(ctx, code.Code(), err.Error(), code.Detail())
			return
		}
		// 其它错误
		f.InternalError(ctx, nil)
		return
	}
	// 无错误但有响应状态码
	if r.Response.Status > 0 && r.Response.Status != http.StatusOK {
		switch r.Response.Status {
		case http.StatusNotFound: // 404
			f.NotFound(ctx, nil)
			return
		default: // 未知的错误状态码
			f.InternalError(ctx, "unsupported http status code")
			return
		}
	}
	var result interface{}
	f.encrypted, result, err = tryEncrypt(r, mime, res)
	if err != nil {
		r.SetError(CodeError(500, err.Error(), nil))
		f.InternalError(ctx, err.Error())
		return
	}
	if mime == "HTML" {
		f.staticTpl = gmeta.Get(res, "x-static-tpl").String() // 从meta读取静态视图
	}
	f.Success(ctx, result)
}
