package restful

import (
	"encoding/json"
	"io/ioutil"
	"strconv"
	"time"

	"github.com/casbin/casbin"

	"git.cloud.top/go/mongodb-adapter"
	"git.cloud.top/go/utility/mongo"
	"git.cloud.top/srp-go/devops-api/pkg/obj"
)

var (
	Addr string
	CE   *casbin.Enforcer
	Conf string
)

func GetDB() *mongo.Mongo {
	return mongo.M()
}

func CloseMongo() {
	mongo.Close()
}

func Dial(addr string, maxPoolLimit int, timeout time.Duration) error {
	Addr = addr
	return mongo.Dial(addr, maxPoolLimit, timeout)
}

func InitDB(conf, modelconf string) {
	m := GetDB()
	var db_map map[string][]string = map[string][]string{
		"user_manage": []string{"user", "role", "casbin_rule", "permissions"},
	}
	for k, v := range db_map {
		for _, i := range v {
			m.SwitchTo(k, i)
			m.CreateTable()
			m.CreateIndex("id")
		}
	}

	a := mongodbadapter.NewAdapter(Addr + "/user_manage")
	CE = casbin.NewEnforcer(modelconf, a)

	//初始化模块权限
	InitPermissions(conf)

	//初始化用户和角色
	PresetUserAndRole()
}

func GetMaxID(dbname, cname string) (error, string) {
	db := GetDB()
	defer db.Close()
	db.SwitchTo(dbname, cname)

	type IDS struct {
		ID string `bson:"_id"`
	}
	var ids []IDS
	err := db.C.Find(nil).All(&ids)
	if err != nil {
		return err, "db err"
	}
	if ids == nil {
		return nil, "1"
	}
	j, err := strconv.Atoi(ids[0].ID)
	for _, i := range ids {
		a, _ := strconv.Atoi(i.ID)
		if a > j {
			j = a
		}
	}
	if err != nil {
		return err, "db err"
	}
	return nil, strconv.Itoa(j + 1)
}

func InitPermissions(conf string) error {

	Conf = conf
	data, err := ioutil.ReadFile(conf)

	if err != nil {
		return err
	}

	ps := []obj.Permission{}
	err = json.Unmarshal(data, &ps)
	if err != nil {
		return err
	}

	db := GetDB()
	defer db.Close()

	db.SwitchTo("user_manage", "permissions")

	for _, p := range ps {
		_, err := db.C.UpsertId(p.ID, p)
		if err != nil {
			return err
		}
	}
	return nil
}

func PresetUserAndRole() error {

	db := GetDB()

	defer db.Close()

	db.SwitchTo("user_manage", "role")

	num, err := db.Count()

	if num != 0 {
		return nil
	}

	err, nid := GetMaxID("user_manage", "role")

	if err != nil {
		return err
	}

	role := obj.Role{
		"role_admin",
		"",
		"preset",
		"All",
		[][]string{
			[]string{"11", "111"},
			[]string{"12", "121"},
			[]string{"3", "31"},
			[]string{"41", "411"},
			[]string{"42", "421"},
			[]string{"43", "431"},
			[]string{"511", "5111"},
			[]string{"512", "5121"},
			[]string{"53", "531"},
			[]string{"54", "541"},
			[]string{"6", "61"},
			[]string{"7", "71"},
			[]string{"81", "811"},
			[]string{"82", "821"},
			[]string{"91", "911"},
			[]string{"92", "921"},
			[]string{"101", "1011"},
			[]string{"102", "1021"},
			[]string{"103", "1031"},
		},
	}

	_, err = db.C.UpsertId(nid, role)

	if err != nil {
		return err
	}

	db.SwitchTo("user_manage", "user")

	err, nid = GetMaxID("user_manage", "user")
	if err != nil {
		return err
	}

	user := obj.User{
		"superman",
		"88627D1FE4D5EF9E8B341F0DBF0370B5",
		"",
		"",
		"1",
		"role_admin",
	}

	_, err = db.C.UpsertId(nid, user)

	if err != nil {
		return err
	}

	// 赋予角色权限
	permissions := role.Ps
	for _, i := range permissions {
		CE.AddPermissionForUser(role.Name, i...)
	}

	return nil
}
