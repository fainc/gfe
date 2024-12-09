package helper

import (
	"context"
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

type ctxUser struct {
}

func CtxUser() *ctxUser {
	return &ctxUser{}
}

type CtxUserInfo struct {
	UID  int64
	UUID string
	//TenantID    int64
	JTI         string
	Exp         time.Time
	RegIP       string
	RegUA       string
	RegDeviceID string
	Subject     string
	Ext         g.Map
}

// Get 获取CTX用户信息
func (rec *ctxUser) Get(ctx context.Context) CtxUserInfo {
	r := g.RequestFromCtx(ctx)
	if r == nil {
		panic("get request from ctx failed")
	}
	return CtxUserInfo{
		UID:  r.GetCtxVar("TOKEN_UID").Int64(),
		UUID: r.GetCtxVar("TOKEN_UUID").String(),
		//TenantID:    r.GetCtxVar("TOKEN_TENANT_ID").Int64(),
		JTI:         r.GetCtxVar("TOKEN_JTI").String(),
		Exp:         r.GetCtxVar("TOKEN_EXP").Time(),
		RegIP:       r.GetCtxVar("TOKEN_REG_IP").String(),
		RegUA:       r.GetCtxVar("TOKEN_REG_UA").String(),
		RegDeviceID: r.GetCtxVar("TOKEN_REG_DEVICE_ID").String(),
		Ext:         r.GetCtxVar("TOKEN_EXT").Map(),
		Subject:     r.GetCtxVar("TOKEN_SUBJECT").String(),
	}
}

// Set 设置CTX用户信息
func (rec *ctxUser) Set(ctx context.Context, u CtxUserInfo) {
	r := g.RequestFromCtx(ctx)
	if r == nil {
		panic("get request from ctx failed")
	}
	r.SetCtxVar("TOKEN_UID", u.UID)
	r.SetCtxVar("TOKEN_UUID", u.UUID)
	//r.SetCtxVar("TOKEN_TENANT_ID", u.TenantID)
	r.SetCtxVar("TOKEN_SUBJECT", u.Subject)
	r.SetCtxVar("TOKEN_JTI", u.JTI)
	r.SetCtxVar("TOKEN_EXP", u.Exp)
	r.SetCtxVar("TOKEN_REG_IP", u.RegIP)
	r.SetCtxVar("TOKEN_REG_UA", u.RegUA)
	r.SetCtxVar("TOKEN_REG_DEVICE_ID", u.RegDeviceID)
	r.SetCtxVar("TOKEN_EXT", u.Ext)
}
