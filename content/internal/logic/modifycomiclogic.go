package logic

import (
	"context"
	"database/sql"
	"errors"
	"github.com/luyb177/XiaoAnBackend/content/internal/middleware"
	"github.com/luyb177/XiaoAnBackend/content/internal/model"
	"github.com/luyb177/XiaoAnBackend/content/pkg/taskqueue/tasks"
	"google.golang.org/protobuf/types/known/anypb"
	"time"

	"github.com/luyb177/XiaoAnBackend/content/internal/svc"
	"github.com/luyb177/XiaoAnBackend/content/pb/content/v1"

	"github.com/zeromicro/go-zero/core/logx"
)

type ModifyComicLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
	ComicDao model.ComicModel
}

func NewModifyComicLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ModifyComicLogic {
	return &ModifyComicLogic{
		ctx:      ctx,
		svcCtx:   svcCtx,
		Logger:   logx.WithContext(ctx),
		ComicDao: model.NewComicModel(svcCtx.Mysql),
	}
}

// ModifyComic 修改漫画
func (l *ModifyComicLogic) ModifyComic(in *v1.ModifyComicRequest) (*v1.Response, error) {
	user := middleware.MustGetUser(l.ctx)
	if user.UID == InvalidUserID || (user.Role != SUPERADMIN && user.Role != STAFF) || user.Status != UserStatusNormal {
		l.Logger.Errorf("ModifyComic err: 用户未登录或者没有权限")

		return &v1.Response{
			Code:    400,
			Message: "用户未登录或者没有权限",
		}, nil
	}

	// 验证参数
	if in.Id <= 0 {
		l.Logger.Errorf("ModifyComic err: 漫画ID不能小于等于0")

		return &v1.Response{
			Code:    400,
			Message: "漫画ID不能小于等于0",
		}, nil
	}
	if in.Name == "" {
		l.Errorf("ModifyComic err: 漫画名称不能为空")

		return &v1.Response{
			Code:    400,
			Message: "漫画名称不能为空",
		}, nil
	}
	if in.Description == "" {
		l.Errorf("ModifyComic err: 漫画描述不能为空")

		return &v1.Response{
			Code:    400,
			Message: "漫画描述不能为空",
		}, nil
	}
	if in.Cover == "" {
		l.Errorf("ModifyComic err: 漫画封面不能为空")

		return &v1.Response{
			Code:    400,
			Message: "漫画封面不能为空",
		}, nil
	}
	if in.Author == "" {
		l.Errorf("ModifyComic err: 漫画作者不能为空")

		return &v1.Response{
			Code:    400,
			Message: "漫画作者不能为空",
		}, nil
	}
	now := time.Now()
	if in.PublishedAt <= 0 {
		in.PublishedAt = now.Unix()
	}
	if in.Tag == nil || len(in.Tag) == 0 {
		in.Tag = []string{"默认标签"}
	}
	if len(in.Tag) > 10 {
		l.Errorf("ModifyComic err: 标签数量不能超过10个")

		return &v1.Response{
			Code:    400,
			Message: "标签数量不能超过10个",
		}, nil
	}

	// 1. 先查询漫画是否存在
	comic, err := l.ComicDao.FindOneWithNotDelete(l.ctx, in.Id)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			l.Logger.Errorf("ModifyComic err: 漫画不存在")

			return &v1.Response{
				Code:    404,
				Message: "漫画不存在",
			}, nil
		}
		l.Logger.Errorf("ModifyComic err: %v", err)

		return &v1.Response{
			Code:    500,
			Message: "查询漫画失败",
		}, nil
	}

	// 2. 更新漫画
	comic.Name = in.Name
	comic.Description = sql.NullString{String: in.Description, Valid: true}
	comic.Cover = in.Cover
	comic.Author = in.Author
	comic.PublishedAt = time.Unix(in.PublishedAt, 0)
	comic.RelationStatus = RelationStatusPending
	comic.LastModifiedBy = sql.NullInt64{Int64: int64(user.UID), Valid: true}

	err = l.ComicDao.Update(l.ctx, comic)
	if err != nil {
		l.Logger.Errorf("ModifyComic err: %v", err)

		return &v1.Response{
			Code:    500,
			Message: "修改漫画失败",
		}, nil
	}

	// 3. 修改标签
	comicRelationTask := &tasks.ComicRelationTask{
		Type:    tasks.ComicRelationModify,
		ComicID: comic.Id,
		Tags:    in.Tag,
	}

	err = l.svcCtx.TaskQueue.Enqueue(l.ctx, comicRelationTask)
	if err != nil {
		l.Logger.Errorf("ModifyComic Enqueue err: %v", err)
	}

	res := &v1.ModifyComicResponse{
		Id:             comic.Id,
		RelationStatus: RelationStatusPending,
	}

	resAny, err := anypb.New(res)
	if err != nil {
		l.Logger.Errorf("ModifyComic err: %v", err)

		return &v1.Response{
			Code:    500,
			Message: "构造返回值失败",
		}, nil
	}

	return &v1.Response{
		Code:    200,
		Message: "修改漫画成功",
		Data:    resAny,
	}, nil
}
