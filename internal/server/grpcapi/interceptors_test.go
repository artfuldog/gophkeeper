package grpcapi

import (
	"testing"

	"github.com/artfuldog/gophkeeper/internal/mocks/mockauth"
	"github.com/artfuldog/gophkeeper/internal/pb"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func TestIsAuthorized(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	authorizer := mockauth.NewMockA(mockCtrl)

	ts, tsErr := NewTestSuiteGRPCServer(t, grpc.UnaryInterceptor(IsAuthorized(authorizer)))
	if tsErr != nil {
		t.Errorf("failed to init test suite: %v", tsErr)
	}
	defer ts.Stop()

	t.Run("Unauth method", func(t *testing.T) {
		req := &pb.CreateUserRequest{
			User: &pb.User{
				Username: "newuser",
			},
			Twofactor: true,
		}
		ts.DB.EXPECT().CreateUser(mockAny, mockAny).Return(nil)
		resp, err := ts.UsersClient.CreateUser(testCtx, req)
		require.NoError(t, err)
		assert.NotEmpty(t, resp)
	})

	t.Run("Missed context", func(t *testing.T) {
		//ts.DB.EXPECT().DeleteItem(mockAny, mockAny, mockAny).Return(nil)
		req := &pb.DeleteItemRequest{}
		_, err := ts.ItemsClient.DeleteItem(testCtx, req)
		assert.Error(t, err)
	})

	authCtx := metadata.AppendToOutgoingContext(testCtx, authUsernameKey, "CorrectUser")
	t.Run("Missed authorization field", func(t *testing.T) {
		//ts.DB.EXPECT().DeleteItem(mockAny, mockAny, mockAny).Return(nil)
		req := &pb.DeleteItemRequest{}
		_, err := ts.ItemsClient.DeleteItem(authCtx, req)
		assert.Error(t, err)
	})

	authCtx = metadata.AppendToOutgoingContext(authCtx, authMetadataKey, "token")

	t.Run("Missed authorization field", func(t *testing.T) {
		authorizer.EXPECT().VerifyToken(mockAny, mockAny).Return(assert.AnError)
		req := &pb.DeleteItemRequest{}
		_, err := ts.ItemsClient.DeleteItem(authCtx, req)
		assert.Error(t, err)
	})

	t.Run("Succesfully authorized", func(t *testing.T) {
		authorizer.EXPECT().VerifyToken(mockAny, mockAny).Return(nil)
		ts.DB.EXPECT().DeleteItem(mockAny, mockAny, mockAny).Return(nil)
		req := &pb.DeleteItemRequest{}
		resp, err := ts.ItemsClient.DeleteItem(authCtx, req)
		require.NoError(t, err)
		assert.NotEmpty(t, resp)
	})
}
