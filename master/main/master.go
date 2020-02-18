package main

import (
	"cronWeb/master"
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
	flag.StringVar(&confFile, "config", "./master.json", "指定配置文件")
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
	if err = master.InitConfig(confFile); err != nil {
		goto ERROR
	}
	// 日志管理器
	if err = master.InitLogMgr(); err != nil {
		goto ERROR
	}
	// 初始化服务发现
	if err = master.InitWorkerMgr(); err != nil {
		goto ERROR
	}
	// 任务管理器
	if err = master.InitJobMgr(); err != nil {
		goto ERROR
	}

	// 启动api http服务
	if err = master.InitApiServer(); err != nil {
		goto ERROR
	}

	for {
		time.Sleep(1 * time.Second)

	}
	return

ERROR:
	fmt.Println(err)
}
