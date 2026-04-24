package logic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// syncContentDisplayAfterUserUpdate 与 content 服务共用 xiaoan 库：在昵称/头像变更后
// 刷新生成内容上的作者名、评论里存的昵称/头像快照。
// 仅按 last_modified_by / user_id 匹配；失败不阻塞用户资料更新，只打日志。
func syncContentDisplayAfterUserUpdate(
	ctx context.Context,
	conn sqlx.SqlConn,
	log logx.Logger,
	userID uint64,
	newName, newAvatar string,
) {
	if newName == "" && newAvatar == "" {
		return
	}
	lm := int64(userID)

	if newName != "" {
		updates := []struct {
			table string
		}{
			{table: "article"},
			{table: "video"},
			{table: "comic"},
			{table: "podcast"},
		}
		for _, u := range updates {
			q := "UPDATE `" + u.table + "` SET `author` = ? WHERE `last_modified_by` = ? AND `deleted_at` = 0"
			if _, err := conn.ExecCtx(ctx, q, newName, lm); err != nil {
				log.Errorf("sync author on %s for user %d: %v", u.table, userID, err)
			}
		}
		qc := "UPDATE `comment` SET `nickname` = ? WHERE `user_id` = ? AND `deleted_at` = 0"
		if _, err := conn.ExecCtx(ctx, qc, newName, userID); err != nil {
			log.Errorf("sync comment nickname for user %d: %v", userID, err)
		}
	}

	if newAvatar != "" {
		qc := "UPDATE `comment` SET `avatar` = ? WHERE `user_id` = ? AND `deleted_at` = 0"
		if _, err := conn.ExecCtx(ctx, qc, newAvatar, userID); err != nil {
			log.Errorf("sync comment avatar for user %d: %v", userID, err)
		}
	}
}
