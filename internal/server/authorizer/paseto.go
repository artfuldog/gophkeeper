package authorizer

import (
	"fmt"
	"time"

	"github.com/aead/chacha20poly1305"
	"github.com/o1egl/paseto"
)

// PasetoAuthorizer represents implementation of Authorizer based on Paseto Tokens.
type PasetoAuthorizer struct {
	paseto        *paseto.V2
	key           []byte
	tokenDuration time.Duration
}

var _ A = (*PasetoAuthorizer)(nil)

// NewPasetoAuthorizer creates new Paseto Authorizer.
func NewPasetoAuthorizer(key string, tokenDuration time.Duration) (*PasetoAuthorizer, error) {
	if len(key) != chacha20poly1305.KeySize {
		return nil, fmt.Errorf("invalid key size: got %d, must be exactly %d characters",
			len(key), chacha20poly1305.KeySize)
	}

	a := &PasetoAuthorizer{
		paseto:        paseto.NewV2(),
		key:           []byte(key),
		tokenDuration: tokenDuration,
	}

	return a, nil
}

// CreateToken creates a token for a specific authorization fields and duration.
func (a *PasetoAuthorizer) CreateToken(fields AuthFields) (string, error) {
	payload, err := NewPayload(fields.Username, a.tokenDuration)
	if err != nil {
		return "", err
	}

	return a.paseto.Encrypt(a.key, payload, nil)
}

// VerifyToken checks if the token is valid or not.
func (a *PasetoAuthorizer) VerifyToken(token string, fields AuthFields) error {
	payload := new(Payload)

	err := a.paseto.Decrypt(token, a.key, payload, nil)
	if err != nil {
		return err
	}

	err = payload.Valid(fields)
	if err != nil {
		return err
	}

	return nil
}
