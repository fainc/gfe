package response

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
)

type writer struct {
	Mime string
}

func Writer(mime ...string) *writer {
	if len(mime) != 1 {
		return &writer{
			Mime: MimeJSON,
		}
	}
	return &writer{mime[0]}
}

// Output 数据输出
func (rec *writer) Output(ctx context.Context, data interface{}, status ...int) {
	statusCode := 200
	if len(status) >= 1 && status[0] != 200 {
		statusCode = status[0]
	}
	r := g.RequestFromCtx(ctx)
	r.Response.WriteStatus(statusCode)
	r.Response.ClearBuffer()
	switch rec.Mime {
	case MimeJSON:
		r.Response.WriteJson(data)
	case MimeXML:
		r.Response.WriteXml(data, "xml")
	case MimeHTML:
		r.Response.Header().Set("Content-Type", "text/html")
		r.Response.Write(data)
	default:
		r.Response.WriteJson(data)
	}
	setServerHeader(r)
	r.ExitAll()
}
