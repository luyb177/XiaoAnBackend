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
	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"google.golang.org/protobuf/types/known/anypb"
)

type ModifyVideoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
	VideoDao model.VideoModel
}

func NewModifyVideoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ModifyVideoLogic {
	return &ModifyVideoLogic{
		ctx:      ctx,
		svcCtx:   svcCtx,
		Logger:   logx.WithContext(ctx),
		VideoDao: model.NewVideoModel(svcCtx.Mysql),
	}
}

// ModifyVideo 修改视频
func (l *ModifyVideoLogic) ModifyVideo(in *v1.ModifyVideoRequest) (*v1.Response, error) {
	user := middleware.MustGetUser(l.ctx)
	if user.UID == InvalidUserID || (user.Role != SUPERADMIN && user.Role != STAFF) || user.Status != UserStatusNormal {
		l.Logger.Errorf("ModifyVideo err: 用户登录状态异常")
		return &v1.Response{
			Code:    400,
			Message: "用户未登录或者登录状态异常",
		}, nil
	}

	// 验证参数
	if in.Id <= 0 {
		l.Logger.Errorf("ModifyVideo err: 视频ID不能小于等于0")
		return &v1.Response{
			Code:    400,
			Message: "视频ID不能小于等于0",
		}, nil
	}
	if in.Name == "" {
		l.Logger.Errorf("ModifyVideo err: 视频名称为空")
		return &v1.Response{
			Code:    400,
			Message: "视频名称为空",
		}, nil
	}
	if in.Author == "" {
		l.Logger.Errorf("ModifyVideo err: 视频作者为空")
		return &v1.Response{
			Code:    400,
			Message: "视频作者为空",
		}, nil
	}
	if in.Description == "" {
		l.Logger.Errorf("ModifyVideo err: 文章摘要为空")
		return &v1.Response{
			Code:    400,
			Message: "文章摘要为空",
		}, nil
	}
	if in.Tag == nil || len(in.Tag) == 0 {
		in.Tag = []string{"默认标签"}
	}
	if len(in.Tag) > 10 {
		l.Logger.Errorf("ModifyVideo err: 标签数量不能超过10")
		return &v1.Response{
			Code:    400,
			Message: "标签数量不能超过10",
		}, nil
	}
	// 检查标签
	for _, tag := range in.Tag {
		if tag == "" {
			l.Logger.Errorf("ModifyVideo err: 标签不能为空")
			return &v1.Response{
				Code:    400,
				Message: "标签不能为空",
			}, nil
		}
	}
	now := time.Now()
	if in.PublishedAt <= 0 {
		in.PublishedAt = now.Unix()
	}

	// 1. 先验证 video 是否存在或者被删除
	video, err := l.VideoDao.FindOne(l.ctx, in.Id)
	if err != nil {
		if errors.Is(err, sqlc.ErrNotFound) {
			l.Logger.Errorf("ModifyVideo err: 视频不存在")
			return &v1.Response{
				Code:    400,
				Message: "视频不存在",
			}, nil
		}

		return &v1.Response{
			Code:    400,
			Message: "修改视频出现错误",
		}, nil
	}

	// 2. 主体部分更新
	video.Name = in.Name
	video.Url = in.Url
	video.Description = sql.NullString{String: in.Description, Valid: true}
	video.Cover = in.Cover
	video.Author = in.Author
	video.PublishedAt = sql.NullTime{Time: time.Unix(in.PublishedAt, 0), Valid: true}
	video.UpdatedAt = now

	err = l.VideoDao.Update(l.ctx, video)
	if err != nil {
		l.Logger.Errorf("ModifyVideo err: %v", err)
		return &v1.Response{
			Code:    400,
			Message: "修改视频出现错误",
		}, nil
	}

	// 3. 标签更新
	videoRelationTask := &tasks.VideoRelationTask{
		Type:    tasks.VideoRelationModify,
		VideoID: video.Id,
		Tags:    in.Tag,
	}
	err = l.svcCtx.TaskQueue.Enqueue(l.ctx, videoRelationTask)
	if err != nil {
		l.Logger.Errorf("ModifyVideo Enqueue err: %v", err)
	}

	// 4. 构造返回结果
	res := &v1.ModifyVideoResponse{
		Id:             video.Id,
		RelationStatus: RelationStatusPending,
	}

	resAny, err := anypb.New(res)
	if err != nil {
		l.Logger.Errorf("ModifyVideo err: %v", err)
		return &v1.Response{
			Code:    500,
			Message: "修改视频出现错误",
		}, nil
	}

	return &v1.Response{
		Code:    200,
		Message: "修改视频成功",
		Data:    resAny,
	}, nil
}
