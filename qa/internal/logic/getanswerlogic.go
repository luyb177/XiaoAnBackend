package logic

import (
	"context"
	"strings"

	"github.com/luyb177/XiaoAnBackend/qa/internal/svc"
	"github.com/luyb177/XiaoAnBackend/qa/pb/qa/v1"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/protobuf/types/known/anypb"
)

type GetAnswerLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetAnswerLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAnswerLogic {
	return &GetAnswerLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetAnswerLogic) GetAnswer(in *v1.GetAnswerRequest) (*v1.Response, error) {
	answer := &v1.GetAnswerResponse{
		Answer: "抱歉，我暂时无法回答这个问题，可以输入“防猥亵”、“自我保护”、“相关法律”等关键词。",
	}

	for _, qa := range l.svcCtx.QAData {
		for _, kw := range qa.Keywords {
			if strings.Contains(in.Question, kw) {
				// 转换类型
				answer.Answer = qa.Answer

				answerAny, err := anypb.New(answer)
				if err != nil {
					return &v1.Response{
						Code:    500,
						Message: "解析错误",
					}, err
				}

				return &v1.Response{
					Code:    200,
					Message: "success",
					Data:    answerAny,
				}, nil
			}
		}
	}

	answerAny, err := anypb.New(answer)
	if err != nil {
		return &v1.Response{
			Code:    500,
			Message: "解析错误",
		}, err
	}
	return &v1.Response{
		Code:    200,
		Message: "success",
		Data:    answerAny,
	}, nil
}
