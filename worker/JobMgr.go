package worker

import (
	"context"
	"cron/common"
	"go.etcd.io/etcd/clientv3"
	"go.etcd.io/etcd/mvcc/mvccpb"
	"time"
)

type JobMgr struct {
	Client  *clientv3.Client
	KV      clientv3.KV
	Lease   clientv3.Lease
	Watcher clientv3.Watcher
}

var (
	JobMgrSingle *JobMgr
)

func NewJobMgr() (err error) {
	var (
		config  clientv3.Config
		client  *clientv3.Client
		kv      clientv3.KV
		lease   clientv3.Lease
		watcher clientv3.Watcher
	)
	//初始化配置
	config = clientv3.Config{
		Endpoints:   ConfSingle.EtcdEndPoints,                                     //集群地址
		DialTimeout: time.Duration(ConfSingle.EtcdDialTimeout) * time.Millisecond, //超时时间
	}

	//建立链接
	if client, err = clientv3.New(config); err != nil {
		return
	}
	//得到KV和Lease的API子集
	kv = clientv3.NewKV(client)
	lease = clientv3.NewLease(client)
	watcher = clientv3.NewWatcher(client)
	JobMgrSingle = &JobMgr{
		Client:  client,
		KV:      kv,
		Lease:   lease,
		Watcher: watcher,
	}
	return
}

//监听KV变化
func (jobMgr *JobMgr) WatchJobs() (err error) {
	var (
		getResp            *clientv3.GetResponse
		kvPair             *mvccpb.KeyValue
		job                *common.Job
		watchStartRevision int64
		watchChan          clientv3.WatchChan
		watchResp          clientv3.WatchResponse
		event              *clientv3.Event
		jobName            string
		jobEvent           *common.JobEvent
	)
	//1.获取/cron/jobs/目录下的所有任务,并且获知当前集群的revision
	if getResp, err = jobMgr.KV.Get(context.TODO(), common.JOB_SAVE_DIR, clientv3.WithPrefix()); err != nil {
		return
	}
	//遍历当前任务
	for _, kvPair = range getResp.Kvs {
		//反序列化json得到job
		if job, err = common.UnpackJob(kvPair.Value); err == nil {
			jobEvent = common.BuildJobEvent(common.JOB_SAVE_EVENT, job)
			// TODO 把job同步给调度的goruntime
			SchedulerSingle.JobEventChan <- jobEvent
		}
	}
	//.从该revision向后监听任务变化,启一个监听协程
	go func() {
		//从get的时刻监听后续版本变化
		watchStartRevision = getResp.Header.Revision + 1
		//启动监听/cron/jobs/目录的后续变化
		watchChan = jobMgr.Watcher.Watch(context.TODO(), common.JOB_SAVE_DIR, clientv3.WithPrefix(), clientv3.WithRev(watchStartRevision))
		for watchResp = range watchChan {
			for _, event = range watchResp.Events {
				switch event.Type {
				case mvccpb.PUT: //保存或更新
					if job, err = common.UnpackJob(event.Kv.Value); err != nil {
						continue
					}
					//构造保存/更新Event
					jobEvent = common.BuildJobEvent(common.JOB_SAVE_EVENT, job)
				case mvccpb.DELETE: //删除
					//提取任务名进行拼接
					jobName = common.ExtractJobName(string(event.Kv.Key))
					job = &common.Job{
						Name: jobName,
					}
					//构造删除Event
					jobEvent = common.BuildJobEvent(common.JOB_DELETE_EVENT, job)
				}
				//TODO 推送更新事件给scheduler
				SchedulerSingle.JobEventChan <- jobEvent
			}
		}
	}()
	return
}

//构建分布式锁来处理并发(etcd原生支持)
func (jobMgr *JobMgr) CreateJobLock(jobName string) (jobLock *JobLock) {
	jobLock = NewJobLock(jobName, jobMgr.KV, jobMgr.Lease)
	return
}
