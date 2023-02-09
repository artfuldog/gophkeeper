package api

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"

	"github.com/artfuldog/gophkeeper/internal/client/config"
	"github.com/artfuldog/gophkeeper/internal/common"
	"github.com/artfuldog/gophkeeper/internal/crypt"
	"github.com/artfuldog/gophkeeper/internal/logger"
	"github.com/artfuldog/gophkeeper/internal/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

// GRPCClient respesents client implementation based on gRPC.
type GRPCClient struct {
	// gRPC-client for Users service.
	usersClient pb.UsersClient
	// gRPC-client for Items service.
	itemsClient pb.ItemsClient

	config *config.Configer
	// parameters GRPCParameters
	Logger logger.L

	// Auth token
	Token string
	// Max secret size
	MaxSecretSize uint32

	// Encryption key. Used for encrypting/decrypting all sended/received
	// sensitive information in Items.
	// encKey stored in encrypted form on server. CLient receives this key
	// after successful authentication and authorization, for decrypting this key
	// client uses secretkey which stored unecrypted on user's side.
	encKey []byte
}

var _ Client = (*GRPCClient)(nil)

// NewGRPCClient creates new GRPCClient instance.
func NewGRPCClient(config *config.Configer, l logger.L) *GRPCClient {
	return &GRPCClient{
		config: config,
		Logger: l,
	}
}

// Connect intiates connections and starts client for gRPC-services.
func (c *GRPCClient) Connect(ctx context.Context) error {
	componentName := "GRPCClient:Connect"

	creds, err := c.getCredentials()
	if err != nil {
		return err
	}

	conn, err := grpc.Dial(c.config.GetServer(),
		grpc.WithTransportCredentials(creds),
		grpc.WithUnaryInterceptor(AuthInterceptor(c.config.GetUser(), &c.Token)))

	if err != nil {
		c.Logger.Error(err, fmt.Sprintf("connect to %s", c.config.GetServer()), componentName)
		return err
	}

	go func() {
		<-ctx.Done()

		errConn := conn.Close()
		if errConn != nil {
			c.Logger.Error(err, "close connection", componentName)
		}
	}()

	c.usersClient = pb.NewUsersClient(conn)
	c.itemsClient = pb.NewItemsClient(conn)

	return nil
}

// getCredentials is a helper function used to configure transport credentials for
// GRPC-connection to server.
func (c *GRPCClient) getCredentials() (credentials.TransportCredentials, error) {
	if c.config.GetTLSDisable() {
		return insecure.NewCredentials(), nil
	}

	certPool, err := x509.SystemCertPool()
	if err != nil {
		return nil, err
	}

	if CAcert := c.config.GetCACert(); CAcert != "" {
		b, err := os.ReadFile(CAcert)
		if err != nil {
			return nil, err
		}

		if !certPool.AppendCertsFromPEM(b) {
			return nil, fmt.Errorf("failed to append CA certificate: %s", CAcert)
		}
	}

	TLSConfig := &tls.Config{
		RootCAs:    certPool,
		MinVersion: tls.VersionTLS12,
	}

	return credentials.NewTLS(TLSConfig), nil
}

// UserLogin login user.
//
// If two-step authentication is enabled for user during, and no verication code were provide
// UserLogin returns ErrSecondFactorRequired.
func (c *GRPCClient) UserLogin(ctx context.Context, username, password, verificationCode string) error {
	req := &pb.UserLoginRequest{
		Username: username,
		Password: password,
		OtpCode:  verificationCode,
	}

	resp, err := c.usersClient.UserLogin(ctx, req)
	if err != nil {
		return err
	}

	if resp.SecondFactor {
		return ErrSecondFactorRequired
	}

	if c.encKey, err = crypt.DecryptAESwithAD([]byte(c.config.GetSecretKey()), resp.Ekey); err != nil {
		return ErrEKeyDecryptionFailed
	}

	c.Token = resp.Token
	c.MaxSecretSize = uint32(resp.ServerLimits.MaxSecretSize)

	return nil
}

// UserRegister registers new user.
//
// During registration process randon 32-byte encryption key is generated. This key is encrypted with
// user sectet key and then sent to server.
func (c *GRPCClient) UserRegister(ctx context.Context, user *NewUser) (*TOTPKey, error) {
	pwdhash, err := crypt.CalculatePasswordHash(user.Password)
	if err != nil {
		return nil, err
	}

	var email *string
	if user.Email != "" {
		email = common.PtrTo(user.Email)
	}

	eKey := crypt.GenerateRandomKey32()

	decryptedEKey, err := crypt.EncryptAESwithAD([]byte(user.SecretKey), eKey)
	if err != nil {
		return nil, ErrEKeyEncryptionFailed
	}

	req := &pb.CreateUserRequest{
		User: &pb.User{
			Username: user.Username,
			Pwdhash:  common.PtrTo(pwdhash),
			Ekey:     decryptedEKey,
			Email:    email,
		},
		Twofactor: user.TwoFactorEnable,
	}

	resp, err := c.usersClient.CreateUser(ctx, req)
	if err != nil {
		return nil, err
	}

	if user.TwoFactorEnable {
		if resp == nil || resp.Totpkey == nil {
			return nil, ErrMissedServerResponce
		}

		return &TOTPKey{
			SecretKey: resp.Totpkey.Secret,
			QRCode:    resp.Totpkey.Qrcode,
		}, nil
	}

	return &TOTPKey{}, nil
}

// GetItemsList returns list with short representation of items.
func (c *GRPCClient) GetItemsList(ctx context.Context) ([]*pb.ItemShort, error) {
	request := &pb.GetItemListRequest{
		Username: c.config.GetUser(),
	}

	resp, err := c.itemsClient.GetItemList(ctx, request)
	if err != nil {
		return nil, c.wrapError(err)
	}

	return resp.Items, nil
}

// GetItem returns all item's content.
func (c *GRPCClient) GetItem(ctx context.Context, itemName, itemType string) (*Item, error) {
	request := &pb.GetItemRequest{
		Username: c.config.GetUser(),
		ItemName: itemName,
		ItemType: itemType,
	}

	resp, err := c.itemsClient.GetItem(ctx, request)
	if err != nil {
		return nil, c.wrapError(err)
	}

	if err := c.DecryptPbItem(resp.Item); err != nil {
		return nil, err
	}

	return NewItemFromPB(resp.Item), nil
}

// SaveItem update existing or create new item.
//
// Which action to take - update or create, based on item ID - for new Item id is always 0,
// for existing >0.
func (c *GRPCClient) SaveItem(ctx context.Context, item *Item) error {
	pbItem := item.ToPB()
	if err := c.EncryptPbItem(pbItem); err != nil {
		return err
	}

	if len(pbItem.Secrets.Secret) > int(c.MaxSecretSize) {
		gotSize := float64(len(pbItem.Secrets.Secret)) / 1024 / 1024
		maxSize := float64(c.MaxSecretSize) / 1024 / 1024

		return fmt.Errorf("%w: uploaded size %.2f Mb, max supported size %.2f Mb",
			ErrSecretTooBig, gotSize, maxSize)
	}

	if item.ID > 0 {
		return c.updateItem(ctx, pbItem)
	}

	return c.createItem(ctx, pbItem)
}

// createItem creates new item.
func (c *GRPCClient) createItem(ctx context.Context, item *pb.Item) error {
	request := &pb.CreateItemRequest{
		Username: c.config.GetUser(),
		Item:     item,
	}

	_, err := c.itemsClient.CreateItem(ctx, request)
	if err != nil {
		return c.wrapError(err)
	}

	return nil
}

// updateItem updates existing item.
func (c *GRPCClient) updateItem(ctx context.Context, item *pb.Item) error {
	request := &pb.UpdateItemRequest{
		Username: c.config.GetUser(),
		Item:     item,
	}

	_, err := c.itemsClient.UpdateItem(ctx, request)
	if err != nil {
		return c.wrapError(err)
	}

	return nil
}

// updateItem updates existing item.
func (c *GRPCClient) DeleteItem(ctx context.Context, item *Item) error {
	request := &pb.DeleteItemRequest{
		Username: c.config.GetUser(),
		Id:       item.ID,
	}

	_, err := c.itemsClient.DeleteItem(ctx, request)
	if err != nil {
		return c.wrapError(err)
	}

	return nil
}

// wrapError wraps well-known returned errors:
//   - server PermissionDenied wraps to ErrSessionExpired, for prompt user to relogin.
func (c *GRPCClient) wrapError(err error) error {
	st, ok := status.FromError(err)
	if ok {
		if st.Code() == codes.PermissionDenied {
			return ErrSessionExpired
		}
	}

	return err
}

// EncryptPbItem encrypts item before send it to the server.
func (c *GRPCClient) EncryptPbItem(item *pb.Item) error {
	if len(item.Secrets.Secret) > 0 {
		encrypted, err := crypt.EncryptAES(c.encKey, item.Secrets.Secret)
		if err != nil {
			return err
		}

		item.Secrets.Secret = encrypted
	}

	if len(item.Secrets.Notes) > 0 {
		encrypted, err := crypt.EncryptAES(c.encKey, item.Secrets.Notes)
		if err != nil {
			return err
		}

		item.Secrets.Notes = encrypted
	}

	if len(item.Additions.Uris) > 0 {
		encrypted, err := crypt.EncryptAES(c.encKey, item.Additions.Uris)
		if err != nil {
			return err
		}

		item.Additions.Uris = encrypted
	}

	if len(item.Additions.CustomFields) > 0 {
		encrypted, err := crypt.EncryptAES(c.encKey, item.Additions.CustomFields)
		if err != nil {
			return err
		}

		item.Additions.CustomFields = encrypted
	}

	return nil
}

// DecryptPbItem decrypts received from server item.
func (c *GRPCClient) DecryptPbItem(item *pb.Item) error {
	if item == nil || item.Secrets == nil {
		return ErrMissedServerResponce
	}

	if len(item.Secrets.Secret) > 0 {
		decrypted, err := crypt.DecryptAES(c.encKey, item.Secrets.Secret)
		if err != nil {
			return err
		}

		item.Secrets.Secret = decrypted
	}

	if len(item.Secrets.Notes) > 0 {
		decrypted, err := crypt.DecryptAES(c.encKey, item.Secrets.Notes)
		if err != nil {
			return err
		}

		item.Secrets.Notes = decrypted
	}

	if len(item.Additions.Uris) > 0 {
		decrypted, err := crypt.DecryptAES(c.encKey, item.Additions.Uris)
		if err != nil {
			return err
		}

		item.Additions.Uris = decrypted
	}

	if len(item.Additions.CustomFields) > 0 {
		decrypted, err := crypt.DecryptAES(c.encKey, item.Additions.CustomFields)
		if err != nil {
			return err
		}

		item.Additions.CustomFields = decrypted
	}

	return nil
}

// TOTPKey reprensent TOTP key data.
type TOTPKey struct {
	SecretKey string
	QRCode    []byte
}
