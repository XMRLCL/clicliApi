package routes

import (
	"clicli/api/v1"
	"clicli/middleware"
	"github.com/gin-gonic/gin"
)

func CollectDanmakuRoutes(r *gin.RouterGroup) {
	danmaku := r.Group("danmaku")
	{
		// 获取弹幕列表
		danmaku.GET("list", api.GetDanmaku)

		//需要用户登录
		auth := danmaku.Group("")
		auth.Use(middleware.Auth())
		{
			// 发送弹幕
			auth.POST("send", api.SendDanmaku)
		}
	}
}
