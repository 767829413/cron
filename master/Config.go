package master

import (
	"encoding/json"
	"io/ioutil"
)

//程序配置
type Config struct {
	ApiPort int `json:"apiPort"`
	ApiReadTimeout int `json:"apiReadTimeout"`
	ApiWriteTimeout int `json:"apiWriteTimeout"`
	EtcdEndPoints []string `json:"etcdEndPoints"`
	EtcdDialTimeout int `json:"etcdDialTimeout"`
}

func NewConfig(fileName string)(conf *Config,err error){
	var(
		content []byte
	)
	//1.读入配置文件
	if content,err = ioutil.ReadFile(fileName);err!=nil{
		return nil,err
	}

	//2.反序列化
	if err = json.Unmarshal(content,&conf);err != nil {
		return nil,err
	}
	//3.赋值单例
	return conf,nil
}