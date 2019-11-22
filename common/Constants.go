package common

const (
	//任务保存目录
	JOB_SAVE_DIR = "/cron/jobs/"
	//任务杀死目录
	JOB_KILL_DIR = "/cron/killer/"
	//任务上锁目录
	JOB_LOCK_DIR = "/cron/lock/"
	//kill操作的租约过期时间
	OP_KILL_EXPIRED = 1

	//保存任务事件
	JOB_SAVE_EVENT   = 1
	JOB_DELETE_EVENT = 2
)
