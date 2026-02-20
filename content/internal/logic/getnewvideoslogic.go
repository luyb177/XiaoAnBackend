package logic

import (
	"context"
	"errors"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/luyb177/XiaoAnBackend/content/internal/model"
	"github.com/luyb177/XiaoAnBackend/content/internal/svc"
	"github.com/luyb177/XiaoAnBackend/content/pb/content/v1"
	"github.com/luyb177/XiaoAnBackend/content/pkg/video/convert"
	"github.com/luyb177/XiaoAnBackend/infra/constants"
	"github.com/luyb177/XiaoAnBackend/infra/middleware"
)

type GetNewVideosLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
	VideoDao model.VideoModel
}

func NewGetNewVideosLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetNewVideosLogic {
	return &GetNewVideosLogic{
		ctx:      ctx,
		svcCtx:   svcCtx,
		Logger:   logx.WithContext(ctx),
		VideoDao: model.NewVideoModel(svcCtx.Mysql),
	}
}

// GetNewVideos 获取最新视频列表
func (l *GetNewVideosLogic) GetNewVideos(in *v1.GetNewVideosRequest) (*v1.Response, error) {
	user, ok := middleware.GetUser(l.ctx)
	if !ok || user.UID == constants.InvalidUserID || user.Status != constants.UserStatusNormal {
		return bad("用户未登录或登录状态异常"), nil
	}

	if in.PageSize <= 0 {
		in.PageSize = 10
	}

	limit := in.PageSize + 1

	var (
		list []*model.Video
		err  error
	)

	if in.Cursor == 0 {
		// 首次查询
		list, err = l.VideoDao.FindManyWithNotDelete(l.ctx, limit)
	} else {
		// 通过游标查询
		list, err = l.VideoDao.FindManyWithNotDeleteByCursor(l.ctx, in.Cursor, limit)
	}
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return notFound("没有更多视频了"), nil
		}
		l.Errorf("FindManyWithNotDeleteByCursor error: %v", err)
		return internal("获取视频列表失败"), nil
	}

	hasMore := int64(len(list)) > in.PageSize
	if hasMore {
		list = list[:in.PageSize]
	}

	nextCursor := uint64(0)
	if len(list) > 0 {
		nextCursor = list[len(list)-1].Id
	}

	// NOTE: 此列表接口仅返回视频基础信息。视频的标签、点赞、收藏状态等关联内容应在用户查看视频详情时通过单独接口获取，避免列表页出现 N+1 查询。
	videosPB := convert.PBFromVideo(list)

	res := &v1.GetNewVideosResponse{
		Videos:     videosPB,
		HasMore:    hasMore,
		NextCursor: nextCursor,
	}

	resAny, err := anypb.New(res)
	if err != nil {
		l.Errorf("GetNewVideosLogic anypb.New error: %v", err)
		return internal("获取视频列表失败"), nil
	}

	return &v1.Response{
		Code:    200,
		Message: "获取视频列表成功",
		Data:    resAny,
	}, nil
}
