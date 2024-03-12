package response

import (
	"context"
	"net/http"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/i18n/gi18n"
)

// CustomRes The mime is "custom" and the response middleware will return the original data.
type CustomRes struct {
	g.Meta `mime:"custom" sm:"custom response data" dc:"The API returns custom data, please contact the developer for details"`
}

// EmptyRes The API returns empty data.
type EmptyRes struct {
	g.Meta `sm:"empty response data" dc:"The API returns empty data"`
}

// CodeError
// Code must be -1(StandError), 401(UnAuthorizedError), or >=1000(CustomCodeError), otherwise a 500 InternalError will result.
func CodeError(code int, message string, ext ...interface{}) error {
	var detail interface{}
	if len(ext) != 0 {
		detail = ext[0]
	}
	return gerror.NewCode(gcode.New(code, message, detail))
}

// CodeErrorTranslate returns an error with i18n message.
// Code must be -1(StandError), 401(UnAuthorizedError), or >=1000(CustomCodeError), otherwise a 500 InternalError will result.
func CodeErrorTranslate(ctx context.Context, code int, message string, ext ...interface{}) error {
	message = gi18n.T(ctx, message)
	return CodeError(code, message, ext...)
}

// CodeErrorTranslateFormat returns an error with i18n format message.
func CodeErrorTranslateFormat(ctx context.Context, code int, format string, values ...interface{}) error {
	message := gi18n.Tf(ctx, format, values)
	return CodeError(code, message)
}

// UnAuthorizedError returns 401 code error.
func UnAuthorizedError(ctx context.Context, ext ...interface{}) error {
	return CodeErrorTranslate(ctx, 401, "Unauthorized", ext)
}

// SignatureError returns 402 code error.
func SignatureError(ctx context.Context, ext ...interface{}) error {
	return CodeErrorTranslate(ctx, 402, "SignatureError", ext)
}

// InternalError returns 500 code error.
func InternalError(ctx context.Context, ext ...interface{}) error {
	return CodeErrorTranslate(ctx, 500, http.StatusText(http.StatusInternalServerError), ext)
}

// StandError returns -1 code error.
func StandError(ctx context.Context, message string, ext ...interface{}) error {
	return CodeErrorTranslate(ctx, -1, message, ext)
}
