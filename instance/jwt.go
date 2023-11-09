package instance

import (
	"context"

	"github.com/fainc/gfe/middleware"
)

var JwtUser = middleware.NewJwt(context.Background(), "default")
