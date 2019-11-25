package worker

import (
	"context"
	"cron/common"
	"os/exec"
	"time"
)

type Excutor struct {
}

//初始化
func NewExcutor() (excutor *Excutor, err error) {
	excutor = &Excutor{}
	return
}

//任务真实执行
func (excutor *Excutor) ExecuteJpb(jobExecuteInfo *common.JobExecuteInfo) {
	//多协程执行任务
	go excutor.run(jobExecuteInfo)
}

func (excutor *Excutor) run(jobExecuteInfo *common.JobExecuteInfo) {
	var (
		cmd     *exec.Cmd
		err     error
		output  []byte
		result  *common.JobExcuteResult
		jobLock *JobLock
		start   time.Time
		end     time.Time
	)
	//首先获取分布式锁(防止并发任务执行,有锁跳过,没锁加锁执行)
	jobLock = JobMgrSingle.CreateJobLock(jobExecuteInfo.Job.Name)
	//记录执行开始时间
	if err = jobLock.TryLock(); err != nil {
		start = time.Now()
		end = start
	} else {
		start = time.Now()
		//执行shell命令
		cmd = exec.CommandContext(context.TODO(), ConfSingle.ShellExcuseBash, ConfSingle.ShellExcuseArg, jobExecuteInfo.Job.Command)
		//执行并捕获输出
		output, err = cmd.CombinedOutput()
		//任务执行完成,构建执行结果,返回schedule,删除excuting table中的执行记录
		end = time.Now()
	}
	result = common.BuildJobExcuteResult(jobExecuteInfo, output, start, end, err)
	SchedulerSingle.PushJobResult(result)
	defer jobLock.UnLock()
}
