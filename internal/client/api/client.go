package api

import (
	"context"
	"errors"

	"github.com/artfuldog/gophkeeper/internal/pb"
)

var (
	// Indicated that session expired, most often because of toke expiration.
	// Used as an indicator for UI to request password/OTP from user.
	ErrSessionExpired       = errors.New("session expired")
	ErrMissedServerResponce = errors.New("missed server response")
	ErrSecondFactorRequired = errors.New("second factor is required")
	ErrEKeyEncryptionFailed = errors.New("failed to encrypt self encryption key")
	ErrEKeyDecryptionFailed = errors.New("failed to decrypt received from server encryption key")
	ErrSecretTooBig         = errors.New("size of secret is too big")
)

// Client represents client intreface for connection and iteraction with server.
type Client interface {
	// Starts connection to server.
	Connect(context.Context) error

	// Performs user login with password-based and OTP(optional) authentication and authorization.
	// Returns ErrSecondFactorRequired if OTP-auth is enabled and required.
	UserLogin(ctx context.Context, username, password, optCode string) error
	// Registers new user.
	UserRegister(ctx context.Context, user *NewUser) (*TOTPKey, error)

	// Returns items' list of active user.
	// Short representation (without sensitive information) of items is used.
	GetItemsList(ctx context.Context) ([]*pb.ItemShort, error)
	// Returns full item's information
	GetItem(ctx context.Context, itemName, itemType string) (*Item, error)
	// Creates or updates item.
	SaveItem(ctx context.Context, item *Item) error
	// Deletes item.
	DeleteItem(ctx context.Context, item *Item) error

	// Encrypts item for sending to server.
	EncryptPbItem(item *pb.Item) error
	// Decrypts item received from server.
	DecryptPbItem(item *pb.Item) error
}

// NewUser represents new user's registration information.
type NewUser struct {
	Username        string
	Password        string
	PasswordConfirm string
	SecretKey       string
	Email           string
	TwoFactorEnable bool
}
