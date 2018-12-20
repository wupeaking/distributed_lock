package distributedlock

import (
	"fmt"
)

// DistributedLock 定义实现分布式锁的接口
type DistributedLock interface {
	// 等待获取锁
	Lock() (bool, error)
	// 尝试获取锁 立刻返回
	TryLock() (bool, error)
	UnLock() (bool, error)
	TryUnLock() (bool, error)
}

//OptionsFn 其他参数配置
type OptionsFn func(interface{})

type newLockFunc func(configure map[string]interface{}, opts OptionsFn) (DistributedLock, error)

var allDistributedPlugin = make(map[string]newLockFunc)

// RegistDistributedLock 注册已经实现的分布式锁
func RegistDistributedLock(name string, fn newLockFunc) {
	allDistributedPlugin[name] = fn
}

// CreateDistributedLock 创建分布式锁
func CreateDistributedLock(lock string, configure map[string]interface{}, opts OptionsFn) (DistributedLock, error) {
	newLockFn, ok := allDistributedPlugin[lock]
	if !ok {
		return nil, fmt.Errorf("未找到基于此后台的分布式锁应用")
	}
	return newLockFn(configure, opts)

}
