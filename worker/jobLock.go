package worker

import (
	"context"
	"cronWeb/common"
	"github.com/coreos/etcd/clientv3"
)

// 分布式锁
type JobLock struct {
	kv         clientv3.KV
	lease      clientv3.Lease
	jobName    string             // 任务名
	cancelFunc context.CancelFunc // 用于终止续租
	leaseId    clientv3.LeaseID   //租约id
	isLocked   bool               // 是否上锁成功
}

// 初始化一把锁
func InitJobLock(jobName string, kv clientv3.KV, lease clientv3.Lease) (jobLock *JobLock) {
	jobLock = &JobLock{
		kv:      kv,
		lease:   lease,
		jobName: jobName,
	}
	return

}

// 尝试上锁
func (joblock *JobLock) TryLock() (err error) {
	var (
		leaseGrantResp *clientv3.LeaseGrantResponse
		cancelCtx      context.Context
		cancelFunc     context.CancelFunc
		leaseid        clientv3.LeaseID
		keepRespChan   <-chan *clientv3.LeaseKeepAliveResponse
		txn            clientv3.Txn
		lockKey        string
		txnResp        *clientv3.TxnResponse
	)
	// 1.创建租约(5秒）
	if leaseGrantResp, err = joblock.lease.Grant(context.TODO(), 5); err != nil {
		return
	}
	// context 用于取消自动续租
	cancelCtx, cancelFunc = context.WithCancel(context.TODO())
	// 租约id
	leaseid = leaseGrantResp.ID
	// 2.自动续租
	if keepRespChan, err = joblock.lease.KeepAlive(cancelCtx, leaseid); err != nil {
		goto FALL
	}
	// 处理续租应答的协程
	go func() {
		var (
			keepResp *clientv3.LeaseKeepAliveResponse
		)
		for {
			select {
			case keepResp = <-keepRespChan: // 自动续租的应答
				if keepResp == nil {
					goto END
				}
			}
		}
	END:
	}()
	// 4.创建事务 tnx
	txn = joblock.kv.Txn(context.TODO())
	// 4.锁路径
	lockKey = common.JOB_LOCK_DIR + joblock.jobName

	// 5.事务抢锁
	txn.If(clientv3.Compare(clientv3.CreateRevision(lockKey), "=", 0)).
		Then(clientv3.OpPut(lockKey, "", clientv3.WithLease(leaseid))).Else(
		clientv3.OpGet(lockKey))
	if txnResp, err = txn.Commit(); err != nil {
		goto FALL

	}

	// 5.成功返回，失败释放租约
	if !txnResp.Succeeded { // 锁被占用
		err = common.ERR_LOCK_ALREADY_REQURED
		goto FALL

	}
	// 抢锁成功
	joblock.leaseId = leaseid
	joblock.cancelFunc = cancelFunc
	joblock.isLocked = true
	return
FALL:
	// 取消自动续租
	cancelFunc()
	// 释放租约
	joblock.lease.Revoke(context.TODO(), leaseid)
	return
}

// 释放锁
func (jobLock *JobLock) UnLock() {
	if jobLock.isLocked {
		jobLock.cancelFunc()                                  // 取消自动续租的协程
		jobLock.lease.Revoke(context.TODO(), jobLock.leaseId) // 释放租约
	}
}