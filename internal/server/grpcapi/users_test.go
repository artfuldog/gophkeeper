package grpcapi

import (
	"testing"

	"github.com/artfuldog/gophkeeper/internal/common"
	"github.com/artfuldog/gophkeeper/internal/crypt"
	"github.com/artfuldog/gophkeeper/internal/mocks/mockauth"
	"github.com/artfuldog/gophkeeper/internal/mocks/mockdb"
	"github.com/artfuldog/gophkeeper/internal/mocks/mocklogger"
	"github.com/artfuldog/gophkeeper/internal/pb"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/metadata"
)

func TestNewUsersService(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	assert.NotEmpty(t, NewUsersService(mockdb.NewMockDB(mockCtrl),
		mocklogger.NewMockLogger(), mockauth.NewMockA(mockCtrl)))
	mockCtrl.Finish()
}

func TestUsersService_CreateUser(t *testing.T) {
	ts, tsErr := NewTestSuiteGRPCServer(t)
	if tsErr != nil {
		t.Errorf("failed to init test suite: %v", tsErr)
	}
	defer ts.Stop()

	t.Run("Empty user", func(t *testing.T) {
		req := &pb.CreateUserRequest{}
		_, err := ts.UsersClient.CreateUser(testCtx, req)
		assert.ErrorIs(t, err, ErrMissedUserInfo)
	})

	t.Run("TOTP key generation error", func(t *testing.T) {
		req := &pb.CreateUserRequest{
			User: &pb.User{
				Email: common.PtrTo("abc@exampl.com"),
			},
			Twofactor: true,
		}
		_, err := ts.UsersClient.CreateUser(testCtx, req)
		assert.Error(t, err)
	})

	t.Run("Database returns error", func(t *testing.T) {
		req := &pb.CreateUserRequest{
			User: &pb.User{
				Username: "newuser",
			},
			Twofactor: true,
		}
		ts.DB.EXPECT().CreateUser(mockAny, mockAny).Return(assert.AnError)
		_, err := ts.UsersClient.CreateUser(testCtx, req)
		assert.Error(t, err)
	})

	t.Run("Succesfully created", func(t *testing.T) {
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

}

func TestUsersService_GetUser(t *testing.T) {
	ts, tsErr := NewTestSuiteGRPCServer(t)
	if tsErr != nil {
		t.Errorf("failed to init test suite: %v", tsErr)
	}
	defer ts.Stop()
	authCtx := metadata.AppendToOutgoingContext(testCtx, authUsernameKey, "CorrectUser")

	t.Run("Missed Context", func(t *testing.T) {
		req := &pb.GetUserRequest{}
		_, err := ts.UsersClient.GetUser(testCtx, req)
		assert.ErrorIs(t, err, permissionDeniedErr("access denied"))
	})

	t.Run("Empty user", func(t *testing.T) {
		req := &pb.GetUserRequest{
			Username: "Wrong username",
		}
		_, err := ts.UsersClient.GetUser(authCtx, req)
		assert.ErrorIs(t, err, permissionDeniedErr("access denied"))
	})

	t.Run("Database returns error", func(t *testing.T) {
		req := &pb.GetUserRequest{
			Username: "CorrectUser",
		}
		ts.DB.EXPECT().GetUserByName(mockAny, mockAny).Return(nil, assert.AnError)
		_, err := ts.UsersClient.GetUser(authCtx, req)
		assert.Error(t, err)
	})

	t.Run("Succesfully get user", func(t *testing.T) {
		req := &pb.GetUserRequest{
			Username: "CorrectUser",
		}
		respUser := &pb.User{
			Username: "CorrectUser",
		}
		ts.DB.EXPECT().GetUserByName(mockAny, mockAny).Return(respUser, nil)
		resp, err := ts.UsersClient.GetUser(authCtx, req)
		require.NoError(t, err)
		assert.NotEmpty(t, resp)
	})
}

func TestUsersService_GetRevision(t *testing.T) {
	ts, tsErr := NewTestSuiteGRPCServer(t)
	if tsErr != nil {
		t.Errorf("failed to init test suite: %v", tsErr)
	}
	defer ts.Stop()

	authCtx := metadata.AppendToOutgoingContext(testCtx, authUsernameKey, "CorrectUser")

	t.Run("Missed Context", func(t *testing.T) {
		req := &pb.GetRevisionRequest{}
		_, err := ts.UsersClient.GetRevision(testCtx, req)
		assert.ErrorIs(t, err, permissionDeniedErr("access denied"))
	})

	t.Run("Empty user", func(t *testing.T) {
		req := &pb.GetRevisionRequest{
			Username: "Wrong username",
		}
		_, err := ts.UsersClient.GetRevision(authCtx, req)
		assert.ErrorIs(t, err, permissionDeniedErr("access denied"))
	})

	t.Run("Database returns error", func(t *testing.T) {
		req := &pb.GetRevisionRequest{
			Username: "CorrectUser",
		}
		ts.DB.EXPECT().GetUserRevision(mockAny, mockAny).Return(nil, assert.AnError)
		_, err := ts.UsersClient.GetRevision(authCtx, req)
		assert.Error(t, err)
	})

	t.Run("Succesfully get user", func(t *testing.T) {
		req := &pb.GetRevisionRequest{
			Username: "CorrectUser",
		}
		ts.DB.EXPECT().GetUserRevision(mockAny, mockAny).Return([]byte("revision"), nil)
		resp, err := ts.UsersClient.GetRevision(authCtx, req)
		require.NoError(t, err)
		assert.NotEmpty(t, resp)
	})
}

func TestUsersService_UpdateUser(t *testing.T) {
	ts, tsErr := NewTestSuiteGRPCServer(t)
	if tsErr != nil {
		t.Errorf("failed to init test suite: %v", tsErr)
	}
	defer ts.Stop()
	authCtx := metadata.AppendToOutgoingContext(testCtx, authUsernameKey, "CorrectUser")

	t.Run("Empty user", func(t *testing.T) {
		req := &pb.UpdateUserRequest{}
		_, err := ts.UsersClient.UpdateUser(testCtx, req)
		assert.ErrorIs(t, err, ErrMissedUserInfo)
	})

	t.Run("Missed Context", func(t *testing.T) {
		req := &pb.UpdateUserRequest{
			User: &pb.User{
				Username: "Wrong username",
			},
		}
		_, err := ts.UsersClient.UpdateUser(testCtx, req)
		assert.ErrorIs(t, err, permissionDeniedErr("access denied"))
	})

	t.Run("Wrong user user", func(t *testing.T) {
		req := &pb.UpdateUserRequest{
			User: &pb.User{
				Username: "Wrong username",
			},
		}
		_, err := ts.UsersClient.UpdateUser(authCtx, req)
		assert.ErrorIs(t, err, permissionDeniedErr("access denied"))
	})

	t.Run("Database returns error", func(t *testing.T) {
		req := &pb.UpdateUserRequest{
			User: &pb.User{
				Username: "CorrectUser",
			},
		}
		ts.DB.EXPECT().UpdateUser(mockAny, mockAny).Return(assert.AnError)
		_, err := ts.UsersClient.UpdateUser(authCtx, req)
		assert.Error(t, err)
	})

	t.Run("Succesfully update user", func(t *testing.T) {
		req := &pb.UpdateUserRequest{
			User: &pb.User{
				Username: "CorrectUser",
			},
		}
		ts.DB.EXPECT().UpdateUser(mockAny, mockAny).Return(nil)
		resp, err := ts.UsersClient.UpdateUser(authCtx, req)
		require.NoError(t, err)
		assert.NotEmpty(t, resp)
	})
}

func TestUsersService_DeleteUser(t *testing.T) {
	ts, tsErr := NewTestSuiteGRPCServer(t)
	if tsErr != nil {
		t.Errorf("failed to init test suite: %v", tsErr)
	}
	defer ts.Stop()
	authCtx := metadata.AppendToOutgoingContext(testCtx, authUsernameKey, "CorrectUser")

	t.Run("Missed Context", func(t *testing.T) {
		req := &pb.DeleteUserRequest{
			Username: "CorrectUser",
		}
		_, err := ts.UsersClient.DeleteUser(testCtx, req)
		assert.ErrorIs(t, err, permissionDeniedErr("access denied"))
	})

	t.Run("Wrong user user", func(t *testing.T) {
		req := &pb.DeleteUserRequest{
			Username: "Wrong user",
		}
		_, err := ts.UsersClient.DeleteUser(authCtx, req)
		assert.ErrorIs(t, err, permissionDeniedErr("access denied"))
	})

	t.Run("Database returns error", func(t *testing.T) {
		req := &pb.DeleteUserRequest{
			Username: "CorrectUser",
		}
		ts.DB.EXPECT().DeleteUserByName(mockAny, mockAny).Return(assert.AnError)
		_, err := ts.UsersClient.DeleteUser(authCtx, req)
		assert.Error(t, err)
	})

	t.Run("Succesfully delete user", func(t *testing.T) {
		req := &pb.DeleteUserRequest{
			Username: "CorrectUser",
		}
		ts.DB.EXPECT().DeleteUserByName(mockAny, mockAny).Return(nil)
		resp, err := ts.UsersClient.DeleteUser(authCtx, req)
		require.NoError(t, err)
		assert.NotEmpty(t, resp)
	})
}

func TestUsersService_UserLogin(t *testing.T) {
	ts, tsErr := NewTestSuiteGRPCServer(t)
	if tsErr != nil {
		t.Errorf("failed to init test suite: %v", tsErr)
	}
	defer ts.Stop()

	testPassword := "TestPassword!@34"
	testPwdHash, _ := crypt.CalculatePasswordHash(testPassword)

	t.Run("Database returns error", func(t *testing.T) {
		req := &pb.UserLoginRequest{}
		ts.DB.EXPECT().GetUserAuthData(mockAny, mockAny).Return("", "", assert.AnError)
		_, err := ts.UsersClient.UserLogin(testCtx, req)
		assert.Error(t, err)
	})

	t.Run("Wrong password hash", func(t *testing.T) {
		req := &pb.UserLoginRequest{}
		ts.DB.EXPECT().GetUserAuthData(mockAny, mockAny).Return("wronghash", "", nil)
		_, err := ts.UsersClient.UserLogin(testCtx, req)
		assert.Error(t, err)
	})

	t.Run("Second factor required", func(t *testing.T) {
		req := &pb.UserLoginRequest{
			Password: testPassword,
		}
		ts.DB.EXPECT().GetUserAuthData(mockAny, mockAny).Return(testPwdHash, "key", nil)
		resp, err := ts.UsersClient.UserLogin(testCtx, req)
		require.NoError(t, err)
		assert.Equal(t, resp.SecondFactor, true)
	})

	t.Run("Wrong verificaion code", func(t *testing.T) {
		req := &pb.UserLoginRequest{
			Password: testPassword,
			OtpCode:  "somekey",
		}
		ts.DB.EXPECT().GetUserAuthData(mockAny, mockAny).Return(testPwdHash, "key", nil)
		_, err := ts.UsersClient.UserLogin(testCtx, req)
		assert.ErrorIs(t, err, ErrWrongVerificationCode)
	})

	t.Run("Create token error", func(t *testing.T) {
		verCode, err := crypt.GenerateVerificationCode("CROOWIM25UJJ5JFJ23UV4QCODWRFKIO2")
		require.NoError(t, err)
		req := &pb.UserLoginRequest{
			Username: "CorrectUser",
			Password: testPassword,
			OtpCode:  verCode,
		}
		ts.DB.EXPECT().GetUserAuthData(mockAny, mockAny).Return(testPwdHash, "CROOWIM25UJJ5JFJ23UV4QCODWRFKIO2", nil)
		ts.Authorizer.EXPECT().CreateToken(mockAny).Return("", assert.AnError)

		_, err = ts.UsersClient.UserLogin(testCtx, req)
		assert.Error(t, err)
	})

	t.Run("Get user encryption key error", func(t *testing.T) {
		verCode, err := crypt.GenerateVerificationCode("CROOWIM25UJJ5JFJ23UV4QCODWRFKIO2")
		require.NoError(t, err)
		req := &pb.UserLoginRequest{
			Username: "CorrectUser",
			Password: testPassword,
			OtpCode:  verCode,
		}
		ts.DB.EXPECT().GetUserAuthData(mockAny, mockAny).Return(testPwdHash, "CROOWIM25UJJ5JFJ23UV4QCODWRFKIO2", nil)
		ts.Authorizer.EXPECT().CreateToken(mockAny).Return("token", nil)
		ts.DB.EXPECT().GetUserEKey(mockAny, "CorrectUser").Return(nil, assert.AnError)

		_, err = ts.UsersClient.UserLogin(testCtx, req)
		assert.Error(t, err)
	})

	t.Run("Succesfull two-factor login", func(t *testing.T) {
		verCode, err := crypt.GenerateVerificationCode("CROOWIM25UJJ5JFJ23UV4QCODWRFKIO2")
		require.NoError(t, err)
		req := &pb.UserLoginRequest{
			Username: "CorrectUser",
			Password: testPassword,
			OtpCode:  verCode,
		}
		ts.DB.EXPECT().GetUserAuthData(mockAny, mockAny).Return(testPwdHash, "CROOWIM25UJJ5JFJ23UV4QCODWRFKIO2", nil)
		ts.Authorizer.EXPECT().CreateToken(mockAny).Return("token", nil)
		ts.DB.EXPECT().GetUserEKey(mockAny, "CorrectUser").Return([]byte("encryption key"), nil)
		ts.DB.EXPECT().GetMaxSecretSize().Return(uint32(12345))

		resp, err := ts.UsersClient.UserLogin(testCtx, req)
		require.NoError(t, err)
		assert.Equal(t, resp.Ekey, []byte("encryption key"))
		assert.Equal(t, resp.ServerLimits.MaxSecretSize, int32(12345))
	})
}
