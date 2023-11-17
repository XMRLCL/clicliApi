package routes

import (
	"clicli/api/v1"
	"clicli/middleware"
	"github.com/gin-gonic/gin"
)

func CollectUploadRoutes(r *gin.RouterGroup) {
	upload := r.Group("upload")
	{
		auth := upload.Group("")
		auth.Use(middleware.Auth())
		{
			auth.POST("image", api.UploadImg)
			auth.POST("video/:vid", api.UploadVideo)
		}
	}
}
