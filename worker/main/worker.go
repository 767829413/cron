package main

import (
	"cron/worker"
	"flag"
	"log"
	"runtime"
	"time"
)

var (
	confFile string //配置文件路径
)

//解析命令行参数
func initArgs() {
	//worker -config ./master.json
	flag.StringVar(&confFile, "config", "./worker.json", "指定master.json")
	flag.Parse()
}

//初始化线程数量
func initEnv() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func main() {
	var (
		err     error
		excutor *worker.Excutor
	)

	//初始化命令行参数
	initArgs()

	//初始化线程
	initEnv()

	//加载配置
	if err = worker.NewConfig(confFile); err != nil {
		goto ERR
	}
	//初始化任务管理器
	if err = worker.NewJobMgr(); err != nil {
		goto ERR
	}
	//启动任务执行器
	if excutor, err = worker.NewExcutor(); err != nil {
		goto ERR
	}
	//启动任务调度器
	if err = worker.NewScheduler(excutor); err != nil {
		goto ERR
	}
	//进行监听
	if err = worker.JobMgrSingle.WatchJobs(); err != nil {
		goto ERR
	}
	worker.SchedulerSingle.ScheduleLoop()
	//正常退出
	for {
		time.Sleep(1 * time.Second)
	}
ERR:
	log.Println(err)
}
