package messageid

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

func NewMessageID(sessionID uint64, userID uint64, clientMsgID string) string {
	//  拼接原始字符串（必须稳定）
	raw := fmt.Sprintf("%d:%d:%s", sessionID, userID, clientMsgID)

	//  SHA-256
	sum := sha256.Sum256([]byte(raw))

	//  转 hex（64 chars）
	return hex.EncodeToString(sum[:])
}
