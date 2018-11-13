package restful

import (
	"net/http"

	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"

	"git.cloud.top/go/rest/rest"
	"git.cloud.top/go/utility/redis"
)

type Config struct {
	Skipper middleware.Skipper
}

func ReSkipper(c rest.Context) bool {
	if c.Path() == "/auth/token" {
		return true
	}
	return false
}

var (
	DefaultConfig = Config{
		Skipper: ReSkipper,
	}
)

func Middleware() echo.MiddlewareFunc {
	return MiddlewareWithConfig(DefaultConfig)
}

func MiddlewareWithConfig(config Config) echo.MiddlewareFunc {
	if config.Skipper == nil {
		config.Skipper = DefaultConfig.Skipper
	}
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if config.Skipper(c) {
				return next(c)
			}
			// 查 token 是否存在 redis
			tokenmd5 := c.Get("user").(*jwt.Token).Raw
			if err, ok := redis.GetRedisClient().IsExists(redis.GetMd5String(tokenmd5)); !ok {
				return &echo.HTTPError{
					Code:     http.StatusUnauthorized,
					Message:  "invalid or expired jwt",
					Internal: err,
				}
			}
			return next(c)
		}
	}
}
