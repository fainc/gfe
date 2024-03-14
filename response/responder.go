package response

import (
	"context"
	"fmt"
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

type errorFormat struct {
	Code    int         `json:"code" msgpack:"code"`
	Message string      `json:"message" msgpack:"message"`
	Detail  interface{} `json:"detail" msgpack:"detail"`
}
type resultFormat struct {
	Ok      bool         `json:"ok" msgpack:"ok"`
	Payload interface{}  `json:"payload,omitempty" msgpack:"payload,omitempty"`
	Error   *errorFormat `json:"error,omitempty" msgpack:"error,omitempty"`
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
	r.Middleware.Next()
	var (
		ctx        = r.Context()
		err        = r.GetError()
		res        = r.GetHandlerResponse()
		code       = gerror.Code(err)
		metaCustom = gmeta.Get(res, "format").String() == FormatCustom
	)
	format := rec.defaultFormat
	if format == FormatCustom || metaCustom {
		SetDefaultResponseHeader(r)
		return
	}
	// Special content, it then exits current handler.
	if r.IsExited() ||
		gstr.Contains(r.RequestURI, "api.json") ||
		gstr.Contains(r.RequestURI, "/debug/pprof/") ||
		r.Response.Writer.Header().Get("Content-Type") == "application/force-download" {
		return
	}
	if err != nil {
		// Normal error code. The CodeValidationFailed 51 is special.
		if code.Code() == -1 || code.Code() >= 1000 || code.Code() == gcode.CodeValidationFailed.Code() || code.Code() >= 400 && code.Code() < 500 {
			rec.Write(ctx, format, http.StatusOK, nil, err)
			return
		}
		// Other unexpected error code, writes http 500 InternalError and removes error details.
		rec.Write(ctx, format, http.StatusInternalServerError, nil, CodeError(500, "InternalServerError", fmt.Sprintf("Unexpected Error Code %v", code.Code())))
		return
	}
	if r.Response.Status > 0 && r.Response.Status != http.StatusOK {
		responseError := CodeError(r.Response.Status, http.StatusText(r.Response.Status), r.RequestURI)
		rec.Write(ctx, format, r.Response.Status, nil, responseError)
		r.SetError(responseError) // Set an error for the next middleware, e.g. logs.
		return
	}
	rec.Write(ctx, format, http.StatusOK, res, nil)
}

// Write formatted data, support json, xml, and msgPack.
func (rec *responder) Write(ctx context.Context, format string, statusCode int, payload interface{}, err error) {
	result := rec.makeResult(payload, err)
	r := g.RequestFromCtx(ctx)
	r.Response.WriteStatus(statusCode) // use http 200
	r.Response.ClearBuffer()
	switch format {
	case FormatXML:
		r.Response.WriteXml(result, "xml")
		r.Response.Header().Set("Content-Type", "application/xml; charset=utf-8") // Overwrite the default XML content type "text/xml" in GF with "application/xml".
	case FormatMsgPack:
		b, _ := msgpack.Marshal(result)
		r.Response.Write(b)
		r.Response.Header().Set("Content-Type", "application/x-msgpack; charset=utf-8") // Set custom content type "application/x-msgpack".
	default:
		r.Response.WriteJson(result)
		r.Response.Header().Set("Content-Type", "application/json; charset=utf-8") // set charset=utf-8
	}
	SetDefaultResponseHeader(r)
	r.ExitAll() // r.IsExited() will return a true value in the next middleware.
}

func (rec *responder) makeResult(payload interface{}, err error) (result *resultFormat) {
	result = &resultFormat{}
	if err != nil {
		ge := gerror.Code(err)
		result.Ok = false
		e := &errorFormat{
			Code:    ge.Code(),
			Message: err.Error(),
			Detail:  ge.Detail(),
		}
		if e.Detail == nil || e.Detail == "" {
			e.Detail = []int{}
		}
		result.Error = e
	} else {
		result.Payload = payload
		result.Ok = true
	}
	return
}
