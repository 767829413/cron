package master

import (
	"cron/common"
	"encoding/json"
	"net"
	"net/http"
	"strconv"
	"time"
)

//任务的http接口
type ApiServer struct {
	httpServer *http.Server
}

//保存任务接口
//POST job={"name":"job1","command":"echo hello","cornExpr":"* * * * *"}
func getHandleJobSaveFunc(jobMgr *JobMgr)( handler func(w http.ResponseWriter,r *http.Request) )  {
	return func (w http.ResponseWriter,r *http.Request){
		var(
			err error
			postJob string
			job *common.Job
			oldJob *common.Job

		)
		//1.解析POST表单
		if err = r.ParseForm();err !=nil{
			goto ERR
		}
		//2.取表单中的job字段
		postJob = r.PostForm.Get("job")
		//3.反序列化job
		if err = json.Unmarshal([]byte(postJob),&job);err !=nil{
			goto ERR
		}
		//4.保存到ETCD
		if oldJob,err = jobMgr.SaveJob(job);err !=nil{
			goto ERR
		}
		//5.返回正常应答{"errno":0,"msg":"","data":{...}}
		w.Write(common.BuildResponse(0,"success",oldJob))
		return
	ERR:
		w.Write(common.BuildResponse(-1,err.Error(),nil))
	}
}

//初始化服务
func NewApiServer(jobMgr *JobMgr,conf *Config)(apiServer *ApiServer,err error){
	var(
		mux *http.ServeMux
		lister net.Listener
		httpServer *http.Server
	)
	//配置路由
	mux = http.NewServeMux()
	mux.HandleFunc("/job/save",getHandleJobSaveFunc(jobMgr))

	//启动TCP监听
	if lister,err = net.Listen("tcp",":"+strconv.Itoa(conf.ApiPort));err != nil {
		return nil,err
	}
	//创建HTTP服务
	httpServer = &http.Server{
		ReadTimeout:time.Duration(conf.ApiReadTimeout)*time.Millisecond,
		WriteTimeout:time.Duration(conf.ApiWriteTimeout)*time.Millisecond,
		Handler:mux,
	}

	go httpServer.Serve(lister)

	return &ApiServer{
		httpServer:httpServer,
	},nil
}