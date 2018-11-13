package restful

import (
	"net/http"

	_ "git.cloud.top/go/rest/i18n"

	"git.cloud.top/go/rest/rest"
	"git.cloud.top/srp-go/devops-api/pkg/obj"
)

type RoleManager struct {
	rest.Resource
}

func (self *RoleManager) Get(c rest.Context) error {

	var resp obj.NewResponse
	resp.Success = true

	var ret []obj.Role
	err := GetRole(&ret)
	if err != nil {
		return rest.Error("get role error")
	}
	return c.JSON(http.StatusOK, ret)
}

func (self *RoleManager) Post(c rest.Context) error {

	var resp obj.NewResponse
	resp.Success = true

	var role obj.Role
	err := c.Bind(&role)

	if err != nil {
		return rest.Error("Bind failed")
	}

	err = AddRole(role)
	if err != nil {
		return rest.Error("add role error")
	}

	e := CE

	// 赋予角色权限
	permissions := role.Ps
	for _, i := range permissions {
		e.AddPermissionForUser(role.Name, i...)
	}

	return c.JSON(http.StatusOK, resp)
}

func (self *RoleManager) Put(c rest.Context) error {

	var resp obj.NewResponse
	resp.Success = true

	var role obj.Role

	name := c.Param("name")

	err := c.Bind(&role)

	if err != nil {
		return rest.Error("Bind failed")
	}

	e := CE

	e.DeletePermissionsForUser(role.Name)

	err = SetRole(name, role)

	if err != nil {
		return rest.Error("set role error")
	}

	// 赋予角色权限
	permissions := role.Ps
	for _, i := range permissions {
		e.AddPermissionForUser(role.Name, i...)
	}

	return c.JSON(http.StatusOK, resp)
}

func (self *RoleManager) Delete(c rest.Context) error {

	var resp obj.NewResponse
	resp.Success = true

	name := c.Param("name")

	e := CE

	// 删除角色对应权限
	e.DeletePermissionsForUser(name)

	err := DeleteRole(name)

	if err != nil {
		return rest.Error("del role error")
	}

	return c.JSON(http.StatusOK, resp)
}
