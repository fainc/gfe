package response

import (
	"errors"
	"reflect"

	cryptor "github.com/fainc/go-crypto/crypto"
	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/net/ghttp"
)

type encryptConf struct {
	Encrypt bool
	Algo    string
	Secret  string
	Hex     bool
}

// ctxEncryptConf 获取上下文数据输出加密配置
func ctxEncryptConf(r *ghttp.Request) (conf encryptConf) {
	conf = encryptConf{
		Encrypt: r.GetCtxVar("response_encrypt", false).Bool(),
		Algo:    r.GetCtxVar("response_encrypt_algo", "").String(),
		Secret:  r.GetCtxVar("response_encrypt_secret", "").String(),
		Hex:     r.GetCtxVar("response_encrypt_hex", false).Bool(),
	}
	return
}

func tryEncrypt(r *ghttp.Request, mime string, res interface{}) (isEncrypted bool, result interface{}, err error) {
	if mime != MimeJSON && mime != MimeXML { // 仅json和xml返回加密数据
		return false, res, nil
	}
	config := ctxEncryptConf(r)
	if !config.Encrypt || interfaceIsNil(res) { // CTX声明不加密或未初始化返回数据时直接返回
		return false, res, nil
	}
	dt := gjson.MustEncodeString(res)                         // 数据转为JSON加密
	if dt == "" || dt == "null" || dt == "{}" || dt == "[]" { // 待加密数据无实质意义直接返回
		return false, res, nil
	}
	if config.Encrypt && config.Algo == "" { // CTX加密算法为空
		err = errors.New("config error:algo")
		return
	}
	if config.Secret == "" { // CTX加密密钥/证书为空
		err = errors.New("config error:secret")
		return
	}
	switch config.Algo {
	case "SM2_C1C3C2", "SM4_CBC", "RSA_PKCS1", "AES_CBC_PKCS7":
		result, err = cryptor.EasyEncrypt(config.Algo, config.Secret, dt, config.Hex)
		if err != nil {
			return
		}
	default:
		err = errors.New("unsupported algo")
		return
	}
	return
}

func interfaceIsNil(i interface{}) bool {
	vi := reflect.ValueOf(i)
	if vi.Kind() == reflect.Ptr {
		return vi.IsNil()
	}
	return false
}
