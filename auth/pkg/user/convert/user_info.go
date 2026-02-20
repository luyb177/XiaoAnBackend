package convert

import (
	"github.com/luyb177/XiaoAnBackend/auth/internal/model"
	v1 "github.com/luyb177/XiaoAnBackend/auth/pb/auth/v1"
)

func PBFromUsers(list []*model.User) []*v1.UserInfo {
	res := make([]*v1.UserInfo, len(list))
	for i, item := range list {
		res[i] = &v1.UserInfo{
			Id:         item.Id,
			Name:       item.Name,
			Email:      item.Email,
			Avatar:     item.Avatar.String,
			Phone:      item.Phone.String,
			Department: item.Department.String,
			Role:       item.Role,
			ClassId:    item.ClassId,
			Status:     item.Status,
			CreatedAt:  item.CreatedAt.Unix(),
			UpdatedAt:  item.UpdatedAt.Unix(),
		}
	}
	return res
}
