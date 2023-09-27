package response

import (
	"net/http"

	"github.com/fainc/go-crypto/gm"
	"github.com/fainc/go-crypto/rsa"
	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/text/gstr"
	"github.com/gogf/gf/v2/util/gmeta"
)

// Handler 默认数据返回中间件
func Handler(r *ghttp.Request) {
	var (
		ctx  = r.Context()
		err  = r.GetError()
		res  = r.GetHandlerResponse()
		code = gerror.Code(err)
	)

	// openapi
	if gstr.Contains(r.RequestURI, "api.json") {
		return
	}

	// 已有err错误
	if err != nil {
		if code.Code() == 50 || code.Code() == 52 || code.Code() == 500 { // 服务器错误
			Json().InternalError(ctx, g.I18n().Translate(ctx, "InternalError"))
			return
		}
		if code.Code() == 401 { // 登录
			Json().UnAuthorizedError(ctx, code.Message(), code.Detail())
			return
		}
		if code.Code() == 402 { // 解密
			Json().DecryptError(ctx, code.Message(), code.Detail())
			return
		}
		if code.Code() == 403 { // 签名
			Json().SignatureError(ctx, code.Message(), code.Detail())
			return
		}
		if code.Code() == 404 { // 404
			Json().NotFoundError(ctx, code.Message())
			return
		}
		Json().Error(ctx, err.Error(), code.Code(), code.Detail()) // 常规错误
		return
	}
	// 已退出程序流程，不走当前中间件
	if r.IsExited() {
		return
	}

	// 自定义输出
	if gmeta.Get(res, "mime").String() == "custom" {
		return
	}

	// 状态码正常且已有buffer内容
	if r.Response.Status == http.StatusOK && r.Response.BufferLength() > 0 {
		return
	}

	// 已有异常响应状态码
	if r.Response.Status > 0 && r.Response.Status != http.StatusOK {
		switch r.Response.Status {
		case http.StatusNotFound: // 404
			Json().NotFoundError(ctx, g.I18n().Translate(ctx, "NotFound"))
			return
		case http.StatusUnauthorized: // 401
			return
		case http.StatusBadRequest: // 400
			return
		default:
			Json().InternalError(ctx, g.I18n().Translate(ctx, "InternalError"))
			return
		}
	}

	/*CTX判断是否加密 etc.*/
	// r.SetCtxVar("response_encrypt", true)
	// r.SetCtxVar("response_encrypt_algorithm", cert.Algorithm)
	// r.SetCtxVar("response_encrypt_key", cert.Public)
	// r.SetCtxVar("response_encrypt_hex", true)
	var (
		encrypt = r.GetCtxVar("response_encrypt", false).Bool()
	)
	if !encrypt || res == nil { // CTX声明不加密或空数据时不处理加密，直接返回
		Json().Success(ctx, res)
		return
	}
	var (
		encryptAlgorithm = r.GetCtxVar("response_encrypt_algorithm", "").String()
	)
	if encrypt && encryptAlgorithm == "" { // CTX加密算法为空
		Json().InternalError(ctx, g.I18n().Translate(ctx, "UnsupportedEncryptAlgorithm"))
		return
	}
	var (
		encryptKey     = r.GetCtxVar("response_encrypt_key", "").String()
		encryptHex     = r.GetCtxVar("response_encrypt_hex", false).Bool() // 是否使用hex，否则输出base64
		encryptSM2Mode = r.GetCtxVar("response_encrypt_sm2_mode", 0).Int()
		encryptSM4Mode = r.GetCtxVar("response_encrypt_sm4_mode", "ECB").String()
	)
	if encryptKey == "" { // CTX加密密钥/证书为空
		Json().InternalError(ctx, g.I18n().Translate(ctx, "EncryptKeyError"))
		return
	}
	switch encryptAlgorithm {
	case "SM2":
		res, err = gm.Sm2().Encrypt(encryptKey, gjson.MustEncodeString(res), encryptHex, encryptSM2Mode)
		if err != nil {
			Json().InternalError(ctx, g.I18n().Translate(ctx, "DataEncryptFailed"))
			return
		}
	case "SM4":
		res, err = gm.Sm4().Encrypt(encryptSM4Mode, encryptKey, gjson.MustEncodeString(res), encryptHex)
		if err != nil {
			Json().InternalError(ctx, g.I18n().Translate(ctx, "DataEncryptFailed"))
			return
		}
	case "RSA_PKCS1":
		res, err = rsa.Encrypt(gjson.MustEncodeString(res), encryptKey, encryptHex)
		if err != nil {
			Json().InternalError(ctx, g.I18n().Translate(ctx, "DataEncryptFailed"))
			return
		}
	default:
		Json().InternalError(ctx, g.I18n().Translate(ctx, "UnsupportedEncryptAlgorithm"))
		return
	}
	Json(JsonOptions{
		Encrypt:          encrypt,
		EncryptAlgorithm: encryptAlgorithm,
	}).Success(ctx, res)
}
