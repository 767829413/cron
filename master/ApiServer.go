package master

import (
	"cron/common"
	"encoding/json"
	"log"
	"net"
	"net/http"
	"strconv"
	"time"
)

//任务的http接口
type ApiServer struct {
	httpServer *http.Server
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
	mux.HandleFunc("/job/delete",getHandleJobDeleteFunc(jobMgr))
	mux.HandleFunc("/job/list",getHandleJobListFunc(jobMgr))
	mux.HandleFunc("/job/kill",getHandleJobKillFunc(jobMgr))

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

	go func() {
		if err = httpServer.Serve(lister);err != nil {
			log.Println(err)
		}
	}()

	return &ApiServer{
		httpServer:httpServer,
	},nil
}

//保存任务接口
//POST /job/save job={"name":"job1","command":"echo hello","cornExpr":"* * * * *"}
func getHandleJobSaveFunc(jobMgr *JobMgr)( handler func(w http.ResponseWriter,r *http.Request) )  {
	return func (w http.ResponseWriter,r *http.Request){
		var(
			err error
			postJob string
			job *common.Job
			oldJob *common.Job
		)
		w.Header().Set("Content-Type","application/json")
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
		if _,err = w.Write(common.BuildResponse(0,"success",oldJob));err != nil {
			goto ERR
		}
		return
	ERR:
		w.Write(common.BuildResponse(-1,err.Error(),nil))
	}
}

//删除任务接口
//POST /job/delete name=job2
func getHandleJobDeleteFunc(mgr *JobMgr) func(http.ResponseWriter, *http.Request) {
	return func (w http.ResponseWriter,r *http.Request){
		var(
			err error
			jobName string
			oldJob *common.Job
		)
		//解析post表单
		if err = r.ParseForm();err != nil {
			goto ERR
		}
		w.Header().Set("Content-Type","application/json")
		//获取任务名称
		jobName = r.PostForm.Get("name")
		//进行删除
		if oldJob,err = mgr.DeleteJob(jobName);err != nil {
			goto ERR
		}
		if _,err = w.Write(common.BuildResponse(0,"success",oldJob));err != nil {
			goto ERR
		}
		return
	ERR:
		w.Write(common.BuildResponse(-1,err.Error(),nil))
	}
}

//获取任务列表(列举所有cron任务)
func getHandleJobListFunc(mgr *JobMgr) func(http.ResponseWriter, *http.Request) {
	return func (w http.ResponseWriter,r *http.Request){
		var(
			jobList []*common.Job
			err error
		)
		w.Header().Set("Content-Type","application/json")
		//获取任务列表
		if jobList,err = mgr.ListJobs(); err != nil {
			goto ERR
		}
		//返回正常应答
		if _,err = w.Write(common.BuildResponse(0,"success",jobList));err != nil{
			goto ERR
		}
		return
		ERR:
			w.Write(common.BuildResponse(-1,err.Error(),nil))
	}
}

//强制杀死某个任务
//POST /job/kill name=job1
func getHandleJobKillFunc(mgr *JobMgr) func(http.ResponseWriter, *http.Request) {
	return func (w http.ResponseWriter,r *http.Request){
		var(
			err error
			jobName string
		)
		w.Header().Set("Content-Type","application/json")
		if err = r.ParseForm();err != nil {
			goto ERR
		}
		jobName = r.PostForm.Get("name")
		if err = mgr.KillJob(jobName);err != nil {
			goto ERR
		}
		//返回正常应答
		if _,err = w.Write(common.BuildResponse(0,"success",nil));err != nil{
			goto ERR
		}
		return
	ERR:
		w.Write(common.BuildResponse(-1,err.Error(),nil))
	}
}