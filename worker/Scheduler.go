package worker

import (
	"cron/common"
	"log"
	"time"
)

var (
	SchedulerSingle *Scheduler
)

type Scheduler struct {
	Excutor          *Excutor                           //执行器
	JobEventChan     chan *common.JobEvent              //etcd任务事件变化
	JobSchedulePlan  map[string]*common.JobSchedulePlan //任务调度计划
	JobExectingTable map[string]*common.JobExecuteInfo  //任务执行状态表
	JobResultChan    chan *common.JobExcuteResult       //任务执行结果通道
}

//初始化Scheduler
func NewScheduler(excutor *Excutor) (err error) {
	SchedulerSingle = &Scheduler{
		Excutor:          excutor,
		JobEventChan:     make(chan *common.JobEvent, 1000),
		JobSchedulePlan:  make(map[string]*common.JobSchedulePlan),
		JobExectingTable: make(map[string]*common.JobExecuteInfo),
		JobResultChan:    make(chan *common.JobExcuteResult, 1000),
	}
	return
}

//调度携程
func (scheduler *Scheduler) ScheduleLoop() {
	var (
		jobEvent      *common.JobEvent
		scheduleAfter time.Duration
		shceduleTimer *time.Timer
		jobResult     *common.JobExcuteResult
	)
	//初始化下次执行时间
	scheduleAfter = scheduler.TrySchedule()
	//调度延迟定时器
	shceduleTimer = time.NewTimer(scheduleAfter)
	for {
		select {
		//监听任务变化事件
		case jobEvent = <-scheduler.JobEventChan: //内存中列表的增删改查
			scheduler.handlerJobEvent(jobEvent)
		case <-shceduleTimer.C: //最近任务到期了
		case jobResult = <-scheduler.JobResultChan: //监听任务执行结果
			scheduler.handlerJobResult(jobResult)
		}
		//调度一次任务
		scheduleAfter = scheduler.TrySchedule()
		//重置调度间隔
		shceduleTimer.Reset(scheduleAfter)
	}
}

//任务执行逻辑
func (scheduler *Scheduler) TrySchedule() (scheduleAfter time.Duration) {
	/**
	TODO
	1.遍历所有任务
	2.过期任务立即执行
	3.统计最近要过期的任务时间
	*/
	var (
		jobPlan  *common.JobSchedulePlan
		now      time.Time
		nearTime *time.Time
	)
	//如果任务为空,睡眠时间随意(默认1秒)
	if len(scheduler.JobSchedulePlan) == 0 {
		scheduleAfter = time.Second * time.Duration(ConfSingle.ScheduleSleepTime)
		return
	}
	now = time.Now()
	for _, jobPlan = range scheduler.JobSchedulePlan {
		if jobPlan.NextTime.Before(now) || jobPlan.NextTime.Equal(now) {
			scheduler.TryRunJob(jobPlan)
			jobPlan.NextTime = jobPlan.Expr.Next(now)
		}
		//统计最近一个要过期的任务时间
		if nearTime == nil || jobPlan.NextTime.Before(*nearTime) {
			nearTime = &jobPlan.NextTime
		}
	}
	//下次调度时间间隔 = 最近的调度时间(不考虑优先级问题) - 当前时间
	scheduleAfter = (*nearTime).Sub(now)
	return
}

//处理事件
func (scheduler *Scheduler) handlerJobEvent(event *common.JobEvent) {
	var (
		jobSchedulePlan *common.JobSchedulePlan
		jobExist        bool
		err             error
	)
	switch event.EventType {
	case common.JOB_SAVE_EVENT:
		//保存|更新任务事件
		if jobSchedulePlan, err = common.BuildJobSchedulePlan(event.Job); err != nil {
			return
		}
		scheduler.JobSchedulePlan[event.Job.Name] = jobSchedulePlan
	case common.JOB_DELETE_EVENT:
		//删除任务事件
		if jobSchedulePlan, jobExist = scheduler.JobSchedulePlan[event.Job.Name]; jobExist {
			delete(scheduler.JobSchedulePlan, event.Job.Name)
		}
	}
}

//执行任务
func (scheduler *Scheduler) TryRunJob(jobPlan *common.JobSchedulePlan) {
	//调度是调度,执行是执行不是一个概念(执行1次,调度可能N次),防止并发
	var (
		jobExceteInfo *common.JobExecuteInfo
		jobExcuting   bool
	)
	//正在执行的任务需要跳过
	if jobExceteInfo, jobExcuting = scheduler.JobExectingTable[jobPlan.Job.Name]; jobExcuting {
		log.Println("尚未完成,跳过执行: ", jobPlan.Job.Name)
		return
	}
	//构建执行状态信息
	jobExceteInfo = common.BuildJobExcuteInfo(jobPlan)
	//保存状态
	scheduler.JobExectingTable[jobPlan.Job.Name] = jobExceteInfo
	//执行任务
	scheduler.Excutor.ExecuteJpb(jobExceteInfo)
	log.Println("执行任务: ", jobExceteInfo.Job.Name, " ", jobExceteInfo.PlanTime, " ", jobExceteInfo.RealTime)
}

//回传执行结果
func (scheduler *Scheduler) PushJobResult(result *common.JobExcuteResult) {
	scheduler.JobResultChan <- result
}

//处理任务结果
func (scheduler *Scheduler) handlerJobResult(result *common.JobExcuteResult) {
	//删除执行状态表里的任务
	delete(scheduler.JobExectingTable, result.ExccuteInfo.Job.Name)
	log.Println("任务执行完成: ", result.ExccuteInfo.Job.Name, string(result.Output), "这是错误: ", result.Err)
}
