package utils

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"google.golang.org/grpc/metadata"
)

func GetUserFromMetadata(ctx context.Context) (uint64, string, int64, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return 0, "", 0, errors.New("metadata missing")
	}

	uidStr := md.Get("user_id")
	role := md.Get("user_role")
	statusStr := md.Get("user_status")

	fmt.Println("uidStr", uidStr)
	fmt.Println("role", role)
	fmt.Println("statusStr", statusStr)

	if len(uidStr) == 0 || len(role) == 0 || len(statusStr) == 0 {
		return 0, "", 0, errors.New("user metadata missing")
	}

	uid, err := strconv.ParseUint(uidStr[0], 10, 64)
	if err != nil {
		return 0, "", 0, err
	}

	status, err := strconv.ParseInt(statusStr[0], 10, 64)
	if err != nil {
		return 0, "", 0, err
	}

	return uid, role[0], status, nil
}
