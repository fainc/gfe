package response

import (
	"context"
	"fmt"
	"net/http"

	"github.com/vmihailenco/msgpack/v5"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/frame/g"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/text/gstr"
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
type resultFormat struct {
	Ok bool `json:"ok" msgpack:"ok"`
	*successFormat
	Error     *errorFormat `json:"error,omitempty" msgpack:"error,omitempty"`
	Encrypted bool         `json:"encrypted,omitempty" msgpack:"encrypted,omitempty"`
}
type responder struct {
	opts Options
}

type Options struct {
	Format    string
	Encryptor func(payload interface{}) (result interface{}, encrypted bool)
}

// NewResponder returns a responder.
// Options.Format supports json / xml / msgPack / custom and works within the scope of middleware.
// The "custom" tag specifies that middleware handler should be skipped, and you can write content to the response buffer yourself.
func NewResponder(opts Options) *responder {
	r := &responder{opts}
	return r
}

// Middleware handler of response.
func (rec *responder) Middleware(r *ghttp.Request) {
	r.Middleware.Next()
	var (
		ctx  = r.Context()
		err  = r.GetError()
		res  = r.GetHandlerResponse()
		code = gerror.Code(err)
	)
	if r.IsExited() ||
		rec.opts.Format == FormatCustom ||
		gstr.Contains(r.RequestURI, "api.json") ||
		gstr.Contains(r.RequestURI, "/debug/pprof/") ||
		r.Response.Writer.Header().Get("Content-Type") == "application/force-download" {
		return
	}
	if err != nil {
		// Normal error code. The CodeValidationFailed 51 is special.
		if code.Code() == -1 || code.Code() >= 1000 || code.Code() == gcode.CodeValidationFailed.Code() || code.Code() >= 400 && code.Code() < 500 {
			rec.Write(ctx, http.StatusOK, nil, err)
			return
		}
		// Other unexpected error code, writes http 500 InternalError and removes error details.
		rec.Write(ctx, http.StatusInternalServerError, nil, CodeError(500, "InternalServerError", fmt.Sprintf("Unexpected Error Code %v", code.Code())))
		return
	}
	if r.Response.Status > 0 && r.Response.Status != http.StatusOK {
		responseError := CodeError(r.Response.Status, http.StatusText(r.Response.Status), r.RequestURI)
		rec.Write(ctx, r.Response.Status, nil, responseError)
		r.SetError(responseError) // Set an error for the next middleware, e.g. logs middleware.
		return
	}
	rec.Write(ctx, http.StatusOK, res, nil)
}

// Write formatted data to response buffer.
func (rec *responder) Write(ctx context.Context, statusCode int, payload interface{}, err error) {
	result := rec.makeResult(payload, err)
	r := g.RequestFromCtx(ctx)
	r.Response.WriteStatus(statusCode) // use http 200
	r.Response.ClearBuffer()
	if rec.opts.Encryptor != nil && result.Ok {
		result.Payload, result.Encrypted = rec.opts.Encryptor(result.Payload) // call the encryptor function.
	}
	switch rec.opts.Format {
	case FormatXML:
		r.Response.WriteXml(result, "xml")
		r.Response.Header().Set("Content-Type", "application/xml; charset=utf-8") // Overwrite the default XML content type "text/xml" in GF with "application/xml".
	case FormatMsgPack:
		marshal, _ := msgpack.Marshal(result)
		r.Response.Write(marshal)
		r.Response.Header().Set("Content-Type", "application/x-msgpack")
	default:
		r.Response.WriteJson(result)
		r.Response.Header().Set("Content-Type", "application/json; charset=utf-8")
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
		result.successFormat = &successFormat{Payload: payload}
		result.Ok = true
	}
	return
}
