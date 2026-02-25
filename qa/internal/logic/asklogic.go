package logic

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/packages/param"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"golang.org/x/sync/errgroup"

	"github.com/luyb177/XiaoAnBackend/infra/constants"
	"github.com/luyb177/XiaoAnBackend/infra/middleware"
	"github.com/luyb177/XiaoAnBackend/qa/internal/model"
	"github.com/luyb177/XiaoAnBackend/qa/internal/repo/chatmessage"
	"github.com/luyb177/XiaoAnBackend/qa/internal/repo/chatsession"
	"github.com/luyb177/XiaoAnBackend/qa/internal/svc"
	"github.com/luyb177/XiaoAnBackend/qa/pb/qa/v1"
	"github.com/luyb177/XiaoAnBackend/qa/pkg/chatmessage/convert"
	"github.com/luyb177/XiaoAnBackend/qa/pkg/chatmessage/messageid"
	"github.com/luyb177/XiaoAnBackend/qa/pkg/taskqueue/tasks"
)

type AskLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
	ChatSessionDao model.ChatSessionModel
	ChatMessageDao model.ChatMessageModel
}

func NewAskLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AskLogic {
	return &AskLogic{
		ctx:            ctx,
		svcCtx:         svcCtx,
		Logger:         logx.WithContext(ctx),
		ChatSessionDao: model.NewChatSessionModel(svcCtx.Mysql),
		ChatMessageDao: model.NewChatMessageModel(svcCtx.Mysql),
	}
}

func (l *AskLogic) Ask(in *v1.AskRequest, stream v1.QAService_AskServer) error {
	// NOTE: stream.Context() 才是最新的
	user, ok := middleware.GetUser(stream.Context())
	if !ok || user.UID == constants.InvalidUserID || user.Role == "" || user.Status != constants.UserStatusNormal {
		return badStream(stream, "用户未登录或状态异常")
	}

	// 验证
	if err := l.validate(in, stream); err != nil {
		return err
	}

	// 1. 确保会话合法有效
	if err := l.ensureSession(in, stream, user.UID); err != nil {
		l.Errorf("ensureSession err: %v", err)
		return err
	}

	// 2. 持久化用户消息 + 构建对话历史上下文
	history, err := l.persistUserAndBuildHistory(in, stream, user.UID)
	if err != nil {
		l.Errorf("persistUserAndBuildHistory err: %v", err)
		return err
	}

	// 3. 创建一个 assistantMessage 占位，状态是生成中，后续流式更新这个消息的内容和状态
	assistantMessage, err := l.createAssistantPlaceholder(in, stream, user.UID)
	if err != nil {
		l.Errorf("createAssistantPlaceholder err: %v", err)
		return err
	}

	err = l.streamAssistant(history, assistantMessage, in, stream)
	return err
}

func (l *AskLogic) validate(in *v1.AskRequest, stream v1.QAService_AskServer) error {
	switch {
	case in.SessionId == 0:
		err := badStream(stream, "session id is required")
		if err != nil {
			return err
		}
		return errors.New("session_id is required")
	case strings.TrimSpace(in.Content) == "":
		err := badStream(stream, "content is required")
		if err != nil {
			return err
		}
		return errors.New("content is required")
	case in.ClientMessageId == "":
		err := badStream(stream, "client_message_id is required")
		if err != nil {
			return err
		}
		return errors.New("client_message_id is required")
	}
	return nil
}

func (l *AskLogic) ensureSession(in *v1.AskRequest, stream v1.QAService_AskServer, uid uint64) error {
	g, ctx := errgroup.WithContext(stream.Context())

	g.Go(func() error {
		// 验证会话是否合法，顺便尽力加载会话归属到 redis，减少后续访问历史消息时对 mysql 的依赖
		return l.validateSession(ctx, in.SessionId, uid, stream)
	})

	g.Go(func() error {
		// 尝试初始化空会话（如果是新建的会话，第一次提问就会初始化；如果是旧会话或者非法会话，则不做任何操作）
		return l.tryInitEmptySession(ctx, in.SessionId, in.Content)
	})

	return g.Wait()
}

func (l *AskLogic) tryInitEmptySession(ctx context.Context, sessionID uint64, userMessage string) error {
	result, err := l.ChatSessionDao.UpdateEmptySlot(ctx, sessionID, time.Now())
	if err != nil {
		return err
	}
	affect, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affect > 0 {
		chatSessionTask := tasks.ChatSessionTask{
			Type:        tasks.ChatSessionUpdateTitle,
			SessionID:   sessionID,
			UserMessage: userMessage,
		}

		err = l.svcCtx.TaskQueue.Enqueue(ctx, &chatSessionTask)
		if err != nil {
			l.Errorf("Enqueue ChatSessionTask err: %v", err)
		}
		return nil
	}
	// 未更新成功
	// - session 不是新建的，已经初始化过了
	// - sessionID 不存在，非法请求
	// - 并发，其他更新成功了
	return nil
}

func (l *AskLogic) validateSession(ctx context.Context, sessionID, uid uint64, stream v1.QAService_AskServer) error {
	// 查 redis
	owner, err := l.svcCtx.ChatSessionRepo.GetSessionOwner(ctx, sessionID)
	if err == nil {
		if owner != uid {
			sseSendErr := badStream(stream, "session owner is different")
			if sseSendErr != nil {
				return sseSendErr
			}
			return errors.New("session owner is different")
		}
		return nil
	}

	if !errors.Is(err, chatsession.ErrSessionNotFound) {
		l.Errorf("redis get session owner failed: %v", err)
		sseSendErr := internalStream(stream, "系统繁忙，请稍后再试")
		if sseSendErr != nil {
			return sseSendErr
		}
		return err
	}
	//  查 mysql 此处只验证归属
	session, err := l.ChatSessionDao.FindOneByIDUserID(ctx, sessionID, uid)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			sseSendErr := notFoundStream(stream, "session not found")
			if sseSendErr != nil {
				return sseSendErr
			}
			return errors.New("session not found")
		}
		l.Errorf("mysql get session failed: %v", err)
		sseSendErr := internalStream(stream, "系统繁忙，请稍后再试")
		if sseSendErr != nil {
			return sseSendErr
		}
		return errors.New("mysql get session failed")
	}

	// 回填 redis（尽力而为）
	err = l.svcCtx.ChatSessionRepo.SetSessionOwner(ctx, session.Id, session.UserId)
	if err != nil {
		l.Errorf("set session owner to redis fail for session %d: %v", session.Id, err)
	}

	return nil
}

func (l *AskLogic) persistUserAndBuildHistory(in *v1.AskRequest, stream v1.QAService_AskServer, uid uint64) ([]openai.ChatCompletionMessageParamUnion, error) {
	var (
		history []openai.ChatCompletionMessageParamUnion
	)

	g, ctx := errgroup.WithContext(stream.Context())

	g.Go(func() error {
		return l.persistUserMessage(ctx, in, uid)
	})

	g.Go(func() error {
		// 构建对话历史上下文，后续用于调用 LLM 接口
		history = l.buildHistory(ctx, in)
		return nil
	})

	return history, g.Wait()
}

func (l *AskLogic) persistUserMessage(ctx context.Context, in *v1.AskRequest, uid uint64) error {
	// 构建 userMessage
	now := time.Now()
	userMessage := &model.ChatMessage{
		MessageId:        in.ClientMessageId,
		SessionId:        in.SessionId,
		UserId:           uid,
		Status:           MessageStatusSuccess,
		PromptTokens:     0,
		CompletionTokens: 0,
		TotalTokens:      0,
		Role:             MessageRoleUser,
		MessageType:      MessageTypeText,
		Content:          in.Content,
		FinishReason:     FinishReasonStop,
		CreatedAt:        now,
		UpdatedAt:        now,
		DeletedAt:        0,
	}
	err := l.svcCtx.Mysql.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
		// 1. 插入用户消息
		result, err := l.ChatMessageDao.InsertWithSession(ctx, session, userMessage)
		if err == nil {
			// 插入成功，获取自增 ID
			userMessageID, err := result.LastInsertId()
			if err != nil {
				return err
			}
			userMessage.Id = uint64(userMessageID)

			// 2. 将 session 的 message_count +1，has_message = 1，last_message_at 更新一下
			result, err = l.ChatSessionDao.IncrementMessageCountWithSession(ctx, session, in.SessionId, now)
			if err != nil {
				return err
			}
			affect, err := result.RowsAffected()
			if err != nil {
				return err
			}
			if affect == 0 {
				// 没有更新成功
				// - sessionID 不存在，非法请求
				// - session 已经被删除了
				// - 并发，其他请求更新了这个 session
				// todo 这里是否返回错误
				return errors.New("session not found or already deleted")
			}
			return nil
		} else if errors.Is(err, model.ErrDuplicateEntry) {
			// message_id 重复了，说明用户客户端重试了同一个请求（可能是网络抖动导致的）,查询一下这个消息，
			userMessage, err = l.ChatMessageDao.FindOneBySessionIdMessageId(ctx, in.SessionId, in.ClientMessageId)
			if err != nil {
				if !errors.Is(err, model.ErrNotFound) {
					return err
				}
				// 没有查到的错误，理论上不应该发生，说明数据不一致了，先记录日志，后续可以加个报警
				l.Errorf("message_id duplicate but not found in mysql, session=%d, message_id=%s", in.SessionId, in.ClientMessageId)
				return nil
			}
			return nil
		}
		return err
	})
	if err != nil {
		return err
	}

	// 将 userMessage 尽力放入缓存
	userMessageCache := chatmessage.MessageCache{
		ID:        userMessage.Id,
		MessageID: userMessage.MessageId,
		Role:      userMessage.Role,
		Content:   userMessage.Content,
		Ts:        userMessage.CreatedAt.Unix(),
	}
	err = l.svcCtx.ChatMessageRepo.UpsertMessage(ctx, in.SessionId, &userMessageCache)
	if err != nil {
		l.Errorf("push user message to redis failed, session=%d, err=%v", in.SessionId, err)
		// 不影响正常使用，所以不返回错误
	}
	return nil
}

func (l *AskLogic) buildHistory(ctx context.Context, in *v1.AskRequest) []openai.ChatCompletionMessageParamUnion {
	// 从历史记录中读取对话上下文：redis -> mysql -> redis
	messagesCache, err := l.svcCtx.ChatMessageRepo.GetRecentMessages(ctx, in.SessionId)
	if err != nil {
		l.Errorf("get recent messages from redis failed, session=%d, err=%v", in.SessionId, err)
	}

	// Redis 命中
	if len(messagesCache) > 0 {
		l.Infof("hit redis messages, session=%d, size=%d", in.SessionId, len(messagesCache))
	} else {
		// Redis 未命中，看 MySQL
		l.Infof("redis miss, fallback to mysql, session=%d", in.SessionId)

		// id asc 排序，保证 messagesCache 中的消息是从旧到新的顺序
		messages, dbErr := l.ChatMessageDao.FindMessagesByChatSessionID(
			ctx,
			in.SessionId,
			chatmessage.MaxMessageListSize,
		)

		if dbErr != nil && !errors.Is(dbErr, model.ErrNotFound) {
			l.Errorf("query messages from mysql failed, session=%d, err=%v", in.SessionId, dbErr)
		} else if len(messages) > 0 {
			//  回填 Redis 从旧到新，保证 Redis 中的消息列表是从旧到新的顺序
			// 顺便转成缓存结构供后续使用
			messagesCache = make([]*chatmessage.MessageCache, 0, len(messages))
			for _, m := range messages {
				cacheMsg := convert.CacheMsg(m)
				messagesCache = append(messagesCache, cacheMsg)
				err = l.svcCtx.ChatMessageRepo.UpsertMessage(ctx, in.SessionId, cacheMsg)
				if err != nil {
					l.Errorf("push message to redis failed, session=%d, err=%v", in.SessionId, err)
					// 不影响正常使用，所以不返回错误
				}
			}
		}
	}

	history := make([]openai.ChatCompletionMessageParamUnion, 0, len(messagesCache)+1)

	for _, m := range messagesCache {
		switch m.Role {
		case MessageRoleUser:
			history = append(history, openai.ChatCompletionMessageParamUnion{
				OfUser: &openai.ChatCompletionUserMessageParam{
					Content: openai.ChatCompletionUserMessageParamContentUnion{
						OfString: param.NewOpt(m.Content),
					},
				},
			})
		case MessageRoleAssistant:
			history = append(history, openai.ChatCompletionMessageParamUnion{
				OfAssistant: &openai.ChatCompletionAssistantMessageParam{
					Content: openai.ChatCompletionAssistantMessageParamContentUnion{
						OfString: param.NewOpt(m.Content),
					},
				},
			})
		case MessageRoleSystem:
			history = append(history, openai.ChatCompletionMessageParamUnion{
				OfSystem: &openai.ChatCompletionSystemMessageParam{
					Content: openai.ChatCompletionSystemMessageParamContentUnion{
						OfString: param.NewOpt(m.Content),
					},
				},
			})
		default:
			l.Errorf("unknown role %d", m.Role)
			continue
		}
	}

	// 最后追加用户的提问
	history = append(history, openai.ChatCompletionMessageParamUnion{
		OfUser: &openai.ChatCompletionUserMessageParam{
			Content: openai.ChatCompletionUserMessageParamContentUnion{
				OfString: param.NewOpt(in.Content),
			},
		},
	})

	return history
}

func (l *AskLogic) createAssistantPlaceholder(in *v1.AskRequest, stream v1.QAService_AskServer, uid uint64) (*model.ChatMessage, error) {
	var assistantMessage *model.ChatMessage
	err := l.svcCtx.Mysql.TransactCtx(stream.Context(), func(ctx context.Context, session sqlx.Session) error {
		now := time.Now()
		assistantMessage = &model.ChatMessage{
			MessageId:        messageid.NewMessageID(in.SessionId, uid, in.ClientMessageId),
			SessionId:        in.SessionId,
			UserId:           uid,
			Status:           MessageStatusGenerating,
			PromptTokens:     0,
			CompletionTokens: 0,
			TotalTokens:      0,
			Role:             MessageRoleAssistant,
			MessageType:      MessageTypeText,
			Content:          "",
			FinishReason:     "",
			CreatedAt:        now,
			UpdatedAt:        now,
			DeletedAt:        0,
		}

		result, err := l.ChatMessageDao.InsertWithSession(ctx, session, assistantMessage)
		if err == nil {
			assistantMessageID, err := result.LastInsertId()
			if err != nil {
				l.Errorf("get last insert id failed for assistant message, session=%d, err=%v", in.SessionId, err)
				return err
			}
			assistantMessage.Id = uint64(assistantMessageID)
			// 消息数 +1
			result, err = l.ChatSessionDao.IncrementMessageCountWithSession(ctx, session, in.SessionId, now)
			if err != nil {
				return err
			}
			affect, err := result.RowsAffected()
			if err != nil {
				return err
			}
			if affect == 0 {
				// 没有更新成功
				// - sessionID 不存在，非法请求
				// - session 已经被删除了
				// - 并发，其他请求更新了这个 session
				// todo 这里是否返回错误
				return errors.New("session not found or already deleted")
			}

		} else if errors.Is(err, model.ErrDuplicateEntry) {
			// message_id 重复了，说明并发了同一个请求（可能是用户重复点击了发送按钮）,查询一下这个消息，
			assistantMessage, err = l.ChatMessageDao.FindOneBySessionIdMessageIdWithSession(ctx, session, in.SessionId, assistantMessage.MessageId)
			if err != nil {
				if !errors.Is(err, model.ErrNotFound) {
					return err
				}
				// 没有查到的错误，理论上不应该发生，说明数据不一致了，先记录日志，后续可以加个报警
				l.Errorf("assistant message_id duplicate but not found in mysql, session=%d, message_id=%s", in.SessionId, assistantMessage.MessageId)
				return nil
			}
			switch assistantMessage.Status {
			case MessageStatusGenerating:
				// 正在生成中，说明是重复请求了
				_ = stream.Send(&v1.AskStreamReply{
					Finished: true,
					Code:     202,
					Message:  "正在生成中",
				})
				// 返回错误是为了退出这次请求
				return errors.New("duplicate request, message is generating")
			case MessageStatusSuccess:
				// 已经生成好了，说明重复请求了，直接将
				_ = stream.Send(&v1.AskStreamReply{
					Delta:    assistantMessage.Content,
					Finished: false,
					Code:     200,
					Message:  "成功",
				})
				_ = stream.Send(&v1.AskStreamReply{
					Finished: true,
					Code:     200,
					Message:  "success",
				})
				// 返回错误是为了退出这次请求
				return errors.New("duplicate request, message already generated")
			case MessageStatusFailed:
				// 上一次请求生成失败了，让客户端更换uuid重试
				_ = badStream(stream, "上一次请求生成失败了，请重试")
				// 返回错误是为了退出这次请求
				return errors.New("duplicate request, previous generation failed")

			default:
				// 其他状态，理论上不应该有，先记录日志，后续可以加个报警
				l.Errorf("assistant message has unexpected status %d, session=%d, message_id=%s", assistantMessage.Status, in.SessionId, assistantMessage.MessageId)
				return nil
			}
		} else {
			// 其他错误，返回
			l.Errorf("insert assistant message to mysql failed, session=%d, err=%v", in.SessionId, err)
			return err
		}
		return nil
	})

	if err != nil {
		l.Errorf("build assistant message failed, session=%d, err=%v", in.SessionId, err)
		return nil, err
	}
	return assistantMessage, nil
}

func (l *AskLogic) streamAssistant(history []openai.ChatCompletionMessageParamUnion, assistantMessage *model.ChatMessage, in *v1.AskRequest, stream v1.QAService_AskServer) error {
	flusher := newAssistantFlusher(l.Logger, l.ChatMessageDao, assistantMessage)
	err := l.svcCtx.LLMClient.ChatCompletionStream(
		stream.Context(),
		history,
		func(chunk openai.ChatCompletionChunk) error {
			if len(chunk.Choices) == 0 {
				return nil
			}

			delta := chunk.Choices[0].Delta.Content
			flusher.append(delta)
			if flusher.shouldFlush() {
				flusher.flush(stream.Context())
			}

			if chunk.Choices[0].FinishReason != "" {
				// 发送玩了，更新 assistantMessage 的状态和 finish_reason
				usage := &openaiUsage{
					PromptTokens:     uint64(chunk.Usage.PromptTokens),
					CompletionTokens: uint64(chunk.Usage.CompletionTokens),
					TotalTokens:      uint64(chunk.Usage.TotalTokens),
				}
				err := flusher.finalize(stream.Context(), MessageStatusSuccess, chunk.Choices[0].FinishReason, usage)
				if err == nil {
					// 尽力加载缓存中的消息内容和状态
					assistantMessageCache := chatmessage.MessageCache{
						ID:        assistantMessage.Id,
						MessageID: assistantMessage.MessageId,
						Role:      assistantMessage.Role,
						Content:   assistantMessage.Content,
						Ts:        assistantMessage.CreatedAt.Unix(),
					}
					err := l.svcCtx.ChatMessageRepo.UpsertMessage(stream.Context(), in.SessionId, &assistantMessageCache)
					if err != nil {
						l.Errorf("push assistant message to redis failed, session=%d, err=%v", in.SessionId, err)
						// 不影响正常使用，所以不返回错误
					}
				} else {
					l.Errorf("finalize assistant message failed, err: %v", err)
				}

				// 发送流式响应，告诉客户端已经结束了
				return stream.Send(&v1.AskStreamReply{
					Delta:    delta,
					Finished: true,
					Code:     200,
					Message:  "success",
				})
			}

			// 还没发送完，继续发送流式响应
			return stream.Send(&v1.AskStreamReply{
				Delta:    delta,
				Finished: false,
				Code:     200,
				Message:  "success",
			})
		},
	)

	if err != nil {
		l.Errorf("llm client chat failed, session=%d, err=%v", in.SessionId, err)
		err := flusher.finalize(stream.Context(), MessageStatusFailed, FinishReasonError, &openaiUsage{})
		if err != nil {
			l.Errorf("flush assistant message failed, session=%d, err=%v", in.SessionId, err)
		}
	}
	return err
}

// tools
func badStream(stream v1.QAService_AskServer, msg string) error {
	return stream.Send(&v1.AskStreamReply{
		Finished: true,
		Code:     400,
		Message:  msg,
	})
}

func internalStream(stream v1.QAService_AskServer, msg string) error {
	return stream.Send(&v1.AskStreamReply{
		Finished: true,
		Code:     500,
		Message:  msg,
	})
}

func notFoundStream(stream v1.QAService_AskServer, msg string) error {
	return stream.Send(&v1.AskStreamReply{
		Finished: true,
		Code:     404,
		Message:  msg,
	})
}
