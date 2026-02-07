package content

import (
	"context"

	content "github.com/luyb177/XiaoAnBackend/content/pb/content/v1"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/svc"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetSubCommentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewGetSubCommentLogic 获取子评论
func NewGetSubCommentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetSubCommentLogic {
	return &GetSubCommentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetSubCommentLogic) GetSubComment(req *types.GetSubCommentRequest) (resp *types.Response, err error) {
	rpcResp, err := l.svcCtx.ContentRpc.GetSubComment(l.ctx, &content.GetSubCommentRequest{
		ContentType:     req.ContentType,
		ContentId:       req.ContentId,
		ParentCommentId: req.ParentCommentId,
		Page:            req.Page,
		PageSize:        req.PageSize,
	})
	if err != nil {
		l.Errorf("rpc GetSubComment err: %s", err.Error())
		return &types.Response{
			Code:    400,
			Message: "获取子评论失败",
			Data:    &types.EmptyResponse{},
		}, nil
	}

	var rpcData = &content.GetSubCommentResponse{}

	if rpcResp.Data != nil {
		if err = rpcResp.Data.UnmarshalTo(rpcData); err != nil {
			l.Errorf("unmarshal GetSubCommentResponse failed: %v", err)
		}
	}
	rpcComments := rpcData.Comments
	if rpcComments == nil {
		rpcComments = []*content.Comment{}
	}

	httpComments := make([]types.Comment, len(rpcComments))
	for i, rpcComment := range rpcComments {
		if rpcComment == nil {
			rpcComment = &content.Comment{}
		}
		httpComments[i] = types.Comment{
			CommentID:      rpcComment.Id,
			ContentType:    rpcComment.Type,
			ContentID:      rpcComment.TargetId,
			UserID:         rpcComment.UserId,
			Nickname:       rpcComment.Nickname,
			Avatar:         rpcComment.Avatar,
			IpLocation:     rpcComment.IpLocation,
			ParentID:       rpcComment.ParentId,
			ReplyCommentID: rpcComment.ReplyCommentId,
			ReplyUserID:    rpcComment.ReplyUserId,
			CommentText:    rpcComment.Content,
			LikeCount:      rpcComment.LikeCount,
			Status:         rpcComment.Status,
			CreatedAt:      rpcComment.CreatedAt,
			UpdatedAt:      rpcComment.UpdatedAt,
			IsLiked:        rpcComment.IsLiked,
		}
	}

	httpData := &types.GetSubCommentResponse{Comments: httpComments}

	return &types.Response{
		Code:    rpcResp.Code,
		Message: rpcResp.Message,
		Data:    httpData,
	}, nil
}
