package response

import (
	"context"
	"net/http"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/vmihailenco/msgpack/v5"

	"github.com/gogf/gf/v2/errors/gcode"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/text/gstr"
	"github.com/gogf/gf/v2/util/gmeta"
)

const (
	FormatJSON    = "json"
	FormatXML     = "xml"
	FormatCustom  = "custom"
	FormatMsgPack = "msgPack" // MessagePack https://msgpack.org
)

type resultFormat struct {
	Code    int         `json:"code"`
	Data    interface{} `json:"data"`
	Message string      `json:"message"`
	Detail  interface{} `json:"detail"`
}

type responder struct {
	defaultFormat string
}

func NewResponder(formatOptions ...string) *responder {
	r := &responder{defaultFormat: FormatJSON}
	if len(formatOptions) == 0 {
		return r
	}
	if len(formatOptions) >= 1 {
		r.defaultFormat = formatOptions[0]
	}
	return r
}

// Middleware of response.
func (rec *responder) Middleware(r *ghttp.Request) {
	r.Middleware.Next() // after middleware
	var (
		ctx    = r.Context()
		err    = r.GetError()
		res    = r.GetHandlerResponse()
		code   = gerror.Code(err)
		format = gmeta.Get(res, "format").String()
	)
	if format == "" {
		format = rec.defaultFormat // set default format
	}
	if format == FormatCustom {
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
	if err != nil {
		// Allowed error code range. The http.StatusUnauthorized 401 is special.
		isAllow := code.Code() == -1 || code.Code() >= 1000 || code.Code() == gcode.CodeValidationFailed.Code()
		if isAllow {
			rec.Write(ctx, format, http.StatusBadRequest, nil, err)
			return
		}
		if http.StatusText(code.Code()) != "" {
			rec.Write(ctx, format, code.Code(), nil, CodeError(code.Code(), http.StatusText(code.Code()), r.RequestURI))
		}
		// Not allowed error code, writes http 500 InternalError and removes error details.
		rec.Write(ctx, format, http.StatusInternalServerError, nil, InternalError(ctx))
		return
	}
	if r.Response.Status > 0 && r.Response.Status != http.StatusOK {
		responseError := CodeError(gcode.CodeNotFound.Code(), http.StatusText(r.Response.Status), r.RequestURI)
		rec.Write(ctx, format, r.Response.Status, nil, responseError)
		r.SetError(responseError) // Set an error for the next middleware, e.g. logs.
		return
	}
	rec.Write(ctx, format, http.StatusOK, res, nil)
}

// Write formatted data, support json, xml, and msgPack.
func (rec *responder) Write(ctx context.Context, format string, statusCode int, data interface{}, err error) {
	result := resultFormat{
		Data:    data,
		Message: "OK",
		Code:    200,
	}
	if err != nil {
		ge := gerror.Code(err)
		result.Code = ge.Code()      // get error code or -1
		result.Message = err.Error() // get error message
		result.Detail = ge.Detail()
	}
	r := g.RequestFromCtx(ctx)
	r.Response.WriteStatus(statusCode) // use http 200
	r.Response.ClearBuffer()
	switch format {
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
