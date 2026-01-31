package worker

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/luyb177/XiaoAnBackend/content/internal/logic"
	"github.com/luyb177/XiaoAnBackend/content/internal/model"
	"github.com/luyb177/XiaoAnBackend/content/internal/repo/redisqueue"
	"github.com/luyb177/XiaoAnBackend/content/internal/svc"
	"github.com/luyb177/XiaoAnBackend/content/pkg/taskqueue"
	"github.com/luyb177/XiaoAnBackend/content/pkg/taskqueue/tasks"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"time"
)

type CommentRelationHandler struct {
	logx.Logger
	svcCtx     *svc.ServiceContext
	ArticleDao model.ArticleModel
	ComicDao   model.ComicModel
	VideoDao   model.VideoModel
	PodcastDao model.PodcastModel
	CommentDao model.CommentModel
}

func NewCommentRelationHandler(svcCtx *svc.ServiceContext, ctx context.Context) *CommentRelationHandler {
	return &CommentRelationHandler{
		svcCtx:     svcCtx,
		Logger:     logx.WithContext(ctx),
		ArticleDao: model.NewArticleModel(svcCtx.Mysql),
		ComicDao:   model.NewComicModel(svcCtx.Mysql),
		VideoDao:   model.NewVideoModel(svcCtx.Mysql),
		PodcastDao: model.NewPodcastModel(svcCtx.Mysql),
		CommentDao: model.NewCommentModel(svcCtx.Mysql),
	}
}
func (h *CommentRelationHandler) Handle(ctx context.Context, task taskqueue.Task) error {
	payload, err := task.Payload()
	if err != nil {
		return err
	}

	var rawTask redisqueue.RawTask
	err = json.Unmarshal(payload, &rawTask)
	if err != nil {
		return err
	}

	commentTask := tasks.CommentRelationTask{}
	err = json.Unmarshal(rawTask.Data, &commentTask)
	if err != nil {
		return err
	}

	h.Infof("processing comment relation task: %+v", commentTask)

	switch commentTask.Type {
	case tasks.CommentRelationAdd:
		return h.handleAdd(ctx, &commentTask)
	case tasks.CommentRelationDelete:
		return h.handleDelete(ctx, &commentTask)
	default:
		h.Errorf("unknown comment relation type: %s", commentTask.Type)
		return nil
	}
}

func (h *CommentRelationHandler) handleAdd(ctx context.Context, task *tasks.CommentRelationTask) error {
	return h.svcCtx.Mysql.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
		// 先尝试标记（CAS）
		result, err := h.CommentDao.MarkCommentAsCountedWithSession(ctx, session, task.CommentID)
		if err != nil {
			return err
		}
		affect, _ := result.RowsAffected()
		if affect == 0 {
			// 已经计数过，直接返回
			return nil
		}

		// 内容评论数 +1
		result, err = h.incrContentCommentCount(ctx, session, task.ContentType, task.ContentID)
		if err != nil {
			return err
		}
		affect, err = result.RowsAffected()
		if err != nil {
			return err
		}
		if affect == 0 {
			// 内容不存在/被删除，直接删除该评论
			err = h.CommentDao.UserSoftDeleteWithSession(ctx, session, task.CommentID, uint64(time.Now().Unix()))
			if err != nil {
				return err
			}
			_, err = h.CommentDao.UnmarkCommentAsCountedWithSession(ctx, session, task.CommentID)
			return err
		}

		// parent 存在才加子评论数
		if task.ParentID != 0 {
			result, err = h.CommentDao.IncrSubCommentCountWithSession(ctx, session, task.ParentID)
			if err != nil {
				return err
			}
			affect, err = result.RowsAffected()
			if err != nil {
				return err
			}
			if affect == 0 {
				// 父评论不存在/被删除，直接删除该评论
				err = h.CommentDao.UserSoftDeleteWithSession(ctx, session, task.CommentID, uint64(time.Now().Unix()))
				if err != nil {
					return err
				}
				_, err = h.CommentDao.UnmarkCommentAsCountedWithSession(ctx, session, task.CommentID)
				return err
			}
		}

		return nil
	})
}

func (h *CommentRelationHandler) handleDelete(ctx context.Context, task *tasks.CommentRelationTask) error {
	return h.svcCtx.Mysql.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
		var commentCount uint64 = 1
		if task.ParentID == 0 {
			// 先 unmark 所有子评论
			result, err := h.CommentDao.UnmarkChildrenCommentAsCountedWithSession(
				ctx, session, task.CommentID,
			)
			if err != nil {
				return err
			}

			childConsumed, err := result.RowsAffected()
			if err != nil {
				return err
			}

			// 再软删除子评论
			_, err = h.CommentDao.CascadeSoftDeleteChildrenWithSession(
				ctx, session, task.CommentID, uint64(time.Now().Unix()),
			)
			if err != nil {
				return err
			}

			// 统计删除评论总数（子 + 父）
			commentCount = uint64(childConsumed + 1)
		} else {
			// 子评论，父评论子评论数 -1
			result, err := h.CommentDao.DecrSubCommentCountWithSession(ctx, session, task.ParentID)
			if err != nil {
				return err
			}
			affect, err := result.RowsAffected()
			if err != nil {
				return err
			}
			if affect == 0 {
				// 父评论不存在/被删除，已经被删除，无需处理
			}
		}

		// 消费 root 评论的 is_counted
		result, err := h.CommentDao.UnmarkCommentAsCountedWithSession(ctx, session, task.CommentID)
		if err != nil {
			return err
		}
		affect, err := result.RowsAffected()
		if err != nil {
			return err
		}
		if affect == 0 {
			// 未计数过，直接返回
			return nil
		}

		result, err = h.decrContentCommentCount(ctx, session, task.ContentType, task.ContentID, commentCount)
		if err != nil {
			return err
		}
		affect, err = result.RowsAffected()
		if err != nil {
			return err
		}
		if affect == 0 {
			// 内容不存在/被删除，已经被删除，无需处理
		}

		return nil
	})
}

func (h *CommentRelationHandler) incrContentCommentCount(ctx context.Context, session sqlx.Session, tp string, contentId uint64) (sql.Result, error) {
	var err error
	var result sql.Result
	switch tp {
	case logic.ContentTypeArticle:
		result, err = h.ArticleDao.IncrCommentCountWithSession(ctx, session, contentId)
	case logic.ContentTypeComic:
		result, err = h.ComicDao.IncrCommentCountWithSession(ctx, session, contentId)
	case logic.ContentTypeVideo:
		result, err = h.VideoDao.IncrCommentCountWithSession(ctx, session, contentId)
	case logic.ContentTypePodcast:
		result, err = h.PodcastDao.IncrCommentCountWithSession(ctx, session, contentId)
	default:
		return nil, errors.New("未知的评论类型")
	}
	return result, err
}

func (h *CommentRelationHandler) decrContentCommentCount(ctx context.Context, session sqlx.Session, tp string, contentId uint64, count uint64) (sql.Result, error) {
	var err error
	var result sql.Result
	switch tp {
	case logic.ContentTypeArticle:
		result, err = h.ArticleDao.DecrCommentCountByCountWithSession(ctx, session, contentId, count)
	case logic.ContentTypeComic:
		result, err = h.ComicDao.DecrCommentCountByCountWithSession(ctx, session, contentId, count)
	case logic.ContentTypeVideo:
		result, err = h.VideoDao.DecrCommentCountByCountWithSession(ctx, session, contentId, count)
	case logic.ContentTypePodcast:
		result, err = h.PodcastDao.DecrCommentCountByCountWithSession(ctx, session, contentId, count)
	default:
		return nil, errors.New("未知的评论类型")
	}
	return result, err
}
