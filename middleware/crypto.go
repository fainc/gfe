package middleware

import (
	"context"
	"errors"
	"fmt"

	cryptor "github.com/fainc/go-crypto/crypto"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"

	"github.com/fainc/gfe/response"
)

type crypto struct {
	ConfigName string
}

func Crypto(configName ...string) *crypto {
	if len(configName) == 0 {
		return &crypto{ConfigName: "default"}
	}
	return &crypto{ConfigName: configName[0]}
}

type cryptoConfig struct {
	Secret string
	Algo   string
	Hex    bool
}

func (rec *crypto) getConfig(ctx context.Context) (c cryptoConfig, err error) {
	c = cryptoConfig{
		Secret: g.Cfg().MustGet(ctx, fmt.Sprintf("crypto.%v.secret", rec.ConfigName)).String(),
		Algo:   g.Cfg().MustGet(ctx, fmt.Sprintf("crypto.%v.algo", rec.ConfigName)).String(),
		Hex:    g.Cfg().MustGet(ctx, fmt.Sprintf("crypto.%v.hex", rec.ConfigName), false).Bool(),
	}
	if c.Algo == "" || c.Secret == "" {
		err = errors.New("crypto config error")
	}
	return
}

func (rec *crypto) ResEncrypt(r *ghttp.Request) {
	c, err := rec.getConfig(r.Context())
	if err != nil {
		r.SetError(response.CodeError(500, "EncryptConfigError", err.Error()))
		return
	}
	r.SetCtxVar("response_encrypt", true)            // 是否返回加密数据
	r.SetCtxVar("response_encrypt_secret", c.Secret) // 加密密钥/证书
	r.SetCtxVar("response_encrypt_algo", c.Algo)     // 加密算法，支持"SM2_C1C3C2", "SM4_CBC", "RSA_PKCS1", "AES_CBC_PKCS7"
	r.SetCtxVar("response_encrypt_hex", c.Hex)       // 是否返回hex
	r.Middleware.Next()
}

func (rec *crypto) ReqDecrypt(r *ghttp.Request) {
	m := r.GetRequestMap()
	if len(m) > 0 {
		dt := r.GetRequest("data").String()
		if len(m) != 1 || dt == "" { // 不允许提交data外的其它数据，防止数据绕过解密
			r.SetError(response.CodeError(400, "EncryptRequestError", "request error"))
			return
		}
		decrypted, err := rec.doDecrypt(r.Context(), dt)
		if err != nil {
			r.SetError(response.CodeError(400, "EncryptRequestError", err.Error()))
			return
		}
		r.SetParam("ID", decrypted)
	}
	r.Middleware.Next()
}

func (rec *crypto) doDecrypt(ctx context.Context, data string) (decrypted string, err error) {
	c, err := rec.getConfig(ctx)
	if err != nil {
		panic(err.Error())
	}
	var s string
	switch c.Algo {
	case "SM4_CBC", "AES_CBC_PKCS7", "SM2_C1C3C2", "RSA_PKCS1":
		s = c.Secret // 对称加密使用密钥解密
	default:
		err = errors.New("unsupported algo")
		return
	}
	decrypted, err = cryptor.EasyEncrypt(c.Algo, s, data, c.Hex)
	return
}
