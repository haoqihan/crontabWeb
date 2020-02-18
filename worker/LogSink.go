package worker

import (
	"context"
	"cronWeb/common"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

// mongoDB 存储日志
type LogSink struct {
	client         *mongo.Client
	logCollection  *mongo.Collection
	logChan        chan *common.JobLog
	autoCommitChan chan *common.LogBatch
}

var (
	G_logSink *LogSink
)

func InitLogSinK() (err error) {
	var (
		client *mongo.Client
	)
	// 设置客户端连接配置
	clientOptions := options.Client().ApplyURI(G_config.MongoDBUrl)
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(G_config.MongodbConnectTimeOut)*time.Millisecond)
	if client, err = mongo.Connect(ctx, clientOptions); err != nil {
		return
	}
	// 选择db和collection
	G_logSink = &LogSink{
		client:         client,
		logCollection:  client.Database("cron").Collection("log"),
		logChan:        make(chan *common.JobLog, 1000),
		autoCommitChan: make(chan *common.LogBatch, 1000),
	}
	go G_logSink.writeLoop()

	return
}

// 批量写入日志
func (logSink *LogSink) saveLogs(batch *common.LogBatch) {
	logSink.logCollection.InsertMany(context.TODO(), batch.Logs)
}

func (logSink *LogSink) writeLoop() {
	var (
		log          *common.JobLog
		logBatch     *common.LogBatch // 当前的批次
		commitTimer  *time.Timer
		timeoutBatch *common.LogBatch // 超时批次
	)
	for {
		select {
		case log = <-logSink.logChan:
			// 将log写入到mongoDB中
			if logBatch == nil {
				logBatch = &common.LogBatch{}
				commitTimer = time.AfterFunc(time.Duration(G_config.JobLogCommitTimeout)*time.Millisecond, func(batch *common.LogBatch) func() {
					// 发出超时通知，不要提交
					return func() {
						logSink.autoCommitChan <- batch
					}
				}(logBatch))
			}
			// 把新的日志追加到批次中
			logBatch.Logs = append(logBatch.Logs, log)
			if len(logBatch.Logs) >= G_config.JobLogBatchSize {
				logSink.saveLogs(logBatch)
				// 清空logBatch
				logBatch = nil
				// 取消定时器
				commitTimer.Stop()
			}
		case timeoutBatch = <-logSink.autoCommitChan: // 过期的批次
			// 判断过期批次是否是当前批次
			if timeoutBatch != logBatch {
				continue // 跳过已经提交的批次
			}
			// 把这个批次写入到mongoDB中
			logSink.saveLogs(timeoutBatch)
			logBatch = nil

		}
	}

}

// 发送日志的API
func (logSink *LogSink) Append(jobLog *common.JobLog) {
	select {
	case logSink.logChan <- jobLog:
	default:
		// 队列满了就丢弃
	}
}
