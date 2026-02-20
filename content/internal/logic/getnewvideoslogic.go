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

	// NOTE: 这里没有获取视频的 tag like collect 等相关内容，前端可以根据视频ID单独请求获取
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
