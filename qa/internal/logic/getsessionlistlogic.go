package logic

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/luyb177/XiaoAnBackend/infra/constants"
	"github.com/luyb177/XiaoAnBackend/infra/middleware"
	"github.com/luyb177/XiaoAnBackend/qa/internal/model"
	"github.com/luyb177/XiaoAnBackend/qa/internal/svc"
	"github.com/luyb177/XiaoAnBackend/qa/pb/qa/v1"
	"github.com/luyb177/XiaoAnBackend/qa/pkg/chatsession/convert"
)

type GetSessionListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
	ChatSessionDao model.ChatSessionModel
}

func NewGetSessionListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetSessionListLogic {
	return &GetSessionListLogic{
		ctx:            ctx,
		svcCtx:         svcCtx,
		Logger:         logx.WithContext(ctx),
		ChatSessionDao: model.NewChatSessionModel(svcCtx.Mysql),
	}
}

// GetSessionList 获取会话列表
func (l *GetSessionListLogic) GetSessionList(in *v1.GetSessionListRequest) (*v1.Response, error) {
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

	sessionCursor, err := decodeCursor(in.Cursor)
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

	var (
		list []*model.ChatSession
	)

	if sessionCursor == nil {
		// 首次查询
		list, err = l.ChatSessionDao.FindManyByUserID(l.ctx, targetUserID, limit)
	} else {
		// 通过游标查询
		list, err = l.ChatSessionDao.FindManyByUserIDWithCursor(l.ctx, targetUserID, sessionCursor.LastMessageAt, sessionCursor.ID, limit)
	}
	if err != nil {
		return bad("查询失败"), err
	}

	hasMore := len(list) > int(in.PageSize)
	if hasMore {
		list = list[:in.PageSize]
	}

	var nextCursorStr string
	if hasMore && len(list) > 0 {
		last := list[len(list)-1]
		nextCursorStr = encodeCursor(&SessionCursor{
			LastMessageAt: last.LastMessageAt,
			ID:            last.Id,
		})
	}

	sessionsPB := convert.PBFromChatSessions(list)

	res := &v1.GetSessionListResponse{
		Sessions:   sessionsPB,
		NextCursor: nextCursorStr,
		HasMore:    hasMore,
	}

	resAny, err := anypb.New(res)
	if err != nil {
		l.Errorf("GetSessionList anypb.New error: %v", err)
		return internal("封装返回结果失败"), nil
	}

	return &v1.Response{
		Code:    200,
		Message: "获取对话列表成功",
		Data:    resAny,
	}, nil
}

type SessionCursor struct {
	LastMessageAt time.Time `json:"last_message_at"`
	ID            uint64    `json:"id"`
}

func encodeCursor(cursor *SessionCursor) string {
	if cursor == nil {
		return ""
	}
	// 版本号 v1
	raw := fmt.Sprintf("v1_%d_%d", cursor.LastMessageAt.UnixMilli(), cursor.ID)
	return base64.StdEncoding.EncodeToString([]byte(raw))
}

func decodeCursor(encodedCursor string) (*SessionCursor, error) {
	// 第一页
	if encodedCursor == "" {
		return nil, nil
	}

	decodedBytes, err := base64.StdEncoding.DecodeString(encodedCursor)
	if err != nil {
		return nil, err
	}

	parts := strings.Split(string(decodedBytes), "_")
	if len(parts) != 3 {
		return nil, errors.New("invalid cursor")
	}

	if parts[0] != "v1" {
		return nil, errors.New("unsupported cursor version")
	}

	ts, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return nil, err
	}

	id, err := strconv.ParseUint(parts[2], 10, 64)
	if err != nil {
		return nil, err
	}

	return &SessionCursor{
		LastMessageAt: time.UnixMilli(ts),
		ID:            id,
	}, nil
}
