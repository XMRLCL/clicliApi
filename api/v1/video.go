package api

import (
	"clicli/cache"
	"clicli/common"
	"clicli/domain/dto"
	"clicli/domain/model"
	"clicli/domain/resp"
	"clicli/domain/valid"
	"clicli/domain/vo"
	"clicli/service"
	"clicli/util/convert"
	"clicli/ws"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"io"
	"net/http"
	"strconv"
	"strings"
)

type UtilsFun struct {
	Code    int               `json:"code"`
	Server  string            `json:"server"`
	Service string            `json:"service"`
	Pid     string            `json:"pid"`
	Streams []UtilsFunStreams `json:"streams"`
}

type UtilsFunStreams struct {
	Id         string                 `json:"id"`
	Name       string                 `json:"name"`
	Vhost      string                 `json:"vhost"`
	App        string                 `json:"app"`
	TcUrl      string                 `json:"tcUrl"`
	Url        string                 `json:"url"`
	Live_ms    int                    `json:"live_ms"`
	Clients    int                    `json:"clients"`
	Frames     int                    `json:"frames"`
	Send_bytes int                    `json:"send_bytes"`
	Recv_bytes int                    `json:"recv_bytes"`
	Kbps       UtilsFunStreamsKbps    `json:"kbps"`
	Publish    UtilsFunStreamsPublish `json:"publish"`
	Video      UtilsFunStreamsVideo   `json:"video"`
	Audio      UtilsFunStreamsAudio   `json:"audio"`
}

type UtilsFunStreamsKbps struct {
	Recv_30s int `json:"recv_30s"`
	Send_30s int `json:"send_30s"`
}

type UtilsFunStreamsPublish struct {
	Active bool   `json:"active"`
	Cid    string `json:"cid"`
}

type UtilsFunStreamsVideo struct {
	Codec   string `json:"codec"`
	Profile string `json:"profile"`
	Level   string `json:"level"`
	Width   int    `json:"width"`
	Height  int    `json:"height"`
}

type UtilsFunStreamsAudio struct {
	Codec       string `json:"codec"`
	Sample_rate int    `json:"sample_rate"`
	Channel     int    `json:"channel"`
	Profile     string `json:"profile"`
}

// 上传视频信息
func UploadVideoInfo(ctx *gin.Context) {
	var uploadVideoDTO dto.UploadVideoDTO
	if err := ctx.Bind(&uploadVideoDTO); err != nil {
		resp.Response(ctx, resp.RequestParamError, "", nil)
		zap.L().Error("请求参数有误")
		return
	}

	// 参数校验
	if !valid.Title(uploadVideoDTO.Title) {
		resp.Response(ctx, resp.RequestParamError, valid.TITLE_ERROR, nil)
		zap.L().Error(valid.TITLE_ERROR)
		return
	}

	userId := ctx.GetUint("userId")
	if cache.GetUploadImage(uploadVideoDTO.Cover) != userId {
		resp.Response(ctx, resp.InvalidLinkError, "", nil)
		zap.L().Error("文件链接无效")
		return
	}

	if !service.IsSubpartition(uploadVideoDTO.Partition) {
		resp.Response(ctx, resp.PartitionError, "", nil)
		zap.L().Error("分区不存在")
		return
	}

	video := dto.UploadVideoDtoToVideo(userId, uploadVideoDTO)
	vid, err := service.InsertVideo(video)
	if err != nil {
		resp.Response(ctx, resp.Error, "创建视频失败", nil)
		zap.L().Error("创建视频失败 " + err.Error())
		return
	}

	// 返回给前端
	resp.OK(ctx, "ok", gin.H{"vid": vid})
}

// 修改视频信息
func ModifyVideoInfo(ctx *gin.Context) {
	var modifyVideoDTO dto.ModifyVideoDTO
	if err := ctx.Bind(&modifyVideoDTO); err != nil {
		resp.Response(ctx, resp.RequestParamError, "", nil)
		zap.L().Error("请求参数有误")
		return
	}

	// 参数校验
	if !valid.Title(modifyVideoDTO.Title) {
		resp.Response(ctx, resp.RequestParamError, valid.TITLE_ERROR, nil)
		zap.L().Error(valid.TITLE_ERROR)
		return
	}

	// 校验用户是否为视频作者
	userId := ctx.GetUint("userId")
	oldVideoInfo := service.GetVideoInfo(modifyVideoDTO.VID)
	if oldVideoInfo.Uid != userId {
		if modifyVideoDTO.Cover != oldVideoInfo.Cover && cache.GetUploadImage(modifyVideoDTO.Cover) != userId {
			resp.Response(ctx, resp.VideoNotExistError, "", nil)
			zap.L().Error("视频不存在")
			return
		}
	}

	// 校验封面图文件是否有效
	if modifyVideoDTO.Cover != oldVideoInfo.Cover && cache.GetUploadImage(modifyVideoDTO.Cover) != userId {
		resp.Response(ctx, resp.InvalidLinkError, "", nil)
		zap.L().Error("文件链接无效")
		return
	}

	// 保存到数据库
	service.UpdateVideoInfo(modifyVideoDTO)
	// 返回给前端
	resp.OK(ctx, "ok", nil)
}

// 获取视频状态
func GetVideoStatus(ctx *gin.Context) {
	videoId := convert.StringToUint(ctx.DefaultQuery("vid", "0"))
	video := service.GetVideoInfo(videoId)

	resources := service.SelectResourceByVideo(videoId, false)

	// 返回给前端
	resp.OK(ctx, "ok", gin.H{"video": vo.ToVideoStatusVO(video, resources)})
}

// 获取视频信息
func GetVideoByID(ctx *gin.Context) {
	ios := convert.StringToUint(ctx.DefaultQuery("ios", "0"))
	vid := convert.StringToUint(ctx.DefaultQuery("vid", "0"))

	video := service.GetVideoInfo(vid)
	if video.ID == 0 || video.Status != common.AUDIT_APPROVED {
		resp.Response(ctx, resp.VideoNotExistError, "", nil)
		zap.L().Error("视频不存在")
		return
	}

	//获取作者信息
	video.Author = service.GetUserInfo(video.Uid)

	//获取视频资源
	resources := service.SelectResourceByVideo(video.ID, true)

	//增加播放量(一个ip在同一个视频下，每30分钟可重新增加1播放量)
	service.AddVideoClicks(video.ID, ctx.ClientIP())

	// 获取播放量
	video.Clicks = service.GetVideoClicks(video.ID)

	// 返回给前端
	resp.OK(ctx, "ok", gin.H{"video": vo.ToVideoVO(video, resources, ios)})
}

// 提交审核
func SubmitReview(ctx *gin.Context) {
	//获取参数
	var idDTO dto.IdDTO
	if err := ctx.Bind(&idDTO); err != nil {
		resp.Response(ctx, resp.RequestParamError, "", nil)
		zap.L().Error("请求参数有误")
		return
	}

	if service.SelectResourceCountByVid(idDTO.ID) == 0 {
		resp.Response(ctx, resp.ResourceNotExistError, "", nil)
		zap.L().Error("资源不存在")
		return
	}

	// 更新视频状态
	service.UpadteVideoStatus(idDTO.ID, common.WAITING_REVIEW)

	// 返回给前端
	resp.OK(ctx, "ok", nil)
}

// 提交密钥
func SubmitFlvkey(ctx *gin.Context) {

	var idDTO dto.IDflv
	zap.L().Info(strconv.Itoa(int(idDTO.ID)))
	zap.L().Info("ID")
	if err := ctx.Bind(&idDTO); err != nil {
		resp.Response(ctx, resp.RequestParamError, "", nil)
		zap.L().Error("请求参数有误")
		return
	}

	if idDTO.Flvkey == "" {
		resp.Response(ctx, resp.KeyNotExistError, "", nil)
		zap.L().Error("密钥为空")
		return
	}

	videoInfo := service.GetVideoInfo(idDTO.ID)

	videoInfo.Flv = "http://ctguqmx.run:8080/live/livestream/" + idDTO.Flvkey + ".flv"
	service.UpdateVideoFlv(videoInfo)

	// 更新视频状态
	service.UpadteVideoStatus(idDTO.ID, common.WAITING_REVIEW)

	// 返回给前端
	resp.OK(ctx, "ok", nil)
}

func GetVideoFlv(ctx *gin.Context) {
	var idDTO dto.IdDTO

	if err := ctx.Bind(&idDTO); err != nil {
		resp.Response(ctx, resp.RequestParamError, "", nil)
		zap.L().Error("请求参数有误")
		return
	}

	videoInfo := service.GetVideoInfo(idDTO.ID)

	resp.OK(ctx, "ok", gin.H{"url": videoInfo.Flv})
}

// 获取收藏视频列表
func GetCollectVideo(ctx *gin.Context) {
	id := convert.StringToUint(ctx.DefaultQuery("id", "0"))

	userId := ctx.GetUint("userId")
	page := convert.StringToInt(ctx.DefaultQuery("page", "1"))
	pageSize := convert.StringToInt(ctx.DefaultQuery("page_size", "10"))

	if pageSize > 30 {
		resp.Response(ctx, resp.TooManyRequestsError, "", nil)
		zap.L().Error("请求数量过多 ")
		return
	}

	collection := service.SelectCollectionByID(id)
	if !collection.Open && collection.Uid != userId {
		resp.Response(ctx, resp.CollectionNotExistError, "", nil)
		zap.L().Error("收藏夹不存在")
		return
	}

	videos, total, err := service.SelectCollectVideo(id, page, pageSize)
	if err != nil {
		resp.Response(ctx, resp.Error, "", nil)
		zap.L().Error("获取收藏视频失败" + err.Error())
		return
	}

	// 返回给前端
	resp.OK(ctx, "ok", gin.H{"total": total, "videos": vo.ToBaseVideoVoList(videos)})
}

// 获取自己的视频
func GetUploadVideoList(ctx *gin.Context) {
	userId := ctx.GetUint("userId")
	page := convert.StringToInt(ctx.Query("page"))
	pageSize := convert.StringToInt(ctx.Query("page_size"))

	if pageSize > 30 {
		resp.Response(ctx, resp.TooManyRequestsError, "", nil)
		zap.L().Error("请求数量过多 ")
		return
	}

	total, videos := service.SelectUploadVideo(userId, page, pageSize)
	// 更新播放量数据
	for i := 0; i < len(videos); i++ {
		videos[i].Clicks = service.GetVideoClicks(videos[i].ID)
	}

	// 返回给前端
	resp.OK(ctx, "ok", gin.H{"total": total, "videos": vo.ToUserUploadVideoVoList(videos)})
}

// 通过用户ID获取视频列表
func GetVideoListByUid(ctx *gin.Context) {
	uid := convert.StringToUint(ctx.Query("uid"))
	page := convert.StringToInt(ctx.Query("page"))
	pageSize := convert.StringToInt(ctx.Query("page_size"))

	if pageSize > 30 {
		resp.Response(ctx, resp.TooManyRequestsError, "", nil)
		zap.L().Error("请求数量过多 ")
		return
	}

	total, videos := service.SelectVideoByUserId(uid, page, pageSize)
	// 更新播放量数据
	for i := 0; i < len(videos); i++ {
		videos[i].Clicks = service.GetVideoClicks(videos[i].ID)
	}

	// 返回给前端
	resp.OK(ctx, "ok", gin.H{"total": total, "videos": vo.ToBaseVideoVoList(videos)})
}

// 删除视频
func DeleteVideo(ctx *gin.Context) {
	var idDTO dto.IdDTO
	if err := ctx.Bind(&idDTO); err != nil {
		resp.Response(ctx, resp.RequestParamError, "", nil)
		zap.L().Error("请求参数有误")
		return
	}

	userId := ctx.GetUint("userId")
	if !service.IsVideoBelongUser(idDTO.ID, userId) {
		resp.Response(ctx, resp.VideoNotExistError, "", nil)
		zap.L().Error("视频不存在")
		return
	}

	service.DeleteVideo(idDTO.ID)

	// 返回给前端
	resp.OK(ctx, "ok", nil)
}

// 获取视频列表
func GetVideoList(ctx *gin.Context) {
	page := convert.StringToInt(ctx.Query("page"))
	pageSize := convert.StringToInt(ctx.Query("page_size"))
	partitionId := convert.StringToUint(ctx.DefaultQuery("partition", "0")) //分区

	if pageSize > 30 {
		resp.Response(ctx, resp.TooManyRequestsError, "", nil)
		zap.L().Error("请求数量过多 ")
		return
	}

	var videos []model.Video
	if partitionId == 0 {
		//不传分区参数默认查询全部
		_, videos = service.SelectVideoListByStatus(page, pageSize, common.AUDIT_APPROVED)
	} else if service.IsSubpartition(partitionId) {
		// 如果为子分区，查询分区下的视频
		_, videos = service.SelectVideoListBySubpartition(partitionId, page, pageSize)
	} else {
		// 获取该分区下的视频
		_, videos = service.SelectVideoListByPartition(partitionId, page, pageSize)
	}

	// 更新播放量数据和作者信息
	for i := 0; i < len(videos); i++ {
		videos[i].Clicks = service.GetVideoClicks(videos[i].ID)
		videos[i].Author = service.GetUserInfo(videos[i].Uid)
	}

	// 返回给前端
	resp.OK(ctx, "ok", gin.H{"videos": vo.ToSearchVideoVoList(videos)})
}

// 获取推荐视频
func GetRecommendedVideo(ctx *gin.Context) {
	pageSize := convert.StringToInt(ctx.DefaultQuery("page_size", "15"))

	if pageSize > 30 {
		resp.Response(ctx, resp.TooManyRequestsError, "", nil)
		zap.L().Error("请求数量过多 ")
		return
	}

	// 没有推荐功能，按点击量查询视频（点击量不是实时数据）
	videos := service.SelectVideoListByClicks(pageSize)
	//zap.L().Error(strconv.Itoa(videos[0].Video))
	// 更新播放量数据和作者信息
	for i := 0; i < len(videos); i++ {
		videos[i].Clicks = service.GetVideoClicks(videos[i].ID)
		videos[i].Author = service.GetUserInfo(videos[i].Uid)
	}
	// 返回给前端
	resp.OK(ctx, "ok", gin.H{"videos": vo.ToSearchVideoVoList(videos)})
}

// 搜索视频
func SearchVideo(ctx *gin.Context) {
	keywords := ctx.Query("keywords")
	page := convert.StringToInt(ctx.Query("page"))
	pageSize := convert.StringToInt(ctx.DefaultQuery("page_size", "15"))

	if pageSize > 30 {
		resp.Response(ctx, resp.TooManyRequestsError, "", nil)
		zap.L().Error("请求数量过多 ")
		return
	}

	var videos []model.Video
	if len(keywords) == 0 {
		_, videos = service.SelectVideoListByStatus(page, pageSize, common.AUDIT_APPROVED)
	} else {
		// 直接用mysql模糊查询，之后可能会更换为es
		videos = service.SelectVideoListByKeywords(keywords, page, pageSize)
	}

	// 更新播放量数据和作者信息
	for i := 0; i < len(videos); i++ {
		videos[i].Clicks = service.GetVideoClicks(videos[i].ID)
		videos[i].Author = service.GetUserInfo(videos[i].Uid)
	}

	// 返回给前端
	resp.OK(ctx, "ok", gin.H{"videos": vo.ToSearchVideoVoList(videos)})
}

// 获取直播
func GetUserLikeList(ctx *gin.Context) {
	pageSize := convert.StringToInt(ctx.DefaultQuery("page_size", "15"))

	if pageSize > 30 {
		resp.Response(ctx, resp.TooManyRequestsError, "", nil)
		zap.L().Error("请求数量过多 ")
		return
	}

	// 没有推荐功能，按点击量查询视频（点击量不是实时数据）
	videos := service.SelectLikeListByClicks(pageSize)
	// 更新播放量数据和作者信息
	for i := 0; i < len(videos); i++ {
		videos[i].Clicks = service.GetVideoClicks(videos[i].ID)
		videos[i].Author = service.GetUserInfo(videos[i].Uid)
	}

	var likes []model.Video

	for f := 0; f < len(videos); f++ {
		if getLike(videos[f].ID) == 1 {
			l := append(likes, videos[f])
			likes = l
		}
	}
	// 返回给前端
	resp.OK(ctx, "ok", gin.H{"videos": vo.ToSearchVideoVoList(likes)})
}

func getLike(id uint) int {

	rsps, err := http.Get("http://api.ctguqmx.run:1985/api/v1/streams")
	if err != nil {
		zap.L().Error("请求SRS失败" + err.Error())
		return 0
	}
	defer rsps.Body.Close()

	rspsbody, err := io.ReadAll(rsps.Body)

	if err != nil {
		zap.L().Error("解析请求错误" + err.Error())
		return 0
	}

	var repsbody UtilsFun

	if err := json.Unmarshal([]byte(rspsbody), &repsbody); err == nil {
	}

	videoInfo := service.GetVideoInfo(id)

	flv := videoInfo.Flv

	m := strings.Split(flv, "http://ctguqmx.run:8080/live/livestream/")
	m = strings.Split(m[1], ".flv")
	flv = m[0]

	for f := 0; f < len(repsbody.Streams); f++ {
		if flv == repsbody.Streams[f].Name {
			return 1
		}
	}

	return 0
}

func GetLike(ctx *gin.Context) {
	var reviewDTO dto.LikeDTO
	if err := ctx.Bind(&reviewDTO); err != nil {
		resp.Response(ctx, resp.RequestParamError, "", nil)
		zap.L().Error("请求参数有误")
		return
	}

	rsps, err := http.Get("http://api.ctguqmx.run:1985/api/v1/streams")
	if err != nil {
		resp.Response(ctx, resp.LikeNotExistError, "", nil)
		zap.L().Error("请求SRS失败" + err.Error())
		return
	}
	defer rsps.Body.Close()

	rspsbody, err := io.ReadAll(rsps.Body)

	if err != nil {
		resp.Response(ctx, resp.LikeNotExistError, "", nil)
		zap.L().Error("解析请求错误" + err.Error())
		return
	}

	var repsbody UtilsFun

	if err := json.Unmarshal([]byte(rspsbody), &repsbody); err == nil {
	}

	videoInfo := service.GetVideoInfo(reviewDTO.ID)

	flv := videoInfo.Flv
	m := strings.Split(flv, "http://ctguqmx.run:8080/live/livestream/")
	m = strings.Split(m[1], ".flv")
	flv = m[0]

	for f := 0; f < len(repsbody.Streams); f++ {
		if flv == repsbody.Streams[f].Name {
			resp.OK(ctx, "ok", nil)
		}
	}

	resp.Response(ctx, resp.LikeNotExistError, "", nil)
	return
}

func GetReviewLiveList(ctx *gin.Context) {
	page := convert.StringToInt(ctx.Query("page"))
	pageSize := convert.StringToInt(ctx.Query("page_size"))

	if pageSize > 30 {
		resp.Response(ctx, resp.TooManyRequestsError, "", nil)
		zap.L().Error("请求数量过多 ")
		return
	}

	total, videos := service.SelectLiveListByStatus(page, pageSize, common.WAITING_REVIEW)

	// 返回给前端
	resp.OK(ctx, "ok", gin.H{"total": total, "videos": vo.ToSearchVideoVoList(videos)})
}

// 获取待审核视频列表
func GetReviewVideoList(ctx *gin.Context) {
	//获取参数
	page := convert.StringToInt(ctx.Query("page"))
	pageSize := convert.StringToInt(ctx.Query("page_size"))

	if pageSize > 30 {
		resp.Response(ctx, resp.TooManyRequestsError, "", nil)
		zap.L().Error("请求数量过多 ")
		return
	}

	total, videos := service.SelectVideoListByStatus(page, pageSize, common.WAITING_REVIEW)

	var m []model.Video

	for f := 0; f < len(videos); f++ {
		if videos[f].Video == 1 {
			n := append(m, videos[f])
			m = n
		}
	}

	videos = m

	// 返回给前端
	resp.OK(ctx, "ok", gin.H{"total": total, "videos": vo.ToSearchVideoVoList(videos)})
}

// 审核视频
func ReviewVideo(ctx *gin.Context) {
	var reviewDTO dto.ReviewDTO
	if err := ctx.Bind(&reviewDTO); err != nil {
		resp.Response(ctx, resp.RequestParamError, "", nil)
		zap.L().Error("请求参数有误")
		return
	}

	if !valid.ReviewStatus(reviewDTO.Status) {
		resp.Response(ctx, resp.RequestParamError, valid.REVIEW_STATUS_ERROR, nil)
		zap.L().Error(valid.REVIEW_STATUS_ERROR)
		return
	}

	//if reviewDTO.Status == common.AUDIT_APPROVED {
	//	if service.SelectResourceCountByStatus(reviewDTO.ID, common.AUDIT_APPROVED) == 0 {
	//		resp.Response(ctx, resp.ResourceNotExistError, "", nil)
	//		zap.L().Error("资源不存在")
	//		return
	//	}
	//}

	service.UpadteVideoStatus(reviewDTO.ID, reviewDTO.Status)

	// 返回给前端
	resp.OK(ctx, "ok", nil)
}

// 审核视频资源
func ReviewResource(ctx *gin.Context) {
	var reviewDTO dto.ReviewDTO
	if err := ctx.Bind(&reviewDTO); err != nil {
		resp.Response(ctx, resp.RequestParamError, "", nil)
		zap.L().Error("请求参数有误")
		return
	}

	if !valid.ReviewStatus(reviewDTO.Status) {
		resp.Response(ctx, resp.RequestParamError, valid.REVIEW_STATUS_ERROR, nil)
		zap.L().Error(valid.REVIEW_STATUS_ERROR)
		return
	}

	service.UpadteResourceStatus(reviewDTO.ID, reviewDTO.Status)

	// 返回给前端
	resp.OK(ctx, "ok", nil)
}

// 通过视频ID获取待审核视频资源
func GetReviewVideoByID(ctx *gin.Context) {
	vid := convert.StringToUint(ctx.Query("vid"))

	resources := service.SelectResourceByVideo(vid, false)

	// 返回给前端
	resp.OK(ctx, "ok", gin.H{"resources": vo.ToResourceVoList(resources)})
}

// 管理员获取视频列表
func AdminGetVideoList(ctx *gin.Context) {
	page := convert.StringToInt(ctx.Query("page"))
	pageSize := convert.StringToInt(ctx.Query("page_size"))
	partitionId := convert.StringToUint(ctx.DefaultQuery("partition", "0")) //分区

	if pageSize > 30 {
		resp.Response(ctx, resp.TooManyRequestsError, "", nil)
		zap.L().Error("请求数量过多 ")
		return
	}

	var total int64
	var videos []model.Video
	if partitionId == 0 {
		//不传分区参数默认查询全部
		total, videos = service.SelectVideoListByStatus(page, pageSize, common.AUDIT_APPROVED)
	} else if service.IsSubpartition(partitionId) {
		// 如果为子分区，查询分区下的视频
		total, videos = service.SelectVideoListBySubpartition(partitionId, page, pageSize)
	} else {
		// 获取该分区下的视频
		total, videos = service.SelectVideoListByPartition(partitionId, page, pageSize)
	}

	// 更新播放量数据和作者信息
	for i := 0; i < len(videos); i++ {
		videos[i].Clicks = service.GetVideoClicks(videos[i].ID)
		videos[i].Author = service.GetUserInfo(videos[i].Uid)
	}

	// 返回给前端
	resp.OK(ctx, "ok", gin.H{"total": total, "videos": vo.ToSearchVideoVoList(videos)})
}

// 管理员搜索视频
func AdminSearchVideo(ctx *gin.Context) {
	keywords := ctx.Query("keywords")
	page := convert.StringToInt(ctx.Query("page"))
	pageSize := convert.StringToInt(ctx.Query("page_size"))

	if pageSize > 30 {
		resp.Response(ctx, resp.TooManyRequestsError, "", nil)
		zap.L().Error("请求数量过多 ")
		return
	}

	total, videos := service.AdminSelectVideoListByKeywords(keywords, page, pageSize)

	// 更新播放量数据和作者信息
	for i := 0; i < len(videos); i++ {
		videos[i].Clicks = service.GetVideoClicks(videos[i].ID)
		videos[i].Author = service.GetUserInfo(videos[i].Uid)
	}

	// 返回给前端
	resp.OK(ctx, "ok", gin.H{"total": total, "videos": vo.ToSearchVideoVoList(videos)})
}

// 删除视频
func AdminDeleteVideo(ctx *gin.Context) {
	var idDTO dto.IdDTO
	if err := ctx.Bind(&idDTO); err != nil {
		resp.Response(ctx, resp.RequestParamError, "", nil)
		zap.L().Error("请求参数有误")
		return
	}

	service.DeleteVideo(idDTO.ID)

	// 返回给前端
	resp.OK(ctx, "ok", nil)
}

// 视频Websocket连接(统计在线人数)
func GetRoomConnect(ctx *gin.Context) {
	vid := convert.StringToUint(ctx.Query("vid"))
	clientId := ctx.Query("client_id")
	if vid == 0 {
		return
	}

	// 升级为websocket长链接
	ws.RoomWsHandler(ctx.Writer, ctx.Request, vid, clientId)
}
