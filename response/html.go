package response

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/gogf/gf/v2/util/gconv"
)

const (
	CtxHTMLRenderTplKey = "html_render_tpl"
)

// SetHTMLRenderTpl 通过上下文设置业务成功的模板文件，优先级高于meta x-static-tpl
func SetHTMLRenderTpl(ctx context.Context, tpl string) {
	g.RequestFromCtx(ctx).SetCtxVar(CtxHTMLRenderTplKey, tpl)
}

func assignsFrameParams(r *ghttp.Request) {
	r.Assigns(g.Map{"Frame": g.Map{"Request": r.Request, "TraceID": gctx.CtxId(r.Context())}})
}
func (rec *format) htmlRender(ctx context.Context, data interface{}) (tpl string, err error) {
	r := g.RequestFromCtx(ctx)
	tplPath := rec.errorTpl // 1.优先读取错误视图
	if tplPath == "" {
		ctxPath := r.GetCtxVar(CtxHTMLRenderTplKey).String() // 2.读取ctx动态视图
		if ctxPath != "" {
			tplPath = ctxPath
		}
	}
	if tplPath == "" {
		tplPath = rec.staticTpl // 3.读取静态视图
	}
	assignsFrameParams(r) // 框架公共信息
	tpl, err = r.Response.ParseTpl(tplPath, gconv.Map(data))
	if err != nil {
		r.SetError(CodeError(500, err.Error(), nil))
	}
	return
}
