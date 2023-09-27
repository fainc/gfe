package middleware

import (
	"github.com/gogf/gf/v2/net/ghttp"

	"github.com/fainc/gfe/response"
	"github.com/fainc/gfe/token"
)

func DemoAuth(r *ghttp.Request) {
	_, _, catch, err := token.Helper().Auth(r, token.AuthParams{
		RevokeAuth: true,
	})
	if err != nil && catch {
		response.Json().UnAuthorizedError(r.Context(), err.Error(), nil)
		return
	}
	r.Middleware.Next()
}
