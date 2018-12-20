package mysqlplugin

import (
	"fmt"
	"github.com/go-sql-driver/mysql" // 初始化mysql驱动
	"github.com/jmoiron/sqlx"
	dlock "github.com/wupeaking/distributed_lock"
	"os"
	//"sync/atomic"
	"time"
	//"unsafe"
)

// 一个mysql实现的分布式锁
// 需要在mysql端执行的SQL语句
/*
创建分布式表distributed_lock
CREATE TABLE `distributed_lock` (
  `lock_id` smallint(5) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '时间戳',
  `process_id` bigint(20) unsigned NOT NULL,
  PRIMARY KEY (`lock_id`)
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8

创建存储过程和定时任务

delimiter //
create PROCEDURE process_distributed_lock_timeout120()
BEGIN
	delete from distributed_lock where TIMESTAMP(create_time) < NOW()-120;
end //
delimiter ;


create event if not exists event_process_distributed_lock on schedule every 30 second on completion preserve enable do call process_distributed_lock_timeout120();

set global event_scheduler = ON;

show VARIABLES like '%event%';

*/

// 此插件使用了sqlx的包

// MySQLLock mysql锁对象
type MySQLLock struct {
	db           *sqlx.DB
	pollInterval int
	lockID       int
}

func init() {
	dlock.RegistDistributedLock("mysql", NewMySQLLock)
}

//NewMySQLLock 创建mysql分布式锁
func NewMySQLLock(configure map[string]interface{}, opts dlock.OptionsFn) (dlock.DistributedLock, error) {
	sqlLock := new(MySQLLock)
	// 如果直接传递SQL连接 则直接使用此连接池
	if obj, ok := configure["db_connection"]; ok {
		sqlLock.db = obj.(*sqlx.DB)
	} else {
		db, err := creatMySQLConnection(configure)
		if err != nil {
			return nil, err
		}
		sqlLock.db = db

	}
	if pollInterval, ok := configure["db_interval"]; ok {
		sqlLock.pollInterval = pollInterval.(int)
	} else {
		sqlLock.pollInterval = 10
	}
	if id, ok := configure["process_id"]; ok {
		sqlLock.lockID = id.(int)
	} else {
		sqlLock.lockID = os.Getpid()
	}
	if opts != nil {
		opts(sqlLock)
	}
	return sqlLock, nil
}

func creatMySQLConnection(configure map[string]interface{}) (*sqlx.DB, error) {

	dbUser, ok := configure["db_user"]
	if !ok {
		return nil, fmt.Errorf("缺少配置项db_user")
	}
	dbPasswd, ok := configure["db_passwd"]
	if !ok {
		return nil, fmt.Errorf("缺少配置项db_passwd")
	}
	dbAddr, ok := configure["db_addr"]
	if !ok {
		return nil, fmt.Errorf("缺少配置项db_addr")
	}
	dbName, ok := configure["db_name"]
	if !ok {
		return nil, fmt.Errorf("缺少配置项db_name")
	}

	dbURI := fmt.Sprintf(
		"%s:%s@tcp(%s)/%s?parseTime=true&charset=utf8",
		dbUser.(string),
		dbPasswd.(string),
		dbAddr.(string),
		dbName.(string))

	DB, err := sqlx.Open("mysql", dbURI)
	if err != nil {
		return nil, err
	}

	DB.SetConnMaxLifetime(time.Millisecond * time.Duration(1000))
	DB.SetMaxIdleConns(1)
	DB.SetMaxOpenConns(2)

	err = DB.Ping()
	if err != nil {
		return nil, err
	}
	return DB, nil
}

// 实现接口

// Lock 尝试获取锁资源
func (sqlLock *MySQLLock) Lock() (bool, error) {
	// if len(sqlLock.lockGet) != 0 {
	// 	return nil, fmt.Errorf("之前有未释放的锁资源")
	// }
	// p := unsafe.Pointer(&sqlLock.isRunning)
	// isRunning := *(*bool)(atomic.LoadPointer(&p))
	// if !isRunning {
	// 	go sqlLock.lockGoroutine()
	// }
	// return sqlLock.lockGet, nil

	for {
		suc, err := sqlLock.insertMysql()
		if err != nil {
			return false, err
		}
		if !suc {
			time.Sleep(time.Duration(sqlLock.pollInterval) * time.Second)
			continue
		}
		return suc, nil
	}
}

// UnLock 解锁
func (sqlLock *MySQLLock) UnLock() (bool, error) {
	for {
		suc, err := sqlLock.deleteMysql()
		if err != nil {
			return false, err
		}
		if !suc {
			time.Sleep(time.Duration(sqlLock.pollInterval) * time.Second)
			continue
		}
		return suc, nil
	}
}

// TryLock 尝试获取锁资源
func (sqlLock *MySQLLock) TryLock() (bool, error) {
	suc, err := sqlLock.insertMysql()
	return suc, err

}

// TryUnLock 解锁
func (sqlLock *MySQLLock) TryUnLock() (bool, error) {
	suc, err := sqlLock.deleteMysql()
	return suc, err
}

func (sqlLock *MySQLLock) insertMysql() (bool, error) {
	sql := `insert into distributed_lock (lock_id, process_id) values (10000, ?);`
	_, err := sqlLock.db.Exec(sql, sqlLock.lockID)
	if err != nil {
		me, _ := err.(*mysql.MySQLError)
		if me.Number == 1062 {
			return false, nil
		}
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (sqlLock *MySQLLock) deleteMysql() (bool, error) {
	sql := `delete from distributed_lock where lock_id=10000 and process_id = ?;`
	_, err := sqlLock.db.Exec(sql, sqlLock.lockID)
	if err != nil {
		return false, err
	}
	return true, nil
}
