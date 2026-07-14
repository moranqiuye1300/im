package routers

import (
	"IM/middleware"
	"IM/service"
	"time"

	// 关键：导入自己项目生成的docs，module名/路径按你go.mod修改
	"IM/docs"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func Router() *gin.Engine {
	r := gin.Default()

	// ========== 修复代理信任警告 ==========
	// 本地开发仅信任本机，线上替换为你的Nginx/负载均衡网段
	_ = r.SetTrustedProxies([]string{"127.0.0.1"})
	// ========== 全局CORS跨域配置 ==========
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"}, // 开发环境允许所有域名，生产替换为前端域名
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// ========== Swagger 全局文档信息配置 ==========
	docs.SwaggerInfo.Title = "IM即时通讯服务API"
	docs.SwaggerInfo.Description = "用户、聊天相关接口文档"
	docs.SwaggerInfo.Version = "v1.0"
	docs.SwaggerInfo.Host = "127.0.0.1:8080" // 访问域名/端口
	docs.SwaggerInfo.BasePath = "/"
	docs.SwaggerInfo.Schemes = []string{"http"}

	// Swagger文档路由
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// ========== 公开路由（无需认证） ==========
	r.GET("/", service.GetIndex)
	r.POST("/user/login", service.Login)
	r.POST("/user/create", service.CreateUser)

	// ========== 需要JWT认证的路由 ==========
	auth := r.Group("/user")
	auth.Use(middleware.JWTAuth())
	{
		auth.GET("/list", service.GetUserList)
		auth.POST("/update", service.UpdateUser)
		auth.POST("/delete", service.DeleteUser)
		auth.GET("/find/name", service.FindUserByName)
		auth.GET("/find/phone", service.FindUserByPhone)
		auth.GET("/find/email", service.FindUserByEmail)
	}
	return r
}
