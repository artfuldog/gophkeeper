package grpcapi

import (
	"context"
	"errors"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/artfuldog/gophkeeper/internal/crypt"
	"github.com/artfuldog/gophkeeper/internal/logger"
	"github.com/artfuldog/gophkeeper/internal/pb"
	"github.com/artfuldog/gophkeeper/internal/server/authorizer"
	"github.com/artfuldog/gophkeeper/internal/server/db"
)

// UsersSever implements all GRPC-method for handling users request and stores service options.
// Used for registering with GRPC-server.
type UsersService struct {
	pb.UnimplementedUsersServer
	db         db.DB
	logger     logger.L
	authorizer authorizer.A
}

// NewGRPCService a constructor for GRPCService.
func NewUsersService(db db.DB, l logger.L, a authorizer.A) *UsersService {
	return &UsersService{
		db:         db,
		logger:     l,
		authorizer: a,
	}
}

// CreateUser creates new user.
func (s *UsersService) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	componentName := "UsersService:CreateUser"
	resp := new(pb.CreateUserResponse)

	if req.User == nil {
		return nil, ErrMissedUserInfo
	}

	var err error

	TOTPKey := new(crypt.TOTPKey)
	if req.Twofactor {
		TOTPKey, err = crypt.GenerateTOTP(req.User.Username, "gophKeeper", 100, 100)
		if err != nil {
			return nil, errors.New("failed to create OTP")
		}
	}

	req.User.OtpKey = &TOTPKey.Secret
	if err := s.db.CreateUser(ctx, req.User); err != nil {
		s.logger.Warn(err, "db error", componentName)
		return nil, wrapErrorToClient(err)
	}

	if req.Twofactor {
		resp.Totpkey = new(pb.TOTPKey)
		resp.Totpkey.Secret = TOTPKey.Secret
		resp.Totpkey.Qrcode = TOTPKey.QRCode
	}

	resp.Info = fmt.Sprintf("successfully create user '%s'", req.User.Username)

	return resp, nil
}

// GetUser returns information about user.
func (s *UsersService) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
	componentName := "UsersService:GetUser"
	resp := new(pb.GetUserResponse)

	if !userPerformSelfOperation(ctx, req.Username) {
		return nil, permissionDeniedErr("access denied")
	}

	var err error
	if resp.User, err = s.db.GetUserByName(ctx, req.Username); err != nil {
		s.logger.Warn(err, "db error", componentName)
		return nil, wrapErrorToClient(err)
	}

	return resp, nil
}

// GetRevision returns user's revision number.
func (s *UsersService) GetRevision(ctx context.Context, req *pb.GetRevisionRequest) (*pb.GetRevisionResponse, error) {
	componentName := "UsersService:GetUserRevision"
	resp := new(pb.GetRevisionResponse)

	if !userPerformSelfOperation(ctx, req.Username) {
		return nil, permissionDeniedErr("access denied")
	}

	var err error
	if resp.Revision, err = s.db.GetUserRevision(ctx, req.Username); err != nil {
		s.logger.Warn(err, "db error", componentName)
		return nil, wrapErrorToClient(err)
	}

	return resp, nil
}

// UpdateUser updates user's information.
func (s *UsersService) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.UpdateUserResponse, error) {
	componentName := "UsersService:UpdateUser"
	resp := new(pb.UpdateUserResponse)

	if req.User == nil {
		return nil, ErrMissedUserInfo
	}

	if !userPerformSelfOperation(ctx, req.User.Username) {
		return nil, permissionDeniedErr("access denied")
	}

	if err := s.db.UpdateUser(ctx, req.User); err != nil {
		s.logger.Warn(err, "db error", componentName)
		return nil, wrapErrorToClient(err)
	}

	resp.Info = fmt.Sprintf("successfully update user '%s'", req.User.Username)

	return resp, nil
}

// DeleteUser deletes user and all related items.
func (s *UsersService) DeleteUser(ctx context.Context, req *pb.DeleteUserRequest) (*pb.DeleteUserResponse, error) {
	componentName := "UsersService:DeleteUser"
	resp := new(pb.DeleteUserResponse)

	if !userPerformSelfOperation(ctx, req.Username) {
		return nil, permissionDeniedErr("access denied")
	}

	if err := s.db.DeleteUserByName(ctx, req.Username); err != nil {
		s.logger.Warn(err, "db error", componentName)
		return nil, wrapErrorToClient(err)
	}

	s.logger.Info(fmt.Sprintf("user '%s' is deleted", req.Username), componentName)
	resp.Info = fmt.Sprintf("successfully delete user '%s'", req.Username)

	return resp, nil
}

// UserLogin performs user authentication and authorization.
//
// When 2-factor authorization is enabled and verification code is not provided returns response
// with SecondFactor flag and nil error. Handling this situation should be implemented on
// client side.
//
// After successful login responses with Token, encryption key and server's limits.
func (s *UsersService) UserLogin(ctx context.Context, req *pb.UserLoginRequest) (*pb.UserLoginResponse, error) {
	componentName := "UsersService:UserLogin"
	resp := new(pb.UserLoginResponse)

	pwdHash, optKey, err := s.db.GetUserAuthData(ctx, req.Username)
	if err != nil {
		s.logger.Warn(err, "db error", componentName)
		return nil, wrapErrorToClient(err)
	}

	if !crypt.CheckPasswordHashStr(req.Password, pwdHash) {
		return nil, status.Error(codes.PermissionDenied, "wrong password")
	}

	if optKey != "" {
		if req.OtpCode == "" {
			resp.SecondFactor = true
			return resp, nil
		}

		if !crypt.ValidateTOTP(req.OtpCode, optKey) {
			return nil, ErrWrongVerificationCode
		}
	}

	fields := authorizer.AuthFields{
		Username: req.Username,
	}
	if resp.Token, err = s.authorizer.CreateToken(fields); err != nil {
		return nil, status.Error(codes.Internal, "failed to generate token")
	}

	if resp.Ekey, err = s.db.GetUserEKey(ctx, req.Username); err != nil {
		return nil, status.Error(codes.Internal, "failed to fetch encryption key")
	}

	resp.ServerLimits = new(pb.ServerLimits)
	resp.ServerLimits.MaxSecretSize = int32(s.db.GetMaxSecretSize())

	return resp, nil
}

// userPerformSelfOperation is helper function which checks if user want to preform operation with his/her
// own account.
func userPerformSelfOperation(ctx context.Context, reqUserName string) bool {
	ctxUsername, ok := mdValueFromContext(ctx, "username")
	if ok {
		return reqUserName == ctxUsername
	}

	return false
}
