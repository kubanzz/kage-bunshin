package db

import (
	"context"

	"github.com/beego/beego/v2/core/logs"
	beego "github.com/beego/beego/v2/server/web"
	"github.com/go-redis/redis/v8"
)

var RedisClient *redis.Client

func init() {

	redisURI, _ := beego.AppConfig.String("redis_uri")
	redisPassword, _ := beego.AppConfig.String("redis_password")
	redisDb, _ := beego.AppConfig.Int("redis_db")

	RedisClient = redis.NewClient(&redis.Options{
		Addr:     redisURI,      // Redis 服务器地址
		Password: redisPassword, // Redis 服务器密码
		DB:       redisDb,       // Redis 数据库索引
	})

	// 检查连接是否成功
	_, err := RedisClient.Ping(context.Background()).Result()
	if err != nil {
		logs.Error("创建Redis失败", err)
	}

	logs.Info("Redis连接成功")
}
