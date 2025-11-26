package logic

import (
	"context"
	"errors"
	"fmt"
	"github.com/luyb177/XiaoAnBackend/content/internal/model"
	"time"

	"github.com/luyb177/XiaoAnBackend/content/internal/svc"
	"github.com/luyb177/XiaoAnBackend/content/pb/content/v1"

	"github.com/zeromicro/go-zero/core/logx"
)

type CollectLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
	contentCollectDao model.ContentCollectModel
}

func NewCollectLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CollectLogic {
	return &CollectLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// Collect 收藏
func (l *CollectLogic) Collect(in *v1.CollectRequest) (*v1.Response, error) {
	userId := l.ctx.Value("user_id").(uint64)
	userRole := l.ctx.Value("user_role").(string)
	userStatus := l.ctx.Value("user_status").(int64)

	if userId == 0 || userRole == "" || userStatus != 1 {
		return &v1.Response{
			Code:    400,
			Message: "用户信息错误",
		}, fmt.Errorf("用户信息错误")
	}

	if in.Type == "" || in.TargetId <= 0 {
		return &v1.Response{
			Code:    400,
			Message: "参数错误",
		}, fmt.Errorf("参数错误")
	}
	now := time.Now().Unix()
	collect, err := l.contentCollectDao.FindOneByUserIdTypeTargetId(l.ctx, userId, in.Type, in.TargetId)
	switch {
	case errors.Is(err, model.ErrNotFound):
		_, err = l.contentCollectDao.Insert(l.ctx, &model.ContentCollect{
			Type:      in.Type,
			TargetId:  in.TargetId,
			UserId:    userId,
			Status:    Valid,
			CreatedAt: now,
			UpdatedAt: now,
		})
		if err != nil {
			return &v1.Response{
				Code:    400,
				Message: "收藏失败",
			}, fmt.Errorf("收藏失败")
		}
		return &v1.Response{
			Code:    200,
			Message: "收藏成功",
		}, nil

	case err != nil:
		return &v1.Response{
			Code:    400,
			Message: "收藏失败",
		}, fmt.Errorf("收藏失败")

	default:
		if collect.Status == Valid {
			// 取消收藏
			collect.Status = Invalid
		} else {
			// 收藏
			collect.Status = Valid
		}
		collect.UpdatedAt = now
		err = l.contentCollectDao.Update(l.ctx, collect)
		if err != nil {
			return &v1.Response{
				Code:    400,
				Message: "收藏失败",
			}, fmt.Errorf("收藏失败")
		}
		return &v1.Response{
			Code:    200,
			Message: "收藏成功",
		}, nil
	}
}
