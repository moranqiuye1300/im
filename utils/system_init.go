package utils

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB
var RedisClient *redis.Client

func InitConfig() {
	viper.SetConfigFile("config/app.yaml")
	err := viper.ReadInConfig()
	if err != nil {
		fmt.Printf("读取配置文件失败，详细错误：%v", err)
	}
	log.Println("读取配置文件成功")
}
func InitMySQL() {
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{ // 日志配置
			SlowThreshold: time.Second, // 慢 SQL 阈值
			LogLevel:      logger.Info, // 日志级别
			Colorful:      true,        // 禁用彩色打印
		},
	)
	sql, err := gorm.Open(mysql.Open(viper.GetString("mysql.dsn")), &gorm.Config{Logger: newLogger})
	if err != nil {
		log.Fatalf("数据库连接失败，详细错误：%v", err)
	}
	DB = sql
}
func InitRedis() {
	RedisClient = redis.NewClient(&redis.Options{
		Addr:         viper.GetString("redis.addr"),
		Password:     viper.GetString("redis.password"),
		DB:           viper.GetInt("redis.db"),
		PoolSize:     viper.GetInt("redis.poolSize"),
		MinIdleConns: viper.GetInt("redis.minIdleConns"),
	})
	ctx := context.Background()
	pong, err := RedisClient.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Redis连接失败，详细错误：%v", err)
	} else {
		log.Println("Redis连接成功，Ping返回：", pong)
	}
}

const (
	PublishKey = "websocket"
)

func Publish(ctx context.Context, channel, message string) error {
	var err error
	err = RedisClient.Publish(ctx, channel, message).Err()
	if err != nil {
		log.Printf("Redis发布消息失败，详细错误：%v", err)
	}
	return err
}

// ListenPatternChannel 持续模式订阅，收到消息执行回调
func ListenPatternChannel(ctx context.Context, pattern string, callback func(string)) error {
	pubsub := RedisClient.PSubscribe(ctx, pattern)
	defer pubsub.Close()

	for {
		msg, err := pubsub.ReceiveMessage(ctx)
		if err != nil {
			log.Printf("订阅接收消息异常: %v", err)
			return err
		}
		// 拿到消息payload执行业务逻辑（推送给websocket客户端）
		callback(msg.Payload)
	}
}
