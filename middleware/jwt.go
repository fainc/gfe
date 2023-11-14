package middleware

import (
	"context"
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
			revoked, redisErr := rec.IsRevoked(c.Redis, tk.ID)
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
		r.SetError(response.CodeError(401, err.Error(), nil))
		return
	}
	if err == nil && tk != nil {
		helper.CtxUser().Set(r.Context(), helper.CtxUserInfo{
			UID:         tk.UID,
			JTI:         tk.ID,
			Exp:         tk.ExpiresAt.Time,
			RegIP:       tk.RegIP,
			RegUA:       tk.RegUA,
			RegDeviceID: tk.RegDeviceID,
			Ext:         tk.Ext,
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

func (rec *jwt) IsRevoked(redisConfig, jti string) (result bool, err error) {
	n, err := g.Redis(redisConfig).Exists(context.Background(), "jwt_block_"+jti)
	if err != nil {
		return
	}
	return n > 0, err
}
