package logic

import (
	"context"
	"errors"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/luyb177/XiaoAnBackend/content/internal/model"
	"github.com/luyb177/XiaoAnBackend/content/internal/svc"
	"github.com/luyb177/XiaoAnBackend/content/pb/content/v1"
	"github.com/luyb177/XiaoAnBackend/content/pkg/article/convert"
	"github.com/luyb177/XiaoAnBackend/infra/constants"
	"github.com/luyb177/XiaoAnBackend/infra/middleware"
)

type GetNewArticlesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
	ArticleDao model.ArticleModel
}

func NewGetNewArticlesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetNewArticlesLogic {
	return &GetNewArticlesLogic{
		ctx:        ctx,
		svcCtx:     svcCtx,
		Logger:     logx.WithContext(ctx),
		ArticleDao: model.NewArticleModel(svcCtx.Mysql),
	}
}

// GetNewArticles 获取最新文章列表
func (l *GetNewArticlesLogic) GetNewArticles(in *v1.GetNewArticlesRequest) (*v1.Response, error) {
	user, ok := middleware.GetUser(l.ctx)
	if !ok || user.UID == constants.InvalidUserID || user.Status != constants.UserStatusNormal {
		return bad("用户未登录或登录状态异常"), nil
	}

	if in.PageSize <= 0 {
		in.PageSize = 10
	}

	limit := in.PageSize + 1

	var (
		list []*model.Article
		err  error
	)

	if in.Cursor == 0 {
		// 首次查询
		list, err = l.ArticleDao.FindManyWithNotDelete(l.ctx, limit)
	} else {
		// 通过游标查询
		list, err = l.ArticleDao.FindManyWithNotDeleteByCursor(l.ctx, in.Cursor, limit)
	}

	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return notFound("没有更多文章了"), nil
		}
		l.Errorf("GetNewArticlesLogic FindManyWithNotDelete error: %v", err)
		return internal("查询文章失败"), nil
	}

	hasMore := int64(len(list)) > in.PageSize
	if hasMore {
		list = list[:in.PageSize]
	}

	nextCursor := uint64(0)
	if len(list) > 0 {
		nextCursor = list[len(list)-1].Id
	}

	// NOTE: 这里没有获取文章的 tag like collect 等相关内容，前端可以根据文章ID单独请求获取
	articlesPB := convert.PBFromArticle(list)

	res := &v1.GetNewArticlesResponse{
		Articles:   articlesPB,
		HasMore:    hasMore,
		NextCursor: nextCursor,
	}

	resAny, err := anypb.New(res)
	if err != nil {
		l.Errorf("GetNewArticlesLogic anypb.New error: %v", err)
		return internal("查询文章失败"), nil
	}

	return &v1.Response{
		Code:    200,
		Message: "查询文章成功",
		Data:    resAny,
	}, nil
}
