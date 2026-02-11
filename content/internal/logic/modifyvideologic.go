package logic

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/luyb177/XiaoAnBackend/content/internal/middleware"
	"github.com/luyb177/XiaoAnBackend/content/internal/model"
	"github.com/luyb177/XiaoAnBackend/content/internal/svc"
	"github.com/luyb177/XiaoAnBackend/content/pb/content/v1"
	"github.com/luyb177/XiaoAnBackend/content/pkg/taskqueue/tasks"
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
	user, ok := middleware.GetUser(l.ctx)
	if !ok || user.UID == InvalidUserID || (user.Role != SUPERADMIN && user.Role != STAFF) || user.Status != UserStatusNormal {
		return bad("用户未登录或状态异常"), nil
	}

	// 验证参数
	validations := []Validation{
		{in.Id > 0, "视频ID不能小于等于0"},
		{in.Name != "", "视频名称为空"},
		{in.Author != "", "视频作者为空"},
		{in.Description != "", "视频描述为空"},
		{len(in.Tag) <= 10, "标签数量不能超过10"},
	}

	for _, v := range validations {
		if !v.Condition {
			l.Errorf("ModifyVideo err: %s", v.Message)

			return &v1.Response{
				Code:    400,
				Message: v.Message,
			}, nil
		}
	}

	if len(in.Tag) == 0 {
		in.Tag = []string{"默认标签"}
	}
	// 检查标签
	for _, tag := range in.Tag {
		if tag == "" {
			l.Errorf("ModifyVideo err: 标签不能为空")
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

	// 先验证 video 是否存在或者被删除
	video, err := l.VideoDao.FindOneWithNotDelete(l.ctx, in.Id)
	if err != nil {
		if errors.Is(err, sqlc.ErrNotFound) {
			l.Errorf("ModifyVideo err: 视频不存在")
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

	// 主体部分更新
	video.Name = in.Name
	video.Url = in.Url
	video.Description = sql.NullString{String: in.Description, Valid: true}
	video.Cover = in.Cover
	video.Author = in.Author
	video.PublishedAt = sql.NullTime{Time: time.Unix(in.PublishedAt, 0), Valid: true}
	video.UpdatedAt = now

	err = l.VideoDao.Update(l.ctx, video)
	if err != nil {
		l.Errorf("ModifyVideo err: %v", err)
		return &v1.Response{
			Code:    400,
			Message: "修改视频出现错误",
		}, nil
	}

	// 标签更新
	videoRelationTask := &tasks.VideoRelationTask{
		Type:    tasks.VideoRelationModify,
		VideoID: video.Id,
		Tags:    in.Tag,
	}
	err = l.svcCtx.TaskQueue.Enqueue(l.ctx, videoRelationTask)
	if err != nil {
		l.Errorf("ModifyVideo Enqueue err: %v", err)
	}

	// 构造返回结果
	res := &v1.ModifyVideoResponse{
		Id:             video.Id,
		RelationStatus: RelationStatusPending,
	}

	resAny, err := anypb.New(res)
	if err != nil {
		l.Errorf("ModifyVideo err: %v", err)
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
