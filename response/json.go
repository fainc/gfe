package response

import (
	"context"
	"os"

	"github.com/gogf/gf/v2/crypto/gmd5"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/i18n/gi18n"
	"github.com/gogf/gf/v2/os/gtime"
)

type json struct {
	JsonOptions
}

// var jsonVar = json{}

type JsonOptions struct {
	Encrypt          bool
	EncryptAlgorithm string
}

func Json(options ...JsonOptions) *json {
	if len(options) == 0 {
		return &json{}
	}
	return &json{options[0]}
}

func (rec *json) Success(ctx context.Context, data interface{}) {
	rec.Writer(ctx, data, "Success", 200, 0, nil)
}

func (rec *json) InternalError(ctx context.Context, message string) {
	rec.Writer(ctx, nil, message, 500, 500, nil)
}

// Error errCode 建议使用400通用错误码（或自定义四位错误码与HTTP状态码区分开），一般不使用公共的401,402,403,404错误码，除非明确输出的是相关内容
func (rec *json) Error(ctx context.Context, message string, errCode int, ext interface{}) {
	rec.Writer(ctx, nil, message, 400, errCode, ext)
}
func (rec *json) UnAuthorizedError(ctx context.Context, message string, ext interface{}) {
	rec.Writer(ctx, nil, message, 401, 401, ext)
}
func (rec *json) DecryptError(ctx context.Context, message string, ext interface{}) {
	rec.Writer(ctx, nil, message, 402, 402, ext)
}
func (rec *json) SignatureError(ctx context.Context, message string, ext interface{}) {
	rec.Writer(ctx, nil, message, 403, 403, ext)
}
func (rec *json) NotFoundError(ctx context.Context, message string) {
	rec.Writer(ctx, nil, message, 404, 404, nil)
}

type JsonFormat struct {
	Code      int         `json:"code"`                // 业务码，用于业务判断（首选），固定 200/400/401/404/500，与http状态码同步，通常取该值判断是否有错误需要处理
	ErrorCode int         `json:"errorCode"`           // 错误码，用于业务判断（可选），-1/400(通用错误)/51(参数验证错误)/401/404/500/other，通常忽略该值，除非业务需要判断详细错误类型（例：交易场景，交易失败返回400业务码时，返回余额不足、账户冻结等详细错误码用于后续业务处理）
	Message   interface{} `json:"message"`             // 消息，业务码非400时固定输出 Success/Authorization/NotFound/InternalError中的一种，400时输出详细错误（可能含i18n转译，建议仅展示或记录信息，不可用于业务判断）
	Data      interface{} `json:"data"`                // 返回数据
	Runtime   int64       `json:"runtime"`             // 当前任务执行耗时（ms）
	Lang      string      `json:"lang"`                // 当前上下文语言
	Ext       interface{} `json:"ext"`                 // 拓展数据（可能含有多个错误详情或其他附加数据，例：复杂登录场景下的401错误返回具体登录地址）
	Encrypted bool        `json:"encrypted"`           // 数据是否加密
	Algorithm string      `json:"algorithm,omitempty"` // 数据加密算法
}

// Writer 标准格式数据输出
func (rec *json) Writer(ctx context.Context, data interface{}, message string, code int, errCode int, ext interface{}) {
	r := g.RequestFromCtx(ctx)
	r.Response.WriteStatus(code)
	r.Response.ClearBuffer()
	serverName, _ := os.Hostname()
	serverID, _ := gmd5.Encrypt(serverName)
	r.Response.WriteJson(JsonFormat{
		Code:      code,
		Message:   message,
		Data:      data,
		Ext:       ext,
		Lang:      gi18n.LanguageFromCtx(ctx),
		ErrorCode: errCode,
		Runtime:   gtime.Now().TimestampMilli() - r.EnterTime,
		Encrypted: rec.Encrypt,
		Algorithm: rec.EncryptAlgorithm,
	})
	r.Response.Header().Set("Content-Type", "application/json;charset=utf-8")
	r.Response.Header().Set("Server-Id", serverID)
	r.ExitAll()
}

// CustomWriter  自定义json输出
func (rec *json) CustomWriter(ctx context.Context, data interface{}, status ...int) {
	statusCode := 200
	if len(status) >= 1 && status[0] != 200 {
		statusCode = status[0]
	}
	serverName, _ := os.Hostname()
	serverID, _ := gmd5.Encrypt(serverName)
	r := g.RequestFromCtx(ctx)
	r.Response.WriteStatus(statusCode)
	r.Response.ClearBuffer()
	r.Response.WriteJson(data)
	r.Response.Header().Set("Content-Type", "application/json;charset=utf-8")
	r.Response.Header().Set("Server-Id", serverID)
	r.ExitAll()
}
