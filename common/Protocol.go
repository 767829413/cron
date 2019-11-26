package common

import (
	"context"
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
	Job           *Job               //任务信息
	PlanTime      time.Time          //理论调度时间
	RealTime      time.Time          //真实调度时间
	CancelContext context.Context    //用于取消任务的context
	CancelFunc    context.CancelFunc //用于取消command的执行函数
}

//任务执行结果
type JobExcuteResult struct {
	ExccuteInfo *JobExecuteInfo //执行状态
	Output      []byte          //执行输出
	Err         error           //执行的错误
	StartTime   time.Time       //启动时间(真实)
	EndTime     time.Time       //结束时间
}

//任务执行日志
type JobLog struct {
	JobName      string `bson:"jobName"`      //任务名称
	Command      string `bson:"command"`      //脚本命令
	Err          string `bson:"err"`          //错误信息
	Output       string `bson:"output"`       //输出信息
	PlanTime     int64  `bson:"planTime"`     //计划开始时间
	ScheduleTime int64  `bson:"scheduleTime"` //任务调度时间
	StartTime    int64  `bson:"startTime"`    //任务执行开始时间
	EndTime      int64  `bson:"endTime"`      //任务执行结束时间
}

//日志批次(做成多批次插入)
type LogBatch struct {
	Logs []interface{} //多条日志
}

// 任务日志过滤条件
type JobLogFilter struct {
	JobName string `bson:"jobName"`
}

// 任务日志排序规则
type SortLogByStartTime struct {
	SortOrder int `bson:"startTime"` // {startTime: -1}
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
func ExtractJobName(jobKey string, jobPreFix string) string {
	return strings.TrimPrefix(jobKey, jobPreFix)
}

func ExtractWorkerIP(regKey string) string {
	return strings.TrimPrefix(regKey, JobWorkerDir)
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
	jobExecuteInfo.CancelContext, jobExecuteInfo.CancelFunc = context.WithCancel(context.TODO())
	return
}

//构建执行结果信息
func BuildJobExcuteResult(info *JobExecuteInfo, output []byte, start time.Time, end time.Time, err error) (jobExcuteResult *JobExcuteResult) {
	return &JobExcuteResult{
		ExccuteInfo: info,
		Output:      output,
		StartTime:   start,
		EndTime:     end,
		Err:         err,
	}
}

//构建执行日志信息
func BuildJobLog(result *JobExcuteResult) (jobLog *JobLog) {
	jobLog = &JobLog{
		JobName:      result.ExccuteInfo.Job.Name,
		Command:      result.ExccuteInfo.Job.Command,
		Output:       string(result.Output),
		PlanTime:     result.ExccuteInfo.PlanTime.UnixNano() / 1000 / 1000,
		ScheduleTime: result.ExccuteInfo.RealTime.UnixNano() / 1000 / 1000,
		StartTime:    result.StartTime.UnixNano() / 1000 / 1000,
		EndTime:      result.EndTime.UnixNano() / 1000 / 1000,
	}
	if result.Err != nil {
		jobLog.Err = result.Err.Error()
	} else {
		jobLog.Err = ""
	}
	return
}
