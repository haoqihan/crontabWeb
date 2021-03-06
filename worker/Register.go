package worker

import (
	"context"
	"cronWeb/common"
	"github.com/coreos/etcd/clientv3"
	"net"
	"time"
)

// 注册节点到etcd  /cron/workers/ip地址
type Register struct {
	client  *clientv3.Client
	kv      clientv3.KV
	lease   clientv3.Lease
	localIP string // 本机ip
}

var (
	G_register *Register
)

// 获取本机ip
func getLocalIP() (ipv4 string, err error) {
	var (
		addrs   []net.Addr
		addr    net.Addr
		ipNet   *net.IPNet
		isIpNet bool
	)
	// 获取所有网卡
	if addrs, err = net.InterfaceAddrs(); err != nil {
		return
	}
	// 取第一个非lo的网卡地址
	for _, addr = range addrs {
		// 这个网络地址是ip地址
		if ipNet, isIpNet = addr.(*net.IPNet); isIpNet && !ipNet.IP.IsLoopback() {
			// 跳过ipv6
			if ipNet.IP.To4() != nil {
				ipv4 = ipNet.IP.String()
				return
			}

		}
	}
	err = common.ERR_NO_LOCAL_IP_FOUNC

	return

}

// 注册到/cron/workers/ip ，并进行自动续租
func (register *Register) KeepOnline() {
	var (
		regKey         string
		leaseGrandResp *clientv3.LeaseGrantResponse
		err            error
		keepAliveChan  <-chan *clientv3.LeaseKeepAliveResponse
		keepAliveResp  *clientv3.LeaseKeepAliveResponse
		cancelCtx      context.Context
		cancelFunc     context.CancelFunc
	)
	for {
		// 注册路径
		regKey = common.JOB_WORKER_DIR + register.localIP
		cancelFunc = nil
		// 创建租约
		if leaseGrandResp, err = register.lease.Grant(context.TODO(), 10); err != nil {
			goto RETRY
		}
		// 自动续租
		if keepAliveChan, err = register.lease.KeepAlive(context.TODO(), leaseGrandResp.ID); err != nil {
			goto RETRY
		}
		cancelCtx, cancelFunc = context.WithCancel(context.TODO())
		// 注册到etcd
		if _, err = register.kv.Put(cancelCtx, regKey, "", clientv3.WithLease(leaseGrandResp.ID)); err != nil {
			goto RETRY
		}
		// 处理续租应答
		for {
			select {
			case keepAliveResp = <-keepAliveChan:
				if keepAliveResp == nil {
					// 续租失败
					goto RETRY
				}

			}
		}

	RETRY:
		time.Sleep(1 * time.Second)
		if cancelFunc != nil {
			cancelFunc()
		}
	}

}

func InitRegister() (err error) {
	var (
		config  clientv3.Config
		client  *clientv3.Client
		kv      clientv3.KV
		lease   clientv3.Lease
		localIP string
	)
	// 初始化配置
	config = clientv3.Config{
		Endpoints:   G_config.EtcdEndpoints,                                    // 集群地址
		DialTimeout: time.Duration(G_config.EtcdDiaTimeout) * time.Millisecond, // 超时时间
	}
	// 建立客户端
	if client, err = clientv3.New(config); err != nil {
		return
	}
	// 得到kv和lease的api子集
	kv = clientv3.NewKV(client)
	lease = clientv3.NewLease(client)

	// 获取本机ip
	if localIP, err = getLocalIP(); err != nil {
		return

	}
	G_register = &Register{
		client:  client,
		kv:      kv,
		lease:   lease,
		localIP: localIP,
	}
	// 注册
	go G_register.KeepOnline()
	return

}
