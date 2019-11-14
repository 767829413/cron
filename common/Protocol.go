package common

import (
	"encoding/json"
)

type Job struct {
	Name string `json:"name"`
	Command string `json:"command"`
	CronExpr string `json:"cronExpr"`
}

//HTTP接口应答
type Response struct {
	Errno int `json:"errno"`
	Msg string `json:"msg"`
	Data interface{} `json:"data"`
}

//应答方法
func BuildResponse(errno int,msg string,data interface{})(response []byte){
	resp := Response{
		Errno:errno,
		Msg:msg,
		Data:data,
	}
	if response,err := json.Marshal(resp);err !=nil{
		panic(err)
	}else{
		return response
	}
}