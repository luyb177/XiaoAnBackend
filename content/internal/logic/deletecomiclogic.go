package logic

import (
	"context"
	"database/sql"
	"errors"
	"github.com/luyb177/XiaoAnBackend/content/internal/middleware"
	"github.com/luyb177/XiaoAnBackend/content/internal/model"
	"github.com/luyb177/XiaoAnBackend/content/pkg/taskqueue/tasks"
	"time"

	"github.com/luyb177/XiaoAnBackend/content/internal/svc"
	"github.com/luyb177/XiaoAnBackend/content/pb/content/v1"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteComicLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
	ComicDao model.ComicModel
}

func NewDeleteComicLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteComicLogic {
	return &DeleteComicLogic{
		ctx:      ctx,
		svcCtx:   svcCtx,
		Logger:   logx.WithContext(ctx),
		ComicDao: model.NewComicModel(svcCtx.Mysql),
	}
}

// DeleteComic 删除漫画
func (l *DeleteComicLogic) DeleteComic(in *v1.DeleteComicRequest) (*v1.Response, error) {
	user := middleware.MustGetUser(l.ctx)
	if user.UID == InvalidUserID || (user.Role != SUPERADMIN && user.Role != STAFF) || user.Status != UserStatusNormal {
		l.Errorf("DeleteComic  err: 用户未登录或登录状态异常")

		return &v1.Response{
			Code:    400,
			Message: "用户未登录或登录状态异常",
		}, nil
	}

	if in.Id <= 0 {
		l.Errorf("DeleteComic err: 参数错误")

		return &v1.Response{
			Code:    400,
			Message: "参数错误",
		}, nil
	}

	// 软删除漫画
	comic, err := l.ComicDao.FindOneWithNotDelete(l.ctx, in.Id)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			l.Errorf("DeleteComic err: 漫画不存在")

			return &v1.Response{
				Code:    404,
				Message: "漫画不存在",
			}, nil
		}
		l.Errorf("DeleteComic err: %v", err)

		return &v1.Response{
			Code:    500,
			Message: "删除漫画失败",
		}, nil
	}

	comic.LastModifiedBy = sql.NullInt64{
		Int64: int64(user.UID),
		Valid: true,
	}
	comic.DeletedAt = uint64(time.Now().Unix())

	err = l.ComicDao.Update(l.ctx, comic)
	if err != nil {
		l.Errorf("DeleteComic err: %v", err)

		return &v1.Response{
			Code:    500,
			Message: "删除漫画失败",
		}, nil
	}

	// 删除标签和对应的章节
	comicRelationTask := &tasks.ComicRelationTask{
		Type:    tasks.ComicRelationDelete,
		ComicID: comic.Id,
		Tags:    nil,
		UID:     user.UID,
	}

	err = l.svcCtx.TaskQueue.Enqueue(l.ctx, comicRelationTask)
	if err != nil {
		l.Errorf("DeleteComic Enqueue err: %v", err)
	}

	return &v1.Response{
		Code:    200,
		Message: "删除漫画成功",
	}, nil
}
