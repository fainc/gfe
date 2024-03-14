package response

import (
	"context"
	"net/http"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/i18n/gi18n"
)

// CustomRes The format is "custom" and the response middleware will return the original data.
type CustomRes struct {
	g.Meta `format:"custom" sm:"custom response data" dc:"The API returns custom data, please contact the developer for details"`
}

// EmptyRes The API returns empty data.
type EmptyRes struct {
	g.Meta `sm:"empty response data" dc:"The API returns empty data"`
}

// CodeError
// Code must be -1(StandError), >=1000(CustomCodeError) or in the range 400 to 499(Special Error such as 401 UnAuthorizedError), otherwise the response middleware will output a 500 InternalError.
func CodeError(code int, message string, detail ...interface{}) error {
	return gerror.NewCode(gcode.New(code, message, getDetailValue(detail...)))
}

// CodeErrorTranslate returns an error with i18n message.
// Code must be -1(StandError), 401(UnAuthorizedError), or >=1000(CustomCodeError), otherwise the response middleware will output a 500 InternalError.
func CodeErrorTranslate(ctx context.Context, code int, message string, detail ...interface{}) error {
	message = gi18n.T(ctx, message)
	return CodeError(code, message, detail...)
}

// CodeErrorTranslateFormat returns an error with i18n format message.
func CodeErrorTranslateFormat(ctx context.Context, code int, format string, values ...interface{}) error {
	message := gi18n.Tf(ctx, format, values)
	return CodeError(code, message, nil)
}

// UnAuthorizedError returns 401 code error.
func UnAuthorizedError(ctx context.Context, detail ...interface{}) error {
	return CodeErrorTranslate(ctx, 401, "Unauthorized", detail...)
}

// SignatureError returns 402 code error.
func SignatureError(ctx context.Context, detail ...interface{}) error {
	return CodeErrorTranslate(ctx, 402, "SignatureError", detail...)
}

// InternalError returns 500 code error. Used by server error.
func InternalError(detail ...interface{}) error {
	return CodeError(500, http.StatusText(500), detail...)
}

// StandError returns -1 code error.
func StandError(ctx context.Context, message string, detail ...interface{}) error {
	return CodeErrorTranslate(ctx, -1, message, detail...)
}

func getDetailValue(detail ...interface{}) interface{} {
	if len(detail) >= 1 {
		return detail
	}
	return []int{} // return empty slice [], not nil (null).
}
