package logic

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/luyb177/XiaoAnBackend/content/internal/middleware"
	"github.com/luyb177/XiaoAnBackend/content/internal/model"
	"github.com/luyb177/XiaoAnBackend/content/internal/svc"
	"github.com/luyb177/XiaoAnBackend/content/pb/content/v1"
	"github.com/luyb177/XiaoAnBackend/content/pkg/taskqueue/tasks"
)

type AddComicChapterLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
	ComicDao        model.ComicModel
	ComicChapterDao model.ComicChapterModel
}

func NewAddComicChapterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddComicChapterLogic {
	return &AddComicChapterLogic{
		ctx:             ctx,
		svcCtx:          svcCtx,
		Logger:          logx.WithContext(ctx),
		ComicDao:        model.NewComicModel(svcCtx.Mysql),
		ComicChapterDao: model.NewComicChapterModel(svcCtx.Mysql),
	}
}

// AddComicChapter 添加漫画章节
func (l *AddComicChapterLogic) AddComicChapter(in *v1.AddComicChapterRequest) (*v1.Response, error) {
	user, ok := middleware.GetUser(l.ctx)
	if !ok || user.UID == InvalidUserID || (user.Role != SUPERADMIN && user.Role != STAFF) || user.Status != UserStatusNormal {
		return bad("用户未登录或状态异常"), nil
	}

	// 校验参数
	validations := []Validation{
		{in.ComicId > 0, "漫画ID不能小于等于0"},
		{in.ChapterNo > 0, "章节号不能为小于等于0"},
		{in.Title != "", "章节标题不能为空"},
		{in.Description != "", "章节描述不能为空"},
		{in.Status == ComicStatusPublished || in.Status == ComicStatusDraft, "章节状态不合法"},
		{in.PageUrls != nil && len(in.PageUrls) > 0, "章节页面不能为空"},
	}
	for _, v := range validations {
		if !v.Condition {
			l.Errorf("AddComicChapter err: %s", v.Message)
			return &v1.Response{
				Code:    400,
				Message: v.Message,
			}, nil
		}
	}

	now := time.Now()
	if in.PublishedAt <= 0 {
		in.PublishedAt = now.Unix()
	}

	for _, url := range in.PageUrls {
		if url == "" {
			l.Errorf("AddComicChapter err: 章节页面URL不能为空")

			return &v1.Response{
				Code:    400,
				Message: "章节页面URL不能为空",
			}, nil
		}
	}

	// 验证漫画存在性
	_, err := l.ComicDao.FindOneWithNotDelete(l.ctx, in.ComicId)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			l.Errorf("AddComicChapter err: 漫画不存在")

			return &v1.Response{
				Code:    404,
				Message: "漫画不存在",
			}, nil
		}
		l.Errorf("AddComicChapter err: %v", err)
		return &v1.Response{
			Code:    500,
			Message: "查询漫画失败",
		}, nil
	}

	// 插入漫画章节
	chapter := model.ComicChapter{
		ComicId:        in.ComicId,
		ChapterNo:      in.ChapterNo,
		Title:          in.Title,
		Description:    sql.NullString{String: in.Description, Valid: true},
		PageCount:      uint64(len(in.PageUrls)),
		Status:         in.Status,
		PublishedAt:    time.Unix(in.PublishedAt, 0),
		RelationStatus: RelationStatusPending,
		LastModifiedBy: sql.NullInt64{Int64: int64(user.UID), Valid: true},
	}

	result, err := l.ComicChapterDao.Insert(l.ctx, &chapter)
	if err != nil {
		if errors.Is(err, model.ErrDuplicateEntry) {
			l.Errorf("AddComicChapter err: 该章节号已存在，err: %v", err)
			return &v1.Response{
				Code:    400,
				Message: "该章节号已存在",
			}, nil
		}

		l.Errorf("AddComicChapter err: 插入漫画章节失败，err: %v", err)

		return &v1.Response{
			Code:    500,
			Message: "添加漫画章节失败",
		}, nil
	}
	chapterId, err := result.LastInsertId()
	if err != nil {
		l.Errorf("AddComicChapter err: 获取插入漫画章节ID失败，err: %v", err)

		return &v1.Response{
			Code:    500,
			Message: "获取插入漫画章节ID失败",
		}, nil
	}
	chapter.Id = uint64(chapterId)

	// 4. 添加漫画章节页面
	comicChapterRelationTask := &tasks.ComicChapterRelationTask{
		Type:      tasks.ComicChapterRelationAdd,
		ComicId:   in.ComicId,
		UID:       user.UID,
		ChapterID: chapter.Id,
		PageUrls:  in.PageUrls,
	}
	err = l.svcCtx.TaskQueue.Enqueue(l.ctx, comicChapterRelationTask)
	if err != nil {
		l.Errorf("AddComicChapter err: 添加漫画章节页面任务入队失败，%v", err)
	}

	// 5. 返回结果
	res := &v1.AddComicChapterResponse{
		Id:             chapter.Id,
		RelationStatus: RelationStatusPending,
	}

	resAny, err := anypb.New(res)
	if err != nil {
		l.Errorf("AddComicChapter err: %v", err)

		return &v1.Response{
			Code:    500,
			Message: "封装返回结果失败",
		}, nil
	}

	return &v1.Response{
		Code:    200,
		Message: "添加漫画章节成功",
		Data:    resAny,
	}, nil
}
