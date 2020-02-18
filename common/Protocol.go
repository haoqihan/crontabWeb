package common

import (
	"context"
	"encoding/json"
	"github.com/gorhill/cronexpr"
	"strings"
	"time"
)

// 定时任务
type Job struct {
	Name     string `json:"name"`     // 任务名
	Command  string `json:"command"`  // shell 命令
	CronExpr string `json:"cronExpr"` //cron 表达式
}

// 任务调度计划
type JobSchedulerPlan struct {
	Job      *Job                 // 要调度的任务信息
	Expr     *cronexpr.Expression // 解析号的cronexpr表达式
	NextTime time.Time            // 下次调度时间

}

// 任务执行状态
type JobExecuteInfo struct {
	Job        *Job               // 任务信息
	PlanTime   time.Time          // 理论上的调度时间
	RealTime   time.Time          // 实际的调度时间
	CancelCtx  context.Context    // 用于取消任务的context
	CancelFunc context.CancelFunc // 用于取消command执行的cancel函数
}

// http接口应答
type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

// 变化事件
type JobEvent struct {
	EventType int // SAVE DELETE
	Job       *Job
}

// 任务执行结果
type JobExecuteResult struct {
	ExecuteInfo *JobExecuteInfo // 执行状态
	OutPut      []byte          // 脚本输出
	Err         error           // 错误原因
	StartTime   time.Time       // 启动时间
	EndTime     time.Time       //结束时间
}

// 任务执行日志
type JobLog struct {
	JobName      string `bson:"JobName"json:"job_name"`           // 任务名字
	Command      string `bson:"command"json:"command"`            // 脚本命令
	Err          string `bson:"err"json:"err"`                    //错误原因
	OutPut       string `bson:"output"json:"out_put"`             // 脚本输出
	PlanTime     int64  `bson:"planTime"json:"plan_time"`         // 计划开始时间
	ScheduleTime int64  `bson:"scheduleTime"json:"schedule_time"` // 实际调度时间
	StartTime    int64  `bson:"startTime"json:"start_time"`       // 任务执行时间
	EndTime      int64  `bson:"endTime"json:"end_time"`           // 任务执行结束的时间

}

// 日志批次
type LogBatch struct {
	Logs []interface{} // 多条日志
}

// 任务日志过滤条件
type JobNameFilter struct {
	JobName string `bson:"JobName"`
}

// 任务日志排序规则
type SortLogByStartTime struct {
	SortOrder int `bson:"startTime"`
}

// 应答方法
func BuildResponse(code int, msg string, data interface{}) (resp []byte, err error) {
	var (
		response Response
	)
	response.Code = code
	response.Msg = msg
	response.Data = data
	// 序列化
	resp, err = json.Marshal(response)
	return
}

// 反序列化job
func UnpackJob(value []byte) (ret *Job, err error) {
	var (
		job *Job
	)
	job = &Job{}
	if err = json.Unmarshal(value, job); err != nil {
		return
	}
	ret = job
	return
}

// 从etcd中的key提取任务名
// /cron/jobs/job1 抹掉 /cron/jobs/
func ExtractJobName(jobKey string) string {
	return strings.TrimPrefix(jobKey, JOB_SAVE_DIR)
}

//  /cron/killer/job10 抹掉 /cron/killer/
func ExtractKillerName(killerKey string) string {
	return strings.TrimPrefix(killerKey, JOB_KILLER_DIR)
}

// 提取worker的ip
func ExtractWorkerIP(key string) string {
	return strings.TrimPrefix(key, JOB_WORKER_DIR)
}

// 任务事件有2中，1更新任务  2删除任务
func BuildJobEvent(eventType int, job *Job) (jobEvent *JobEvent) {
	return &JobEvent{
		EventType: eventType,
		Job:       job,
	}
}

// 构造任务执行计划
func BuildJobSchedulePlan(job *Job) (jobSchedulerPlan *JobSchedulerPlan, err error) {
	var (
		expr *cronexpr.Expression
	)
	// 解析job的cron表达式
	if expr, err = cronexpr.Parse(job.CronExpr); err != nil {
		return
	}
	// 生成任务调度计划
	jobSchedulerPlan = &JobSchedulerPlan{
		Job:      job,
		Expr:     expr,
		NextTime: expr.Next(time.Now()),
	}
	return
}

// 构造执行状态信息
func BuildJobExecuteInfo(jobScheduler *JobSchedulerPlan) (jobExecuteInfo *JobExecuteInfo) {
	jobExecuteInfo = &JobExecuteInfo{
		Job:      jobScheduler.Job,
		PlanTime: jobScheduler.NextTime, // 计划调度时间
		RealTime: time.Now(),            // 真实调度时间
	}
	jobExecuteInfo.CancelCtx, jobExecuteInfo.CancelFunc = context.WithCancel(context.TODO())
	return
}
