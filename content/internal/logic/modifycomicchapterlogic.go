package logic

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/luyb177/XiaoAnBackend/content/internal/middleware"
	"github.com/luyb177/XiaoAnBackend/content/internal/model"
	"github.com/luyb177/XiaoAnBackend/content/internal/svc"
	"github.com/luyb177/XiaoAnBackend/content/pb/content/v1"
	"github.com/luyb177/XiaoAnBackend/content/pkg/taskqueue/tasks"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/protobuf/types/known/anypb"
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
	user := middleware.MustGetUser(l.ctx)
	if user.UID == InvalidUserID || (user.Role != SUPERADMIN && user.Role != STAFF) || user.Status != UserStatusNormal {
		l.Logger.Errorf("ModifyComicChapter err: 用户未登录或者没有权限")

		return &v1.Response{
			Code:    400,
			Message: "用户未登录或者没有权限",
		}, nil
	}

	// 验证参数
	if in.Id <= 0 {
		l.Errorf("AddComicChapter err: 章节ID不能为空")

		return &v1.Response{
			Code:    400,
			Message: "章节ID不能为空",
		}, nil
	}
	if in.ComicId <= 0 {
		l.Errorf("AddComicChapter err: 漫画ID不能为空")

		return &v1.Response{
			Code:    400,
			Message: "漫画ID不能为空",
		}, nil
	}
	if in.ChapterNo <= 0 {
		l.Errorf("AddComicChapter err: 章节号不能为空")

		return &v1.Response{
			Code:    400,
			Message: "章节号不能为空",
		}, nil
	}
	if in.Title == "" {
		l.Errorf("AddComicChapter err: 章节标题不能为空")

		return &v1.Response{
			Code:    400,
			Message: "章节标题不能为空",
		}, nil
	}
	if in.Description == "" {
		l.Errorf("AddComicChapter err: 章节描述不能为空")

		return &v1.Response{
			Code:    400,
			Message: "章节描述不能为空",
		}, nil
	}
	if in.Status != ComicStatusPublished && in.Status != ComicStatusDraft {
		l.Errorf("AddComicChapter err: 章节状态不合法")

		return &v1.Response{
			Code:    400,
			Message: "章节状态不合法",
		}, nil
	}
	now := time.Now()
	if in.PublishedAt <= 0 {
		in.PublishedAt = now.Unix()
	}
	if in.PageUrls == nil || len(in.PageUrls) == 0 {
		l.Errorf("AddComicChapter err: 章节页面不能为空")

		return &v1.Response{
			Code:    400,
			Message: "章节页面不能为空",
		}, nil
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

	// 1. 验证漫画存在性
	_, err := l.ComicDao.FindOneWithNotDelete(l.ctx, in.ComicId)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			l.Errorf("ModifyComicChapter err: 漫画不存在")

			return &v1.Response{
				Code:    404,
				Message: "漫画不存在",
			}, nil
		}
		l.Errorf("ModifyComicChapter err: %v", err)

		return &v1.Response{
			Code:    500,
			Message: "修改漫画章节失败",
		}, nil
	}

	// 2. 验证章节存在性
	chapter, err := l.ComicChapterDao.FindOneWithNotDelete(l.ctx, in.Id)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			l.Errorf("ModifyComicChapter err: 章节不存在")

			return &v1.Response{
				Code:    404,
				Message: "章节不存在",
			}, nil
		}
		l.Errorf("ModifyComicChapter err: %v", err)

		return &v1.Response{
			Code:    500,
			Message: "修改漫画章节失败",
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
