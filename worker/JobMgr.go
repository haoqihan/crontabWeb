package worker

import (
	"context"
	"cronWeb/common"
	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"time"
)

// 任务管理器
type JobMgr struct {
	client  *clientv3.Client
	kv      clientv3.KV
	lease   clientv3.Lease
	watcher clientv3.Watcher
}

var (
	G_jobMgr *JobMgr
)

// 初始化管理器
func InitJobMgr() (err error) {
	var (
		config  clientv3.Config
		client  *clientv3.Client
		kv      clientv3.KV
		lease   clientv3.Lease
		watcher clientv3.Watcher
	)
	// 初始化配置
	config = clientv3.Config{
		Endpoints:   G_config.EtcdEndpoints,                                    // 集群地址
		DialTimeout: time.Duration(G_config.EtcdDiaTimeout) * time.Millisecond, // 超时时间
	}
	// 建立客户端
	if client, err = clientv3.New(config); err != nil {
		return
	}
	// 得到kv和lease的api子集
	kv = clientv3.NewKV(client)
	lease = clientv3.NewLease(client)
	watcher = clientv3.NewWatcher(client)

	// 赋值单例
	G_jobMgr = &JobMgr{
		client:  client,
		kv:      kv,
		lease:   lease,
		watcher: watcher,
	}
	// 启动任务监听
	G_jobMgr.watchJobs()

	// 启动监听killer
	G_jobMgr.watchKiller()

	return
}

// 监听强杀任务通知
func (jobMgr *JobMgr) watchKiller() {
	var (
		watchChan     clientv3.WatchChan
		watchResponse clientv3.WatchResponse
		watchEvent    *clientv3.Event
		jobEvent      *common.JobEvent
		jobName       string
		job           *common.Job
	)

	// 从该reversion向后监听变化事件
	go func() {
		// 启动监听/cron/killer 目录的后续变化
		watchChan = jobMgr.watcher.Watch(context.TODO(), common.JOB_KILLER_DIR, clientv3.WithPrefix())
		// 处理监听事件
		for watchResponse = range watchChan {
			for _, watchEvent = range watchResponse.Events {
				switch watchEvent.Type {
				case mvccpb.PUT: // 杀死任务事件
					jobName = common.ExtractKillerName(string(watchEvent.Kv.Key))
					job = &common.Job{Name: jobName}
					jobEvent = common.BuildJobEvent(common.JOB_EVENT_KILL, job)
					G_scheduler.PushJobEvent(jobEvent)
				case mvccpb.DELETE: // killer标记过期，被自动删除
				}
			}
		}
	}()

	return

}

// 监听任务变化
func (jobMgr *JobMgr) watchJobs() (err error) {
	var (
		getResp            *clientv3.GetResponse
		kvpair             *mvccpb.KeyValue
		job                *common.Job
		watchStartRevision int64
		watchChan          clientv3.WatchChan
		watchResponse      clientv3.WatchResponse
		watchEvent         *clientv3.Event
		jobName            string
		jobEvent           *common.JobEvent
	)
	// get一下/cron/jobs 目录下的所有任务，并且获知当前集群的reversion
	if getResp, err = jobMgr.kv.Get(context.TODO(), common.JOB_SAVE_DIR, clientv3.WithPrefix()); err != nil {
		return
	}
	// 当前有哪些任务
	for _, kvpair = range getResp.Kvs {
		// 序列化json得到job
		if job, err = common.UnpackJob(kvpair.Value); err == nil {
			jobEvent = common.BuildJobEvent(common.JOB_EVENT_SAVE, job)
			// TODO:把这个job同步给scheduler(调度协程)
			G_scheduler.PushJobEvent(jobEvent)
		}
	}
	// 从该reversion向后监听变化事件
	go func() {
		// 从get时刻的后续版本监听变化
		watchStartRevision = getResp.Header.Revision + 1
		// 启动监听/cron/jobs 目录的后续变化
		watchChan = jobMgr.watcher.Watch(context.TODO(), common.JOB_SAVE_DIR, clientv3.WithRev(watchStartRevision), clientv3.WithPrefix())
		// 处理监听事件
		for watchResponse = range watchChan {
			for _, watchEvent = range watchResponse.Events {
				switch watchEvent.Type {
				case mvccpb.PUT: // 任务保存事件
					// TODO:反序列化job，推送更新事件给scheduler
					if job, err = common.UnpackJob(watchEvent.Kv.Value); err != nil {
						continue
					}
					// 构建一个更新的Event事件
					jobEvent = common.BuildJobEvent(common.JOB_EVENT_SAVE, job)
				case mvccpb.DELETE: // 任务删除事件
					// TODO:推送删除事件给scheduler
					jobName = common.ExtractJobName(string(watchEvent.Kv.Key))
					job = &common.Job{
						Name: jobName,
					}
					// 构造一个删除的event事件
					jobEvent = common.BuildJobEvent(common.JOB_EVENT_DELETE, job)
				}

				// TODO:推送给scheduler
				G_scheduler.PushJobEvent(jobEvent)

			}

		}

	}()

	return
}

// 创建任务执行锁
func (jobMgr *JobMgr) CreateJobLock(jobName string) (jobLock *JobLock) {
	// 返回一把锁
	jobLock = InitJobLock(jobName, jobMgr.kv, jobMgr.lease)
	return

}
