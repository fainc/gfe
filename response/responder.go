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
type successFormat struct {
	Payload interface{} `json:"payload" msgpack:"payload"`
}
type v2resultFormat struct {
	Ok bool `json:"ok" msgpack:"ok"`
	*successFormat
	Error     *errorFormat `json:"error,omitempty" msgpack:"error,omitempty"`
	Encrypted bool         `json:"encrypted,omitempty" msgpack:"encrypted,omitempty"`
}
type responder struct {
	defaultFormat string
	version       int
	encryptFunc   func(payload interface{}) (result interface{}, encrypted bool)
}

type Options struct {
	Format  string
	Version int
}

// NewResponder returns a responder to bind middleware.
// Options.Format supports json / xml / msgPack / custom.
// Options.Version supports 1 or 2. Default is 2.
// V2 demo: V2FormatSuccessDemo, V2FormatErrorDemo.
// Version 1 demo: V1FormatSuccessDemo, V1FormatErrorDemo.
// The encryptFunc receives the raw payload and returns the encrypted payload and encryption status(encrypted or not).
func NewResponder(opts Options, encryptFunc ...func(payload interface{}) (result interface{}, encrypted bool)) *responder {
	v := 2
	if opts.Version == 1 {
		v = 1
	}
	r := &responder{defaultFormat: opts.Format, version: v}
	if len(encryptFunc) == 1 {
		r.encryptFunc = encryptFunc[0]
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
	result := rec.makeV2Result(payload, err)
	r := g.RequestFromCtx(ctx)
	r.Response.WriteStatus(statusCode) // use http 200
	r.Response.ClearBuffer()
	if rec.encryptFunc != nil && result.Ok {
		result.Payload, result.Encrypted = rec.encryptFunc(result.Payload) // call the encryptor function.
	}
	switch format {
	case FormatXML:
		r.Response.WriteXml(result, "xml")
		r.Response.Header().Set("Content-Type", "application/xml; charset=utf-8") // Overwrite the default XML content type "text/xml" in GF with "application/xml".
	case FormatMsgPack:
		marshal, _ := msgpack.Marshal(result)
		r.Response.Write(marshal)
		r.Response.Header().Set("Content-Type", "application/x-msgpack; charset=utf-8") // Set custom content type "application/x-msgpack".
	default:
		r.Response.WriteJson(result)
		r.Response.Header().Set("Content-Type", "application/json; charset=utf-8") // set charset=utf-8
	}
	SetDefaultResponseHeader(r)
	r.ExitAll() // r.IsExited() will return a true value in the next middleware.
}

func (rec *responder) makeV2Result(payload interface{}, err error) (result *v2resultFormat) {
	result = &v2resultFormat{}
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
		result.successFormat = &successFormat{Payload: payload}
		result.Ok = true
	}
	return
}
