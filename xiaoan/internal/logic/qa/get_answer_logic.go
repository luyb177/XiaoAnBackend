package qa

import (
	"context"

	qa "github.com/luyb177/XiaoAnBackend/qa/pb/qa/v1"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/svc"
	"github.com/luyb177/XiaoAnBackend/xiaoan/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetAnswerLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取答案
func NewGetAnswerLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAnswerLogic {
	return &GetAnswerLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetAnswerLogic) GetAnswer(req *types.GetAnswerRequest) (resp *types.Response, err error) {
	if req.Question == "" {
		return &types.Response{
			Code:    400,
			Message: "请输入问题",
		}, nil
	}
	res, err := l.svcCtx.QARpc.GetAnswer(l.ctx, &qa.GetAnswerRequest{Question: req.Question})
	if err != nil {
		return &types.Response{
			Code:    500,
			Message: "获取答案失败",
		}, nil
	}
	// 解析数据
	answer := &qa.GetAnswerResponse{}
	err = res.Data.UnmarshalTo(answer)
	if err != nil {
		return &types.Response{
			Code:    500,
			Message: "数据解析失败",
		}, nil
	}
	return &types.Response{
		Code:    200,
		Message: "success",
		Data:    answer,
	}, nil
}
