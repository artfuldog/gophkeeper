package api

import (
	"context"
	"strings"

	"github.com/artfuldog/gophkeeper/internal/client/config"
	"github.com/artfuldog/gophkeeper/internal/common"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// List of methods which not required authorization
var unAuthMethods = []string{
	"UserLogin",
	"UserRegister",
}

// Metadata keys
const (
	authMetadataKey = "authorization"
	authUsernameKey = "username"
)

// AuthInterceptor insert into egress requests authorization information - username and token.
//
// Methods, for which inserting authorizaion information is not required, described in unAuthMethods slice.
func AuthInterceptor(config config.Configer, token *string) grpc.UnaryClientInterceptor {
	return func(ctx context.Context,
		method string,
		req, reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption) error {

		if common.Contains(common.Last(strings.Split(method, "/")), unAuthMethods) {
			return invoker(ctx, method, req, reply, cc, opts...)
		}

		authCtx := metadata.AppendToOutgoingContext(ctx, authUsernameKey, config.GetUser())
		authCtx = metadata.AppendToOutgoingContext(authCtx, authMetadataKey, *token)
		return invoker(authCtx, method, req, reply, cc, opts...)
	}
}
