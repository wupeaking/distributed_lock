package mysqlplugin

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestMysqlLock(t *testing.T) {
	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		defer wg.Done()
		configure := make(map[string]interface{})
		configure["db_addr"] = "192.168.9.148:3306"
		configure["db_user"] = "tiger"
		configure["db_passwd"] = "tigerisnotcat"
		configure["db_name"] = "blockchain_eth"
		configure["process_id"] = 1000

		lock, _ := NewMySQLLock(configure, nil)
		time.Sleep(time.Duration(2) * time.Second)
		suc, err := lock.Lock()
		if err != nil {
			t.Fatalf(err.Error())
		}
		if !suc {
			t.Fatalf("获取锁失败")
		}

		fmt.Printf("goutineid: %d, 成功获取到🔐资源", 1000)

		ok, err := lock.UnLock()
		if err != nil {
			t.Fatalf(err.Error())
		}
		if !ok {
			t.Fatalf("释放锁失败")
		} else {
			t.Logf("goutineid: %d, 释放锁成功", 1000)
		}
	}()

	go func() {
		configure := make(map[string]interface{})
		configure["db_addr"] = "192.168.9.148:3306"
		configure["db_user"] = "tiger"
		configure["db_passwd"] = "tigerisnotcat"
		configure["db_name"] = "blockchain_eth"
		configure["process_id"] = 2000
		defer wg.Done()

		lock, _ := NewMySQLLock(configure, nil)

		suc, err := lock.Lock()
		if err != nil {
			t.Fatalf(err.Error())
		}
		if !suc {
			t.Fatalf("获取锁失败")
		}

		t.Logf("goutineid: %d, 成功获取到🔐资源", 2000)

		ok, err := lock.UnLock()
		if err != nil {
			t.Fatalf(err.Error())
		}
		if !ok {
			t.Fatalf("释放锁失败")
		} else {
			t.Logf("goutineid: %d, 释放锁成功", 2000)
		}
	}()

	wg.Wait()
}
