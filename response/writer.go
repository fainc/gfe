package response

import (
	"context"
	"fmt"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/util/gconv"
)

const (
	MimeJSON = "JSON"
	MimeXML  = "XML"
	MimeHTML = "HTML"
)

type format struct {
	Mime string // 输出类型
}

// FormatWriter 格式化输出，全局后置response中间件也是调用这个方法，不建议单独调用该方法输出（提前输出），最佳实践是通过 r.SetError 统一处理。
func FormatWriter(mime ...string) *format {
	if len(mime) == 0 {
		return &format{Mime: MimeJSON}
	}
	return &format{mime[0]}
}

func (rec *format) Success(ctx context.Context, data interface{}, tpl ...string) {
	rec.writer(ctx, 200, nil, data, tpl...)
}

// StandardError 400 业务级标准错误输出
func (rec *format) StandardError(ctx context.Context, err error, tpl ...string) {
	if err == nil {
		err = UnknownError(ctx)
	}
	// 设置 html渲染默认模板为 error/400.html
	viewTpl := fmt.Sprintf("error/%v.html", 400)
	if len(tpl) != 0 {
		viewTpl = tpl[0]
	}
	rec.writer(ctx, 400, err, nil, viewTpl)
}

// SyncHTTPCodeError 同步http状态码错误输出
func (rec *format) SyncHTTPCodeError(ctx context.Context, err error, tpl ...string) {
	if err == nil {
		err = UnknownError(ctx)
	}
	e := gerror.Code(err)
	// 设置 html渲染默认模板为 error/对应错误码.html
	var viewTpl = fmt.Sprintf("error/%v.html", e.Code())
	if len(tpl) != 0 {
		viewTpl = tpl[0]
	}
	rec.writer(ctx, e.Code(), err, nil, viewTpl)
}

type errFormat struct {
	Code    int         `json:"code"`    // 错误码，为-1（通用）或特定错误码
	Message string      `json:"message"` // 错误信息，从err.Error()读取
	Detail  interface{} `json:"detail"`  // 错误详情，为一个列表或null（使用错误封装才能附带错误列表，普通错误直接捕获detail为null）
}
type resultFormat struct {
	Code  int         `json:"code"`  // 业务状态码 与 http 状态码同步
	Data  interface{} `json:"data"`  // 	业务成功返回数据 or null
	Error *errFormat  `json:"error"` // 错误信息 or null
}

// Writer 标准格式数据输出, tpl 为html格式化输出专属的参数
func (rec *format) writer(ctx context.Context, code int, error error, data interface{}, viewTpl ...string) {
	var e *errFormat
	if error != nil {
		ge := gerror.Code(error)
		e = &errFormat{
			Code:    ge.Code(),
			Message: error.Error(), // 直接从error读，未封装的普通err 使用ge.Message() 错误信息为空
			Detail:  ge.Detail(),   // list or null
		}
	}
	result := resultFormat{
		Code:  code,
		Data:  data,
		Error: e,
	}
	r := g.RequestFromCtx(ctx)
	r.Response.WriteStatus(code)
	r.Response.ClearBuffer()
	switch rec.Mime {
	case MimeXML:
		r.Response.WriteXml(result, "xml")
	case MimeHTML:
		t := ""
		if len(viewTpl) != 0 {
			t = viewTpl[0]
		}
		view, err := r.Response.ParseTpl(t, gconv.Map(result))
		if err != nil {
			// 模板渲染直接输出错误兜底
			r.SetError(CodeError(500, err.Error()))
			r.Response.WriteStatus(500, "html render tpl error")
		} else {
			r.Response.Write(view)
		}
		r.Response.Header().Set("Content-Type", "text/html")
	default:
		r.Response.WriteJson(result)
		r.Response.Header().Set("Content-Type", "application/json;charset=utf-8") // 部分浏览器如果没设置charset可能会中文乱码
	}
	setServerHeader(r)
	r.ExitAll() // 强行中断当前执行流程，当前执行方法的后续逻辑以及后续所有的逻辑方法将不再执行，后续中间件可以通过 r.IsExited() 判断是否已经退出。PS: 不建议单独调用输出（提前输出），最佳实践是通过 r.SetError 统一处理。
}
