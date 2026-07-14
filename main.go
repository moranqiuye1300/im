package main

import (
	router "IM/routers"
	"IM/utils"
)

func main() {
	utils.InitConfig()
	utils.InitMySQL()
	r := router.Router()
	r.Run(":8080")
}
