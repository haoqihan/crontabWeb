package worker

import (
	"encoding/json"
	"io/ioutil"
)

var (
	G_config *Config
)

// 程序配置
type Config struct {
	EtcdEndpoints         []string `json:"etcdEndpoints"`
	EtcdDiaTimeout        int      `json:"etcdDiaTimeout"`
	MongoDBUrl            string   `json:"mongoDBUrl"`
	MongodbConnectTimeOut int      `json:"mongodbConnectTimeOut"`
	JobLogBatchSize       int      `json:"jobLogBatchSize"`
	JobLogCommitTimeout   int      `json:"jobLogCommitTimeout"`
}

func InitConfig(filename string) (err error) {

	var (
		content []byte
		conf    Config
	)
	// 读取配置文件
	if content, err = ioutil.ReadFile(filename); err != nil {
		return
	}
	// json反序列化
	if err = json.Unmarshal(content, &conf); err != nil {
		return
	}
	// 赋值单例
	G_config = &conf
	return
}
