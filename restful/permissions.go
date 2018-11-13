package restful

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"

	_ "git.cloud.top/go/rest/i18n"

	"git.cloud.top/go/rest/rest"
	"git.cloud.top/srp-go/devops-api/pkg/obj"
)

type UserPS struct {
	rest.Resource
}

type RolePS struct {
	rest.Resource
}

type AllPS struct {
	rest.Resource
}

func (self *UserPS) Get(c rest.Context) error {

	uname := c.Param("name")

	rname := GetRoleByUserName(uname)

	var cj obj.Role

	GetRoleByRoleName(rname, &cj)

	ppp := make(map[string]interface{})

	var plist []int

	for _, j := range cj.Ps {
		wxw, _ := strconv.Atoi(j[1])
		plist = append(plist, wxw)
	}

	ppp["client_id"] = cj.ClientID
	ppp["cs"] = GetGroupNameByID(cj.ClientID)
	ppp["ps"] = plist

	return c.JSON(http.StatusOK, ppp)
}

func (self *RolePS) Get(c rest.Context) error {

	uname := c.Param("name")

	var cj obj.Role
	GetRoleByRoleName(uname, &cj)

	ppp := make(map[string]interface{})

	var plist []int

	for _, j := range cj.Ps {
		wxw, _ := strconv.Atoi(j[1])
		plist = append(plist, wxw)
	}

	ppp["client_id"] = cj.ClientID
	ppp["cs"] = GetGroupNameByID(cj.ClientID)
	ppp["ps"] = plist

	return c.JSON(http.StatusOK, ppp)
}

func (self *AllPS) Get(c rest.Context) error {

	data, _ := ioutil.ReadFile(Conf)

	ps := []obj.Permission{}

	json.Unmarshal(data, &ps)

	return c.JSON(http.StatusOK, ps)
}
