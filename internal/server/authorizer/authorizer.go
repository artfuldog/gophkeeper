// Package authorizer represents interface for entity performing user authozriation.
package authorizer

import (
	"errors"
	"fmt"
	"time"

	"github.com/artfuldog/gophkeeper/internal/logger"
)

// Supported authorization methods.
const (
	TypeJWT    = "jwt"
	TypePaseto = "paseto"
	TypeYesMan = "yesman"
)

// Errors.
var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("expired token")
)

// A represents general authorizer inteface.
type A interface {
	// Creates new token
	CreateToken(fields AuthFields) (string, error)
	// Verify token
	VerifyToken(token string, fields AuthFields) error
}

// AuthorizeItems contains possible parameters for authorization.
type AuthFields struct {
	Username string
}

// New is a fabric method for create Authorizer with provided type.
func New(aType string, key string, tokenDuration time.Duration, logger logger.L) (A, error) {
	switch aType {
	case TypeJWT:
		return nil, errors.New("unimplemented")
	case TypePaseto:
		return NewPasetoAuthorizer(key, tokenDuration)
	case TypeYesMan:
		return NewYesManAuthorizer(logger), nil
	default:
		return nil, fmt.Errorf("unknown authorizer type: %s", aType)
	}
}
