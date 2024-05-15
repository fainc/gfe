package helper

import (
	"context"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

type runTime struct {
	t time.Time
}

func RunTime() *runTime {
	return &runTime{}
}

// GetRequestRunTime 获取请求进入框架后的实时运行时间
func (rec *runTime) GetRequestRunTime(ctx context.Context) int64 {
	r := g.RequestFromCtx(ctx)
	return gtime.Now().Sub(r.EnterTime).Milliseconds()
}

// Start 开始计算
func (rec *runTime) Start() {
	rec.t = time.Now()
}

// Since 获取开始计算时间至当前时间用时
func (rec *runTime) Since() time.Duration {
	return time.Since(rec.t)
}
