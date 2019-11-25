package master

import (
	"context"
	"cron/common"
	"encoding/json"
	"go.etcd.io/etcd/clientv3"
	"go.etcd.io/etcd/mvcc/mvccpb"
	"time"
)

type JobMgr struct {
	Client *clientv3.Client
	KV     clientv3.KV
	Lease  clientv3.Lease
}

var (
	JobMgrSingle *JobMgr
)

func NewJobMgr() (err error) {
	var (
		config clientv3.Config
		client *clientv3.Client
		kv     clientv3.KV
		lease  clientv3.Lease
	)
	//初始化配置
	config = clientv3.Config{
		Endpoints:   ConfigSingle.EtcdEndPoints,                                     //集群地址
		DialTimeout: time.Duration(ConfigSingle.EtcdDialTimeout) * time.Millisecond, //超时时间
	}

	//建立链接
	if client, err = clientv3.New(config); err != nil {
		return
	}
	//得到KV和Lease的API子集
	kv = clientv3.NewKV(client)
	lease = clientv3.NewLease(client)
	JobMgrSingle = &JobMgr{
		Client: client,
		KV:     kv,
		Lease:  lease,
	}
	return
}

//保存任务
func (jobMgr *JobMgr) SaveJob(job *common.Job) (oldJob *common.Job, err error) {
	var (
		jobValue []byte
		putRes   *clientv3.PutResponse
	)
	//拼接任务名称
	jobKey := common.JobSaveDir + job.Name
	//任务信息json化
	if jobValue, err = json.Marshal(job); err != nil {
		return
	}
	if putRes, err = jobMgr.KV.Put(context.TODO(), jobKey, string(jobValue), clientv3.WithPrevKV()); err != nil {
		return
	}
	//如果更新,返回旧值
	if putRes.PrevKv != nil {
		if err = json.Unmarshal(putRes.PrevKv.Value, &oldJob); err != nil {
			err = nil
			return
		}
	}
	return oldJob, nil
}

//删除任务
func (jobMgr *JobMgr) DeleteJob(jobName string) (oldJob *common.Job, err error) {
	var (
		oldJobObj common.Job
		delRes    *clientv3.DeleteResponse
	)
	//拼接任务名称
	jobKey := common.JobSaveDir + jobName
	//从etcd删除
	if delRes, err = jobMgr.KV.Delete(context.TODO(), jobKey, clientv3.WithPrevKV()); err != nil {
		return
	}
	//返回删除的任务信息
	if len(delRes.PrevKvs) != 0 {
		if err = json.Unmarshal(delRes.PrevKvs[0].Value, &oldJobObj); err != nil {
			err = nil
			return
		}
		oldJob = &oldJobObj
	}
	return
}

//列举所有任务
func (jobMgr *JobMgr) ListJobs() (jobList []*common.Job, err error) {
	var (
		dirKey  string
		getResp *clientv3.GetResponse
		kvPair  *mvccpb.KeyValue
		job     *common.Job
	)
	dirKey = common.JobSaveDir
	if getResp, err = jobMgr.KV.Get(context.TODO(), dirKey, clientv3.WithPrefix()); err != nil {
		return
	}
	jobList = make([]*common.Job, 0)
	//遍历,反序列化
	for _, kvPair = range getResp.Kvs {
		job = &common.Job{}
		if err = json.Unmarshal(kvPair.Value, job); err != nil {
			err = nil
			continue
		}
		jobList = append(jobList, job)
	}
	return
}

//杀死任务
func (jobMgr *JobMgr) KillJob(jobName string) (err error) {
	//更新一下key=/cron/killer/任务名
	var (
		killerKey      string
		leaseGrantResp *clientv3.LeaseGrantResponse
	)
	//拼装被杀死任务的key
	killerKey = common.JobKillDir + jobName

	//让worker监听到一次put操作,同时将这个put操作绑定一个租约,过期自动销毁
	if leaseGrantResp, err = jobMgr.Lease.Grant(context.TODO(), common.OpKillExpired); err != nil {
		return
	}
	//设置kill标记
	if _, err = jobMgr.KV.Put(context.TODO(), killerKey, "", clientv3.WithLease(leaseGrantResp.ID)); err != nil {
		return
	}
	return
}
