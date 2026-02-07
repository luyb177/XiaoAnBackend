package convert

import (
	"github.com/luyb177/XiaoAnBackend/auth/internal/model"
	v1 "github.com/luyb177/XiaoAnBackend/auth/pb/auth/v1"
)

func PBFromInviteCode(codes []*model.InviteCode) []*v1.InviteCode {
	res := make([]*v1.InviteCode, len(codes))
	for i, code := range codes {
		res[i] = &v1.InviteCode{
			Code:        code.Code,
			CreatorId:   code.CreatorId,
			CreatorName: code.CreatorName.String,
			Department:  code.Department.String,
			MaxUses:     code.MaxUses,
			UsedCount:   code.UsedCount,
			Remark:      code.Remark.String,
			CreatedAt:   code.CreatedAt.Unix(),
			UpdatedAt:   code.UpdatedAt.Unix(),
			ExpiresAt:   code.ExpiresAt.Time.Unix(),
			TargetRole:  code.TargetRole,
			ClassId:     code.ClassId,
		}
	}
	return res
}
