用户、角色、权限、WEB 访问控制的 API 以及基本操作

支持该包需在主文件初始化 mongodb、redis 连接

在对应文件里注册对应的路由即可

## Demo
```
package auth

import (
        "git.cloud.top/go/rest/rest"
        srpauth "git.cloud.top/srp-go/auth/restful"
)

func init() {
        rest.Register(nil, "/auth/token", &srpauth.Auth{})
        rest.Register(nil, "/auth/fresh_token", &srpauth.Fresh{})
        rest.Register(nil, "/auth/logout", &srpauth.Logout{})
}

```

## 初始化 mongodb
```
conf := config.GetSection("mongodb")
addr := fmt.Sprintf("mongodb://%s%s%s:%s", conf["username"], conf["password"], conf["ip"], conf["port"])
max, _ := strconv.Atoi(conf["max_pool_limit"])
gg, _ := strconv.Atoi(conf["socket_time"])
st := time.Duration(gg)
srpauth.Dial(addr, max, st)
srpauth.InitDB()
defer srpauth.CloseMongo()
```

## 初始化 redis(存储 token) 并启用插件 jwt 认证 之后再次验证 token 有效性

```
redisConf := config.GetSection("redis")
cc, err := redis.NewHTTPClient(redisConf["ip"] + ":" + redisConf["port"])
defer cc.RedisClient.Close()
if err != nil {
    fmt.Println(err.Error())
    return
}
App.Use(srpauth.Middleware())
```
## WEB 访问控制定期清除解除锁定的IP
```
err = srpauth.Check(authConf["conf"])
if err != nil {
        fmt.Println(err.Error())
        return
}
t, _ := strconv.Atoi(authConf["check_period"])
srpauth.Start(t)
```
## WEB 访问控制是对配置文件的操作实现，该文件参数如下
```
{
  "access_options": {
    "lock_mins": 50,
    "max_retries": 5,
    "timeout_mins": 1
  },
  "lock_ips": {}
}
```
分别对应锁定时长、最大尝试次数、token 有效时长，锁定的IP
