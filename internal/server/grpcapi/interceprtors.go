package grpcapi

import (
	"context"
	"strings"

	"github.com/artfuldog/gophkeeper/internal/common"
	"github.com/artfuldog/gophkeeper/internal/server/authorizer"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// List of methods which not required authorization
var unAuthMethods = []string{
	"CreateUser",
	"UserLogin",
}

const (
	authMetadataKey = "authorization"
	authUsernameKey = "username"
)

// isAuthorized is gRPC interceptor.
func IsAuthorized(auth authorizer.A) grpc.UnaryServerInterceptor {
	return func(ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler) (interface{}, error) {

		method := common.Last(strings.Split(info.FullMethod, "/"))
		if common.Contains(method, unAuthMethods) {
			return handler(ctx, req)
		}

		var username, token string
		username, ok := mdValueFromContext(ctx, authUsernameKey)
		if !ok {
			return nil, status.Error(codes.PermissionDenied, "cannot retrieve user name")
		}
		token, ok = mdValueFromContext(ctx, authMetadataKey)
		if !ok {
			return nil, status.Error(codes.PermissionDenied, "cannot retrieve token")
		}

		fields := authorizer.AuthFields{
			Username: username,
		}
		if err := auth.VerifyToken(token, fields); err != nil {
			return nil, status.Error(codes.PermissionDenied, err.Error())
		}

		return handler(ctx, req)
	}
}
