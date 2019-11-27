package main

import (
	"cron/master"
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
	//master -config ./master.json
	flag.StringVar(&confFile, "config", "./master.json", "指定master.json")
	flag.Parse()
}

//初始化线程数量
func initEnv() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func main() {
	var (
		err error
	)
	//初始化命令行参数
	initArgs()

	//初始化线程
	initEnv()

	//加载配置
	if err = master.NewConfig(confFile); err != nil {
		goto ERR
	}
	//加载日志解析器
	if err = master.NewLogMgr(); err != nil {
		goto ERR
	}
	//加载工作节点解析器
	if err = master.NewWorkerMgr(); err != nil {
		goto ERR
	}

	if err = master.NewJobMgr(); err != nil {
		goto ERR
	}

	//启动Api HTTP服务
	if _, err = master.NewApiServer(); err != nil {
		goto ERR
	}
	//正常退出
	for {
		time.Sleep(1 * time.Second)
	}
ERR:
	log.Println(err)
}
