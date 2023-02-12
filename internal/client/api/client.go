package api

import (
	"context"
	"errors"

	"github.com/artfuldog/gophkeeper/internal/client/storage"
	"github.com/artfuldog/gophkeeper/internal/pb"
)

var (
	// Indicated that session expired, most often because of toke expiration.
	// Used as an indicator for UI to request password/OTP from user.
	ErrSessionExpired       = errors.New("session expired")
	ErrMissedServerResponse = errors.New("missed server response")
	ErrSecondFactorRequired = errors.New("second factor is required")
	ErrEKeyEncryptionFailed = errors.New("failed to encrypt self encryption key")
	ErrEKeyDecryptionFailed = errors.New("failed to decrypt received from server encryption key")
	ErrSecretTooBig         = errors.New("size of secret is too big")
	ErrOutOfSync            = errors.New("local and server's information are out of sync")
)

// Client is a general API-Client interface.
type Client interface {
	Executor
	UsersInteractor
	ItemsInteractor
	Cryptor
	Storager
}

// Executor defines methods for controlling and maintaining client and its connections.
type Executor interface {
	// Starts connection to server.
	// Context and control channel should be provided for correct client stop.
	Connect(context.Context, chan<- struct{}) error
}

// UsersInteractor defines methods for processing users-related events, such as login, AAA, registration.
type UsersInteractor interface {
	// Performs user login with password-based and OTP(optional) authentication and authorization.
	// Returns ErrSecondFactorRequired if OTP-auth is enabled and required.
	UserLogin(ctx context.Context, username, password, optCode string) error
	// Registers new user.
	UserRegister(context.Context, *NewUser) (*TOTPKey, error)
}

// ItemsInteractor defines methods for processing items-related events (CRUD).
type ItemsInteractor interface {
	// Returns items' list of active user.
	// Short representation (without sensitive information) of items is used.
	GetItemsList(context.Context) ([]*pb.ItemShort, error)
	// Returns full item's information
	GetItem(ctx context.Context, itemName, itemType string) (*Item, error)
	// Returns several item's information
	GetItemsForStorage(ctx context.Context, itemsIDs []int64) (storage.Items, error)
	// Creates or updates item.
	SaveItem(context.Context, *Item) error
	// Deletes item.
	DeleteItem(context.Context, *Item) error
}

// Cryptor defines methods for encrypt/decrypt data.
type Cryptor interface {
	// Encrypts item for sending to server.
	EncryptPbItem(*pb.Item) error
	// Decrypts item received from server.
	DecryptPbItem(*pb.Item) error
}

// Storager defines methods for prefrom operation with local storage.
type Storager interface {
	// Setup local storage.
	StorageInit(context.Context, chan<- string) error
	// Performs regular synchrinization with server and passed status through channel
	Sync(context.Context, chan<- string)
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
