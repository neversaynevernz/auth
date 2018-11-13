package restful

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	_ "git.cloud.top/go/rest/i18n"

	"git.cloud.top/go/rest/rest"
	"git.cloud.top/srp-go/devops-api/pkg/obj"
)

type WebAccess struct {
	rest.Resource
}

type LockObjs struct {
	rest.Resource
}

type AccessConf struct {
	MaxRetries  int `json:"max_retries"`
	LockMins    int `json:"lock_mins"`
	TimeoutMins int `json:"timeout_mins"`
}

func (self *LockObjs) Get(c rest.Context) error {

	var resp obj.NewResponse
	resp.Success = true

	err, objs, _ := GetlockObjs()

	if err != nil {
		return rest.Error("parse conf error")
	}

	var dealret []map[string]string

	for k, v := range objs {

		one := make(map[string]string)

		one["ip"] = strings.Split(k, "ip-")[1]

		ltime, _ := strconv.ParseInt(v.(map[string]interface{})["lock_time"].(string), 10, 64)
		lutime := time.Unix(ltime, 0)
		one["lock_time"] = lutime.Format("MST 2006-01-02 15:04:05")

		untime, _ := strconv.ParseInt(v.(map[string]interface{})["unlock_time"].(string), 10, 64)
		unutime := time.Unix(untime, 0)
		one["unlock_time"] = unutime.Format("MST 2006-01-02 15:04:05")

		dealret = append(dealret, one)
	}

	return c.JSON(http.StatusOK, dealret)
}

func (self *LockObjs) Put(c rest.Context) error {

	var resp obj.NewResponse
	resp.Success = true

	ip := c.Param("ip")

	err := UnlockObj(ip)

	if err != nil {
		return rest.Error("set conf error")
	}

	return c.JSON(http.StatusOK, resp)
}

func (self *WebAccess) Get(c rest.Context) error {

	var resp obj.NewResponse
	resp.Success = true

	err, attrs, _ := GetAttrs()

	if err != nil {
		return rest.Error("parse conf error")
	}

	return c.JSON(http.StatusOK, attrs)
}

func (self *WebAccess) Put(c rest.Context) error {

	var resp obj.NewResponse
	resp.Success = true

	var ret AccessConf
	err := c.Bind(&ret)

	if err != nil {
		return rest.Error("Bind falied")
	}

	err, _, data := GetAttrs()

	attrs := map[string]interface{}{
		"lock_mins":    ret.LockMins,
		"max_retries":  ret.MaxRetries,
		"timeout_mins": ret.TimeoutMins,
	}

	err = WriteAttrs(attrs, data)

	if err != nil {
		return rest.Error("set conf error")
	}

	return c.JSON(http.StatusOK, resp)
}
