package common

import "errors"

var (
	ERR_LOCK_ALREADY_REQURED = errors.New("锁已被占用")
	ERR_NO_LOCAL_IP_FOUNC    = errors.New("没有找到网卡ip")
)
