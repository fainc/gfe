package response

import (
	"github.com/gogf/gf/v2/frame/g"
)

// CustomRes 自定义返回数据标注，使用该类型数据返回时，全局后置中间件ResponseHandler将不再处理返回数据，请自行提前输出
type CustomRes struct {
	g.Meta `mime:"custom" sm:"自定义数据返回" dc:"本接口使用自定义数据返回，非OPEN API v3规范，具体返回数据字段请联系管理员获取"`
}
