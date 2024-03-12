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

// Success write formatted data with 200 http status code.
func (rec *format) Success(ctx context.Context, data interface{}) {
	rec.write(ctx, 200, nil, data)
}

// StandardError  write formatted data with 200 http status code.
func (rec *format) StandardError(ctx context.Context, err error) {
	rec.write(ctx, 200, err, nil)
}

// InternalError  write formatted data with 500 http status code.
func (rec *format) InternalError(ctx context.Context, err error) {
	rec.write(ctx, 500, err, nil)
}

// SyncHTTPCodeError write formatted data with http status code.
func (rec *format) SyncHTTPCodeError(ctx context.Context, err error) {
	e := gerror.Code(err)
	rec.write(ctx, e.Code(), err, nil)
}

// Write formatted data, support json, xml, and msgPack.
func (rec *format) write(ctx context.Context, statusCode int, err error, data interface{}) {
	result := resultFormat{
		Data: data,
	}
	if err != nil {
		ge := gerror.Code(err)
		result.Code = ge.Code()      // get error code or -1
		result.Message = err.Error() // get error message
	}
	r := g.RequestFromCtx(ctx)
	r.Response.WriteStatus(statusCode) // use http 200
	r.Response.ClearBuffer()
	switch rec.mime {
	case FormatXML:
		r.Response.WriteXml(result, "xml")
		r.Response.Header().Set("Content-Type", "application/xml; charset=utf-8") // Overwrite the default XML content type "text/xml" in GF with "application/xml".
	case FormatMsgPack:
		b, marshalError := msgpack.Marshal(result)
		if marshalError != nil {
			panic(marshalError) // todo test panic
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
