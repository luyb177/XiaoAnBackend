package convert

import (
	"github.com/luyb177/XiaoAnBackend/content/internal/model"
	v1 "github.com/luyb177/XiaoAnBackend/content/pb/content/v1"
)

func PBFromComment(comments []*model.Comment) []*v1.Comment {
	res := make([]*v1.Comment, len(comments))
	for i, comment := range comments {
		res[i] = &v1.Comment{
			Id:              comment.Id,
			Type:            comment.Type,
			TargetId:        comment.TargetId,
			UserId:          comment.UserId,
			Nickname:        comment.Nickname,
			Avatar:          comment.Avatar,
			IpLocation:      comment.IpLocation,
			ParentId:        comment.ParentId,
			ReplyCommentId:  comment.ReplyCommentId,
			ReplyUserId:     comment.ReplyUserId,
			Content:         comment.Content,
			LikeCount:       comment.LikeCount,
			SubCommentCount: comment.SubCommentCount,
			Status:          comment.Status,
			CreatedAt:       comment.CreatedAt.Unix(),
			UpdatedAt:       comment.UpdatedAt.Unix(),
			IsLiked:         false,
		}
	}
	return res
}
