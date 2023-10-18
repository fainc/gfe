package middleware

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/fainc/gojwt"
	"github.com/gogf/gf/v2/container/garray"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/util/gconv"

	"github.com/fainc/gfe/response"
	"github.com/fainc/gfe/util"
)

type jwt struct {
	ConfigName string
}

func Jwt(configName ...string) *jwt {
	if len(configName) == 0 {
		return &jwt{ConfigName: "default"}
	}
	return &jwt{ConfigName: configName[0]}
}

// jwtConfig  config file: supported hot reload, also supported dynamically set config content use g.Cfg().GetAdapter().(*gcfg.AdapterFile).SetContent()
// jwt:
//  default: // this is configName
//    secret: "xxx"
//    algo: "HS256"
//    subject: "xxx"
//    crypto: "jwt" // link to crypto.jwt
//    redis: "jwt" // link to redis.jwt
//
// crypto:
//  jwt:
//    secret: "xxx"
//    algo: "AES_CBC_PKCS7"
// 	  hex: true
//  other:
//    ...
//
// redis:
//  jwt:
//    address: 127.0.0.1:6379
//    db:      1
//  other:
//    ...

type jwtConfig struct {
	JwtAlgo      string
	JwtSecret    string
	Subject      string
	Crypto       string // 关联配置crypto组，不设置即不启用
	cryptoAlgo   string // 自动从关联配置crypto组读取
	cryptoSecret string // 自动从关联配置crypto组读取
	cryptoHex    bool   // 自动从关联配置crypto组读取
	Redis        string // 关联配置redis组，不设置即不启用
}

func (rec *jwt) GetConfig(ctx context.Context) (cfg jwtConfig, err error) {
	cfg = jwtConfig{
		JwtAlgo:   g.Cfg().MustGet(ctx, fmt.Sprintf("jwt.%v.algo", rec.ConfigName)).String(),
		JwtSecret: g.Cfg().MustGet(ctx, fmt.Sprintf("jwt.%v.secret", rec.ConfigName)).String(),
		Subject:   g.Cfg().MustGet(ctx, fmt.Sprintf("jwt.%v.subject", rec.ConfigName)).String(),
		Crypto:    g.Cfg().MustGet(ctx, fmt.Sprintf("jwt.%v.crypto", rec.ConfigName)).String(),
		Redis:     g.Cfg().MustGet(ctx, fmt.Sprintf("jwt.%v.redis", rec.ConfigName)).String(),
	}
	if cfg.JwtAlgo == "" || cfg.JwtSecret == "" {
		err = errors.New("jwt config error")
		return
	}
	if cfg.Crypto != "" {
		cfg.cryptoAlgo = g.Cfg().MustGet(ctx, fmt.Sprintf("crypto.%v.algo", cfg.Crypto)).String()
		cfg.cryptoSecret = g.Cfg().MustGet(ctx, fmt.Sprintf("crypto.%v.secret", cfg.Crypto)).String()
		cfg.cryptoHex = g.Cfg().MustGet(ctx, fmt.Sprintf("crypto.%v.hex", cfg.Crypto)).Bool()
		if cfg.cryptoSecret == "" || cfg.cryptoAlgo == "" {
			err = errors.New("jwt crypto config error")
			return
		}
	}
	return
}
func (rec *jwt) Auth(r *ghttp.Request) {
	var whiteTables g.SliceStr
	if util.GetReqMetaStr(r, "x-jwt-pass") == "true" {
		whiteTables = append(whiteTables, r.URL.Path) // req 定义免验证 自动加入白名单
	}
	cfg, err := rec.GetConfig(r.Context())
	if err != nil {
		r.SetError(response.CodeError(500, err.Error(), nil))
		return
	}
	passErr, err := rec.parser(r, cfg, whiteTables)
	if err != nil && !passErr {
		r.SetError(response.CodeError(401, err.Error(), nil))
		return
	}
	r.Middleware.Next()
}

func (rec *jwt) parser(r *ghttp.Request, cfg jwtConfig, whiteTables g.SliceStr) (passErr bool, err error) {
	c, err := gojwt.Parser(gojwt.ParserConf{
		JwtAlgo:      cfg.JwtAlgo,
		JwtSecret:    cfg.JwtSecret,
		CryptoSecret: cfg.cryptoSecret,
	}).Validate(gojwt.ValidateParams{
		Token:    r.GetHeader("Authorization"),
		Subject:  cfg.Subject,
		Audience: "",
	})
	var revoked = false
	if err == nil && cfg.Redis != "" {
		if revoked, err = rec.IsRedisRevoked(cfg.Redis, c.ID); err != nil {
			panic(err.Error())
		}
	}
	if err != nil || revoked {
		if len(whiteTables) != 0 {
			whiteTable := garray.NewStrArrayFrom(whiteTables)
			if whiteTable.ContainsI(r.URL.Path) {
				passErr = true
			}
		}
		return
	}
	r.SetCtxVar("TOKEN_UID", c.UserID)
	r.SetCtxVar("TOKEN_JTI", c.ID)
	r.SetCtxVar("TOKEN_EXP", c.ExpiresAt)
	return
}

// Publish 代理方法，简化签发
func (rec *jwt) Publish(ctx context.Context, UID int64, duration time.Duration) (tk, jti string, err error) {
	cfg, err := rec.GetConfig(ctx)
	if err != nil {
		return
	}
	tk, jti, err = gojwt.Issuer(gojwt.IssuerConf{
		JwtAlgo:      cfg.JwtAlgo,
		JwtSecret:    cfg.JwtSecret,
		CryptoAlgo:   cfg.cryptoAlgo,
		CryptoSecret: cfg.cryptoSecret,
	}).Publish(&gojwt.IssueParams{
		Subject:  cfg.Subject,
		UserID:   gconv.String(UID),
		Duration: duration,
		Audience: nil,
	})
	return
}

type CtxJwtUser struct {
	ID      int64
	TokenID string
	Exp     time.Time
}

// GetCtxUser 获取CTX用户ID信息
func (rec *jwt) GetCtxUser(ctx context.Context) CtxJwtUser {
	r := g.RequestFromCtx(ctx)
	if r == nil {
		panic("get request from ctx failed")
	}
	return CtxJwtUser{
		ID:      r.GetCtxVar("TOKEN_UID").Int64(),
		TokenID: r.GetCtxVar("TOKEN_JTI").String(),
		Exp:     r.GetCtxVar("TOKEN_EXP").Time(),
	}
}
func (rec *jwt) IsRedisRevoked(redisConfig, jti string) (result bool, err error) {
	n, err := g.Redis(redisConfig).Exists(context.Background(), "jwt_block_"+jti)
	if err != nil {
		return
	}
	return n > 0, err
}
