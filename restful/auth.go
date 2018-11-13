package restful

import (
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"

	"git.cloud.top/go/rest/auth"
	"git.cloud.top/go/rest/rest"

	"git.cloud.top/go/utility/redis"
)

type BaseAuthData struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Auth struct {
	rest.Resource
}

// auth.conf timeout_mins
// default 1 minute
func tokenExpertion() time.Duration {
	err, t := GetTokenExpertion()
	if err != nil {
		return 1 * time.Minute
	}
	return t
}

func (self *Auth) Options(c rest.Context) error {
	return nil
}

func (self *Auth) Post(c rest.Context) error {

	var data BaseAuthData
	err := c.Bind(&data)

	// store `username` into context
	c.Set("username", data.Username)

	if err != nil {
		return rest.Error("Bind failed")
	}

	// u := user.Fake()
	// if data.Username != u.Name || data.Password != u.Password {
	//         return rest.NewHTTPError(http.StatusUnauthorized)
	// }

	cip := c.RealIP()
	notLock := CheckStatus(cip)

	if !notLock {
		return rest.Error("the IP is locked")
	}

	// mongodb 认证
	err, pw, status := GetInfoByUserName(data.Username)

	if err != nil {
		LoginFailed(cip)
		return rest.Error("invalid username")
	}

	if data.Password != pw {
		LoginFailed(cip)
		return rest.Error("invalid password")
	}

	if status == "0" {
		LoginFailed(cip)
		return rest.Error("user is not active")
	}

	LoginSuccess(cip)

	c.Logger().Warnj(rest.Log{"message": "user login", "username": data.Username})

	claims := make(jwt.MapClaims)
	// ExpiresAt: time.Now().Add(time.Hour * 72).Unix(),
	now := time.Now()
	// iat: token创建时间
	claims["iat"] = now.Unix()
	// exp: token到期时间
	claims["exp"] = now.Add(tokenExpertion()).Unix()
	// nbf: token生效时间
	claims["nbf"] = now.Unix()
	// sip: token创建者(ip)
	claims["sip"] = c.Request().Host
	// cip: token使用者(ip)
	claims["cip"] = c.RealIP()
	// iss: token创建者(名称)
	// role: 角色权限管理(TODO)

	// sub: token使用者(用户名)
	claims["sub"] = data.Username

	// Create token with claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Generate encoded token and send it as response.
	t, err := token.SignedString(auth.GetSecretKey())
	if err != nil {
		return rest.Error("token signature error")
	}

	// 存储 token
	// key: value md5(token): token
	tid := redis.GetMd5String(t)
	redis.GetRedisClient().Set(tid, tid, tokenExpertion())

	return c.JSON(http.StatusOK, rest.Map{"token": t})
}

type Fresh struct {
	rest.Resource
}

func (self *Fresh) Post(c rest.Context) error {

	ut := c.Get("user")

	if ut == nil {
		return rest.NewHTTPError(http.StatusUnauthorized, auth.ErrJWTInvalid.Message)
	}

	token := c.Get("user").(*jwt.Token)
	claims := token.Claims.(jwt.MapClaims)

	// 只更新到期时间
	// claims["exp"] = time.Now().Add(time.Minute * 1).Unix()
	claims["exp"] = time.Now().Add(tokenExpertion()).Unix()

	// 重新生成token
	token = jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	t, err := token.SignedString(auth.GetSecretKey())

	if err != nil {
		return rest.Error("token signature error")
	}

	// 老token立即失效
	newkey := redis.GetMd5String(t)
	redis.GetRedisClient().Set(newkey, newkey, tokenExpertion())
	return c.JSON(http.StatusOK, rest.Map{"token": t})
}

type Logout struct {
	rest.Resource
}

func (self *Logout) Post(c rest.Context) error {
	// 使token立即失效
	token := c.Get("user").(*jwt.Token).Raw
	redis.GetRedisClient().Del(redis.GetMd5String(token))
	return c.JSON(http.StatusOK, rest.Map{"success": true})
}
