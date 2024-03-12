package response

import (
	"context"
	"net/http"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/i18n/gi18n"
)

// CustomRes 自定义返回数据标注，使用该类型数据返回时，全局后置中间件ResponseHandler将不再处理返回数据，请自行提前输出
type CustomRes struct {
	g.Meta `mime:"custom" sm:"自定义数据返回" dc:"本接口使用自定义数据返回，非OPEN API v3规范，具体返回数据字段请联系管理员获取"`
}

// EmptyRes 空数据
type EmptyRes struct {
	g.Meta `sm:"空数据返回" dc:"本接口返回空数据"`
}

func newCodeError(code int, message string, ext ...interface{}) error {
	var detail interface{}
	if len(ext) != 0 {
		detail = ext[0]
	}
	return gerror.NewCode(gcode.New(code, message, detail))
}

// CodeError 返回带错误码和错误详情的错误
func CodeError(code int, message string, ext ...interface{}) error {
	return newCodeError(code, message, ext)
}

func CodeErrorTranslate(ctx context.Context, code int, message string, ext ...interface{}) error {
	message = gi18n.T(ctx, message)
	return newCodeError(code, message, ext...)
}

// CodeErrorTranslateFormat 返回带错误码和模板格式化翻译信息的错误
func CodeErrorTranslateFormat(ctx context.Context, code int, format string, values ...interface{}) error {
	message := gi18n.Tf(ctx, format, values)
	return newCodeError(code, message)
}

// UnAuthorizedError 未授权错误 401
func UnAuthorizedError(ctx context.Context, ext ...interface{}) error {
	return CodeErrorTranslate(ctx, http.StatusUnauthorized, "Unauthorized", ext)
}

// SignatureError 签名错误 404
func SignatureError(ctx context.Context, ext ...interface{}) error {
	return CodeErrorTranslate(ctx, http.StatusPaymentRequired, "SignatureError", ext)
}

// InternalError 服务器内部错误
func InternalError(ctx context.Context, ext ...interface{}) error {
	return CodeErrorTranslate(ctx, http.StatusInternalServerError, "InternalServerError", ext)
}

// StandError 常规错误，最常用，返回一个统一code为-1的错误，支持多错误列表输出
func StandError(ctx context.Context, message string, ext ...interface{}) error {
	return CodeErrorTranslate(ctx, -1, message, ext)
}

// fixFrameError 补全框架的gcode自带错误码信息
func fixFrameError(ctx context.Context, code int, message string, ext ...interface{}) error {
	return CodeErrorTranslate(ctx, code, message, ext)
}
