package openai

import (
	"context"

	"github.com/luyb177/XiaoAnBackend/qa/internal/config"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/packages/param"
	"github.com/zeromicro/go-zero/core/logx"
)

type LLMClient interface {
	ChatCompletionStream(ctx context.Context, history []openai.ChatCompletionMessageParamUnion, onDelta func(chunk openai.ChatCompletionChunk) error) error
	ChatCompletionToTitle(ctx context.Context, userMessage *openai.ChatCompletionUserMessageParam) (string, error)
}

type LLMClientImpl struct {
	logx.Logger
	cfg    config.LLMClientConfig
	Client openai.Client
}

func NewLLMClient(cfg config.LLMClientConfig) LLMClient {
	return &LLMClientImpl{
		Logger: logx.WithContext(context.Background()),
		cfg:    cfg,
		Client: openai.NewClient(
			option.WithAPIKey(cfg.APIKey),
			option.WithBaseURL(cfg.BaseURL),
		),
	}
}

var XiaoAnSystemMessage = &openai.ChatCompletionSystemMessageParam{
	Content: openai.ChatCompletionSystemMessageParamContentUnion{
		OfString: param.NewOpt(SystemPrompt),
	},
}

var TitleSystemMessage = &openai.ChatCompletionSystemMessageParam{
	Content: openai.ChatCompletionSystemMessageParamContentUnion{
		OfString: param.NewOpt(TitlePrompt),
	},
}

// ChatCompletionToTitle 单次聊天获取标题使用
func (c *LLMClientImpl) ChatCompletionToTitle(ctx context.Context, userMessage *openai.ChatCompletionUserMessageParam) (string, error) {
	res, err := c.Client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			{
				OfSystem: TitleSystemMessage,
			},
			{
				OfUser: userMessage,
			},
		},
		Model: c.cfg.Model,
	})
	return res.Choices[0].Message.Content, err
}

// ChatCompletionStream 聊天补全流式接口
func (c *LLMClientImpl) ChatCompletionStream(ctx context.Context, history []openai.ChatCompletionMessageParamUnion, onDelta func(chunk openai.ChatCompletionChunk) error) error {
	// 支持多轮对话，传入历史消息
	if len(history) == 0 {
		history = []openai.ChatCompletionMessageParamUnion{}
	}

	messages := make([]openai.ChatCompletionMessageParamUnion, 0, len(history)+1)

	// 系统消息放在第一条，作为对话的背景和指导
	messages = append(messages, openai.ChatCompletionMessageParamUnion{
		OfSystem: XiaoAnSystemMessage,
	})

	// 将历史消息追加到系统消息之后，保持对话的连续性
	messages = append(messages, history...)

	stream := c.Client.Chat.Completions.NewStreaming(ctx, openai.ChatCompletionNewParams{
		Messages: messages,
		Model:    c.cfg.Model,
	})

	defer stream.Close()

	for stream.Next() {

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		chunk := stream.Current()

		if len(chunk.Choices) == 0 {
			continue
		}

		// 回调
		if err := onDelta(chunk); err != nil {
			// todo 如果客户端断开了
			return err
		}
	}

	return stream.Err()
}
