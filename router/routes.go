package router

import (
	"laporinchat/controllers/admin"
	admin_webhook "laporinchat/controllers/admin/webhook"
	"laporinchat/controllers/webhook"
	"laporinchat/middlewares"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(router *gin.Engine) {

	newrouter := router.Group("")
	newrouter.Use(middlewares.RequestLogging())
	{
		router_webhook := newrouter.Group("/webhook")
		{
			router_webhook.POST("", webhook.Inbound)
		}

		router_admin := newrouter.Group("/admin")
		{
			router_admin.POST("/register", admin.Register)
			router_admin_webhook := router_admin.Group("/webhook")
			{
				router_admin_webhook.POST("", admin_webhook.Inbound)
			}
		}
	}

}
