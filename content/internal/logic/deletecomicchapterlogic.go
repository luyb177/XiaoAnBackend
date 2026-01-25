package logic

import (
	"context"
	"database/sql"
	"errors"
	"github.com/luyb177/XiaoAnBackend/content/internal/middleware"
	"github.com/luyb177/XiaoAnBackend/content/internal/model"
	"github.com/luyb177/XiaoAnBackend/content/internal/svc"
	"github.com/luyb177/XiaoAnBackend/content/pb/content/v1"
	"github.com/luyb177/XiaoAnBackend/content/pkg/taskqueue/tasks"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteComicChapterLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
	ComicDao        model.ComicModel
	ComicChapterDao model.ComicChapterModel
}

func NewDeleteComicChapterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteComicChapterLogic {
	return &DeleteComicChapterLogic{
		ctx:             ctx,
		svcCtx:          svcCtx,
		Logger:          logx.WithContext(ctx),
		ComicDao:        model.NewComicModel(svcCtx.Mysql),
		ComicChapterDao: model.NewComicChapterModel(svcCtx.Mysql),
	}
}

// DeleteComicChapter 删除漫画章节
func (l *DeleteComicChapterLogic) DeleteComicChapter(in *v1.DeleteComicChapterRequest) (*v1.Response, error) {
	user := middleware.MustGetUser(l.ctx)
	if user.UID == InvalidUserID || (user.Role != SUPERADMIN && user.Role != STAFF) || user.Status != UserStatusNormal {
		l.Errorf("DeleteComicChapter  err: 用户未登录或登录状态异常")

		return &v1.Response{
			Code:    400,
			Message: "用户未登录或登录状态异常",
		}, nil
	}

	if in.Id <= 0 {
		l.Errorf("DeleteComicChapter err: 参数错误")

		return &v1.Response{
			Code:    400,
			Message: "参数错误",
		}, nil
	}
	if in.ComicId <= 0 {
		l.Errorf("DeleteComicChapter err: 参数错误")

		return &v1.Response{
			Code:    400,
			Message: "参数错误",
		}, nil
	}

	// 检查漫画是否存在
	_, err := l.ComicDao.FindOneWithNotDelete(l.ctx, in.ComicId)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			l.Errorf("DeleteComicChapter err: 漫画不存在")

			return &v1.Response{
				Code:    404,
				Message: "漫画不存在",
			}, nil
		}
		l.Errorf("DeleteComicChapter err: %v", err)

		return &v1.Response{
			Code:    500,
			Message: "删除漫画章节失败",
		}, nil
	}

	// 软删除漫画章节
	comicChapter, err := l.ComicChapterDao.FindOneWithNotDelete(l.ctx, in.Id)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			l.Errorf("DeleteComicChapter err: 漫画章节不存在")

			return &v1.Response{
				Code:    404,
				Message: "漫画章节不存在",
			}, nil
		}
		l.Errorf("DeleteComicChapter err: %v", err)

		return &v1.Response{
			Code:    500,
			Message: "删除漫画章节失败",
		}, nil
	}

	deletedAt := uint64(time.Now().Unix())
	modifier := sql.NullInt64{Int64: int64(user.UID), Valid: true}
	err = l.ComicChapterDao.SoftDelete(l.ctx, comicChapter.Id, deletedAt, modifier)
	if err != nil {
		return &v1.Response{
			Code:    500,
			Message: "删除漫画章节失败",
		}, nil
	}

	// 删除对应章节的内容
	comicChapterRelationTask := &tasks.ComicChapterRelationTask{
		Type:      tasks.ComicChapterRelationDelete,
		ComicId:   in.ComicId,
		UID:       user.UID,
		ChapterID: comicChapter.Id,
	}
	err = l.svcCtx.TaskQueue.Enqueue(l.ctx, comicChapterRelationTask)
	if err != nil {
		l.Errorf("DeleteComicChapter Enqueue err: %v", err)
	}

	return &v1.Response{
		Code:    200,
		Message: "删除漫画章节成功",
	}, nil
}
