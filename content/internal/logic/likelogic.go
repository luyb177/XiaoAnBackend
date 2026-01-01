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

const (
	Valid   = "valid"
	Invalid = "invalid"
)

type LikeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
	contentLikeDao model.ContentLikeModel
}

func NewLikeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LikeLogic {
	return &LikeLogic{
		ctx:            ctx,
		svcCtx:         svcCtx,
		Logger:         logx.WithContext(ctx),
		contentLikeDao: model.NewContentLikeModel(svcCtx.Mysql),
	}
}

// Like 点赞
func (l *LikeLogic) Like(in *v1.LikeRequest) (*v1.Response, error) {
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

	// 先查询
	now := time.Now()

	like, err := l.contentLikeDao.FindOneByUserIdTypeTargetId(l.ctx, userId, in.Type, in.TargetId)

	switch {
	case errors.Is(err, model.ErrNotFound):
		_, err = l.contentLikeDao.Insert(l.ctx, &model.ContentLike{
			Type:     in.Type,
			TargetId: in.TargetId,
			UserId:   userId,

			CreatedAt: now,
			UpdatedAt: now,
		})
		if err != nil {
			return &v1.Response{
				Code:    400,
				Message: "点赞失败",
			}, fmt.Errorf("点赞失败")
		}
		return &v1.Response{
			Code:    200,
			Message: "点赞成功",
		}, nil

	case err != nil:
		return &v1.Response{
			Code:    400,
			Message: "点赞失败",
		}, fmt.Errorf("点赞失败")

	default:
		//// 有 查看状态
		//if like.Status == Valid {
		//	// 取消点赞
		//	like.Status = Invalid
		//} else {
		//	// 点赞
		//	like.Status = Valid
		//}
		like.UpdatedAt = now
		// 更新
		err = l.contentLikeDao.Update(l.ctx, like)
		if err != nil {
			return &v1.Response{
				Code:    400,
				Message: "点赞失败",
			}, fmt.Errorf("点赞失败")
		}
		return &v1.Response{
			Code:    200,
			Message: "点赞成功",
		}, nil
	}
}
