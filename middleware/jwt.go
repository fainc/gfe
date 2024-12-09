package middleware

import (
	"errors"

	"github.com/gogf/gf/v2/container/garray"
	"github.com/gogf/gf/v2/crypto/gmd5"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"

	"github.com/fainc/gfe/helper"
	"github.com/fainc/gfe/ins"
	"github.com/fainc/gfe/response"
	"github.com/fainc/gfe/util"
)

type jwt struct {
	i *ins.JwtIns
}

// Jwt 使用jwt实例注册中间件
func Jwt(jwtIns *ins.JwtIns) *jwt {
	return &jwt{i: jwtIns}
}

func (rec *jwt) Register(r *ghttp.Request) {
	var whiteTables g.SliceStr
	if util.GetReqMetaStr(r, "x-jwt-ignore") == "true" {
		whiteTables = append(whiteTables, r.URL.Path) // req 定义免验证 自动加入白名单
	}
	inWhite := rec.inWhiteTable(whiteTables, r.URL.Path)
	c := rec.i.Cfg
	tk, err := rec.i.Validate(r.GetHeader("Authorization"))
	if err == nil {
		if c.Redis != "" {
			revoked, redisErr := rec.i.IsRevoked(r.Context(), tk.ID)
			if redisErr != nil {
				panic(redisErr.Error())
			}
			if revoked {
				err = errors.New("token is revoked")
			}
		}
		if err == nil && c.AuthUA && tk.RegUA != gmd5.MustEncrypt(r.Request.UserAgent()) {
			err = errors.New("current ua is not trusted")
		}
		if err == nil && c.AuthIP && tk.RegIP != r.GetClientIp() {
			err = errors.New("current ip is not trusted")
		}
	}
	if err != nil && !inWhite {
		r.SetError(response.UnAuthorizedError(r.Context(), err.Error()))
		return
	}
	if err == nil && tk != nil {
		helper.CtxUser().Set(r.Context(), helper.CtxUserInfo{
			UID:         tk.UID,
			UUID:        tk.UUID,
			TenantID:    tk.Ext["tenantID"].(int64), // todo update tenantID
			JTI:         tk.ID,
			Exp:         tk.ExpiresAt.Time,
			RegIP:       tk.RegIP,
			RegUA:       tk.RegUA,
			RegDeviceID: tk.RegDeviceID,
			Ext:         tk.Ext,
			Subject:     rec.i.Cfg.Subject,
		})
	}
	r.Middleware.Next()
}
func (rec *jwt) inWhiteTable(whiteTables g.SliceStr, url string) (res bool) {
	if len(whiteTables) != 0 {
		whiteTable := garray.NewStrArrayFrom(whiteTables)
		if whiteTable.ContainsI(url) {
			res = true
		}
	}
	return
}
