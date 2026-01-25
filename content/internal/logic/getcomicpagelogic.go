package logic

import (
	"context"
	"github.com/luyb177/XiaoAnBackend/content/internal/model"
	"github.com/luyb177/XiaoAnBackend/content/pkg/comic/convert"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/luyb177/XiaoAnBackend/content/internal/svc"
	"github.com/luyb177/XiaoAnBackend/content/pb/content/v1"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetComicPageLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
	ComicChapterDao model.ComicChapterModel
	ComicPageDao    model.ComicPageModel
}

func NewGetComicPageLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetComicPageLogic {
	return &GetComicPageLogic{
		ctx:             ctx,
		svcCtx:          svcCtx,
		Logger:          logx.WithContext(ctx),
		ComicChapterDao: model.NewComicChapterModel(svcCtx.Mysql),
		ComicPageDao:    model.NewComicPageModel(svcCtx.Mysql),
	}
}

// GetComicPage 获取漫画章节页面
func (l *GetComicPageLogic) GetComicPage(in *v1.GetComicPageRequest) (*v1.Response, error) {
	if in.ComicChapterId <= 0 {
		l.Errorf("GetComicPage err: 漫画章节ID不能小于等于0")

		return &v1.Response{
			Code:    400,
			Message: "漫画章节ID不能小于等于0",
		}, nil
	}
	if in.Page <= 0 {
		in.Page = 1
	}
	if in.PageSize <= 0 {
		in.PageSize = 10
	}

	// 获取漫画章节
	offset := (in.Page - 1) * in.PageSize
	pageModels, err := l.ComicPageDao.FindManyByChapterIDOrderByPageNo(l.ctx, in.ComicChapterId, offset, in.PageSize)
	if err != nil {
		l.Errorf("GetComicPage err: %v", err)
		return &v1.Response{
			Code:    500,
			Message: "获取漫画章节页面失败",
		}, nil
	}

	if len(pageModels) == 0 {
		l.Errorf("GetComicPage err: 漫画章节页面不存在")

		return &v1.Response{
			Code:    404,
			Message: "漫画章节页面不存在",
		}, nil
	}

	pagesPB := convert.PBFromComicPages(pageModels)

	res := &v1.GetComicPageResponse{Pages: pagesPB}

	resAny, err := anypb.New(res)
	if err != nil {
		l.Errorf("GetComicPage err: %v", err)

		return &v1.Response{
			Code:    500,
			Message: "获取漫画章节页面失败",
		}, nil
	}

	return &v1.Response{
		Code:    200,
		Message: "获取漫画章节页面成功",
		Data:    resAny,
	}, nil
}
