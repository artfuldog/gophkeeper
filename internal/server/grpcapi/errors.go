package grpcapi

import (
	"errors"

	"github.com/artfuldog/gophkeeper/internal/server/db"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ErrMissedUserInfo        = status.Error(codes.InvalidArgument, "missed user information")
	ErrWrongVerificationCode = status.Error(codes.PermissionDenied, "wrong verification code")
)

// permissionDeniedErr is helper function for return error with status code PermissionDenied and
// info message.
func permissionDeniedErr(message string) error {
	return status.Errorf(codes.PermissionDenied, message)
}

// wrapErrorToClient wraps known type of Database' errors for sending to client.
//
// If wrapErrorToClient receives already wrapped error, it extracts original error message and
// wraps it to new error with defined status code, i.e. substracts from error all internal details.
// If received error is not wrapped wrapErrorToClient leaves error's text description as is in message.
func wrapErrorToClient(err error) error {
	message := err.Error()

	if unwrappedErr := errors.Unwrap(err); unwrappedErr != nil {
		message = unwrappedErr.Error()
	}

	if errors.Is(err, db.ErrNotFound) {
		return status.Error(codes.NotFound, message)
	}
	if errors.Is(err, db.ErrDuplicateEntry) {
		return status.Error(codes.InvalidArgument, message)
	}
	if errors.Is(err, db.ErrTransactionFailed) {
		return status.Error(codes.Internal, message)
	}
	if errors.Is(err, db.ErrBadSQLQuery) {
		return status.Error(codes.InvalidArgument, message)
	}

	return status.Error(codes.Unknown, err.Error())
}
