package token

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
)

func IsRedisRevoked(jti string) (result bool, err error) {
	n, err := g.Redis().Exists(context.Background(), "jwt_block_"+jti)
	if err != nil {
		return
	}
	return n > 0, err
}
