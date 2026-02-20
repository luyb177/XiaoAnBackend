package logic

import (
	"context"
	"errors"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/luyb177/XiaoAnBackend/content/internal/model"
	"github.com/luyb177/XiaoAnBackend/content/internal/svc"
	"github.com/luyb177/XiaoAnBackend/content/pb/content/v1"
	"github.com/luyb177/XiaoAnBackend/content/pkg/comic/convert"
	"github.com/luyb177/XiaoAnBackend/infra/constants"
	"github.com/luyb177/XiaoAnBackend/infra/middleware"
)

type GetNewComicsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
	ComicModel model.ComicModel
}

func NewGetNewComicsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetNewComicsLogic {
	return &GetNewComicsLogic{
		ctx:        ctx,
		svcCtx:     svcCtx,
		Logger:     logx.WithContext(ctx),
		ComicModel: model.NewComicModel(svcCtx.Mysql),
	}
}

// GetNewComics 获取最新漫画列表
func (l *GetNewComicsLogic) GetNewComics(in *v1.GetNewComicsRequest) (*v1.Response, error) {
	user, ok := middleware.GetUser(l.ctx)
	if !ok || user.UID == constants.InvalidUserID || user.Status != constants.UserStatusNormal {
		return bad("用户未登录或登录状态异常"), nil
	}
	if in.PageSize <= 0 {
		in.PageSize = 10
	}

	limit := in.PageSize + 1

	var (
		list []*model.Comic
		err  error
	)

	if in.Cursor == 0 {
		// 首次查询
		list, err = l.ComicModel.FindManyWithNotDelete(l.ctx, limit)
	} else {
		// 通过游标查询
		list, err = l.ComicModel.FindManyWithNotDeleteByCursor(l.ctx, in.Cursor, limit)
	}

	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return notFound("没有更多漫画了"), nil
		}
		l.Errorf("GetNewComics err: %v", err)
		return internal("获取最新漫画列表失败"), nil
	}

	hasMore := int64(len(list)) > in.PageSize
	if hasMore {
		list = list[:in.PageSize]
	}

	nextCursor := uint64(0)
	if len(list) > 0 {
		nextCursor = list[len(list)-1].Id
	}

	// NOTE: 这里没有获取漫画的 tag like collect 等相关内容，前端可以根据漫画ID单独请求获取
	comicsPB := convert.PBFromComics(list)

	res := &v1.GetNewComicsResponse{
		Comics:     comicsPB,
		HasMore:    hasMore,
		NextCursor: nextCursor,
	}

	resAny, err := anypb.New(res)
	if err != nil {
		l.Errorf("GetNewComics err: %v", err)
		return internal("获取最新漫画列表失败"), nil
	}

	return &v1.Response{
		Code:    200,
		Message: "获取最新漫画列表成功",
		Data:    resAny,
	}, nil
}
