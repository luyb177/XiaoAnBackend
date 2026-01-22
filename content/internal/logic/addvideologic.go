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
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"google.golang.org/protobuf/types/known/anypb"
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
	// 目前添加视频也只能由超级管理员或员工添加
	user := middleware.MustGetUser(l.ctx)
	if user.UID == InvalidUserID || (user.Role != SUPERADMIN && user.Role != STAFF) || user.Status != UserStatusNormal {
		l.Errorf("AddVideo err: 用户未登录或无权限")

		return &v1.Response{
			Code:    400,
			Message: "用户未登录或无权限",
		}, nil
	}

	// 校验参数
	if in.Name == "" || in.Url == "" {
		return &v1.Response{
			Code:    400,
			Message: "视频名称或视频URL不能为空",
		}, nil
	}
	if in.Description == "" {
		return &v1.Response{
			Code:    400,
			Message: "视频描述不能为空",
		}, nil
	}
	if in.Cover == "" {
		return &v1.Response{
			Code:    400,
			Message: "视频封面不能为空",
		}, nil
	}
	if in.Author == "" {
		return &v1.Response{
			Code:    400,
			Message: "视频作者不能为空",
		}, nil
	}

	now := time.Now()
	if in.PublishedAt <= 0 {
		in.PublishedAt = now.Unix()
	}
	if in.Tag == nil {
		in.Tag = []string{"默认标签"}
	}

	// 添加视频
	// 事务添加
	// todo 依旧不需要事务
	var video model.Video
	err := l.svcCtx.Mysql.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		// 1. 构造
		video = model.Video{
			Name:        in.Name,
			Url:         in.Url,
			Description: sql.NullString{String: in.Description, Valid: true},
			Cover:       in.Cover,
			Author:      in.Author,
			// todo 这里的 sql.NullTime 类型可以支持未来的草稿，未发布
			RelationStatus: RelationStatusPending,
			LastModifiedBy: sql.NullInt64{Int64: int64(user.UID), Valid: true},
			PublishedAt:    sql.NullTime{Time: time.Unix(in.PublishedAt, 0), Valid: true},
			CreatedAt:      now,
			UpdatedAt:      now,
			LikeCount:      0,
			ViewCount:      0,
			CollectCount:   0,
		}

		// 2. 插入
		ret, err := l.videoDao.InsertWithSession(ctx, session, &video)
		if err != nil {
			return err
		}

		// 3. 回写
		id, err := ret.LastInsertId()
		if err != nil {
			return err
		}
		video.Id = uint64(id)

		return nil
	})

	if err != nil {
		l.Errorf("AddVideo err: %v", err)

		return &v1.Response{
			Code:    500,
			Message: "添加失败",
		}, err
	}

	videoRelationTask := &tasks.VideoRelationTask{
		Type:    tasks.VideoRelationAdd,
		VideoID: video.Id,
		Tags:    in.Tag,
	}

	// 任务
	err = l.svcCtx.TaskQueue.Enqueue(l.ctx, videoRelationTask)
	if err != nil {
		l.Errorf("AddVideo Enqueue err: %v", err)
	}

	// 构造返回内容
	res := &v1.AddVideoResponse{
		Id:             video.Id,
		RelationStatus: RelationStatusPending,
	}

	resAny, err := anypb.New(res)
	if err != nil {
		l.Errorf("AddVideo err: %v", err)

		return &v1.Response{
			Code:    500,
			Message: "转换类型失败",
		}, err
	}

	return &v1.Response{
		Code:    200,
		Message: "添加成功",
		Data:    resAny,
	}, err
}
