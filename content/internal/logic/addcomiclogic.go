package logic

import (
	"context"
	"database/sql"
	"time"

	"github.com/luyb177/XiaoAnBackend/content/internal/middleware"
	"github.com/luyb177/XiaoAnBackend/content/internal/model"
	"github.com/luyb177/XiaoAnBackend/content/internal/svc"
	"github.com/luyb177/XiaoAnBackend/content/pb/content/v1"
	"github.com/luyb177/XiaoAnBackend/content/pkg/taskqueue/tasks"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/protobuf/types/known/anypb"
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
	user := middleware.MustGetUser(l.ctx)
	if user.UID == InvalidUserID || (user.Role != SUPERADMIN && user.Role != STAFF) || user.Status != UserStatusNormal {
		l.Errorf("AddComic err: 用户未登录或无权限")

		return &v1.Response{
			Code:    400,
			Message: "用户未登录或无权限",
		}, nil
	}

	// 校验参数
	if in.Name == "" {
		l.Errorf("AddComic err: 漫画名称不能为空")

		return &v1.Response{
			Code:    400,
			Message: "漫画名称不能为空",
		}, nil
	}
	if in.Description == "" {
		l.Errorf("AddComic err: 漫画描述不能为空")

		return &v1.Response{
			Code:    400,
			Message: "漫画描述不能为空",
		}, nil
	}
	if in.Cover == "" {
		l.Errorf("AddComic err: 漫画封面不能为空")

		return &v1.Response{
			Code:    400,
			Message: "漫画封面不能为空",
		}, nil
	}
	if in.Author == "" {
		l.Errorf("AddComic err: 漫画作者不能为空")

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
		l.Errorf("AddComic err: 标签数量不能超过10个")

		return &v1.Response{
			Code:    400,
			Message: "标签数量不能超过10个",
		}, nil
	}

	// 添加漫画主体
	// 1. 构造
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

	// 2. 写入数据库
	result, err := l.ComicDao.Insert(l.ctx, &comic)
	if err != nil {
		l.Errorf("AddComic err: 添加漫画主体失败，%v", err)

		return &v1.Response{
			Code:    500,
			Message: "添加漫画主体失败",
		}, nil
	}

	// 3. 回写
	comicId, err := result.LastInsertId()
	if err != nil {
		l.Errorf("AddComic err: 获取漫画ID失败，%v", err)

		return &v1.Response{
			Code:    500,
			Message: "获取漫画ID失败",
		}, nil
	}
	comic.Id = uint64(comicId)

	// 4. 添加标签
	comicRelationTask := tasks.ComicRelationTask{
		Type:    tasks.ComicRelationAdd,
		ComicID: comic.Id,
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
