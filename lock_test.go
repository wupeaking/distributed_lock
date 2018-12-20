package distributedlock_test

import (
	dl "github.com/wupeaking/distributed_lock"
	_ "github.com/wupeaking/distributed_lock/plugin/mysql"
	"log"
	"testing"
	"time"
)

func TestMysqlLock(t *testing.T) {
	configure := make(map[string]interface{})
	configure["db_addr"] = "192.168.9.148:3306"
	configure["db_user"] = "tiger"
	configure["db_passwd"] = "tigerisnotcat"
	configure["db_name"] = "blockchain_eth"
	configure["process_id"] = 1000

	lock, err := dl.CreateDistributedLock("mysql", configure, nil)
	if err != nil {
		t.Fatalf("创建mysql分布式锁失败, %s", err.Error())
	}

	// 准备获取锁
	_, err = lock.Lock()
	if err != nil {
		t.Fatalf("获取锁出错, %s", err.Error())
	}
	log.Printf("process_id: %d 成功获取到锁资源", 1000)
	// 做一些运算
	log.Println("获取锁之后的一些运算")
	time.Sleep(time.Duration(30) * time.Second)

	// 释放锁
	_, err = lock.UnLock()
	if err != nil {
		t.Fatalf("释放锁出错, %s", err.Error())
	}

	log.Printf("测试完成")
}
