package tasks

import (
	"encoding/json"
	"fmt"
)

type InviteCodeRelationTaskType string

const ()

type InviteCodeTask struct {
	Type InviteCodeRelationTaskType `json:"type"`
	UID  uint64                     `json:"uid"`
	Code string                     `json:"code"`
}

func (t *InviteCodeTask) ID() string {
	return fmt.Sprintf("%s:%s:%d:%s", InviteCodeRelationTaskPrefix, t.Type, t.UID, t.Code)
}

func (t *InviteCodeTask) Payload() ([]byte, error) {
	return json.Marshal(t)
}
