package routes

import (
	"clicli/api/v1"
	"clicli/middleware"
	"github.com/gin-gonic/gin"
)

func CollectResourceRoutes(r *gin.RouterGroup) {
	resource := r.Group("resource")
	{
		//需要用户登录
		auth := resource.Group("")
		auth.Use(middleware.Auth())
		{
			auth.POST("title/modify", api.ModifyResourceTitle)
			auth.POST("delete", api.DeleteResource)
		}
	}
}
