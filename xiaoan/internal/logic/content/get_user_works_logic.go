package content

import (
	"context"
	"database/sql"
	"time"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/luyb177/XiaoAnBackend/infra/constants"
	"github.com/luyb177/XiaoAnBackend/infra/middleware"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/logic"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/svc"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/types"
)

// 与 content 表结构一致的扫描用结构（仅本接口使用）
type userWorksArticleRow struct {
	Id             uint64         `db:"id"`
	Name           string         `db:"name"`
	Url            string         `db:"url"`
	Description    sql.NullString `db:"description"`
	Cover          string         `db:"cover"`
	Content        sql.NullString `db:"content"`
	Author         string         `db:"author"`
	PublishedAt    sql.NullTime   `db:"published_at"`
	RelationStatus int64          `db:"relation_status"`
	LastModifiedBy sql.NullInt64  `db:"last_modified_by"`
	LikeCount      uint64         `db:"like_count"`
	ViewCount      uint64         `db:"view_count"`
	CommentCount   uint64         `db:"comment_count"`
	CollectCount   uint64         `db:"collect_count"`
	CreatedAt      time.Time      `db:"created_at"`
	UpdatedAt      time.Time      `db:"updated_at"`
	DeletedAt      uint64         `db:"deleted_at"`
}

type userWorksVideoRow struct {
	Id             uint64         `db:"id"`
	Name           string         `db:"name"`
	Url            string         `db:"url"`
	Description    sql.NullString `db:"description"`
	Cover          string         `db:"cover"`
	Author         string         `db:"author"`
	RelationStatus int64          `db:"relation_status"`
	LastModifiedBy sql.NullInt64  `db:"last_modified_by"`
	PublishedAt    sql.NullTime   `db:"published_at"`
	CreatedAt      time.Time      `db:"created_at"`
	UpdatedAt      time.Time      `db:"updated_at"`
	DeletedAt      uint64         `db:"deleted_at"`
	LikeCount      uint64         `db:"like_count"`
	ViewCount      uint64         `db:"view_count"`
	CollectCount   uint64         `db:"collect_count"`
	CommentCount   uint64         `db:"comment_count"`
}

type userWorksPodcastRow struct {
	Id             uint64         `db:"id"`
	Name           string         `db:"name"`
	Url            string         `db:"url"`
	Description    sql.NullString `db:"description"`
	Cover          string         `db:"cover"`
	Author         string         `db:"author"`
	PublishedAt    sql.NullTime   `db:"published_at"`
	RelationStatus int64          `db:"relation_status"`
	LastModifiedBy sql.NullInt64  `db:"last_modified_by"`
	LikeCount      uint64         `db:"like_count"`
	ViewCount      uint64         `db:"view_count"`
	CollectCount   uint64         `db:"collect_count"`
	CommentCount   uint64         `db:"comment_count"`
	Channel        string         `db:"channel"`
	Status         int64          `db:"status"`
	CreatedAt      time.Time      `db:"created_at"`
	UpdatedAt      time.Time      `db:"updated_at"`
	DeletedAt      uint64         `db:"deleted_at"`
}

type userWorksComicRow struct {
	Id             uint64         `db:"id"`
	Name           string         `db:"name"`
	Description    sql.NullString `db:"description"`
	Cover          string         `db:"cover"`
	Author         string         `db:"author"`
	PublishedAt    sql.NullTime   `db:"published_at"`
	RelationStatus int64          `db:"relation_status"`
	LastModifiedBy sql.NullInt64  `db:"last_modified_by"`
	ChapterCount   uint64         `db:"chapter_count"`
	LikeCount      uint64         `db:"like_count"`
	ViewCount      uint64         `db:"view_count"`
	CollectCount   uint64         `db:"collect_count"`
	CommentCount   uint64         `db:"comment_count"`
	CreatedAt      time.Time      `db:"created_at"`
	UpdatedAt      time.Time      `db:"updated_at"`
	DeletedAt      uint64         `db:"deleted_at"`
}

type GetUserWorksLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetUserWorksLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserWorksLogic {
	return &GetUserWorksLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetUserWorksLogic) GetUserWorks(req *types.GetUserWorksRequest) (resp *types.Response, err error) {
	if l.svcCtx.Mysql == nil {
		return logic.BadResponse("网关未配置 MysqlConf，无法查询用户作品"), nil
	}
	u, ok := middleware.GetUser(l.ctx)
	if !ok || u.UID == constants.InvalidUserID || u.Status != constants.UserStatusNormal {
		return logic.BadResponse("用户未登录或状态异常"), nil
	}
	if req.UserID == 0 {
		return logic.BadResponse("user_id 无效"), nil
	}

	uid := int64(req.UserID)
	ctx := l.ctx
	conn := l.svcCtx.Mysql

	var articles []userWorksArticleRow
	qArticle := `SELECT * FROM ` + "`article`" + ` WHERE ` + "`last_modified_by`" + ` = ? AND ` + "`deleted_at`" + ` = 0 ORDER BY ` + "`id`" + ` DESC`
	// 使用 Partial，避免 go-zero 对 SELECT * 的 strict 与历史表缺列时报 not matching destination；DateTime 建议 DSN 加 parseTime=true
	if e := conn.QueryRowsPartialCtx(ctx, &articles, qArticle, uid); e != nil {
		l.Errorf("get user works articles: %v", e)
		return logic.BadResponse(queryFailMsg("查询文章", e)), nil
	}

	var videos []userWorksVideoRow
	qVideo := `SELECT * FROM ` + "`video`" + ` WHERE ` + "`last_modified_by`" + ` = ? AND ` + "`deleted_at`" + ` = 0 ORDER BY ` + "`id`" + ` DESC`
	if e := conn.QueryRowsPartialCtx(ctx, &videos, qVideo, uid); e != nil {
		l.Errorf("get user works videos: %v", e)
		return logic.BadResponse(queryFailMsg("查询视频", e)), nil
	}

	var podcasts []userWorksPodcastRow
	qPod := `SELECT * FROM ` + "`podcast`" + ` WHERE ` + "`last_modified_by`" + ` = ? AND ` + "`deleted_at`" + ` = 0 ORDER BY ` + "`id`" + ` DESC`
	if e := conn.QueryRowsPartialCtx(ctx, &podcasts, qPod, uid); e != nil {
		l.Errorf("get user works podcasts: %v", e)
		return logic.BadResponse(queryFailMsg("查询播客", e)), nil
	}

	var comics []userWorksComicRow
	qComic := `SELECT * FROM ` + "`comic`" + ` WHERE ` + "`last_modified_by`" + ` = ? AND ` + "`deleted_at`" + ` = 0 ORDER BY ` + "`id`" + ` DESC`
	if e := conn.QueryRowsPartialCtx(ctx, &comics, qComic, uid); e != nil {
		l.Errorf("get user works comics: %v", e)
		return logic.BadResponse(queryFailMsg("查询漫画", e)), nil
	}

	out := types.GetUserWorksResponse{
		Articles:  make([]types.ArticleInfo, 0, len(articles)),
		Videos:    make([]types.VideoInfo, 0, len(videos)),
		Podcasts:  make([]types.PodcastInfo, 0, len(podcasts)),
		Comics:    make([]types.ComicInfo, 0, len(comics)),
	}
	for i := range articles {
		a := &articles[i]
		articlePub := int64(0)
		if a.PublishedAt.Valid {
			articlePub = a.PublishedAt.Time.Unix()
		}
		out.Articles = append(out.Articles, types.ArticleInfo{
			ArticleID:      a.Id,
			Name:           a.Name,
			Url:            a.Url,
			Description:    a.Description.String,
			Cover:          a.Cover,
			Content:        a.Content.String,
			Author:         a.Author,
			PublishedAt:    articlePub,
			CreatedAt:      a.CreatedAt.Unix(),
			UpdatedAt:      a.UpdatedAt.Unix(),
			LikeCount:      a.LikeCount,
			ViewCount:      a.ViewCount,
			CollectCount:   a.CollectCount,
			CommentCount:   a.CommentCount,
			LastModifiedBy: nullInt64Val(a.LastModifiedBy),
			RelationStatus: a.RelationStatus,
		})
	}
	for i := range videos {
		v := &videos[i]
		published := int64(0)
		if v.PublishedAt.Valid {
			published = v.PublishedAt.Time.Unix()
		}
		out.Videos = append(out.Videos, types.VideoInfo{
			VideoID:        v.Id,
			Name:           v.Name,
			Url:            v.Url,
			Description:    v.Description.String,
			Cover:          v.Cover,
			Author:         v.Author,
			PublishedAt:    published,
			CreatedAt:      v.CreatedAt.Unix(),
			UpdatedAt:      v.UpdatedAt.Unix(),
			LikeCount:      v.LikeCount,
			ViewCount:      v.ViewCount,
			CollectCount:   v.CollectCount,
			CommentCount:   v.CommentCount,
			LastModifiedBy: nullInt64Val(v.LastModifiedBy),
			RelationStatus: v.RelationStatus,
		})
	}
	for i := range podcasts {
		p := &podcasts[i]
		pub := int64(0)
		if p.PublishedAt.Valid {
			pub = p.PublishedAt.Time.Unix()
		}
		out.Podcasts = append(out.Podcasts, types.PodcastInfo{
			PodcastID:      p.Id,
			Name:           p.Name,
			Url:            p.Url,
			Description:    p.Description.String,
			Cover:          p.Cover,
			Author:         p.Author,
			Channel:        p.Channel,
			Status:         p.Status,
			PublishedAt:    pub,
			CreatedAt:      p.CreatedAt.Unix(),
			UpdatedAt:      p.UpdatedAt.Unix(),
			LikeCount:      p.LikeCount,
			ViewCount:      p.ViewCount,
			CollectCount:   p.CollectCount,
			CommentCount:   p.CommentCount,
			LastModifiedBy: nullInt64Val(p.LastModifiedBy),
			RelationStatus: p.RelationStatus,
		})
	}
	for i := range comics {
		c := &comics[i]
		comicPub := int64(0)
		if c.PublishedAt.Valid {
			comicPub = c.PublishedAt.Time.Unix()
		}
		out.Comics = append(out.Comics, types.ComicInfo{
			ComicID:        c.Id,
			Name:           c.Name,
			Description:    c.Description.String,
			Cover:          c.Cover,
			Author:         c.Author,
			PublishedAt:    comicPub,
			CreatedAt:      c.CreatedAt.Unix(),
			UpdatedAt:      c.UpdatedAt.Unix(),
			LikeCount:      c.LikeCount,
			ViewCount:      c.ViewCount,
			CollectCount:   c.CollectCount,
			CommentCount:   c.CommentCount,
			ChapterCount:   c.ChapterCount,
			LastModifiedBy: nullInt64Val(c.LastModifiedBy),
			RelationStatus: c.RelationStatus,
		})
	}

	return &types.Response{
		Code:    200,
		Message: "ok",
		Data:    out,
	}, nil
}

func nullInt64Val(n sql.NullInt64) int64 {
	if !n.Valid {
		return 0
	}
	return n.Int64
}

func queryFailMsg(what string, e error) string {
	msg := e.Error()
	if len(msg) > 220 {
		msg = msg[:220] + "..."
	}
	return what + "失败: " + msg
}
