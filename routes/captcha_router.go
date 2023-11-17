package routes

import (
	"clicli/api/v1"
	"github.com/gin-gonic/gin"
)

func CollectCaptchaRoutes(r *gin.RouterGroup) {
	captcha := r.Group("captcha")
	{
		// 获取滑块验证
		captcha.GET("get", api.GetSliderCaptcha)
		// 验证滑块
		captcha.POST("validate", api.ValidateSlider)
	}

}
