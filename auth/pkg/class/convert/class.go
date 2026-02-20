package convert

import (
	"github.com/luyb177/XiaoAnBackend/auth/internal/model"
	v1 "github.com/luyb177/XiaoAnBackend/auth/pb/auth/v1"
)

func PBFromClass(list []*model.Class) []*v1.Class {
	res := make([]*v1.Class, len(list))
	for i, item := range list {
		res[i] = &v1.Class{
			Id:           item.Id,
			Name:         item.Name,
			Description:  item.Description.String,
			AdminId:      item.AdminId,
			StudentCount: item.StudentCount,
			Status:       item.Status,
			CreatedAt:    item.CreatedAt.Unix(),
			UpdatedAt:    item.UpdatedAt.Unix(),
		}
	}
	return res
}
