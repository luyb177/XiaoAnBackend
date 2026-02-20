package logic

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/luyb177/XiaoAnBackend/content/internal/model"
	"github.com/luyb177/XiaoAnBackend/content/internal/svc"
	"github.com/luyb177/XiaoAnBackend/content/pb/content/v1"
	"github.com/luyb177/XiaoAnBackend/content/pkg/taskqueue/tasks"
	"github.com/luyb177/XiaoAnBackend/infra/constants"
	"github.com/luyb177/XiaoAnBackend/infra/middleware"
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
	user, ok := middleware.GetUser(l.ctx)
	if !ok || user.UID == constants.InvalidUserID || (user.Role != constants.SUPERADMIN && user.Role != constants.STAFF) || user.Status != constants.UserStatusNormal {
		return bad("用户未登录或状态异常"), nil
	}

	validations := []Validation{
		{in.Id > 0, "漫画ID不能小于等于0"},
	}
	for _, v := range validations {
		if !v.Condition {
			l.Errorf("DeleteComic err: %s", v.Message)

			return &v1.Response{
				Code:    400,
				Message: v.Message,
			}, nil
		}
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

	deletedAt := uint64(time.Now().Unix())
	modifier := sql.NullInt64{Int64: int64(user.UID), Valid: true}
	err = l.ComicDao.SoftDelete(l.ctx, comic.Id, deletedAt, modifier)
	if err != nil {
		l.Errorf("DeleteComic SoftDelete err: %v", err)

		return &v1.Response{
			Code:    500,
			Message: "删除漫画失败",
		}, nil
	}

	// 删除标签和对应的章节
	comicRelationTask := &tasks.ComicRelationTask{
		Type:    tasks.ComicRelationDelete,
		ComicID: comic.Id,
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
