package worker

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/luyb177/XiaoAnBackend/content/internal/logic"
	"github.com/luyb177/XiaoAnBackend/content/internal/model"
	"github.com/luyb177/XiaoAnBackend/content/internal/repo/redisqueue"
	"github.com/luyb177/XiaoAnBackend/content/internal/svc"
	"github.com/luyb177/XiaoAnBackend/content/pkg/taskqueue"
	"github.com/luyb177/XiaoAnBackend/content/pkg/taskqueue/tasks"
)

type LikeRelationHandler struct {
	logx.Logger
	svcCtx         *svc.ServiceContext
	ContentLikeDao model.ContentLikeModel

	ArticleDao model.ArticleModel
	PodcastDao model.PodcastModel
	VideoDao   model.VideoModel
	ComicDao   model.ComicModel
	CommentDao model.CommentModel
}

func NewLikeRelationHandler(svcCtx *svc.ServiceContext, ctx context.Context) *LikeRelationHandler {
	return &LikeRelationHandler{
		svcCtx:         svcCtx,
		Logger:         logx.WithContext(ctx),
		ContentLikeDao: model.NewContentLikeModel(svcCtx.Mysql),
		ArticleDao:     model.NewArticleModel(svcCtx.Mysql),
		PodcastDao:     model.NewPodcastModel(svcCtx.Mysql),
		VideoDao:       model.NewVideoModel(svcCtx.Mysql),
		ComicDao:       model.NewComicModel(svcCtx.Mysql),
		CommentDao:     model.NewCommentModel(svcCtx.Mysql),
	}
}

func (h *LikeRelationHandler) Handle(ctx context.Context, task taskqueue.Task) error {
	payload, err := task.Payload()
	if err != nil {
		return err
	}

	var rawTask redisqueue.RawTask
	err = json.Unmarshal(payload, &rawTask)
	if err != nil {
		return err
	}

	likeTask := tasks.LikeRelationTask{}
	err = json.Unmarshal(rawTask.Data, &likeTask)
	if err != nil {
		return err
	}

	h.Infof("processing like relation task: %+v", likeTask)

	switch likeTask.Type {
	case tasks.LikeRelationAdd:
		return h.handleAdd(ctx, &likeTask)
	case tasks.LikeRelationDelete:
		return h.handleDelete(ctx, &likeTask)
	default:
		h.Errorf("unknown like relation task type: %s", likeTask.Type)
		return nil
	}
}

func (h *LikeRelationHandler) handleAdd(ctx context.Context, task *tasks.LikeRelationTask) error {
	return h.svcCtx.Mysql.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
		contentLike := &model.ContentLike{
			Type:      task.ContentType,
			TargetId:  task.ContentID,
			UserId:    task.UID,
			IsCounted: 0,
			DeletedAt: 0,
		}
		// 确保存在性
		_, err := h.ContentLikeDao.UpsertWithSession(ctx, session, contentLike)
		if err != nil {
			return err
		}
		// 获取ID
		contentLike, err = h.ContentLikeDao.FindOneByUserIdTypeTargetIdWithSession(ctx, session, task.UID, task.ContentType, task.ContentID)
		if err != nil {
			return err
		}

		// 增加对应的点赞数据 标记
		result, err := h.ContentLikeDao.MarkLikeAsCountedWithSession(ctx, session, contentLike.Id)
		if err != nil {
			return err
		}
		affect, err := result.RowsAffected()
		if err != nil {
			return err
		}
		if affect == 0 {
			// 已经计数过，直接返回
			return nil
		}

		// 内容 like_count + 1
		result, err = h.incrContentLikeCount(ctx, session, task.ContentType, task.ContentID)
		if err != nil {
			return err
		}
		affect, err = result.RowsAffected()
		if err != nil {
			return err
		}
		if affect == 0 {
			// 此时没有对应内容，删除 like
			_, err = h.ContentLikeDao.UnmarkLikeAsCountedWithSession(ctx, session, contentLike.Id)
			if err != nil {
				return err
			}
			_, err = h.ContentLikeDao.SoftDeleteWithSession(ctx, session, contentLike.Id, uint64(time.Now().Unix()))
			return err
		}

		return nil
	})
}

func (h *LikeRelationHandler) handleDelete(ctx context.Context, task *tasks.LikeRelationTask) error {
	return h.svcCtx.Mysql.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
		// 先 find
		contentLike, err := h.ContentLikeDao.FindOneByUserIdTypeTargetIdWithSession(ctx, session, task.UID, task.ContentType, task.ContentID)
		if err != nil {
			if errors.Is(err, model.ErrNotFound) {
				return nil
			}
			return err
		}

		// 标记
		result, err := h.ContentLikeDao.UnmarkLikeAsCountedWithSession(ctx, session, contentLike.Id)
		if err != nil {
			return err
		}
		affect, err := result.RowsAffected()
		if err != nil {
			return err
		}
		if affect > 0 {
			// 计数过
			// 内容 like - 1
			result, err = h.decrContentLikeCount(ctx, session, task.ContentType, task.ContentID)
			if err != nil {
				return err
			}
			affect, err = result.RowsAffected()
			if err != nil {
				return err
			}
			if affect == 0 {
				// 内容不存在
			}
		}

		// 删除
		_, err = h.ContentLikeDao.SoftDeleteWithSession(ctx, session, contentLike.Id, uint64(time.Now().Unix()))
		return err
	})
}

func (h *LikeRelationHandler) incrContentLikeCount(ctx context.Context, session sqlx.Session, contentType string, contentID uint64) (sql.Result, error) {
	var result sql.Result
	var err error
	switch contentType {
	case logic.ContentTypeArticle:
		result, err = h.ArticleDao.IncrLikeCountWithSession(ctx, session, contentID)
	case logic.ContentTypePodcast:
		result, err = h.PodcastDao.IncrLikeCountWithSession(ctx, session, contentID)
	case logic.ContentTypeVideo:
		result, err = h.VideoDao.IncrLikeCountWithSession(ctx, session, contentID)
	case logic.ContentTypeComic:
		result, err = h.ComicDao.IncrLikeCountWithSession(ctx, session, contentID)
	case logic.ContentTypeComment:
		result, err = h.CommentDao.IncrLikeCountWithSession(ctx, session, contentID)
	default:
		return nil, errors.New("unknown content type")
	}
	return result, err
}

func (h *LikeRelationHandler) decrContentLikeCount(ctx context.Context, session sqlx.Session, contentType string, contentID uint64) (sql.Result, error) {
	var result sql.Result
	var err error
	switch contentType {
	case logic.ContentTypeArticle:
		result, err = h.ArticleDao.DecrLikeCountWithSession(ctx, session, contentID)
	case logic.ContentTypePodcast:
		result, err = h.PodcastDao.DecrLikeCountWithSession(ctx, session, contentID)
	case logic.ContentTypeVideo:
		result, err = h.VideoDao.DecrLikeCountWithSession(ctx, session, contentID)
	case logic.ContentTypeComic:
		result, err = h.ComicDao.DecrLikeCountWithSession(ctx, session, contentID)
	case logic.ContentTypeComment:
		result, err = h.CommentDao.DecrLikeCountWithSession(ctx, session, contentID)
	default:
		return nil, errors.New("unknown content type")
	}
	return result, err
}
