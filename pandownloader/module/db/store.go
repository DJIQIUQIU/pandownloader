package db

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
	_ "github.com/go-sql-driver/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"git-sec.com/pandownloader/logger"
	"git-sec.com/pandownloader/module/utils"
)

// mysql 全局变量
var Db *gorm.DB

/**
 * 初始化配置文件
 */
func init() {
	err := utils.GetConfigIni("./config/config.ini")
	if err != nil {
		logger.GetLogger().Print("get config error", err)
	}
}

/**
 * 初始化mysql
 */
func init() {
	var err error
	host, err := utils.Cfg.GetValue("mysql", "host")
	port, err := utils.Cfg.GetValue("mysql", "port")
	username, err := utils.Cfg.GetValue("mysql", "username")
	password, err := utils.Cfg.GetValue("mysql", "password")
	database, err := utils.Cfg.GetValue("mysql", "database")
	dataSource := username + ":" + password + "@tcp(" + host + ":" + port + ")/" + database
	Db, err = gorm.Open(mysql.Open(dataSource), &gorm.Config{})
	if err != nil {
		logger.GetLogger().Print("err:", err.Error())
	}
}

/**
 * 初始化redis
 */
var (
	RedisClient        *redis.Client
	RedisClusterClient *redis.ClusterClient
	Ctx                = context.Background()
	RedisModel         string
)

func init() {
	RedisModel, _ = utils.Cfg.GetValue("main", "redis_model")
	if RedisModel == "pool" {
		host, _ := utils.Cfg.GetValue("redis", "host")
		port, _ := utils.Cfg.GetValue("redis", "port")
		RedisClient = redis.NewClient(&redis.Options{
			Addr:         host + ":" + port,
			Password:     "", // no password set
			DB:           0,  // use default DB
			PoolSize:     10,
			MaxRetries:   3,
			IdleTimeout:  30 * time.Second,
			MinIdleConns: 30,
            DialTimeout:  5000 * time.Microsecond, // 设置连接超时
		})
	} else {
		// 连接redis集群
		RedisClusterClient = redis.NewClusterClient(&redis.ClusterOptions{
			Addrs: []string{ // 填写master主机
				"192.168.21.22:30001",
				"192.168.21.22:30002",
				"192.168.21.22:30003",
			},
			Password:     "123456",              // 设置密码
			DialTimeout:  50 * time.Microsecond, // 设置连接超时
			ReadTimeout:  50 * time.Microsecond, // 设置读取超时
			WriteTimeout: 50 * time.Microsecond, // 设置写入超时
		})
	}
}

// redis get
func Get(key string) *redis.StringCmd {
	if RedisModel == "pool" {
		s := RedisClient.Get(Ctx, key)
		return s
	} else {
		s := RedisClusterClient.Get(Ctx, key)
		return s
	}
}

// redis set
func Set(key string, value interface{}) {
	if RedisModel == "pool" {
		RedisClient.Set(Ctx, key, value, 0)
	} else {
		RedisClusterClient.Set(Ctx, key, value, 0)
	}
}

// redis HGetAll
func HGetAll(key string) *redis.StringStringMapCmd {
	if RedisModel == "pool" {
		s := RedisClient.HGetAll(Ctx, key)
		return s
	} else {
		s := RedisClusterClient.HGetAll(Ctx, key)
		return s
	}
}

// redis HGet
func HGet(key string, field string) *redis.StringCmd {
	if RedisModel == "pool" {
		s := RedisClient.HGet(Ctx, key, field)
		return s
	} else {
		s := RedisClusterClient.HGet(Ctx, key, field)
		return s
	}
}

// redis HSet
func HSet(key string, value ...interface{}) {
	if RedisModel == "pool" {
		RedisClient.HSet(Ctx, key, value)
	} else {
		RedisClusterClient.HSet(Ctx, key, value)
	}
}

// redis del
func Del(key string) {
	if RedisModel == "pool" {
		RedisClient.Del(Ctx, key)
	} else {
		RedisClusterClient.Del(Ctx, key)
	}
}

// Publish
func Publish(topic string, message string) error {
	return RedisClient.Publish(Ctx, topic, message).Err()
}

// Subscribe
type Callback func([]byte) error

func Subscribe(topic string, handle Callback) error {
	pubsub := RedisClient.Subscribe(Ctx, topic)
	_, err := pubsub.Receive(Ctx)
	if err != nil {
		return err
	}

	ch := pubsub.Channel()
	for msg := range ch {
		logger.GetLogger().Printf(
			"[Channel:%s] Receive `%s`: %s ",
			topic, msg.Channel, msg.Payload)
		switch msg.Channel {
		case topic:
			err = handle([]byte(msg.Payload))
			if err != nil {
				logger.GetLogger().Printf(
					"[Channel:%s] Handle Error: %s %v",
					topic, msg.Payload, err)
			}
		default:
			logger.GetLogger().Printf("[Channel:%s] Cannot handle `%s`", topic, msg.Channel)
		}
		// logger.GetLogger().Printf("Receive `%s`: %s ", msg.Channel, msg.Payload)
		// if msg.Channel != topic { continue }
		// logger.GetLogger().Printf("%s %s", topic, msg.Payload)
		// sendMail([]byte(msg.Payload))
	}
	return nil
}
