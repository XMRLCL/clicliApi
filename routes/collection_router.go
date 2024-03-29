package routes

import (
	"clicli/api/v1"
	"clicli/middleware"
	"github.com/gin-gonic/gin"
)

func CollectCollectionRoutes(r *gin.RouterGroup) {
	collection := r.Group("collection")
	{
		auth := collection.Group("")
		auth.Use(middleware.Auth())
		{
			// 获取收藏夹列表
			auth.GET("list", api.GetCollectionList)
			// 获取收藏夹信息
			auth.GET("info", api.GetCollectionInfo)
			// 添加收藏夹
			auth.POST("add", api.CreateCollection)
			// 修改收藏夹
			auth.POST("modify", api.ModifyCollection)
			// 删除收藏夹
			auth.POST("delete", api.DeleteCollection)
		}

	}

}
