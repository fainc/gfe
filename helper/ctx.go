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
	UID         int64
	JTI         string
	Exp         time.Time
	RegIP       string
	RegUA       string
	RegDeviceID string
	Ext         g.Map
}

// GetUser 获取CTX用户信息
func (rec *ctx) GetUser() CtxUser {
	r := g.RequestFromCtx(rec.c)
	if r == nil {
		panic("get request from ctx failed")
	}
	return CtxUser{
		UID:         r.GetCtxVar("TOKEN_UID").Int64(),
		JTI:         r.GetCtxVar("TOKEN_JTI").String(),
		Exp:         r.GetCtxVar("TOKEN_EXP").Time(),
		RegIP:       r.GetCtxVar("TOKEN_REG_IP").String(),
		RegUA:       r.GetCtxVar("TOKEN_REG_UA").String(),
		RegDeviceID: r.GetCtxVar("TOKEN_REG_DEVICE_ID").String(),
		Ext:         r.GetCtxVar("TOKEN_EXT").Map(),
	}
}

// SetUser 设置CTX用户信息
func (rec *ctx) SetUser(u CtxUser) {
	r := g.RequestFromCtx(rec.c)
	if r == nil {
		panic("get request from ctx failed")
	}
	r.SetCtxVar("TOKEN_UID", u.UID)
	r.SetCtxVar("TOKEN_JTI", u.JTI)
	r.SetCtxVar("TOKEN_EXP", u.Exp)
	r.SetCtxVar("TOKEN_REG_IP", u.RegIP)
	r.SetCtxVar("TOKEN_REG_UA", u.RegUA)
	r.SetCtxVar("TOKEN_REG_DEVICE_ID", u.RegDeviceID)
	r.SetCtxVar("TOKEN_EXT", u.Ext)
}
