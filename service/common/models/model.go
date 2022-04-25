package models

import (
	"time"

	"github.com/go-redis/redis"
	"github.com/go-xorm/xorm"
	. "mdm/common/qm"
	"xorm.io/core"
)

type CDBConfigData struct {
	DriverName           string `toml:"driver_name"`
	SourceName           string `toml:"source_name"`
	SourceNameDB         string `toml:"source_name_db"`
	ConnKeepAliveMinutes int    `toml:"conn_keep_alive_minutes"`
	MaxOpenConns         int    `toml:"max_open_conns"`
	MaxIdleConns         int    `toml:"max_idle_conns"`
	SourceStatisticsDB   string `toml:"source_statistics_db"`
	DnsSourceName        string `toml:"dns_source_name"`
	RedisIPPort          string `toml:"redis_ipport"`
	RedisPassword        string `toml:"redis_password"`
}

var RedisClient *redis.Client

// 初始化redis
func InitRedis(addr, password string, db int) (*redis.Client, error) {
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})
	_, err := RedisClient.Ping().Result()
	if err != nil {
		return nil, err
	}
	return RedisClient, err
}

const (
	GRedisDefaultDb = 0
	GRedisNotifyDb  = 1
	GRedisNgnDb     = 2
)

//var GRedisClient *redis.Client
//
//// 初始化redis哨兵连接
//func InitSentinelRedis(db int) error {
//	// TODO 这里加载哨兵要替换
//	//GRedisClient = GetRedisSentinelClient(db)
//	GRedisClient = GetRedisSentinelClientEx(db)
//	RedisClient = GetRedisSentinelClientEx(db)
//	var err error
//	if GRedisClient == nil || RedisClient == nil {
//		err = errors.New("GetRedisSentinelClient  error")
//	}
//	return err
//}
//
//func InitDefaultRedisClient(db int) error {
//	var config CRedisConfig
//	if !GetConfigStruct(&config) {
//		LOG_ERROR("get conf failed")
//		return nil
//	}
//	GRedisClient = redis.NewClient(&redis.Options{
//		Addr:     config.Data.IPPort,
//		Password: config.Data.Password,
//		DB:       db,
//	})
//	_, err := GRedisClient.Ping().Result()
//	if err != nil {
//		LOG_ERROR_F("init redis failed:%s, ", err.Error())
//		return err
//	}
//	return err
//}

const (
	maxIdleConns      = 100
	maxOpenConns      = 100
	DefaultLimitCount = 10
)

var (
	ExcelAuthUserImportFileName string
	ExcelAuthUserExportFileName string
	ExcelAuthUserSheelName      string
)

var (
	Db *xorm.Engine
)

func InitDB(psqlInfo string) error {
	var err error
	Db, err = xorm.NewEngine("postgres", psqlInfo)
	if err != nil {
		return err
	}

	if err = Db.Ping(); err != nil {
		return err
	}
	Db.SetMaxIdleConns(maxIdleConns)
	Db.SetMaxOpenConns(maxOpenConns)
	Db.SetLogLevel(core.LOG_INFO)
	Db.SetConnMaxLifetime(time.Minute * 5)
	//Db.ShowSQL(true)

	go keepDbAlived(Db)
	return err
}

func keepDbAlived(engine *xorm.Engine) {
	t := time.Tick(180 * time.Second)
	for {
		<-t
		engine.Ping()
	}
}

func DbStop() {
	err := Db.Close()
	if err != nil {
		LOG_ERROR(err)
	}
}

func RedisStop() {
	err := RedisClient.Close()
	if err != nil {
		LOG_ERROR(err)
	}
}

func Stop() {
	DbStop()
	RedisStop()
}

func GetDBEngine(psqlInfo string) (Db *xorm.Engine, err error) {
	Db, err = xorm.NewEngine("postgres", psqlInfo)
	if err != nil {
		LOG_ERROR_F("connect db failed:%s, ", err.Error())
		return
	}

	if err = Db.Ping(); err != nil {
		LOG_ERROR_F("ping db failed:%s, ", err.Error())
		return
	}
	Db.SetMaxIdleConns(maxIdleConns)
	Db.SetMaxOpenConns(maxOpenConns)
	Db.SetConnMaxLifetime(time.Minute * 5)

	go keepDbAlived(Db)
	return
}

type GinLogger int

func (ginLogger *GinLogger) Write(p []byte) (n int, err error) {
	LOG_INFO(string(p))
	return len(p), nil
}
