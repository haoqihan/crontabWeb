package master

import (
	"encoding/json"
	"io/ioutil"
)

var (
	G_config *Config
)

// 程序配置
type Config struct {
	ApiPort               int      `json:"apiPort"`
	ApiReadTimeout        int      `json:"apiReadTimeOut"`
	ApiWriteTimeout       int      `json:"apiWriteTimeOut"`
	EtcdEndpoints         []string `json:"etcdEndpoints"`
	EtcdDiaTimeout        int      `json:"etcdDiaTimeout"`
	WebRoot               string   `json:"webroot"`
	MongoDBUrl            string   `json:"mongoDBUrl"`
	MongodbConnectTimeOut int      `json:"mongodbConnectTimeOut"`
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
