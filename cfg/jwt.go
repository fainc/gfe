package cfg

import (
	"context"
	"fmt"

	"github.com/gogf/gf/v2/frame/g"
)

type JwtCfg struct {
	Private string // 私钥加签
	Public  string // 公钥验签
	Subject string
	Redis   string // 关联配置redis组，不设置即不启用
	AuthUA  bool   // 是否校验客户端信息（UA相对稳定，可选）
	AuthIP  bool   // 是否强校验IP（严苛内部IP场景使用，面向客户端一般不启用）
}

// NewJwtCfgFromFrame 从框架读取JWT配置 可在 config.yaml 配置
func NewJwtCfgFromFrame(ctx context.Context, configName ...string) JwtCfg {
	cname := "default"
	if len(configName) != 0 && configName[0] != "" {
		cname = configName[0]
	}
	return JwtCfg{
		Public:  g.Cfg().MustGet(ctx, fmt.Sprintf("jwt.%v.public", cname)).String(),
		Private: g.Cfg().MustGet(ctx, fmt.Sprintf("jwt.%v.private", cname)).String(),
		Subject: g.Cfg().MustGet(ctx, fmt.Sprintf("jwt.%v.subject", cname)).String(),
		Redis:   g.Cfg().MustGet(ctx, fmt.Sprintf("jwt.%v.redis", cname)).String(),
		AuthUA:  g.Cfg().MustGet(ctx, fmt.Sprintf("jwt.%v.authUA", cname)).Bool(),
		AuthIP:  g.Cfg().MustGet(ctx, fmt.Sprintf("jwt.%v.authIP", cname)).Bool(),
	}
}
