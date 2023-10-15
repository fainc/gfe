package middleware

import (
	"fmt"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"

	"github.com/fainc/gfe/util"
)

func GetSignMap(r *ghttp.Request) (str string) {
	/*If header data is empty, it will be converted to the 'null' string*/
	headerMap := map[string]string{}
	headerMap["HTTPMethod"] = r.Method                                  // HTTPMethod etc. GET(toUpper)
	headerMap["RequestURL"] = GetSignUrl(r)                             // RequestURL with out query params(get method)
	headerMap["AppId"] = r.GetHeader("AppId")                           // Common Api AppId
	headerMap["Authorization"] = r.GetHeader("Authorization")           // Common Jwt Authorization
	headerMap["NonceStr"] = r.GetHeader("NonceStr")                     // Common nonce random string (max 32)
	headerMap["Timestamp"] = r.GetHeader("Timestamp")                   // Timestamp millisecond
	headerMap["DeviceId"] = r.GetHeader("DeviceId")                     // DeviceId
	headerMap["AppVersionCode"] = r.GetHeader("AppVersionCode")         // AppVersionCode
	headerMap["Platform"] = r.GetHeader("Platform")                     // Platform
	headerStr := util.SignatureStr(headerMap, nil, nil)                 // header map to string
	payloadStr := util.SignatureComplexStr(r.GetRequestMap(), nil, nil) // payload map to string
	str = "header=" + headerStr + "&payload=" + payloadStr              // header + payload
	return
}
func GetSignUrl(r *ghttp.Request) string {
	return fmt.Sprintf(`%s%s`, g.Cfg().MustGet(r.Context(), "server.signHost"), r.URL.Path)
}
