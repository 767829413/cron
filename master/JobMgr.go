package master

import (
	"context"
	"cron/common"
	"encoding/json"
	"go.etcd.io/etcd/clientv3"
	"time"
)

var(
	jobKey string = "/cron/jobs/"
	jobValue []byte
)

type JobMgr struct {
	Client *clientv3.Client
	KV clientv3.KV
	Lease clientv3.Lease
}

func NewJobMgr(conf *Config)(mgr *JobMgr,err error){
	var(
		config clientv3.Config
		client *clientv3.Client
		kv clientv3.KV
		lease clientv3.Lease
	)
	//初始化配置
	config = clientv3.Config{
		Endpoints:conf.EtcdEndPoints,//集群地址
		DialTimeout:time.Duration(conf.EtcdDialTimeout)*time.Millisecond,//超时时间
	}

	//建立链接
	if client,err = clientv3.New(config);err != nil{
		return nil,err
	}
	//得到KV和Lease的API子集
	kv = clientv3.NewKV(client)
	lease = clientv3.NewLease(client)
	return &JobMgr{
		Client:client,
		KV:kv,
		Lease:lease,
	},nil
}

//保存任务
func (jobMgr *JobMgr)SaveJob(job *common.Job)(oldJob *common.Job,err error){
	var (
		putRes *clientv3.PutResponse
	)
	//把任务保存到/cron/jobs/任务名 -> json
	jobKey = jobKey + job.Name
	//任务信息json化
	if jobValue,err = json.Marshal(job);err != nil{
		return
	}
	if putRes,err = jobMgr.KV.Put(context.TODO(),jobKey,string(jobValue),clientv3.WithPrevKV());err != nil {
		return
	}
	//如果更新,返回旧值
	if putRes.PrevKv != nil {
		if err = json.Unmarshal(putRes.PrevKv.Value,&oldJob);err != nil {
			err = nil
			return
		}
	}
	return oldJob,nil

}