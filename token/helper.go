package token

import (
	"context"
	"errors"
	"time"

	"github.com/fainc/gojwt"
	"github.com/gogf/gf/v2/container/garray"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/util/gconv"
)

type helper struct{}

var helperVar = helper{}

func Helper() *helper {
	return &helperVar
}

type AuthParams struct {
	Aud         string
	WhiteTables g.SliceStr
	RevokeAuth  bool
}
type PublishParams struct {
	Aud      []string
	UserID   int64
	Duration time.Duration
}

var jwtSecret string

func init() {
	jwtSecret = g.Cfg().MustGet(context.Background(), "jwt.secret").String()
}

// Publish 发布HS256 NO CRYPTO auth token(可直接使用或作为示例自行开发)
func (rec *helper) Publish(p PublishParams) (token, jti string, err error) {
	token, jti, err = gojwt.Issuer(gojwt.IssuerConf{
		JwtAlgo:   gojwt.AlgoHS256,
		JwtSecret: jwtSecret,
	}).Publish(&gojwt.IssueParams{
		Subject:  "Auth",
		UserID:   gconv.String(p.UserID),
		Duration: p.Duration,
		Audience: p.Aud,
	})
	return
}

// Auth  HS256 NO CRYPTO auth 通用jwt验证和ctx写入(可直接使用或作为示例自行开发),通过err和catchErr判断拦截
func (rec *helper) Auth(r *ghttp.Request, p AuthParams) (userID int64, aud string, catchErr bool, err error) {
	c, err := gojwt.Parser(gojwt.ParserConf{
		JwtAlgo:   gojwt.AlgoHS256,
		JwtSecret: jwtSecret,
	}).Validate(gojwt.ValidateParams{
		Token:    r.GetHeader("Authorization"),
		Subject:  "Auth",
		Audience: p.Aud,
	})
	if err != nil {
		catchErr = true
		if len(p.WhiteTables) != 0 {
			whiteTable := garray.NewStrArrayFrom(p.WhiteTables)
			if whiteTable.ContainsI(r.URL.Path) {
				catchErr = false
			}
		}
		return
	}
	if p.RevokeAuth {
		var revoked bool
		if revoked, err = IsRedisRevoked(c.ID); err != nil {
			panic(err)
		}
		if revoked {
			return 0, "", true, errors.New("当前登录凭证已被注销")
		}
	}
	userID = gconv.Int64(c.UserID)
	aud = p.Aud
	r.SetCtxVar("TOKEN_UID", userID)
	r.SetCtxVar("TOKEN_JTI", c.ID)
	r.SetCtxVar("TOKEN_AUD", aud)
	r.SetCtxVar("TOKEN_EXP", c.ExpiresAt)
	return
}

type CtxJwtUser struct {
	ID      int64
	TokenID string
	Aud     string
	Exp     time.Time
}

// GetCtxUser 获取CTX用户ID信息
func (*helper) GetCtxUser(ctx context.Context) CtxJwtUser {
	r := g.RequestFromCtx(ctx)
	if r == nil {
		return CtxJwtUser{}
	}
	return CtxJwtUser{
		ID:      r.GetCtxVar("TOKEN_UID").Int64(),
		TokenID: r.GetCtxVar("TOKEN_JTI").String(),
		Aud:     r.GetCtxVar("TOKEN_AUD").String(),
		Exp:     r.GetCtxVar("TOKEN_EXP").Time(),
	}
}
