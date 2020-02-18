package worker

import (
	"cronWeb/common"
	"math/rand"
	"os/exec"
	"time"
)

// 任务执行器
type Executor struct {
}

var G_executor *Executor

// 执行一个任务的
func (executor *Executor) ExecuteJob(info *common.JobExecuteInfo) {

	go func() {
		var (
			cmd     *exec.Cmd
			err     error
			outPut  []byte
			result  *common.JobExecuteResult
			jobLock *JobLock
		)
		// 任务结果
		result = &common.JobExecuteResult{
			ExecuteInfo: info,
			OutPut:      make([]byte, 0),
		}
		// 初始化锁
		jobLock = G_jobMgr.CreateJobLock(info.Job.Name)
		// 记录任务开始时间
		result.StartTime = time.Now()
		// 随机睡眠
		time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)
		// 上锁
		err = jobLock.TryLock()
		// 释放锁
		defer jobLock.UnLock()
		if err != nil { // 上锁失败
			result.Err = err
			result.EndTime = time.Now()
		} else {
			// 上锁成功后，重置启动时间
			result.StartTime = time.Now()
			// 执行shell命令
			cmd = exec.CommandContext(info.CancelCtx, "/bin/bash", "-c", info.Job.Command)
			// 执行并捕获输出
			outPut, err = cmd.CombinedOutput()
			// 任务的结束时间
			result.EndTime = time.Now()
			result.OutPut = outPut
			result.Err = err
		}
		// 任务执行完成后，把执行结果返回给scheduler scheduler会从executingTable 中删除结果
		G_scheduler.PushJobResult(result)
	}()

}

// 初始化执行器
func InitExecutor() (err error) {
	G_executor = &Executor{}
	return

}
