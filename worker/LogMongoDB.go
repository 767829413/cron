package worker

import (
	"context"
	"cron/common"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"time"
)

//Mongodb存储
type LogMongoDB struct {
	client         *mongo.Client
	logCollection  *mongo.Collection
	LogChan        chan *common.JobLog
	autoCommitChan chan *common.LogBatch
}

var (
	LogMongoSingle *LogMongoDB
)

func NewLogMongoDB() (err error) {
	var (
		client        *mongo.Client
		clientOptions *options.ClientOptions
	)
	//建立连接
	clientOptions = options.Client().ApplyURI(ConfSingle.MongoUri).SetConnectTimeout(time.Millisecond * time.Duration(ConfSingle.MongoConnectTimeout))
	if client, err = mongo.Connect(context.TODO(), clientOptions); err != nil {
		return
	}
	//选择db和collection
	LogMongoSingle = &LogMongoDB{
		client:         client,
		logCollection:  client.Database(common.LogDbName).Collection(common.LogCollectionName),
		LogChan:        make(chan *common.JobLog, 1000),
		autoCommitChan: make(chan *common.LogBatch, 1000),
	}

	//启动一个gorounte来消费
	go LogMongoSingle.writeLoop()
	return
}

func (logMongo *LogMongoDB) Join(jobLog *common.JobLog) {
	logMongo.LogChan <- jobLog
}

func (logMongo *LogMongoDB) writeLoop() {
	var (
		jobLog       *common.JobLog
		logBatch     *common.LogBatch
		commitTimer  *time.Timer
		timeoutBatch *common.LogBatch
	)
	for {
		select {
		//每次插入需要经过mongodb请求往返,比较耗时
		case jobLog = <-logMongo.LogChan:
			if logBatch == nil {
				logBatch = &common.LogBatch{}

				//让这个批次超时自动提交
				commitTimer = time.AfterFunc(time.Duration(common.LogTimeOutTime)*time.Millisecond, func(batch *common.LogBatch) func() {
					return func() {
						//发出超时通知
						logMongo.autoCommitChan <- logBatch
					}
				}(logBatch))
			}
			logBatch.Logs = append(logBatch.Logs, jobLog)
			if len(logBatch.Logs) > common.LogInsertNum {
				logMongo.save(logBatch)
				logBatch = nil
				commitTimer.Stop()
			}
		case timeoutBatch = <-logMongo.autoCommitChan:
			if timeoutBatch == logBatch {
				continue
			}
			logMongo.save(timeoutBatch)
			logBatch = nil
		}
	}
}

func (logMongo *LogMongoDB) save(logBatch *common.LogBatch) {
	log.Println(logBatch.Logs)
	//入库
	_, _ = logMongo.logCollection.InsertMany(context.TODO(), logBatch.Logs)
}
