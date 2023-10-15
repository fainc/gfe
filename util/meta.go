package util

import (
	"reflect"

	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/util/gmeta"
)

func GetReqMetaStr(r *ghttp.Request, key string) string {
	if r.GetServeHandler() != nil && r.GetServeHandler().Handler.Info.Type.String() != "func(*ghttp.Request)" {
		var objectReq = reflect.New(r.GetServeHandler().Handler.Info.Type.In(1))
		return gmeta.Get(objectReq, key).String()
	}
	return ""
}
