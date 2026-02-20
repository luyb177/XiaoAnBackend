package logic

import (
	"context"
	"errors"

	"github.com/luyb177/XiaoAnBackend/content/internal/model"
	"github.com/luyb177/XiaoAnBackend/content/internal/svc"
	"github.com/luyb177/XiaoAnBackend/content/pb/content/v1"
	"github.com/luyb177/XiaoAnBackend/content/pkg/podcast/convert"
	"github.com/luyb177/XiaoAnBackend/infra/constants"
	"github.com/luyb177/XiaoAnBackend/infra/middleware"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetNewPodcastsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
	PodcastDao model.PodcastModel
}

func NewGetNewPodcastsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetNewPodcastsLogic {
	return &GetNewPodcastsLogic{
		ctx:        ctx,
		svcCtx:     svcCtx,
		Logger:     logx.WithContext(ctx),
		PodcastDao: model.NewPodcastModel(svcCtx.Mysql),
	}
}

// GetNewPodcasts 获取最新播客列表
func (l *GetNewPodcastsLogic) GetNewPodcasts(in *v1.GetNewPodcastsRequest) (*v1.Response, error) {
	user, ok := middleware.GetUser(l.ctx)
	if !ok || user.UID == constants.InvalidUserID || user.Status != constants.UserStatusNormal {
		return bad("用户未登录或登录状态异常"), nil
	}
	if in.PageSize <= 0 {
		in.PageSize = 10
	}

	limit := in.PageSize + 1

	var (
		list []*model.Podcast
		err  error
	)

	if in.Cursor == 0 {
		// 首次查询
		list, err = l.PodcastDao.FindManyWithNotDelete(l.ctx, limit)
	} else {
		// 通过游标查询
		list, err = l.PodcastDao.FindManyWithNotDeleteByCursor(l.ctx, in.Cursor, limit)
	}

	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return notFound("没有更多播客了"), nil
		}
		l.Errorf("GetNewPodcasts err: %v", err)
		return internal("获取播客列表失败"), nil
	}

	hasMore := int64(len(list)) > in.PageSize
	if hasMore {
		list = list[:in.PageSize]
	}

	nextCursor := uint64(0)
	if len(list) > 0 {
		nextCursor = list[len(list)-1].Id
	}

	// NOTE: 这里没有获取播客的 tag highlight like collect 等相关内容，前端可以根据播客ID单独请求获取
	podcastsPB := convert.PBFromPodcasts(list)

	res := &v1.GetNewPodcastsResponse{
		Podcasts:   podcastsPB,
		HasMore:    hasMore,
		NextCursor: nextCursor,
	}

	resAny, err := anypb.New(res)
	if err != nil {
		l.Errorf("GetNewPodcasts err: %v", err)
		return internal("获取播客列表失败"), nil
	}

	return &v1.Response{
		Code:    200,
		Message: "获取播客列表成功",
		Data:    resAny,
	}, nil
}
