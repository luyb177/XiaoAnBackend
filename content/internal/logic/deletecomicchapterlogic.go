package logic

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/luyb177/XiaoAnBackend/content/internal/middleware"
	"github.com/luyb177/XiaoAnBackend/content/internal/model"
	"github.com/luyb177/XiaoAnBackend/content/internal/svc"
	"github.com/luyb177/XiaoAnBackend/content/pb/content/v1"
	"github.com/luyb177/XiaoAnBackend/content/pkg/taskqueue/tasks"
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
	user, ok := middleware.GetUser(l.ctx)
	if !ok || user.UID == InvalidUserID || (user.Role != SUPERADMIN && user.Role != STAFF) || user.Status != UserStatusNormal {
		return bad("用户未登录或状态异常"), nil
	}

	validations := []Validation{
		{in.Id > 0, "章节ID不能小于等于0"},
		{in.ComicId > 0, "漫画ID不能小于等于0"},
	}

	for _, v := range validations {
		if !v.Condition {
			l.Errorf("DeleteComicChapter err: %s", v.Message)

			return &v1.Response{
				Code:    400,
				Message: v.Message,
			}, nil
		}
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
	comicChapter, err := l.ComicChapterDao.FindOneByComicIDAndChapterID(l.ctx, in.ComicId, in.Id)
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
		ComicID:   in.ComicId,
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
