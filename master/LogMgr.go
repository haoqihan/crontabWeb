package master

import (
	"context"
	"cronWeb/common"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

// mongo日志管理
type LogMgr struct {
	client        *mongo.Client
	logCollection *mongo.Collection
}

var (
	G_logMgr *LogMgr
)

func InitLogMgr() (err error) {
	var (
		client *mongo.Client
	)
	// 设置客户端连接配置
	clientOptions := options.Client().ApplyURI(G_config.MongoDBUrl)
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(G_config.MongodbConnectTimeOut)*time.Millisecond)
	if client, err = mongo.Connect(ctx, clientOptions); err != nil {
		return
	}
	G_logMgr = &LogMgr{
		client:        client,
		logCollection: client.Database("cron").Collection("log"),
	}

	return
}

// 查看任务日志
func (logMgr *LogMgr) ListMgr(name string, skip, limit int) (logArray []*common.JobLog, err error) {
	var (
		filter  *common.JobNameFilter
		logSort *common.SortLogByStartTime
		cursor  *mongo.Cursor
		jobLog  *common.JobLog
	)

	logArray = make([]*common.JobLog, 0)

	// 过滤条件
	filter = &common.JobNameFilter{JobName: name}
	// 按照任务开始时间倒序排列
	logSort = &common.SortLogByStartTime{SortOrder: -1}
	findOptions := options.Find()
	findOptions.SetSort(logSort)
	findOptions.SetSkip(int64(skip))
	findOptions.SetLimit(int64(limit))
	// 查询
	if cursor, err = logMgr.logCollection.Find(context.TODO(), filter, findOptions); err != nil {
		return
	}
	// 延迟释放游标
	defer cursor.Close(context.TODO())

	for cursor.Next(context.TODO()) {
		jobLog = &common.JobLog{}
		// 反序列化BSON
		if err = cursor.Decode(jobLog); err != nil {
			continue
		}
		logArray = append(logArray, jobLog)
	}
	return

}

//1581946508132
//1581946512892

// 1581946508132 2020-02-17 21:35:08
// 1581946512892 2020-02-17 21:35:12
