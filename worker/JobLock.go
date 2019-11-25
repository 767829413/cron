package worker

import (
	"context"
	"cron/common"
	"go.etcd.io/etcd/clientv3"
)

//分布式锁的一种实现(利用ETCD的TXN事务实现)
type JobLock struct {
	KV        clientv3.KV
	Lease     clientv3.Lease
	JobName   string             //任务名
	CanceFunc context.CancelFunc //取消租约(终止自动续租)
	LeaseID   clientv3.LeaseID
	IsLock    bool
}

//初始化(未上锁)
func NewJobLock(jobName string, kv clientv3.KV, lease clientv3.Lease) (jobLock *JobLock) {
	jobLock = &JobLock{
		KV:      kv,
		Lease:   lease,
		JobName: jobName,
	}
	return
}

//尝试上锁(乐观锁)
func (jobLock *JobLock) TryLock() (err error) {
	var (
		leaseGrantResp *clientv3.LeaseGrantResponse
		cancelCtx      context.Context
		cancelFunc     context.CancelFunc
		keepRespChan   <-chan *clientv3.LeaseKeepAliveResponse
		txn            clientv3.Txn
		lockKey        string
		txnResp        *clientv3.TxnResponse
	)
	/**
	TODO
	1.创建租约 5秒
	2.自动续租
	3.创建事务 txn
	4.利用事务抢锁
	5.成功放回,失败释放租约
	*/
	if leaseGrantResp, err = jobLock.Lease.Grant(context.TODO(), int64(ConfSingle.EtcdLeaseTimeout)); err != nil {
		return
	}
	cancelCtx, cancelFunc = context.WithCancel(context.TODO())
	if keepRespChan, err = jobLock.Lease.KeepAlive(cancelCtx, leaseGrantResp.ID); err != nil {
		goto FAIL
	}
	go func() {
		var (
			keepResp *clientv3.LeaseKeepAliveResponse
		)
		for {
			select {
			case keepResp = <-keepRespChan:
				if keepResp == nil {
					goto END
				}
			}
		}
	END:
	}()

	txn = jobLock.KV.Txn(context.TODO())
	lockKey = common.JobLockDir + jobLock.JobName
	txn.If(clientv3.Compare(clientv3.CreateRevision(lockKey), "=", 0)).Then(clientv3.OpPut(lockKey, "", clientv3.WithLease(leaseGrantResp.ID))).Else(clientv3.OpGet(lockKey))
	if txnResp, err = txn.Commit(); err != nil {
		goto FAIL
	}
	if !txnResp.Succeeded {
		err = common.ErrLockAlreadyRequired
		goto FAIL
	}
	jobLock.LeaseID = leaseGrantResp.ID
	jobLock.CanceFunc = cancelFunc
	jobLock.IsLock = true
	return
FAIL:
	cancelFunc()
	_,_ = jobLock.Lease.Revoke(context.TODO(), leaseGrantResp.ID)
	return
}

//释放锁
func (jobLock *JobLock) UnLock() {
	if jobLock.IsLock {
		jobLock.CanceFunc()
		_,_ = jobLock.Lease.Revoke(context.TODO(), jobLock.LeaseID)
	}
}
