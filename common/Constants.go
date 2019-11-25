package common

const (
	//任务保存目录
	JobSaveDir = "/cron/jobs/"
	//任务杀死目录
	JobKillDir = "/cron/killer/"
	//任务上锁目录
	JobLockDir = "/cron/lock/"
	//kill操作的租约过期时间
	OpKillExpired = 1

	//保存任务事件
	JobSaveEvent   = 1
	JobDeleteEvent = 2
)
