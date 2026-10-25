package service

import (
	"context"
	"strconv"
	"strings"

	"github.com/gofreego/goutils/logger"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// getProfileID returns the caller's profile from the gateway's headers: X-Profile-Id is the
// profile acting, X-Profile-Ids the profiles the signed-in user owns. (Same check as
// learnerservice.)
func getProfileID(ctx context.Context) (int64, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return 0, status.Error(codes.Unauthenticated, "metadata is not provided")
	}
	values := md.Get("X-Profile-Id")
	if len(values) == 0 {
		return 0, status.Error(codes.InvalidArgument, "X-Profile-Id header is missing")
	}
	profileID, err := strconv.ParseInt(values[0], 10, 64)
	if err != nil {
		return 0, status.Error(codes.InvalidArgument, "invalid X-Profile-Id")
	}

	owned := md.Get("X-Profile-Ids")
	if len(owned) == 0 {
		logger.Error(ctx, "getProfileID: X-Profile-Ids header is missing")
		return 0, status.Error(codes.Unauthenticated, "unauthenticated")
	}
	for _, id := range strings.Split(owned[0], ",") {
		if strings.TrimSpace(id) == values[0] {
			return profileID, nil
		}
	}
	logger.Warn(ctx, "getProfileID: profile %d not among the user's (%s)", profileID, owned[0])
	return 0, status.Error(codes.PermissionDenied, "permission denied")
}
