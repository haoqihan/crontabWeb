package main

import (
	"context"
	"fmt"
	"os/exec"
	"time"
)

type result struct {
	err    error
	output []byte
}

func main() {
	var (
		ctx        context.Context
		cancelFunc context.CancelFunc
		cmd        *exec.Cmd
		resultChan chan *result
		res        *result
	)
	// 创建结果队列
	resultChan = make(chan *result, 1000)

	// 创建一个上下文
	ctx, cancelFunc = context.WithCancel(context.TODO())
	go func() {
		var (
			output []byte
			err    error
		)

		cmd = exec.CommandContext(ctx, "/bin/bash", "-c", "sleep 2;ls -l")

		// 执行任务，捕获输出结果
		output, err = cmd.CombinedOutput()
		resultChan <- &result{
			err:    err,
			output: output,
		}

	}()
	// 暂停6秒
	time.Sleep(1 * time.Second)

	// 关闭上下文
	cancelFunc()

	// 获取执行结果
	res = <- resultChan

	// 打印任务执行结果
	fmt.Print(res.err,string(res.output))


}
