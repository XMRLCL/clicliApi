package api

import (
	"clicli/domain/resp"
	"clicli/domain/vo"
	"clicli/service"
	"clicli/util/convert"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func GetLikeMessage(ctx *gin.Context) {
	page := convert.StringToInt(ctx.Query("page"))
	pageSize := convert.StringToInt(ctx.Query("page_size"))

	if pageSize > 30 {
		resp.Response(ctx, resp.TooManyRequestsError, "", nil)
		zap.L().Error("请求数量过多 ")
		return
	}

	userId := ctx.GetUint("userId")

	messages := service.SelectLikeMessage(userId, page, pageSize)
	// 查询对应的用户和视频
	for i := 0; i < len(messages); i++ {
		messages[i].User = service.GetUserInfo(messages[i].Fid)
		messages[i].Video = service.GetVideoInfo(messages[i].Vid)
	}

	// 返回给前端
	resp.OK(ctx, "ok", gin.H{"messages": vo.ToLikeMessageVoList(messages)})
}
