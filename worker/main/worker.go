package main

import (
	"cronWeb/worker"
	"flag"
	"fmt"
	"runtime"
	"time"
)

var (
	confFile string // 配置文件路径
)

// 解析命令行参数
func initArgs() {
	// worker -config ./worker.json
	flag.StringVar(&confFile, "config", "./worker.json", "指定配置文件")
	flag.Parse()

}

// 定义线程数，最大线程数和cpu的核数相关
func initEnv() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func main() {
	var err error
	// 初始化命令行参数
	initArgs()
	// 初始化线程
	initEnv()
	// 加载配置
	if err = worker.InitConfig(confFile); err != nil {
		goto ERROR
	}
	// 服务注册
	if err = worker.InitRegister(); err != nil {
		goto ERROR
	}
	// 启动日志协程
	if err = worker.InitLogSinK(); err != nil {
		goto ERROR
	}
	// 启动执行器
	if err = worker.InitExecutor(); err != nil {
		goto ERROR
	}
	// 启动调度器
	if err = worker.InitScheduler(); err != nil {
		goto ERROR
	}
	// 初始化任务管理器
	if err = worker.InitJobMgr(); err != nil {
		goto ERROR
	}

	for {
		time.Sleep(1 * time.Second)

	}
	return

ERROR:
	fmt.Println(err)
}
