package worker

import (
	"encoding/json"
	"io/ioutil"
)

//程序配置
type Config struct {
	EtcdEndPoints     []string `json:"etcdEndPoints"`
	EtcdDialTimeout   int      `json:"etcdDialTimeout"`
	EtcdLeaseTimeout  int      `json:"etcdLeaseTimeout"`
	ScheduleSleepTime int      `json:"scheduleSleepTime"`
	ShellExcuseBash   string   `json:"shellExcuseBash"`
	ShellExcuseArg    string   `json:"shellExcuseArg"`
}

var (
	ConfSingle *Config
)

func NewConfig(fileName string) (err error) {
	var (
		conf    *Config
		content []byte
	)
	//1.读入配置文件
	if content, err = ioutil.ReadFile(fileName); err != nil {
		return
	}

	//2.反序列化
	if err = json.Unmarshal(content, &conf); err != nil {
		return
	}
	//3.赋值单例
	ConfSingle = conf
	return
}
