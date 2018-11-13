package restful

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"reflect"
	"time"

	"github.com/robfig/cron"
)

type DNF struct {
	RWer *os.File
}

func Loads(b interface{}, r []byte) error {
	err := json.Unmarshal(r, &b)
	if err != nil {
		return err
	}
	return nil
}

func Load(b interface{}, r io.Reader) error {
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}
	return Loads(b, data)
}

func Dumps(b interface{}) ([]byte, error) {
	if s, err := json.MarshalIndent(b, "", "  "); err == nil {
		return s, nil
	} else {
		return nil, err
	}
}

func Dump(b interface{}, w io.Writer) error {
	if s, err := Dumps(b); err == nil {
		w.Write(s)
		return nil
	} else {
		return err
	}
}

func (so *DNF) CreateLockName(ip string) string {
	return "ip-" + ip
}

func (so *DNF) LoginSuccess(ip string) {
	_, objs, data := so.GetLockObjs()
	lock_ip := so.CreateLockName(ip)
	if objs[lock_ip] != nil {
		delete(objs, lock_ip)
		so.WriteLockObjs(objs, data)
	}
}

func (so *DNF) LoginFailed(ip string) {

	var lock_obj map[string]interface{}

	var max_retry_config int
	var lock_min_config int
	var lock_obj_count int

	lock_ip := so.CreateLockName(ip)

	err, objs, _ := so.GetLockObjs()

	if err != nil {
		panic(err)
	}

	v := reflect.ValueOf(objs[lock_ip])

	switch v.Kind() {
	case reflect.Map:
		lock_obj = objs[lock_ip].(map[string]interface{})
	case reflect.Invalid:
		lock_obj = so.CreateObj(ip)[lock_ip].(map[string]interface{})
	}

	cj := reflect.ValueOf(lock_obj["login_failed_count"])

	switch cj.Kind() {

	case reflect.Float64:
		lock_obj_count = int(lock_obj["login_failed_count"].(float64))

	case reflect.Int:
		lock_obj_count = lock_obj["login_failed_count"].(int)
	}

	lock_obj["login_failed_count"] = lock_obj_count + 1

	_, objs, data := so.GetLockObjs()

	options := data["access_options"]
	max, ok := options["max_retries"]

	if ok {
		max_retry_config = int(max.(float64))
	} else {
		max_retry_config = 5
	}

	lock_min, ok := options["lock_mins"]

	if ok {
		lock_min_config = int(lock_min.(float64))
	} else {
		lock_min_config = 1
	}

	if lock_obj_count >= max_retry_config-1 {

		lock_obj["enable_login"] = false
		lock_time := int(time.Now().Unix())
		lock_seconds := lock_min_config * 60
		unlock_time := lock_time + lock_seconds

		lock_obj["lock_time"] = fmt.Sprintf("%d", lock_time)
		lock_obj["unlock_time"] = fmt.Sprintf("%d", unlock_time)

		// TODO 一段时间后解锁
		// 改为周期检查
	}

	objs[lock_ip] = lock_obj
	so.WriteLockObjs(objs, data)
}

func (so *DNF) CreateObj(ip string) (attrs map[string]interface{}) {
	attrs = map[string]interface{}{
		so.CreateLockName(ip): map[string]interface{}{
			"lock_time":          "",
			"unlock_time":        "",
			"enable_login":       true,
			"login_failed_count": 0,
		}}
	return
}

func (so *DNF) CheckObj(ip string) bool {

	err, objs, _ := so.GetLockObjs()

	if err != nil {
		return false
	}

	lock_ip := so.CreateLockName(ip)

	if objs != nil {
		for k, v := range objs {
			if k == lock_ip && v.(map[string]interface{})["enable_login"] == false {
				return false
			}
		}
	}
	return true
}

func (so *DNF) UnlockObj(ip string) (err error) {

	err, objs, data := so.GetLockObjs()

	if err != nil {
		return
	}

	delete(objs, so.CreateLockName(ip))

	so.WriteLockObjs(objs, data)

	return
}

func (so *DNF) GetLockObjs() (err error, objs map[string]interface{}, data map[string]map[string]interface{}) {

	defer so.RWer.Seek(0, 0)

	err = Load(&data, so.RWer)

	objs = data["lock_ips"]

	if objs == nil {
		objs = map[string]interface{}{"lock_ips": nil}
	}

	return
}

func (so *DNF) WriteLockObjs(objs map[string]interface{}, data map[string]map[string]interface{}) (err error) {

	// 清空文件内容回到文件开始
	so.RWer.Truncate(0)
	so.RWer.Seek(0, 0)

	data["lock_ips"] = objs
	err = Dump(data, so.RWer)
	return
}

func (so *DNF) GetAttrs() (err error, attrs map[string]interface{}, data map[string]map[string]interface{}) {
	defer so.RWer.Seek(0, 0)
	err = Load(&data, so.RWer)
	attrs = data["access_options"]
	return
}

func (so *DNF) SetAttrs(attrs map[string]interface{}, data map[string]map[string]interface{}) (err error) {
	so.RWer.Truncate(0)
	so.RWer.Seek(0, 0)
	data["access_options"] = attrs
	err = Dump(data, so.RWer)
	return
}

func LoginFailed(ip string) {
	e, _ := os.OpenFile(filepath, os.O_RDWR|os.O_CREATE, 0666)
	dd := DNF{RWer: e}
	defer e.Close()
	dd.LoginFailed(ip)
}

func LoginSuccess(ip string) {
	e, _ := os.OpenFile(filepath, os.O_RDWR|os.O_CREATE, 0666)
	dd := DNF{RWer: e}
	defer e.Close()
	dd.LoginSuccess(ip)
}

func CheckStatus(ip string) bool {
	e, _ := os.OpenFile(filepath, os.O_RDONLY|os.O_CREATE, 0666)
	defer e.Close()
	dd := DNF{RWer: e}
	return dd.CheckObj(ip)
}

func periodCheck() {
	_, objs, data := read()
	str_now := fmt.Sprintf("%d", int(time.Now().Unix()))
	newobjs := make(map[string]interface{})
	for k, v := range objs {
		tt := v.(map[string]interface{})["unlock_time"].(string)
		isban := v.(map[string]interface{})["enable_login"] == false
		if tt == "" {
			newobjs[k] = v
			continue
		}
		if str_now < tt && isban {
			newobjs[k] = v
		}
	}
	write(newobjs, data)
}

func read() (err error, objs map[string]interface{}, data map[string]map[string]interface{}) {
	e, _ := os.OpenFile(filepath, os.O_RDONLY|os.O_CREATE, 0666)
	defer e.Close()
	dd := DNF{RWer: e}
	err, objs, data = dd.GetLockObjs()
	return
}

func write(objs map[string]interface{}, data map[string]map[string]interface{}) {
	e, _ := os.OpenFile(filepath, os.O_WRONLY|os.O_CREATE, 0666)
	defer e.Close()
	dd := DNF{RWer: e}
	err := dd.WriteLockObjs(objs, data)
	if err != nil {
		panic(err)
	}
	return
}

func GetlockObjs() (err error, objs map[string]interface{}, data map[string]map[string]interface{}) {
	e, err := os.OpenFile(filepath, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return
	}
	defer e.Close()
	dd := DNF{RWer: e}
	err, objs, data = dd.GetLockObjs()
	return
}

func UnlockObj(ip string) error {
	e, err := os.OpenFile(filepath, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	dd := DNF{RWer: e}
	defer e.Close()
	err = dd.UnlockObj(ip)
	return err
}

func GetAttrs() (err error, attrs map[string]interface{}, data map[string]map[string]interface{}) {
	e, err := os.OpenFile(filepath, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return
	}
	defer e.Close()
	dd := DNF{RWer: e}
	err, attrs, data = dd.GetAttrs()
	return
}

func WriteAttrs(attrs map[string]interface{}, data map[string]map[string]interface{}) error {
	e, err := os.OpenFile(filepath, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer e.Close()
	dd := DNF{RWer: e}
	err = dd.SetAttrs(attrs, data)
	return err
}

func GetTokenExpertion() (err error, t time.Duration) {
	e, err := os.OpenFile(filepath, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return
	}
	defer e.Close()
	dd := DNF{RWer: e}
	err, attrs, _ := dd.GetAttrs()
	cj := int(attrs["timeout_mins"].(float64))
	t = time.Duration(cj) * time.Minute
	return
}

var filepath string

func pathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func Check(fname string) error {
	filepath = fname
	exist, _ := pathExists(filepath)
	if !exist {
		return errors.New(fmt.Sprintf("no such auth conf: %s ", filepath))
	}
	return nil
}

func Start(t int) {
	c := cron.New()
	// remove ip when time > unlocktime per t s
	spec := fmt.Sprintf("*/%d * * * * ?", t)
	c.AddFunc(spec, periodCheck)
	c.Start()
}
