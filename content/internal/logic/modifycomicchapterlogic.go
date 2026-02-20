package logic

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/luyb177/XiaoAnBackend/content/internal/model"
	"github.com/luyb177/XiaoAnBackend/content/internal/svc"
	"github.com/luyb177/XiaoAnBackend/content/pb/content/v1"
	"github.com/luyb177/XiaoAnBackend/content/pkg/taskqueue/tasks"
	"github.com/luyb177/XiaoAnBackend/infra/constants"
	"github.com/luyb177/XiaoAnBackend/infra/middleware"
)

type ModifyComicChapterLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
	ComicDao        model.ComicModel
	ComicChapterDao model.ComicChapterModel
}

func NewModifyComicChapterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ModifyComicChapterLogic {
	return &ModifyComicChapterLogic{
		ctx:             ctx,
		svcCtx:          svcCtx,
		Logger:          logx.WithContext(ctx),
		ComicDao:        model.NewComicModel(svcCtx.Mysql),
		ComicChapterDao: model.NewComicChapterModel(svcCtx.Mysql),
	}
}

// ModifyComicChapter 修改漫画章节
func (l *ModifyComicChapterLogic) ModifyComicChapter(in *v1.ModifyComicChapterRequest) (*v1.Response, error) {
	user, ok := middleware.GetUser(l.ctx)
	if !ok || user.UID == constants.InvalidUserID || (user.Role != constants.SUPERADMIN && user.Role != constants.STAFF) || user.Status != constants.UserStatusNormal {
		return bad("用户未登录或状态异常"), nil
	}

	// 验证参数
	validations := []Validation{
		{in.Id > 0, "章节ID不能小于等于0"},
		{in.ComicId > 0, "漫画ID不能为小于等于0"},
		{in.ChapterNo > 0, "章节号不能为小于等于0"},
		{in.Title != "", "章节标题不能为空"},
		{in.Description != "", "章节描述不能为空"},
		{in.Status == ComicStatusPublished || in.Status == ComicStatusDraft, "章节状态不合法"},
		{len(in.PageUrls) > 0, "章节页面不能为空"},
	}

	for _, v := range validations {
		if !v.Condition {
			l.Errorf("ModifyComicChapter err: %s", v.Message)
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
			l.Errorf("ModifyComicChapter err: 章节页面URL不能为空")

			return &v1.Response{
				Code:    400,
				Message: "章节页面URL不能为空",
			}, nil
		}
	}

	// 验证漫画章节存在性
	chapter, err := l.ComicChapterDao.FindOneByComicIDAndChapterID(l.ctx, in.ComicId, in.Id)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			l.Errorf("ModifyComicChapter err: 该漫画章节不存在")

			return &v1.Response{
				Code:    404,
				Message: "该漫画章节不存在",
			}, nil
		}
		l.Errorf("ModifyComicChapter err: %v", err)

		return &v1.Response{
			Code:    500,
			Message: "查询漫画章节失败",
		}, nil
	}

	// 3. 修改章节
	chapter.ChapterNo = in.ChapterNo
	chapter.Title = in.Title
	chapter.Description = sql.NullString{String: in.Description, Valid: true}
	chapter.Status = in.Status
	chapter.PublishedAt = time.Unix(in.PublishedAt, 0)
	chapter.LastModifiedBy = sql.NullInt64{Int64: int64(user.UID), Valid: true}
	chapter.RelationStatus = RelationStatusPending
	chapter.PageCount = uint64(len(in.PageUrls))

	err = l.ComicChapterDao.Update(l.ctx, chapter)
	if err != nil {
		l.Errorf("ModifyComicChapter err: %v", err)

		return &v1.Response{
			Code:    400,
			Message: "修改漫画章节失败",
		}, nil
	}

	// 4. 修改章节页面
	comicChapterRelationTask := tasks.ComicChapterRelationTask{
		Type:      tasks.ComicChapterRelationModify,
		ChapterID: chapter.Id,
		PageUrls:  in.PageUrls,
		UID:       user.UID,
	}

	err = l.svcCtx.TaskQueue.Enqueue(l.ctx, &comicChapterRelationTask)
	if err != nil {
		l.Errorf("ModifyComicChapter err: 添加漫画章节页面任务入队失败，%v", err)
	}

	// 5. 返回结果
	res := &v1.ModifyComicChapterResponse{
		Id:             chapter.Id,
		RelationStatus: RelationStatusPending,
	}

	resAny, err := anypb.New(res)
	if err != nil {
		l.Errorf("ModifyComicChapter err: %v", err)

		return &v1.Response{
			Code:    500,
			Message: "封装返回结果失败",
		}, nil
	}

	return &v1.Response{
		Code:    200,
		Message: "修改漫画章节成功",
		Data:    resAny,
	}, nil
}
