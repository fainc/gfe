package helper

import (
	"context"
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

type ctx struct {
	c context.Context
}

func Ctx(c context.Context) *ctx {
	return &ctx{c: c}
}

type CtxUser struct {
	ID      int64
	TokenID string
	Exp     time.Time
}

// GetUser 获取CTX用户ID信息
func (rec *ctx) GetUser() CtxUser {
	r := g.RequestFromCtx(rec.c)
	if r == nil {
		panic("get request from ctx failed")
	}
	return CtxUser{
		ID:      r.GetCtxVar("TOKEN_UID").Int64(),
		TokenID: r.GetCtxVar("TOKEN_JTI").String(),
		Exp:     r.GetCtxVar("TOKEN_EXP").Time(),
	}
}
