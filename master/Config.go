package master

import (
	"encoding/json"
	"io/ioutil"
)

//程序配置
type Config struct {
	ApiPort         int      `json:"apiPort"`
	ApiReadTimeout  int      `json:"apiReadTimeout"`
	ApiWriteTimeout int      `json:"apiWriteTimeout"`
	EtcdEndPoints   []string `json:"etcdEndPoints"`
	EtcdDialTimeout int      `json:"etcdDialTimeout"`
	WebRoot         string   `json:"webroot"`
}

var (
	ConfigSingle *Config
)

func NewConfig(fileName string) (err error) {
	var (
		content []byte
		conf    *Config
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
	ConfigSingle = conf
	return
}
