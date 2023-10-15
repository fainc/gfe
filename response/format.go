package response

import (
	"context"
	"fmt"

	"github.com/gogf/gf/v2/i18n/gi18n"
)

const (
	MimeJSON = "JSON"
	MimeXML  = "XML"
	MimeHTML = "HTML"
)

type format struct {
	FormatOptions
}

type FormatOptions struct {
	encrypted bool   // 是否加密 handler
	Mime      string // 输出类型
	staticTpl string // HTML meta 指定静态渲染模板，优先级最低，handler 自动从 g.meta 读取
	errorTpl  string // HTML 错误信息模板，优先级最高，根据错误码自动指定
}

// Format 格式化输出
func Format(options ...FormatOptions) *format {
	if len(options) == 0 {
		return &format{}
	}
	return &format{options[0]}
}

// 以下是中间件专用方法(正常流程返回数据和错误可自动处理，非中间件不要直接调用这类方法)

func (rec *format) Success(ctx context.Context, data interface{}) {
	rec.writer(ctx, data, "OK", 200, 0, nil)
}

// Error 400通用错误
func (rec *format) Error(ctx context.Context, errCode int, message string, ext interface{}) {
	rec.errorTpl = fmt.Sprintf("error/%v.html", 400)
	rec.writer(ctx, nil, message, 400, errCode, ext)
}

// errorSyncHttpStatus 同步http状态码错误输出
func (rec *format) errorSyncHTTPStatus(ctx context.Context, code int, message string, ext interface{}) {
	rec.errorTpl = fmt.Sprintf("error/%v.html", code)
	rec.writer(ctx, nil, message, code, code, ext)
}
func (rec *format) UnAuthorized(ctx context.Context, ext interface{}) {
	rec.errorSyncHTTPStatus(ctx, 401, gi18n.Translate(ctx, "UnAuthorized"), ext)
}
func (rec *format) SignatureError(ctx context.Context, ext interface{}) {
	rec.errorSyncHTTPStatus(ctx, 402, gi18n.Translate(ctx, "SignatureError"), ext)
}
func (rec *format) Forbidden(ctx context.Context, ext interface{}) {
	rec.errorSyncHTTPStatus(ctx, 403, gi18n.Translate(ctx, "Forbidden"), ext)
}
func (rec *format) NotFound(ctx context.Context, ext interface{}) {
	rec.errorSyncHTTPStatus(ctx, 404, gi18n.Translate(ctx, "NotFound"), ext)
}
func (rec *format) InternalError(ctx context.Context, ext interface{}) {
	rec.errorSyncHTTPStatus(ctx, 500, gi18n.Translate(ctx, "InternalError"), ext)
}

type DataFormat struct {
	Code      int         `json:"code"`      // 业务状态码 与http状态码同步
	ErrorCode int         `json:"errorCode"` // 错误码，-1/400(通用错误)/51(参数验证错误)/401/404/500/other，通常忽略该值，除非业务需要判断详细错误类型（例：交易场景，交易失败返回400业务码时，返回余额不足、账户冻结等详细错误码用于后续业务处理）
	Message   interface{} `json:"message"`   // 错误消息
	Data      interface{} `json:"data"`      // 返回数据
	Ext       interface{} `json:"ext"`       // 拓展数据（可能含有多个错误详情或其他附加数据）
	Encrypted bool        `json:"encrypted"` // 数据是否加密
}

// Writer 标准格式数据输出
func (rec *format) writer(ctx context.Context, data interface{}, message string, code int, errCode int, ext interface{}) {
	w := Writer(rec.Mime)
	if rec.Mime == MimeHTML {
		tpl, err := rec.htmlRender(ctx, data)
		if err != nil {
			code = 500
			tpl = "Parse template file error"
		}
		w.Output(ctx, tpl, code)
		return
	}
	w.Output(ctx, DataFormat{
		Code:      code,
		Message:   message,
		Data:      data,
		Ext:       ext,
		ErrorCode: errCode,
		Encrypted: rec.encrypted,
	}, code)
}
