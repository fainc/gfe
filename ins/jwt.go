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
	"github.com/gogf/gf/v2/frame/g"
	"github.com/golang-jwt/jwt/v5"

	"github.com/fainc/gfe/cfg"
)

type JwtIns struct {
	Cfg    cfg.JwtCfg       // 暴露实例配置
	client *gojwt.JwtClient // 暴露底层GoJwt
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
	return &JwtIns{Cfg: c, client: client}
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
func (rec *JwtIns) Publish(ctx context.Context, uid int64, audience []string, ext map[string]interface{}, duration time.Duration) (tk, jti string, err error) {
	r := g.RequestFromCtx(ctx)
	if err != nil {
		return
	}
	tk, jti, err = rec.client.Publish(&gojwt.IssueParams{
		Subject:  rec.Cfg.Subject,
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
