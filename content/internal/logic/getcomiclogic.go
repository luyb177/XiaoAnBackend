package logic

import (
	"context"
	"errors"

	"github.com/luyb177/XiaoAnBackend/content/internal/model"
	"github.com/luyb177/XiaoAnBackend/content/internal/svc"
	"github.com/luyb177/XiaoAnBackend/content/pb/content/v1"
	"github.com/luyb177/XiaoAnBackend/content/pkg/comic/convert"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/protobuf/types/known/anypb"
)

type GetComicLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
	ComicDao    model.ComicModel
	ComicTagDao model.ComicTagModel
}

func NewGetComicLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetComicLogic {
	return &GetComicLogic{
		ctx:         ctx,
		svcCtx:      svcCtx,
		Logger:      logx.WithContext(ctx),
		ComicDao:    model.NewComicModel(svcCtx.Mysql),
		ComicTagDao: model.NewComicTagModel(svcCtx.Mysql),
	}
}

// GetComic 获取漫画
func (l *GetComicLogic) GetComic(in *v1.GetComicRequest) (*v1.Response, error) {
	if in.Id <= 0 {
		l.Errorf("GetComic err: 漫画ID不能小于等于0")

		return &v1.Response{
			Code:    400,
			Message: "漫画ID不能小于等于0",
		}, nil
	}

	// 异步获取tag
	type TagResult struct {
		tags []*model.ComicTag
		err  error
	}
	tagCh := make(chan TagResult, 1)

	go func() {
		t, err := l.ComicTagDao.FindManyByComicId(l.ctx, in.Id)
		tagCh <- TagResult{
			tags: t,
			err:  err,
		}
	}()

	// 获取漫画
	comic, err := l.ComicDao.FindOneWithNotDelete(l.ctx, in.Id)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			l.Errorf("GetComic err: 漫画不存在")

			return &v1.Response{
				Code:    404,
				Message: "漫画不存在",
			}, nil
		}
		l.Errorf("GetComic err: %v", err)

		return &v1.Response{
			Code:    500,
			Message: "获取漫画失败",
		}, nil
	}

	// 等待tag结果
	tagResult := <-tagCh
	if tagResult.err != nil {
		l.Errorf("GetComic err: 获取漫画标签失败, %v", tagResult.err)
	}

	tags := convert.StringsFromComicTags(tagResult.tags)

	res := &v1.GetComicResponse{Comic: &v1.Comic{
		Id:           comic.Id,
		Name:         comic.Name,
		Tag:          tags,
		Description:  comic.Description.String,
		Cover:        comic.Cover,
		Author:       comic.Author,
		PublishedAt:  comic.PublishedAt.Unix(),
		CreatedAt:    comic.CreatedAt.Unix(),
		UpdatedAt:    comic.UpdatedAt.Unix(),
		LikeCount:    comic.LikeCount,
		ViewCount:    comic.ViewCount,
		CollectCount: comic.CollectCount,
		ChapterCount: comic.ChapterCount,
	}}

	reaAny, err := anypb.New(res)
	if err != nil {
		l.Errorf("GetComic err: anypb.New failed, %v", err)

		return &v1.Response{
			Code:    500,
			Message: "封装相应失败",
		}, nil
	}

	msg := "获取漫画成功"

	if comic.RelationStatus == RelationStatusPending {
		msg = "漫画相关内容正在关联中，请稍后刷新查看"
	}

	return &v1.Response{
		Code:    200,
		Message: msg,
		Data:    reaAny,
	}, nil
}
