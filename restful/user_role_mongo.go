package restful

import (
	"gopkg.in/mgo.v2/bson"

	"git.cloud.top/srp-go/devops-api/pkg/obj"
)

func GetRoleByUserName(name string) string {

	db := GetDB()
	defer db.Close()

	db.SwitchTo("user_manage", "user")

	type role struct {
		Name string `bson:"role_name"`
	}

	var result role
	err := db.C.Find(bson.M{"username": name}).One(&result)
	if err != nil {
		return ""
	}
	return result.Name
}

func GetRoleByRoleName(name string, role interface{}) error {

	db := GetDB()

	defer db.Close()

	db.SwitchTo("user_manage", "role")

	err := db.C.Find(bson.M{"role_name": name}).One(role)

	return err
}

func GetPermissionByID(id string) map[string]string {

	db := GetDB()
	defer db.Close()

	db.SwitchTo("user_manage", "permissions")

	var r obj.Permission

	err := db.C.Find(bson.M{"_id": id}).One(&r)

	if err != nil {
		return nil
	}

	p := make(map[string]string)
	p["permission_eng"] = r.PsEN
	p["permission_chn"] = r.PsCN
	return p
}

func AddUser(user obj.User) error {

	// 前端加密后传
	// w := md5.New()
	// w.Write([]byte(user.PassWord))
	// w.Write([]byte(user.Salt))
	// md := strings.ToUpper(fmt.Sprintf("%x", w.Sum(nil)))
	// user.PassWord = md

	db := GetDB()
	defer db.Close()

	db.SwitchTo("user_manage", "user")

	err, nid := GetMaxID("user_manage", "user")
	if err != nil {
		return err
	}
	_, err = db.C.UpsertId(nid, user)
	if err != nil {
		return err
	}
	return nil
}

func GetUser(ret interface{}) error {

	db := GetDB()
	defer db.Close()
	db.SwitchTo("user_manage", "user")

	err := db.C.Find(nil).All(ret)

	return err
}

func SetUserStatus(name string, item interface{}) error {
	db := GetDB()
	defer db.Close()
	db.SwitchTo("user_manage", "user")
	err := db.C.Update(bson.M{"username": name}, bson.M{"$set": item})
	return err
}

func SetUser(name string, item interface{}) error {
	db := GetDB()
	defer db.Close()
	db.SwitchTo("user_manage", "user")
	err := db.C.Update(bson.M{"username": name}, bson.M{"$set": item})
	return err
}

func DeleteUser(name string) error {
	db := GetDB()
	defer db.Close()
	db.SwitchTo("user_manage", "user")
	err := db.C.Remove(bson.M{"username": name})
	return err
}

func GetRole(ret interface{}) error {

	db := GetDB()
	defer db.Close()
	db.SwitchTo("user_manage", "role")

	err := db.C.Find(nil).All(ret)

	return err
}

func SetRole(name string, item interface{}) error {
	db := GetDB()
	defer db.Close()
	db.SwitchTo("user_manage", "role")
	err := db.C.Update(bson.M{"role_name": name}, bson.M{"$set": item})
	return err
}

func DeleteRole(name string) error {
	db := GetDB()
	defer db.Close()
	db.SwitchTo("user_manage", "role")
	err := db.C.Remove(bson.M{"role_name": name})
	return err
}

func AddRole(role obj.Role) error {
	db := GetDB()
	defer db.Close()
	db.SwitchTo("user_manage", "role")
	err, nid := GetMaxID("user_manage", "role")
	if err != nil {
		return err
	}
	_, err = db.C.UpsertId(nid, role)
	if err != nil {
		return err
	}
	return nil
}

func GetGroupNameByID(id string) string {

	db := GetDB()

	defer db.Close()

	db.SwitchTo("devices", "groups")

	type name struct {
		GroupName string `bson:"name"`
	}

	var ss name

	err := db.C.Find(bson.M{"_id": id}).One(&ss)
	if err != nil {
		return ""
	}
	return ss.GroupName
}

func GetInfoByUserName(username string) (error, string, string) {

	db := GetDB()
	defer db.Close()

	db.SwitchTo("user_manage", "user")

	type role struct {
		Password string `bson:"password"`
		IsActive string `bson:"is_active"`
	}
	var result role
	err := db.C.Find(bson.M{"username": username}).One(&result)
	if err != nil {
		return err, "", ""
	}
	return nil, result.Password, result.IsActive
}
