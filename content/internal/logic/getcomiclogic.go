package logic

import (
	"context"
	"errors"
	"github.com/luyb177/XiaoAnBackend/content/pkg/taskqueue/tasks"

	"github.com/luyb177/XiaoAnBackend/content/internal/middleware"
	"github.com/luyb177/XiaoAnBackend/content/internal/model"
	"github.com/luyb177/XiaoAnBackend/content/internal/svc"
	"github.com/luyb177/XiaoAnBackend/content/pb/content/v1"
	"github.com/luyb177/XiaoAnBackend/content/pkg/comic/convert"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/protobuf/types/known/anypb"
)

type GetComicLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
	ComicDao    model.ComicModel
	ComicTagDao model.ComicTagModel
}

func NewGetComicLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetComicLogic {
	return &GetComicLogic{
		ctx:         ctx,
		svcCtx:      svcCtx,
		Logger:      logx.WithContext(ctx),
		ComicDao:    model.NewComicModel(svcCtx.Mysql),
		ComicTagDao: model.NewComicTagModel(svcCtx.Mysql),
	}
}

// GetComic 获取漫画
func (l *GetComicLogic) GetComic(in *v1.GetComicRequest) (*v1.Response, error) {
	user, ok := middleware.GetUser(l.ctx)
	if !ok || user.UID == InvalidUserID || user.Status != UserStatusNormal {
		return bad("用户未登录或登录状态异常"), nil
	}

	if in.Id == 0 {
		return bad("漫画ID不合法"), nil
	}

	// 获取漫画
	comic, err := l.ComicDao.FindOneWithNotDelete(l.ctx, in.Id)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			l.Errorf("GetComic err: 漫画不存在")

			return &v1.Response{
				Code:    404,
				Message: "漫画不存在",
			}, nil
		}
		l.Errorf("GetComic err: %v", err)

		return internal("获取漫画失败"), nil
	}

	// 入队
	comicRelationTask := &tasks.ComicRelationTask{
		Type:    tasks.ComicRelationGet,
		ComicID: in.Id,
		UID:     user.UID,
		Tags:    nil,
	}

	err = l.svcCtx.TaskQueue.Enqueue(l.ctx, comicRelationTask)
	if err != nil {
		l.Errorf("GetComic err: 入队获取漫画相关内容失败, %v", err)
	}

	// 异步获取tag like
	type TagResult struct {
		tags []*model.ComicTag
		err  error
	}
	type LikeResult struct {
		liked bool
		err   error
	}
	type CollectResult struct {
		collected bool
		err       error
	}

	tagCh := make(chan TagResult, 1)
	likeCh := make(chan LikeResult, 1)
	collectCh := make(chan CollectResult, 1)

	go func() {
		t, err := l.ComicTagDao.FindManyByComicId(l.ctx, in.Id)
		tagCh <- TagResult{
			tags: t,
			err:  err,
		}
	}()

	go func() {
		liked, err := l.svcCtx.LikeRepo.HasLiked(l.ctx, user.UID, ContentTypeComic, in.Id)
		likeCh <- LikeResult{
			liked: liked,
			err:   err,
		}
	}()

	go func() {
		collected, err := l.svcCtx.CollectRepo.HasCollect(l.ctx, user.UID, ContentTypeComic, in.Id)
		collectCh <- CollectResult{collected: collected, err: err}
	}()

	// 等待tag结果
	tagResult := <-tagCh
	if tagResult.err != nil {
		l.Errorf("GetComic err: 获取漫画标签失败, %v", tagResult.err)
	}

	tags := convert.StringsFromComicTags(tagResult.tags)

	// 等待 like 结果
	likeResult := <-likeCh
	if likeResult.err != nil {
		l.Errorf("GetComic err: 获取漫画点赞情况失败,%v", likeResult.err)
	}

	// 等待 collect 结果
	collectRes := <-collectCh
	if collectRes.err != nil {
		l.Errorf("GetComic err: %v", collectRes.err)
	}

	res := &v1.GetComicResponse{Comic: &v1.Comic{
		Id:             comic.Id,
		Name:           comic.Name,
		Tag:            tags,
		Description:    comic.Description.String,
		Cover:          comic.Cover,
		Author:         comic.Author,
		PublishedAt:    comic.PublishedAt.Unix(),
		CreatedAt:      comic.CreatedAt.Unix(),
		UpdatedAt:      comic.UpdatedAt.Unix(),
		LikeCount:      comic.LikeCount,
		ViewCount:      comic.ViewCount,
		CollectCount:   comic.CollectCount,
		ChapterCount:   comic.ChapterCount,
		CommentCount:   comic.CommentCount,
		LastModifiedBy: comic.LastModifiedBy.Int64,
		RelationStatus: comic.RelationStatus,
		IsLiked:        likeResult.liked,
		IsCollected:    collectRes.collected,
	}}

	reaAny, err := anypb.New(res)
	if err != nil {
		l.Errorf("GetComic err: anypb.New failed, %v", err)

		return &v1.Response{
			Code:    500,
			Message: "封装相应失败",
		}, nil
	}

	msg := "获取漫画成功"

	if comic.RelationStatus == RelationStatusPending {
		msg = "漫画相关内容正在关联中，请稍后刷新查看"
	}

	return &v1.Response{
		Code:    200,
		Message: msg,
		Data:    reaAny,
	}, nil
}
