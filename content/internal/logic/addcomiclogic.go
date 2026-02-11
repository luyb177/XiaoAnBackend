package logic

import (
	"context"
	"database/sql"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/luyb177/XiaoAnBackend/content/internal/middleware"
	"github.com/luyb177/XiaoAnBackend/content/internal/model"
	"github.com/luyb177/XiaoAnBackend/content/internal/svc"
	"github.com/luyb177/XiaoAnBackend/content/pb/content/v1"
	"github.com/luyb177/XiaoAnBackend/content/pkg/taskqueue/tasks"
)

type AddComicLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
	ComicDao model.ComicModel
}

func NewAddComicLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddComicLogic {
	return &AddComicLogic{
		ctx:      ctx,
		svcCtx:   svcCtx,
		Logger:   logx.WithContext(ctx),
		ComicDao: model.NewComicModel(svcCtx.Mysql),
	}
}

// AddComic 添加漫画
func (l *AddComicLogic) AddComic(in *v1.AddComicRequest) (*v1.Response, error) {
	// 添加漫画只有 超级管理员 和 员工 才能添加
	user, ok := middleware.GetUser(l.ctx)
	if !ok || user.UID == InvalidUserID || (user.Role != SUPERADMIN && user.Role != STAFF) || user.Status != UserStatusNormal {
		return bad("用户未登录或状态异常"), nil
	}

	// 校验参数
	validations := []Validation{
		{in.Name != "", "漫画名称不能为空"},
		{in.Description != "", "漫画描述不能为空"},
		{in.Cover != "", "漫画封面不能为空"},
		{in.Author != "", "漫画作者不能为空"},
		{len(in.Tag) <= 10, "漫画标签不能超过10个"},
	}

	for _, v := range validations {
		if !v.Condition {
			l.Errorf("AddComic err: %s", v.Message)

			return &v1.Response{
				Code:    400,
				Message: v.Message,
			}, nil
		}
	}
	// 添加默认值
	now := time.Now()
	if in.PublishedAt <= 0 {
		in.PublishedAt = now.Unix()
	}
	if len(in.Tag) == 0 {
		in.Tag = []string{"默认标签"}
	}

	// 添加漫画主体
	comic := model.Comic{
		Name:           in.Name,
		Description:    sql.NullString{String: in.Description, Valid: true},
		Cover:          in.Cover,
		Author:         in.Author,
		PublishedAt:    time.Unix(in.PublishedAt, 0),
		RelationStatus: RelationStatusPending,
		LastModifiedBy: sql.NullInt64{Int64: int64(user.UID), Valid: true},
		ChapterCount:   0,
		LikeCount:      0,
		ViewCount:      0,
		CollectCount:   0,
	}

	// 写入数据库
	result, err := l.ComicDao.Insert(l.ctx, &comic)
	if err != nil {
		l.Errorf("AddComic err: 添加漫画主体失败，%v", err)

		return &v1.Response{
			Code:    500,
			Message: "添加漫画主体失败",
		}, nil
	}

	// 回写
	comicID, err := result.LastInsertId()
	if err != nil {
		l.Errorf("AddComic err: 获取漫画ID失败，%v", err)

		return &v1.Response{
			Code:    500,
			Message: "获取漫画ID失败",
		}, nil
	}
	comic.Id = uint64(comicID)

	// 添加标签
	comicRelationTask := tasks.ComicRelationTask{
		Type:    tasks.ComicRelationAdd,
		ComicID: comic.Id,
		UID:     user.UID,
		Tags:    in.Tag,
	}
	err = l.svcCtx.TaskQueue.Enqueue(l.ctx, &comicRelationTask)
	if err != nil {
		l.Errorf("AddComic err: 添加漫画标签任务入队失败，%v", err)
	}

	res := &v1.AddComicResponse{
		Id:             comic.Id,
		RelationStatus: RelationStatusPending,
	}

	resAny, err := anypb.New(res)
	if err != nil {
		l.Errorf("AddComic err: 响应封装失败，%v", err)

		return &v1.Response{
			Code:    500,
			Message: "响应封装失败",
		}, nil
	}

	return &v1.Response{
		Code:    200,
		Message: "添加漫画成功",
		Data:    resAny,
	}, nil
}
