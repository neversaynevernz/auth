package restful

import (
	"net/http"

	_ "git.cloud.top/go/rest/i18n"

	"git.cloud.top/go/rest/rest"
	"git.cloud.top/srp-go/devops-api/pkg/obj"
)

type UserManager struct {
	rest.Resource
}

type StatusManager struct {
	rest.Resource
}

func (self *StatusManager) Put(c rest.Context) error {

	var resp obj.NewResponse
	resp.Success = true

	name := c.Param("name")

	type Status struct {
		Status string `json:"is_active"`
	}
	var status Status
	err := c.Bind(&status)
	if err != nil {
		return rest.Error("Bind failed")
	}

	user := map[string]string{"is_active": status.Status}
	err = SetUserStatus(name, user)

	if err != nil {
		return rest.Error("set user status err")
	}

	return c.JSON(http.StatusOK, resp)
}

func (self *UserManager) Get(c rest.Context) error {

	var resp obj.NewResponse
	resp.Success = true

	var ret []obj.User
	err := GetUser(&ret)
	if err != nil {
		return rest.Error("get user error")
	}
	return c.JSON(http.StatusOK, ret)
}

func (self *UserManager) Post(c rest.Context) error {

	var resp obj.NewResponse
	resp.Success = true

	var user obj.User

	err := c.Bind(&user)
	if err != nil {
		return rest.Error("Bind failed")
	}

	err = AddUser(user)

	if err != nil {
		return rest.Error("add user error")
	}

	// 用户赋予角色
	CE.AddRoleForUser(user.UserName, user.RoleName)

	return c.JSON(http.StatusOK, resp)
}

func (self *UserManager) Put(c rest.Context) error {

	var resp obj.NewResponse
	resp.Success = true

	var user obj.User

	err := c.Bind(&user)

	if err != nil {
		return rest.Error("Bind failed")
	}

	// 删除用户角色
	CE.DeleteRolesForUser(user.UserName)

	name := c.Param("name")
	err = SetUser(name, user)

	if err != nil {
		return rest.Error("set user error")
	}

	// 赋予用户角色
	CE.AddRoleForUser(user.UserName, user.RoleName)

	return c.JSON(http.StatusOK, resp)
}

func (self *UserManager) Delete(c rest.Context) error {

	var resp obj.NewResponse
	resp.Success = true

	name := c.Param("name")

	CE.DeleteRolesForUser(name)

	err := DeleteUser(name)

	if err != nil {
		return rest.Error("del user err")
	}

	return c.JSON(http.StatusOK, resp)
}
