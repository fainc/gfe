package response

import (
	"context"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/vmihailenco/msgpack/v5"
)

const (
	FormatJSON    = "json"
	FormatXML     = "xml"
	FormatCustom  = "custom"
	FormatMsgPack = "msgPack" // MessagePack https://msgpack.org
)

type format struct {
	mime string // 输出类型
}

// FormatWriter 格式化输出
func FormatWriter(mime ...string) *format {
	if len(mime) == 0 {
		return &format{mime: FormatJSON}
	}
	return &format{mime[0]}
}

type resultFormat struct {
	Code    int         `json:"code"`
	Data    interface{} `json:"data"`
	Message string      `json:"message"` // error message
}

func (rec *format) Success(ctx context.Context, data interface{}) {
	rec.writer(ctx, 200, nil, data)
}

// StandardError  业务级标准错误输出
func (rec *format) StandardError(ctx context.Context, err error) {
	rec.writer(ctx, 200, err, nil)
}

// InternalError  服务器级错误
func (rec *format) InternalError(ctx context.Context, err error) {
	rec.writer(ctx, 500, err, nil)
}

// SyncHTTPCodeError 同步http状态码错误输出,不建议错误处理单独调用该方法输出（直接输出错误不会被日志系统捕获），最佳实践是通过 r.SetError 统一处理。
func (rec *format) SyncHTTPCodeError(ctx context.Context, err error) {
	e := gerror.Code(err)
	rec.writer(ctx, e.Code(), err, nil)
}

// Writer 标准格式数据输出
func (rec *format) writer(ctx context.Context, statusCode int, error error, data interface{}) {
	result := resultFormat{
		Data: data,
	}
	if error != nil {
		ge := gerror.Code(error)
		result.Code = ge.Code()
		result.Message = error.Error()
	}
	r := g.RequestFromCtx(ctx)
	r.Response.WriteStatus(statusCode) // use http 200
	r.Response.ClearBuffer()
	switch rec.mime {
	case FormatXML:
		r.Response.WriteXml(result, "xml")
		r.Response.Header().Set("Content-Type", "application/xml; charset=utf-8") // Overwrite the default XML content type "text/xml" in GF with "application/xml".
	case FormatMsgPack:
		b, err := msgpack.Marshal(result)
		if err != nil {
			panic(err) // todo test panic
		}
		r.Response.Write(b)
		r.Response.Header().Set("Content-Type", "application/x-msgpack; charset=utf-8") // Set custom content type "application/x-msgpack".
	default:
		r.Response.WriteJson(result)
		r.Response.Header().Set("Content-Type", "application/json; charset=utf-8") // set charset=utf-8
	}
	SetDefaultResponseHeader(r)
	r.ExitAll() // r.IsExited() will return a true value in the next middleware.
}
