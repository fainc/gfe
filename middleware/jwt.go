package middleware

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/fainc/gojwt"
	"github.com/gogf/gf/v2/container/garray"
	"github.com/gogf/gf/v2/crypto/gmd5"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"

	"github.com/fainc/gfe/response"
	"github.com/fainc/gfe/util"
)

type jwt struct {
	ConfigName string
}

func Jwt(configName ...string) *jwt {
	if len(configName) == 0 || configName[0] == "" {
		return &jwt{ConfigName: "default"}
	}
	return &jwt{ConfigName: configName[0]}
}

type jwtConfig struct {
	JwtAlgo   string
	JwtSecret string
	Subject   string
	Redis     string // 关联配置redis组，不设置即不启用
	AuthUA    bool   // 是否校验客户端信息（UA相对稳定，可选）
	AuthIP    bool   // 是否强校验IP（严苛内部IP场景使用，面向客户端一般不启用）
}

func (rec *jwt) GetConfig(ctx context.Context) (cfg jwtConfig, err error) {
	cfg = jwtConfig{
		JwtAlgo:   g.Cfg().MustGet(ctx, fmt.Sprintf("jwt.%v.algo", rec.ConfigName)).String(),
		JwtSecret: g.Cfg().MustGet(ctx, fmt.Sprintf("jwt.%v.secret", rec.ConfigName)).String(),
		Subject:   g.Cfg().MustGet(ctx, fmt.Sprintf("jwt.%v.subject", rec.ConfigName)).String(),
		Redis:     g.Cfg().MustGet(ctx, fmt.Sprintf("jwt.%v.redis", rec.ConfigName)).String(),
		AuthUA:    g.Cfg().MustGet(ctx, fmt.Sprintf("jwt.%v.authUA", rec.ConfigName)).Bool(),
		AuthIP:    g.Cfg().MustGet(ctx, fmt.Sprintf("jwt.%v.authIP", rec.ConfigName)).Bool(),
	}
	if cfg.JwtAlgo == "" || cfg.JwtSecret == "" {
		err = errors.New("jwt config error")
		return
	}
	return
}
func (rec *jwt) Auth(r *ghttp.Request) {
	var whiteTables g.SliceStr
	if util.GetReqMetaStr(r, "x-jwt-pass") == "true" {
		whiteTables = append(whiteTables, r.URL.Path) // req 定义免验证 自动加入白名单
	}
	inWhite := rec.inWhiteTable(whiteTables, r.URL.Path)
	cfg, err := rec.GetConfig(r.Context())
	if err != nil {
		r.SetError(response.CodeError(500, err.Error(), nil))
		return
	}
	tk, err := rec.parser(r.GetHeader("Authorization"), cfg)
	if err == nil {
		if cfg.Redis != "" {
			revoked, redisErr := rec.IsRedisRevoked(cfg.Redis, tk.ID)
			if redisErr != nil {
				panic(redisErr.Error())
			}
			if revoked {
				err = errors.New("token is revoked")
			}
		}
		if err == nil && cfg.AuthUA && tk.UA != gmd5.MustEncrypt(r.Request.UserAgent()) {
			err = errors.New("current ua is not trusted")
		}
		if err == nil && cfg.AuthIP && tk.IP != r.GetClientIp() {
			err = errors.New("current ip is not trusted")
		}
	}
	if err != nil && !inWhite {
		r.SetError(response.CodeError(401, err.Error(), nil))
		return
	}
	if err == nil && tk != nil {
		r.SetCtxVar("TOKEN_UID", tk.UID)
		r.SetCtxVar("TOKEN_JTI", tk.ID)
		r.SetCtxVar("TOKEN_EXP", tk.ExpiresAt)
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
	c, err = gojwt.NewJwt(gojwt.JwtConfig{
		JwtAlgo:   cfg.JwtAlgo,
		JwtSecret: cfg.JwtSecret,
	}).Validate(&gojwt.ValidateParams{
		Token:   token,
		Subject: cfg.Subject,
	})
	if err != nil { // 通用化错误
		err = errors.New("token is invalid")
	}
	return
}

// Publish 代理方法，简化签发
func (rec *jwt) Publish(ctx context.Context, uid int64, duration time.Duration) (tk, jti string, err error) {
	r := g.RequestFromCtx(ctx)
	cfg, err := rec.GetConfig(ctx)
	if err != nil {
		return
	}
	tk, jti, err = gojwt.NewJwt(gojwt.JwtConfig{
		JwtAlgo:   "HS256",
		JwtSecret: cfg.JwtSecret,
	}).Publish(&gojwt.IssueParams{
		Subject:  cfg.Subject,
		UID:      uid,
		Duration: duration,
		IP:       r.GetClientIp(),
		UA:       gmd5.MustEncrypt(r.Request.UserAgent()),
	})
	return
}

func (rec *jwt) IsRedisRevoked(redisConfig, jti string) (result bool, err error) {
	n, err := g.Redis(redisConfig).Exists(context.Background(), "jwt_block_"+jti)
	if err != nil {
		return
	}
	return n > 0, err
}
