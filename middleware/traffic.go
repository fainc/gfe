package middleware

import (
	"context"
	"fmt"
	"time"

	"github.com/gogf/gf/v2/container/garray"
	"github.com/gogf/gf/v2/database/gredis"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gtime"

	"github.com/fainc/gfe/response"
)

type traffic struct {
	GlobalIPQPS       int           // IP级QPS 每秒漏桶上限 单独接口通过 req meta 定义 x-ip-qps  0为不限制  超出支持惩罚
	GlobalIPQPM       int           // IP级QPM 每分钟漏桶上限 单独接口通过 req meta 定义 x-ip-qpm  0为不限制
	GlobalIPQPH       int           // IP级QPH 每小时漏桶上限 单独接口通过 req meta 定义 x-ip-qph  0为不限制
	GlobalIPQTD       int           // IP级QTD 当日合计上限 单独接口通过 req meta 定义 x-ip-qtd  0为不限制 超出后当日自动锁定 次日重置 适用接口日配额场景
	GlobalIPQTM       int           // IP级QTM 当月合计上限 单独接口通过 req meta 定义 x-ip-qtm  0为不限制 超出后当月自动锁定 次月重置 适用接口月配额场景
	GlobalIPQPSPunish int           // IP QPS超限 惩罚沉默秒数时长（QPS粒度最细、周期最短、危害度最大，QPM、QPH和QT*均属长周期，长周期内自带锁定，无需额外沉默惩罚）
	GlobalIPExclude   *garray.Array // IP豁免白名单
}

var rds *gredis.Redis

func NewTraffic() *traffic {
	return &traffic{ // 默认配置
		GlobalIPQPS:       5,
		GlobalIPQPM:       10,
		GlobalIPQPH:       100, // 每个 IP 地址的每分钟请求数 (QPS) 为 100 次。
		GlobalIPQPSPunish: 10,  // IP超过任意QP周期阈值惩罚沉默60秒
		GlobalIPQTD:       1000,
		GlobalIPQTM:       1000,
	}
}
func (rec *traffic) SetRedisAdapter(r *gredis.Redis) *traffic {
	if rds != nil {
		panic("traffic: don't set redis adapter repeatedly")
	}
	if r == nil {
		panic("traffic: rds is nil")
	}
	_, err := r.Get(context.Background(), "test_ping") // 测试是否连通
	if err != nil {
		panic("traffic: test ping error :" + err.Error())
	}
	rds = r
	return rec
}

func (rec *traffic) getRdsClient() (r *gredis.Redis) {
	if rds == nil {
		panic("redis is not initialize")
	}
	return rds
}

// RateLimit 接口速率限制
func (rec *traffic) RateLimit(r *ghttp.Request) {
	ctx := r.Context()
	IP := r.GetClientIp()
	if rec.notExclude(IP, rec.GlobalIPExclude) {
		var ipQTM, ipQTD int
		if rec.GlobalIPQTM != 0 {
			ipQTM = rec.getIntValue(ctx, fmt.Sprintf("gb_ip_qtm_%v", IP))
			if ipQTM >= rec.GlobalIPQTM {
				r.SetError(response.CodeError(429, "IP punished, try again next month", nil))
				return
			}
		}
		if rec.GlobalIPQTD != 0 {
			ipQTD = rec.getIntValue(ctx, fmt.Sprintf("gb_ip_qtd_%v", IP))
			if ipQTD >= rec.GlobalIPQTD {
				r.SetError(response.CodeError(429, "IP punished, try again next day", nil))
				return
			}
		}
		punish := rec.getPTTL(ctx, fmt.Sprintf("gb_ip_punish_%v", IP))
		if punish > 0 || punish == -1 { // 惩罚剩余PTTL -1 永不过期
			r.SetError(response.CodeError(429, fmt.Sprintf("IP punished, try again in %vs", (time.Duration(punish)*time.Millisecond).Seconds()), nil))
			return
		}
		// QPH 小时级
		if rec.GlobalIPQPH != 0 {
			err := rec.ipLevel(ctx, IP, "qph", rec.GlobalIPQPH, time.Hour, time.Duration(rec.GlobalIPQPSPunish)*time.Second)
			if err != nil {
				r.SetError(response.CodeError(429, err.Error(), nil))
				return
			}
		}
		// QPM 分钟级
		if rec.GlobalIPQPM != 0 {
			err := rec.ipLevel(ctx, IP, "qpm", rec.GlobalIPQPM, time.Minute, time.Duration(rec.GlobalIPQPSPunish)*time.Second)
			if err != nil {
				r.SetError(response.CodeError(429, err.Error(), nil))
				return
			}
		}
		// QPS 秒级
		if rec.GlobalIPQPS != 0 {
			err := rec.ipLevel(ctx, IP, "qps", rec.GlobalIPQPS, time.Second, time.Duration(rec.GlobalIPQPSPunish)*time.Second)
			if err != nil {
				r.SetError(response.CodeError(429, err.Error(), nil))
				return
			}
		}

		// 最后更新，上文被拦截的请求不计入
		if rec.GlobalIPQTM != 0 {
			rec.setOrUpdateWithExpired(ctx, fmt.Sprintf("gb_ip_qtm_%v", IP), ipQTM+1, gtime.Now().EndOfMonth().TimestampMilli())
		}
		if rec.GlobalIPQTD != 0 {
			rec.setOrUpdateWithExpired(ctx, fmt.Sprintf("gb_ip_qtd_%v", IP), ipQTD+1, gtime.Now().EndOfDay().TimestampMilli())
		}
	}
	r.Middleware.Next()
}
func (rec *traffic) notExclude(key string, target *garray.Array) bool {
	if target != nil && target.Len() != 0 && target.Contains(key) {
		return false
	}
	return true
}
func (rec *traffic) ipLevel(ctx context.Context, ip string, level string, max int, lifecycle, punishDuration time.Duration) (err error) {
	cur := rec.getIntValue(ctx, fmt.Sprintf("gb_ip_%v_%v", level, ip))
	if cur >= max {
		if punishDuration != 0 && level == "qps" { // QPS 超出惩罚沉默
			rec.setOrUpdateWithPTTL(ctx, fmt.Sprintf("gb_ip_punish_%v", ip), cur+1, punishDuration)
		}
		err = fmt.Errorf("IP %v restricted", level)
		return
	}
	rec.setOrUpdateWithPTTL(ctx, fmt.Sprintf("gb_ip_%v_%v", level, ip), cur+1, lifecycle)
	return
}

func (rec *traffic) getIntValue(ctx context.Context, key string) int {
	v, err := rec.getRdsClient().Get(ctx, key)
	if err != nil {
		panic(err.Error())
	}
	return v.Int()
}

func (rec *traffic) getPTTL(ctx context.Context, key string) int64 {
	v, err := rec.getRdsClient().PTTL(ctx, key)
	if err != nil {
		panic(err.Error())
	}
	return v
}

func (rec *traffic) setOrUpdateWithPTTL(ctx context.Context, key string, value interface{}, setDuration time.Duration) {
	pttl, err := rec.getRdsClient().PTTL(ctx, key)
	if err != nil {
		panic(err.Error())
	}
	if pttl > 0 || pttl == -1 { // 延续有效期并更新值，KEEPTTL 6.0以上才支持暂不使用
		_, err = rec.getRdsClient().Do(ctx, "SET", key, value, "KEEPTTL")
		if err != nil {
			panic(err.Error())
		}
	}
	if pttl == 0 || pttl == -2 { // 过期或不存在，直接SET并覆盖新TTL
		_, err = rec.getRdsClient().Do(ctx, "SET", key, value, "PX", setDuration.Milliseconds()) // 使用毫秒级有效期
		if err != nil {
			panic(err.Error())
		}
	}
}

func (rec *traffic) setOrUpdateWithExpired(ctx context.Context, key string, value interface{}, expired int64) {
	pttl, err := rec.getRdsClient().PTTL(ctx, key)
	if err != nil {
		panic(err.Error())
	}
	if pttl > 0 || pttl == -1 { // 延续有效期并更新值，KEEPTTL 6.0以上才支持暂不使用
		_, err = rec.getRdsClient().Do(ctx, "SET", key, value, "KEEPTTL")
		if err != nil {
			panic(err.Error())
		}
	}
	if pttl == 0 || pttl == -2 { // 过期或不存在，直接SET并覆盖新TTL
		_, err = rec.getRdsClient().Do(ctx, "SET", key, value, "PXAT", expired) // 使用毫秒级有效期
		if err != nil {
			panic(err.Error())
		}
	}
}
