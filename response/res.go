package response

import (
	"context"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/i18n/gi18n"
)

// CustomRes 自定义返回数据标注，使用该类型数据返回时，全局后置中间件ResponseHandler将不再处理返回数据，请自行提前输出
type CustomRes struct {
	g.Meta `mime:"custom" sm:"自定义数据返回" dc:"本接口使用自定义数据返回，非OPEN API v3规范，具体返回数据字段请联系管理员获取"`
}

// CodeError 返回带错误码错误，建议 code >= 1000 || == -1 (使用400标准错误http状态码, 400-500的错误码会和http状态码同步，其它< 1000的状态码统一500服务器未知错误)
func CodeError(code int, message string, ext interface{}) error {
	return gerror.NewCode(gcode.New(code, message, ext))
}

// CodeErrorTranslate 返回带错误码和翻译信息的错误，建议 code >= 1000 || == -1 (使用400标准错误http状态码, 400-500的错误码会和http状态码同步，其它< 1000的状态码统一500服务器未知错误)
func CodeErrorTranslate(ctx context.Context, code int, message string, ext interface{}) error {
	message = gi18n.T(ctx, message)
	return CodeError(code, message, ext)
}

// CodeErrorTranslateFormat 返回带错误码和模板格式化翻译信息的错误，建议 code >= 1000 || == -1 (使用400标准错误http状态码, 400-500的错误码会和http状态码同步，其它< 1000的状态码统一500服务器未知错误)
func CodeErrorTranslateFormat(ctx context.Context, code int, format string, values ...interface{}) error {
	message := gi18n.Tf(ctx, format, values)
	return CodeError(code, message, nil)
}
