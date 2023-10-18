package middleware

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"

	"github.com/fainc/gfe/util"
)

func SignAuth(r *ghttp.Request) {
	g.Dump(GetServerSign(r))
	r.Middleware.Next()
}

// GetServerSign 获取服务端签名
func GetServerSign(r *ghttp.Request) (str string) {
	/*If header data is empty, it will be converted to the 'null' string*/
	headerMap := map[string]string{}
	// Base params 基础header签名信息
	headerMap["Method"] = r.Method                                      // HTTPMethod etc. GET(toUpper)
	headerMap["Uri"] = r.URL.Path                                       // RequestURL with out query params(get method)
	headerMap["Authorization"] = r.GetHeader("Authorization")           // Common Jwt Authorization
	headerMap["NonceStr"] = r.GetHeader("NonceStr")                     // Common nonce random string (max 32)
	headerMap["Timestamp"] = r.GetHeader("Timestamp")                   // Timestamp millisecond
	headerStr := util.SignatureStr(headerMap, nil, nil)                 // header map to string
	payloadStr := util.SignatureComplexStr(r.GetRequestMap(), nil, nil) // payload map to string
	str = "header=" + headerStr + "&payload=" + payloadStr              // header + payload
	return
}
