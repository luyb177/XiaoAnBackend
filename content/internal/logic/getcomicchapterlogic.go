package logic

import (
	"context"
	"errors"
	"fmt"
	"github.com/luyb177/XiaoAnBackend/content/internal/model"
	"github.com/luyb177/XiaoAnBackend/content/pkg/comic/convert"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/luyb177/XiaoAnBackend/content/internal/svc"
	"github.com/luyb177/XiaoAnBackend/content/pb/content/v1"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetComicChapterLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
	ComicDao        model.ComicModel
	ComicChapterDao model.ComicChapterModel
}

func NewGetComicChapterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetComicChapterLogic {
	return &GetComicChapterLogic{
		ctx:             ctx,
		svcCtx:          svcCtx,
		Logger:          logx.WithContext(ctx),
		ComicDao:        model.NewComicModel(svcCtx.Mysql),
		ComicChapterDao: model.NewComicChapterModel(svcCtx.Mysql),
	}
}

// GetComicChapter 获取漫画章节
func (l *GetComicChapterLogic) GetComicChapter(in *v1.GetComicChapterRequest) (*v1.Response, error) {
	if in.ComicId <= 0 {
		l.Errorf("GetComicChapter err: 参数错误")

		return &v1.Response{
			Code:    400,
			Message: "参数错误",
		}, nil
	}
	if in.Page <= 0 {
		in.Page = 1
	}
	if in.PageSize <= 0 {
		in.PageSize = 10
	}

	// 看漫画是否存在
	_, err := l.ComicDao.FindOneWithNotDelete(l.ctx, in.ComicId)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			l.Errorf("GetComicChapter err: 漫画不存在")

			return &v1.Response{
				Code:    404,
				Message: "漫画不存在",
			}, nil
		}
		l.Errorf("GetComicChapter err: %v", err)

		return &v1.Response{
			Code:    500,
			Message: "获取漫画失败",
		}, nil
	}

	offset := (in.Page - 1) * in.PageSize

	chapterModels, err := l.ComicChapterDao.FindManyByComicIDOrderByChapterNo(l.ctx, in.ComicId, offset, in.PageSize)
	if err != nil {
		l.Errorf("GetComicChapter err: %v", err)

		return &v1.Response{
			Code:    500,
			Message: "获取漫画章节失败",
		}, nil
	}

	type MsgResult struct {
		msg string
	}

	msgCh := make(chan MsgResult, 1)

	go func() {
		msg := "获取漫画章节成功"
		for _, chapter := range chapterModels {
			if chapter.RelationStatus == RelationStatusPending {
				msg = fmt.Sprintf("章节%d页面图片正在关联中，请稍后刷新查看", chapter.ChapterNo)
				break
			}
		}
		msgCh <- MsgResult{msg: msg}
	}()

	comicChaptersPB := convert.PBFromComicChapter(chapterModels)

	res := &v1.GetComicChapterResponse{Chapters: comicChaptersPB}

	resAny, err := anypb.New(res)
	if err != nil {
		l.Errorf("GetComicChapter err: %v", err)

		return &v1.Response{
			Code:    500,
			Message: "获取漫画章节失败",
		}, nil
	}

	msgResult := <-msgCh

	return &v1.Response{
		Code:    200,
		Message: msgResult.msg,
		Data:    resAny,
	}, nil
}
