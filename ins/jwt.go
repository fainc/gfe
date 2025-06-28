package ins

import (
	"context"
	ecdsa2 "crypto/ecdsa"
	"encoding/base64"
	"errors"
	"time"

	"github.com/fainc/go-crypto/ecdsa"
	"github.com/fainc/gojwt"
	"github.com/gogf/gf/v2/crypto/gmd5"
	"github.com/gogf/gf/v2/database/gredis"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/golang-jwt/jwt/v5"

	"github.com/fainc/gfe/cfg"
)

type JwtIns struct {
	Cfg    cfg.JwtCfg       // 暴露实例配置
	client *gojwt.JwtClient // 底层GoJwt
	rds    *gredis.Redis    // redis
}

// NewJwt 新的JWT实例
func NewJwt(jwtCfg cfg.JwtCfg) *JwtIns {
	c := jwtCfg
	if c.Public == "" && c.Private == "" {
		panic("jwt private / public key missing")
	}
	var pub *ecdsa2.PublicKey
	var pri *ecdsa2.PrivateKey
	if c.Public != "" {
		pubDer, err := base64.StdEncoding.DecodeString(c.Public)
		if err != nil {
			panic(err.Error())
		}
		pub, err = ecdsa.ParsePublicKeyFromDer(pubDer)
		if err != nil {
			panic(err.Error())
		}
	}
	if c.Private != "" {
		priDer, err := base64.StdEncoding.DecodeString(c.Private)
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
	var rds *gredis.Redis
	if c.Redis != "" {
		rds = g.Redis(c.Redis)
		_, err := rds.Get(context.Background(), "ping")
		if err != nil {
			panic(err.Error())
		}
	}
	return &JwtIns{Cfg: c, client: client, rds: rds}
}

// Validate Token核验
func (rec *JwtIns) Validate(token string) (claims *gojwt.TokenClaims, err error) {
	claims, err = rec.client.Validate(&gojwt.ValidateParams{
		Token:   token,
		Subject: rec.Cfg.Subject,
	})
	if err != nil { // 通用化错误
		err = errors.New("token is invalid")
	}
	return
}

// ParseRaw Validate无法正常处理时 解析原始token数据
func (rec *JwtIns) ParseRaw(token string) (claims jwt.MapClaims, err error) {
	claims, err = rec.client.ParseRaw(token)
	return
}

// Publish Token签发
func (rec *JwtIns) Publish(ctx context.Context, uid int64, uuid string, tenantId int64, audience []string, ext map[string]interface{}, duration time.Duration) (tk, jti string, err error) {
	r := g.RequestFromCtx(ctx)
	tk, jti, err = rec.client.Publish(&gojwt.IssueParams{
		Subject:  rec.Cfg.Subject,
		Audience: audience,
		Duration: duration,
		PayloadClaims: gojwt.PayloadClaims{
			UID:      uid,
			UUID:     uuid,
			TenantId: tenantId,
			RegIP:    r.GetClientIp(),
			RegUA:    gmd5.MustEncrypt(r.Request.UserAgent()),
			Ext:      ext,
		},
	})
	return
}

// IsRevoked 通过redis判断jwt是否吊销
func (rec *JwtIns) IsRevoked(ctx context.Context, jti string) (result bool, err error) {
	if rec.rds == nil {
		err = errors.New("redis is not init for current jwt instance")
		return
	}
	n, err := rec.rds.Exists(ctx, "jwt_block_"+jti)
	if err != nil {
		return
	}
	return n > 0, err
}

// Revoke 通过redis吊销redis
func (rec *JwtIns) Revoke(ctx context.Context, jti string, exp time.Time) (err error) {
	t := time.Until(exp).Seconds()
	if rec.rds == nil {
		err = errors.New("redis is not init for current jwt instance")
		return
	}
	err = rec.rds.SetEX(ctx, "jwt_block_"+jti, 1, gconv.Int64(t))
	return
}
