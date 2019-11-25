package master

import (
	"cron/common"
	"encoding/json"
	"net/http"
)

//保存任务接口
//POST /job/save job={"name":"job1","command":"echo hello","cornExpr":"* * * * *"}
func GetHandleJobSaveFunc() (handler func(w http.ResponseWriter, r *http.Request)) {
	return func(w http.ResponseWriter, r *http.Request) {
		var (
			err     error
			postJob string
			job     *common.Job
			oldJob  *common.Job
		)
		w.Header().Set("Content-Type", "application/json")
		//1.解析POST表单
		if err = r.ParseForm(); err != nil {
			goto ERR
		}
		//2.取表单中的job字段
		postJob = r.PostForm.Get("job")
		//3.反序列化job
		if err = json.Unmarshal([]byte(postJob), &job); err != nil {
			goto ERR
		}
		//4.保存到ETCD
		if oldJob, err = JobMgrSingle.SaveJob(job); err != nil {
			goto ERR
		}
		//5.返回正常应答{"errno":0,"msg":"","data":{...}}
		if _, err = w.Write(common.BuildResponse(0, "success", oldJob)); err != nil {
			goto ERR
		}
		return
	ERR:
		w.Write(common.BuildResponse(-1, err.Error(), nil))
	}
}

//删除任务接口
//POST /job/delete name=job2
func GetHandleJobDeleteFunc() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var (
			err     error
			jobName string
			oldJob  *common.Job
		)
		//解析post表单
		if err = r.ParseForm(); err != nil {
			goto ERR
		}
		w.Header().Set("Content-Type", "application/json")
		//获取任务名称
		jobName = r.PostForm.Get("name")
		//进行删除
		if oldJob, err = JobMgrSingle.DeleteJob(jobName); err != nil {
			goto ERR
		}
		if _, err = w.Write(common.BuildResponse(0, "success", oldJob)); err != nil {
			goto ERR
		}
		return
	ERR:
		w.Write(common.BuildResponse(-1, err.Error(), nil))
	}
}

//获取任务列表(列举所有cron任务)
func GetHandleJobListFunc() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var (
			jobList []*common.Job
			err     error
		)
		w.Header().Set("Content-Type", "application/json")
		//获取任务列表
		if jobList, err = JobMgrSingle.ListJobs(); err != nil {
			goto ERR
		}
		//返回正常应答
		if _, err = w.Write(common.BuildResponse(0, "success", jobList)); err != nil {
			goto ERR
		}
		return
	ERR:
		w.Write(common.BuildResponse(-1, err.Error(), nil))
	}
}

//强制杀死某个任务
//POST /job/kill name=job1
func GetHandleJobKillFunc() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var (
			err     error
			jobName string
		)
		w.Header().Set("Content-Type", "application/json")
		if err = r.ParseForm(); err != nil {
			goto ERR
		}
		jobName = r.PostForm.Get("name")
		if err = JobMgrSingle.KillJob(jobName); err != nil {
			goto ERR
		}
		//返回正常应答
		if _, err = w.Write(common.BuildResponse(0, "success", nil)); err != nil {
			goto ERR
		}
		return
	ERR:
		w.Write(common.BuildResponse(-1, err.Error(), nil))
	}
}