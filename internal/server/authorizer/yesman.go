package authorizer

import (
	"github.com/artfuldog/gophkeeper/internal/logger"
)

// YesManAuthorizer represents test implementation of Authorizer.
//
// Should be used only for tested purposes. Always succesfully authorizes user.
type YesManAuthorizer struct {
	logger logger.L
}

var _ A = (*YesManAuthorizer)(nil)

// NewYesManAuthorizer is used to create new YesManAuthorizer instance.
func NewYesManAuthorizer(logger logger.L) *YesManAuthorizer {
	return &YesManAuthorizer{
		logger: logger,
	}
}

// CreateToken is a dummy function for create token. Always return nil-error.
func (a YesManAuthorizer) CreateToken(fields AuthFields) (string, error) {
	a.logger.Info("yes", "createtoken")
	return "", nil
}

// VerifyToken is a dummy function for verifying token. Always return nil-error.
func (a YesManAuthorizer) VerifyToken(token string, fields AuthFields) error {
	a.logger.Info("yes", "VerifyToken")
	return nil
}
