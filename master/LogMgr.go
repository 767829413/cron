package master

import (
	"context"
	"cron/common"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

// mongodb日志管理
type LogMgr struct {
	client        *mongo.Client
	logCollection *mongo.Collection
}

var (
	LogMgrSingle *LogMgr
)

func NewLogMgr() (err error) {
	var (
		client        *mongo.Client
		clientOptions *options.ClientOptions
	)

	// 建立mongodb连接
	clientOptions = options.Client().ApplyURI(ConfigSingle.MongoUri).SetConnectTimeout(time.Millisecond * time.Duration(ConfigSingle.MongoConnectTimeout))
	if client, err = mongo.Connect(context.TODO(), clientOptions); err != nil {
		return
	}
	LogMgrSingle = &LogMgr{
		client:        client,
		logCollection: client.Database(common.LogDbName).Collection(common.LogCollectionName),
	}
	return
}

// 查看任务日志
func (logMgr *LogMgr) ListLog(name string, skip int, limit int) (logArr []*common.JobViewLog, err error) {
	var (
		filter      *common.JobLogFilter
		logSort     *common.SortLogByStartTime
		cursor      *mongo.Cursor
		jobLog      *common.JobViewLog
		findOptions *options.FindOptions
	)

	// len(logArr)
	logArr = make([]*common.JobViewLog, 0)

	// 过滤条件
	filter = &common.JobLogFilter{JobName: name}

	// 按照任务开始时间倒排
	logSort = &common.SortLogByStartTime{SortOrder: -1}

	//构建查询参数
	findOptions = &options.FindOptions{}
	findOptions.SetSkip(int64(skip))
	findOptions.SetSort(logSort)
	findOptions.SetLimit(int64(limit))
	// 查询
	if cursor, err = logMgr.logCollection.Find(context.TODO(), filter, findOptions); err != nil {
		return
	}
	// 延迟释放游标
	defer cursorClose(cursor)

	for cursor.Next(context.TODO()) {
		jobLog = &common.JobViewLog{}

		// 反序列化BSON
		if err = cursor.Decode(jobLog); err != nil {
			continue // 有日志不合法
		}
		logArr = append(logArr, jobLog)
	}
	return
}

func cursorClose(cursor *mongo.Cursor) {
	_ = cursor.Close(context.TODO())
}
