package middleware

import (
	"context"
	ecdsa2 "crypto/ecdsa"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/fainc/go-crypto/ecdsa"
	"github.com/fainc/gojwt"
	"github.com/gogf/gf/v2/container/garray"
	"github.com/gogf/gf/v2/crypto/gmd5"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"

	"github.com/fainc/gfe/helper"
	"github.com/fainc/gfe/response"
	"github.com/fainc/gfe/util"
)

type jwt struct {
	cfg    jwtConfig
	client *gojwt.JwtClient
}

// NewJwt 建议使用单例，降低处理密钥开销
func NewJwt(ctx context.Context, configName ...string) *jwt {
	cname := "default"
	if len(configName) != 0 && configName[0] != "" {
		cname = configName[0]
	}
	cfg := jwtConfig{
		Public:  g.Cfg().MustGet(ctx, fmt.Sprintf("jwt.%v.public", cname)).String(),
		Private: g.Cfg().MustGet(ctx, fmt.Sprintf("jwt.%v.private", cname)).String(),
		Subject: g.Cfg().MustGet(ctx, fmt.Sprintf("jwt.%v.subject", cname)).String(),
		Redis:   g.Cfg().MustGet(ctx, fmt.Sprintf("jwt.%v.redis", cname)).String(),
		AuthUA:  g.Cfg().MustGet(ctx, fmt.Sprintf("jwt.%v.authUA", cname)).Bool(),
		AuthIP:  g.Cfg().MustGet(ctx, fmt.Sprintf("jwt.%v.authIP", cname)).Bool(),
	}
	if cfg.Public == "" && cfg.Private == "" {
		panic("jwt private / public key missing")
	}
	var pub *ecdsa2.PublicKey
	var pri *ecdsa2.PrivateKey
	if cfg.Public != "" {
		pubDer, err := base64.StdEncoding.DecodeString(cfg.Public)
		if err != nil {
			panic(err.Error())
		}
		pub, err = ecdsa.ParsePublicKeyFromDer(pubDer)
		if err != nil {
			panic(err.Error())
		}
	}
	if cfg.Private != "" {
		priDer, err := base64.StdEncoding.DecodeString(cfg.Private)
		if err != nil {
			panic(err.Error())
		}
		pri, err = ecdsa.ParsePrivateKeyFromDer(priDer)
		if err != nil {
			panic(err.Error())
		}
	}

	client := gojwt.NewJwt(gojwt.JwtConfig{
		JwtAlgo:    "ES256",
		JwtPublic:  pub,
		JwtPrivate: pri,
	})
	return &jwt{cfg: cfg, client: client}
}

type jwtConfig struct {
	Private string // 私钥加签
	Public  string // 公钥验签
	Subject string
	Redis   string // 关联配置redis组，不设置即不启用
	AuthUA  bool   // 是否校验客户端信息（UA相对稳定，可选）
	AuthIP  bool   // 是否强校验IP（严苛内部IP场景使用，面向客户端一般不启用）
}

func (rec *jwt) Auth(r *ghttp.Request) {
	var whiteTables g.SliceStr
	if util.GetReqMetaStr(r, "x-jwt-ignore") == "true" {
		whiteTables = append(whiteTables, r.URL.Path) // req 定义免验证 自动加入白名单
	}
	inWhite := rec.inWhiteTable(whiteTables, r.URL.Path)
	cfg := rec.cfg
	tk, err := rec.parser(r.GetHeader("Authorization"), cfg)
	if err == nil {
		if cfg.Redis != "" {
			revoked, redisErr := rec.IsRevoked(cfg.Redis, tk.ID)
			if redisErr != nil {
				panic(redisErr.Error())
			}
			if revoked {
				err = errors.New("token is revoked")
			}
		}
		if err == nil && cfg.AuthUA && tk.RegUA != gmd5.MustEncrypt(r.Request.UserAgent()) {
			err = errors.New("current ua is not trusted")
		}
		if err == nil && cfg.AuthIP && tk.RegIP != r.GetClientIp() {
			err = errors.New("current ip is not trusted")
		}
	}
	if err != nil && !inWhite {
		r.SetError(response.CodeError(401, err.Error(), nil))
		return
	}
	if err == nil && tk != nil {
		helper.Ctx(r.Context()).SetUser(helper.CtxUser{
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
func (rec *jwt) parser(token string, cfg jwtConfig) (c *gojwt.TokenClaims, err error) {
	c, err = rec.client.Validate(&gojwt.ValidateParams{
		Token:   token,
		Subject: cfg.Subject,
	})
	if err != nil { // 通用化错误
		err = response.CodeError(-1, "xxx", "")
		panic(err)
	}
	return
}

// Publish 代理方法，简化签发
func (rec *jwt) Publish(ctx context.Context, uid int64, audience []string, ext map[string]interface{}, duration time.Duration) (tk, jti string, err error) {
	r := g.RequestFromCtx(ctx)
	cfg := rec.cfg
	if err != nil {
		return
	}
	tk, jti, err = rec.client.Publish(&gojwt.IssueParams{
		Subject:  cfg.Subject,
		Audience: audience,
		Duration: duration,
		PayloadClaims: gojwt.PayloadClaims{
			UID:   uid,
			RegIP: r.GetClientIp(),
			RegUA: gmd5.MustEncrypt(r.Request.UserAgent()),
			Ext:   ext,
		},
	})
	return
}

func (rec *jwt) IsRevoked(redisConfig, jti string) (result bool, err error) {
	n, err := g.Redis(redisConfig).Exists(context.Background(), "jwt_block_"+jti)
	if err != nil {
		return
	}
	return n > 0, err
}
