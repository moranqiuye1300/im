package main

import (
	router "IM/routers"
	"IM/service"
	"IM/utils"
	"context"
	"log"
)

func main() {
	utils.InitConfig()
	utils.InitMySQL()
	utils.InitRedis()
	go func() {
		ctx := context.Background()
		err := utils.ListenPatternChannel(ctx, "websocket", service.Broadcast)
		if err != nil {
			log.Println("redis订阅失败：", err)
		}
	}()
	r := router.Router()
	r.Run(":8080")
}
