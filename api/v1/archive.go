package api

import (
	"clicli/domain/resp"
	"clicli/domain/vo"
	"clicli/service"
	"clicli/util/convert"
	"github.com/gin-gonic/gin"
)

// 通过视频ID获取点赞收藏数据
func GetArchiveStat(ctx *gin.Context) {
	videoId := convert.StringToUint(ctx.DefaultQuery("vid", "0"))

	likeCount, _ := service.SelectLikeCount(videoId)
	collectCount, _ := service.SelectCollectCount(videoId)

	// 返回给前端
	resp.OK(ctx, "ok", gin.H{"stat": vo.ArchiveStatVO{
		Like:    int64(likeCount),
		Collect: int64(collectCount),
	}})
}
