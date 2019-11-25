package common

import (
	"encoding/json"
	"github.com/gorhill/cronexpr"
	"strings"
	"time"
)

//约定的任务
type Job struct {
	Name     string `json:"name"`
	Command  string `json:"command"`
	CronExpr string `json:"cronExpr"`
}

//调度计划
type JobSchedulePlan struct {
	Job      *Job
	Expr     *cronexpr.Expression //解析完毕的cron expr 表达式
	NextTime time.Time            //任务下次执行时间
}

//执行状态
type JobExecuteInfo struct {
	Job      *Job      //任务信息
	PlanTime time.Time //理论调度时间
	RealTime time.Time //真实调度时间
}

//任务执行结果
type JobExcuteResult struct {
	ExccuteInfo *JobExecuteInfo //执行状态
	Output      []byte          //执行输出
	Err         error           //执行的错误
	StartTime   time.Time       //启动时间(真实)
	EndTime     time.Time       //结束时间

}

//HTTP接口应答
type Response struct {
	Errno int         `json:"errno"`
	Msg   string      `json:"msg"`
	Data  interface{} `json:"data"`
}

//变化事件
type JobEvent struct {
	EventType int //事件类型 SAVE DELETE
	Job       *Job
}

//应答方法
func BuildResponse(errno int, msg string, data interface{}) (response []byte) {
	resp := Response{
		Errno: errno,
		Msg:   msg,
		Data:  data,
	}
	if response, err := json.Marshal(resp); err != nil {
		panic(err)
	} else {
		return response
	}
}

//反序列化方法
func UnpackJob(value []byte) (res *Job, err error) {
	var (
		job *Job
	)
	job = &Job{}
	if err = json.Unmarshal(value, job); err != nil {
		return
	}
	res = job
	return
}

//提取etcd的key中任务名称
func ExtractJobName(jobKey string) string {
	return strings.TrimPrefix(jobKey, JobSaveDir)
}

//任务变化的事件,1.更新 2.删除
func BuildJobEvent(eventType int, job *Job) *JobEvent {
	return &JobEvent{
		EventType: eventType,
		Job:       job,
	}
}

//构造任务执行计划
func BuildJobSchedulePlan(job *Job) (jobSchedulePlan *JobSchedulePlan, err error) {
	var (
		expre *cronexpr.Expression
	)
	//解析job的cron表达式
	if expre, err = cronexpr.Parse(job.CronExpr); err != nil {
		return
	}
	//生成调度计划任务对象
	jobSchedulePlan = &JobSchedulePlan{
		Job:      job,
		Expr:     expre,
		NextTime: expre.Next(time.Now()),
	}
	return
}

//构建执行状态信息
func BuildJobExcuteInfo(jobSchedulePlanlan *JobSchedulePlan) (jobExecuteInfo *JobExecuteInfo) {
	jobExecuteInfo = &JobExecuteInfo{
		Job:      jobSchedulePlanlan.Job,
		PlanTime: jobSchedulePlanlan.NextTime, //计划调度时间
		RealTime: time.Now(),                  //真实调度时间
	}
	return
}

//构建执行结果信息
func BuildJobExcuteResult(info *JobExecuteInfo, start time.Time, end time.Time, err error) (jobExcuteResult *JobExcuteResult) {
	return &JobExcuteResult{
		ExccuteInfo: info,
		Output:      make([]byte, 0),
		StartTime:   start,
		EndTime:     end,
		Err:         err,
	}
}
