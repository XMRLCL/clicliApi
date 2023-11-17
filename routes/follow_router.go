package routes

import (
	"clicli/api/v1"
	"clicli/middleware"
	"github.com/gin-gonic/gin"
)

func CollectFollowRoutes(r *gin.RouterGroup) {
	follow := r.Group("follow")
	{
		// 关注列表
		follow.GET("/following", api.GetFollowings)
		// 粉丝列表
		follow.GET("/follower", api.GetFollowers)
		// 获取关注数和粉丝数
		follow.GET("/count", api.GetFollowCount)

		auth := follow.Group("")
		auth.Use(middleware.Auth())
		{
			// 获取关注状态
			auth.GET("/status", api.GetFollowStatus)
			// 关注
			auth.POST("/add", api.Follow)
			// 取消关注
			auth.POST("/cancel", api.UnFollow)
		}
	}
}
