package content

import (
	"context"
	content "github.com/luyb177/XiaoAnBackend/content/pb/content/v1"

	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/svc"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetRootCommentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewGetRootCommentLogic 获取根评论
func NewGetRootCommentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetRootCommentLogic {
	return &GetRootCommentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetRootCommentLogic) GetRootComment(req *types.GetRootCommentRequest) (resp *types.Response, err error) {
	rpcResp, err := l.svcCtx.ContentRpc.GetRootComment(l.ctx, &content.GetRootCommentRequest{
		ContentType: req.ContentType,
		ContentId:   req.ContentId,
		Page:        req.Page,
		PageSize:    req.PageSize,
	})

	if err != nil {
		l.Errorf("rpc GetRootComment err: %v", err)
		return &types.Response{
			Code:    400,
			Message: "获取根评论失败",
			Data:    &types.EmptyResponse{},
		}, nil
	}

	var rpcData = &content.GetCommentResponse{}
	if rpcResp.Data != nil {
		if err = rpcResp.Data.UnmarshalTo(rpcData); err != nil {
			l.Errorf("unmarshal GetCommentResponse failed: %v", err)
		}
	}
	rpcComments := rpcData.Comments
	if rpcComments == nil {
		rpcComments = []*content.Comment{}
	}

	// HTTP 返回数据（对前端稳定）
	httpComments := make([]types.Comment, len(rpcComments))
	for i, comment := range rpcComments {
		if comment == nil {
			comment = &content.Comment{}
		}
		httpComments[i] = types.Comment{
			CommentID:      comment.Id,
			ContentType:    comment.Type,
			ContentID:      comment.TargetId,
			UserID:         comment.UserId,
			Nickname:       comment.Nickname,
			Avatar:         comment.Avatar,
			IpLocation:     comment.IpLocation,
			ParentID:       comment.ParentId,
			ReplyCommentID: comment.ReplyCommentId,
			ReplyUserID:    comment.ReplyUserId,
			CommentText:    comment.Content,
			LikeCount:      comment.LikeCount,
			Status:         comment.Status,
			CreatedAt:      comment.CreatedAt,
			UpdatedAt:      comment.UpdatedAt,
			IsLiked:        comment.IsLiked,
		}
	}

	httpData := &types.GetRootCommentResponse{Comments: httpComments}

	return &types.Response{
		Code:    rpcResp.Code,
		Message: rpcResp.Message,
		Data:    httpData,
	}, nil
}
