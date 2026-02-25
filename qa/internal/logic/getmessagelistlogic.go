package logic

import (
	"context"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"

	"github.com/luyb177/XiaoAnBackend/infra/constants"
	"github.com/luyb177/XiaoAnBackend/infra/middleware"
	"github.com/luyb177/XiaoAnBackend/qa/internal/model"
	"github.com/luyb177/XiaoAnBackend/qa/internal/svc"
	"github.com/luyb177/XiaoAnBackend/qa/pb/qa/v1"
	"github.com/luyb177/XiaoAnBackend/qa/pkg/chatmessage/convert"
	"github.com/luyb177/XiaoAnBackend/qa/pkg/reverse"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetMessageListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
	ChatMessageDao model.ChatMessageModel
}

func NewGetMessageListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMessageListLogic {
	return &GetMessageListLogic{
		ctx:            ctx,
		svcCtx:         svcCtx,
		Logger:         logx.WithContext(ctx),
		ChatMessageDao: model.NewChatMessageModel(svcCtx.Mysql),
	}
}

// GetMessageList 获取消息列表
func (l *GetMessageListLogic) GetMessageList(in *v1.GetMessageListRequest) (*v1.Response, error) {
	user, ok := middleware.GetUser(l.ctx)
	if !ok || user.UID == constants.InvalidUserID || user.Role == "" || user.Status != constants.UserStatusNormal {
		return bad("用户未登录或状态异常"), nil
	}

	if in.PageSize <= 0 {
		in.PageSize = DefaultPageSize
	}
	if in.PageSize > MaxPageSize {
		in.PageSize = MaxPageSize
	}

	messageCursor, err := decodeMessageCursor(in.Cursor)
	if err != nil {
		return bad("无效的游标"), nil
	}

	var targetUserID uint64
	if user.UID != in.UserId && in.UserId != constants.InvalidUserID {
		if user.Role != constants.SUPERADMIN && user.Role != constants.STAFF {
			return bad("没有权限查询其他用户的对话列表"), nil
		}
		targetUserID = in.UserId
	} else {
		targetUserID = user.UID
	}

	// 查询多一条记录，来判断是否有下一页
	limit := in.PageSize + 1

	var list []*model.ChatMessage

	if messageCursor == nil {
		// 获取最新的消息
		list, err = l.ChatMessageDao.FindManyByUserIDSessionID(l.ctx, targetUserID, in.SessionId, limit)
	} else {
		list, err = l.ChatMessageDao.FindManyByUserIDSessionIDWithCursor(l.ctx, targetUserID, in.SessionId, messageCursor.ID, limit)
	}

	if err != nil {
		return bad("查询消息列表失败"), nil
	}

	hasMore := len(list) > int(in.PageSize)
	if hasMore {
		list = list[:in.PageSize]
	}

	var nextCursor string
	if len(list) > 0 {
		last := list[len(list)-1]
		nextCursor = encodeMessageCursor(&MessageCursor{
			ID: last.Id,
		})
	}

	// 反转一下
	reverse.Slice(list)

	// 转换成pb
	messagesPB := convert.PBFromChatMessages(list)

	res := &v1.GetMessageListResponse{
		Messages:   messagesPB,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	}

	resAny, err := anypb.New(res)
	if err != nil {
		return bad("响应数据错误"), nil
	}

	return &v1.Response{
		Code:    200,
		Message: "获取消息列表成功",
		Data:    resAny,
	}, nil
}

type MessageCursor struct {
	ID uint64 `json:"id"`
}

func encodeMessageCursor(c *MessageCursor) string {
	if c == nil {
		return ""
	}
	raw := fmt.Sprintf("v1_%d", c.ID)
	return base64.StdEncoding.EncodeToString([]byte(raw))
}

func decodeMessageCursor(s string) (*MessageCursor, error) {
	if s == "" {
		return nil, nil
	}

	decodeBytes, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return nil, err
	}

	parts := strings.Split(string(decodeBytes), "_")
	if len(parts) != 2 || parts[0] != "v1" {
		return nil, fmt.Errorf("invalid cursor format")
	}

	id, err := strconv.ParseUint(parts[1], 10, 64)
	if err != nil {
		return nil, err
	}

	return &MessageCursor{ID: id}, nil
}
