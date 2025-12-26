package logic

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/luyb177/XiaoAnBackend/content/internal/model"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"time"

	"github.com/luyb177/XiaoAnBackend/content/internal/svc"
	"github.com/luyb177/XiaoAnBackend/content/pb/content/v1"

	"github.com/zeromicro/go-zero/core/logx"
)

const (
	SUPERADMIN = "superadmin"
	CLASSADMIN = "classadmin"
	STUDENT    = "student"
	STAFF      = "staff"
)

type AddVideoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
	videoDao    model.VideoModel
	videoTagDao model.VideoTagModel
}

func NewAddVideoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddVideoLogic {
	return &AddVideoLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// AddVideo 添加视频
func (l *AddVideoLogic) AddVideo(in *v1.AddVideoRequest) (*v1.Response, error) {
	creatorId := l.ctx.Value("user_id").(uint64)
	creatorRole := l.ctx.Value("user_role").(string)
	creatorStatus := l.ctx.Value("user_status").(int64)

	if creatorId == 0 || creatorRole == "" || creatorStatus != 1 {
		return &v1.Response{
			Code:    400,
			Message: "用户信息错误",
		}, fmt.Errorf("用户信息错误")
	}

	// 验证身份
	if creatorRole == STUDENT {
		return &v1.Response{
			Code:    400,
			Message: "学生无权限添加视频",
		}, fmt.Errorf("学生无权限添加视频")
	}

	if in.Name == "" || in.Url == "" {
		return &v1.Response{
			Code:    400,
			Message: "视频名称或视频URL不能为空",
		}, fmt.Errorf("视频名称或视频URL不能为空")
	}

	if in.Tag == nil {
		return &v1.Response{
			Code:    400,
			Message: "视频标签不能为空",
		}, fmt.Errorf("视频标签不能为空")
	}

	if in.Description == "" {
		return &v1.Response{
			Code:    400,
			Message: "视频描述不能为空",
		}, fmt.Errorf("视频描述不能为空")
	}

	if in.Cover == "" {
		return &v1.Response{
			Code:    400,
			Message: "视频封面不能为空",
		}, fmt.Errorf("视频封面不能为空")
	}

	if in.Author == "" {
		return &v1.Response{
			Code:    400,
			Message: "视频作者不能为空",
		}, fmt.Errorf("视频作者不能为空")
	}

	if in.CreateTime == 0 {
		return &v1.Response{
			Code:    400,
			Message: "视频创建时间不能为空",
		}, fmt.Errorf("视频创建时间不能为空")
	}
	now := time.Now().Unix()

	video := model.Video{
		Name:         in.Name,
		Url:          in.Url,
		Description:  sql.NullString{String: in.Description, Valid: true},
		Cover:        in.Cover,
		Author:       in.Author,
		CreateTime:   in.CreateTime,
		CreatedAt:    now,
		UpdateTime:   now,
		LikeCount:    0,
		ViewCount:    0,
		CollectCount: 0,
	}

	videoTag := make([]model.VideoTag, 0, len(in.Tag))

	// 事务添加
	err := l.svcCtx.Mysql.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		// 添加视频主体信息
		ret, err := l.videoDao.InsertWithSession(ctx, session, &video)
		if err != nil {
			return err
		}
		// 回写
		id, err := ret.LastInsertId()
		if err != nil {
			return err
		}
		video.Id = uint64(id)

		// 构造
		for _, tag := range in.Tag {
			videoTag = append(videoTag, model.VideoTag{
				VideoId:   video.Id,
				Tag:       tag,
				CreatedAt: now,
			})
		}
		_, err = l.videoTagDao.BatchInsertWithSession(ctx, session, videoTag)
		if err != nil {
			return err
		}
		return nil
	})

	return &v1.Response{
		Code:    200,
		Message: "添加成功",
	}, err
}
