//nolint:forbidigo
package it

import (
	"context"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/artfuldog/gophkeeper/internal/client/api"
	"github.com/artfuldog/gophkeeper/internal/client/config"
	"github.com/artfuldog/gophkeeper/internal/common"
	"github.com/artfuldog/gophkeeper/internal/mocks/mocklogger"
	"github.com/artfuldog/gophkeeper/internal/server"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type E2ETestSuit struct {
	suite.Suite
	SomeNumber   int
	Server       *server.Server
	ServerStop   context.CancelFunc
	ClientConfig *config.Configer
	Client       api.Client
	wg           *sync.WaitGroup
}

func (s *E2ETestSuit) SetupSuite() {
	fmt.Println("=== > Setup server")

	os.Setenv("GK_DB_DSN", "127.0.0.1:5432/gophkeeper_db_inttests")
	os.Setenv("GK_DB_USER", "gksa")
	os.Setenv("GK_LOG_LEVEL", "error")
	os.Setenv("GK_SERVER_KEY", "123456789f123456789q123456789pQ1")
	os.Setenv("GK_TLS_CERT", "certs/service.pem")
	os.Setenv("GK_TLS_KEY", "certs/service.key")

	cfg, err := server.NewConfig()
	require.NoError(s.T(), err)

	s.Server, err = server.NewServer(cfg)
	require.NoError(s.T(), err)

	ctx, cancel := context.WithCancel(context.Background())
	srvStatusCh := make(chan error)
	s.wg = new(sync.WaitGroup)

	go s.Server.Run(ctx, srvStatusCh)

	s.wg.Add(1)
	go s.ControlServer(ctx, srvStatusCh, s.wg)

	s.ServerStop = cancel

	fmt.Println("=== > Setup client")

	s.ClientConfig, err = config.NewConfiger(&config.Flags{CustomConfigPath: `client_config.yaml`})
	require.NoError(s.T(), err)
	logger := mocklogger.NewMockLogger()

	s.Client = api.NewGRPCClient(s.ClientConfig, logger)

	clientStopCh := make(chan struct{})
	clientSyncCh := make(chan string)
	err = s.Client.Connect(ctx, clientStopCh)
	require.NoError(s.T(), err)

	err = s.ClientConfig.CreateAppDir()
	require.NoError(s.T(), err)
	err = s.Client.StorageInit(ctx, clientSyncCh)
	require.NoError(s.T(), err)

	s.wg.Add(1)
	go s.ControlClient(ctx, clientSyncCh, s.wg)

	time.Sleep(3 * time.Second) // give server some time to start

	fmt.Println("=== > Setup all successfully")
}

func (s *E2ETestSuit) ControlServer(ctx context.Context, ctrlCh <-chan error, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case err := <-ctrlCh:
			if err != nil {
				require.NoError(s.T(), err)
			}
		}
	}
}

func (s *E2ETestSuit) ControlClient(ctx context.Context, statusCh <-chan string, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case <-statusCh:
			continue
		}
	}
}

func (s *E2ETestSuit) TearDownSuite() {
	fmt.Println("=== > Clear client's artefacts")

	os.RemoveAll(s.ClientConfig.GetAppConfigDir())

	fmt.Println("=== > Clear server's artefacts")

	dbClearCtx, dbClearCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer dbClearCancel()

	s.Server.DB.Clear(dbClearCtx)
	s.ServerStop()

	s.wg.Wait()

	fmt.Println("=== > Clear all artefacts successfully")
}

func TestE2ETestSuit(t *testing.T) {
	suite.Run(t, new(E2ETestSuit))
}

func (s *E2ETestSuit) Test01UserReg() {
	user := &api.NewUser{
		Username:        usr01Username,
		Password:        usr01Password,
		PasswordConfirm: usr01Password,
		SecretKey:       usr01SecretKey,
		Email:           "someone@example.com",
		TwoFactorEnable: false,
	}

	_, err := s.Client.UserRegister(context.Background(), user)
	require.NoError(s.T(), err)

	_, err = s.Client.UserRegister(context.Background(), user)
	assertErrWithStatusCode(s.T(), err, codes.InvalidArgument)
}

func (s *E2ETestSuit) Test02UserLogin() {
	err := s.Client.UserLogin(context.Background(), usr01Username, "wrongpassword", "")
	assertErrWithStatusCode(s.T(), err, codes.PermissionDenied)

	err = s.Client.UserLogin(context.Background(), usr01Username, usr01Password, "")
	require.NoError(s.T(), err)
}

func (s *E2ETestSuit) Test03CreateItems() {
	err := s.Client.SaveItem(context.Background(), TestE2ENewLoginItem())
	require.NoError(s.T(), err)

	err = s.Client.SaveItem(context.Background(), TestE2ENewCardItem())
	require.NoError(s.T(), err)

	err = s.Client.SaveItem(context.Background(), TestE2ENewSecNoteItem())
	require.NoError(s.T(), err)

	err = s.Client.SaveItem(context.Background(), TestE2ENewSecDataItem())
	require.NoError(s.T(), err)

	err = s.Client.SaveItem(context.Background(), TestE2ENewSecNoteItem())
	assertErrWithStatusCode(s.T(), err, codes.InvalidArgument)
}

func (s *E2ETestSuit) Test04GetItems() {
	itemsShort, err := s.Client.GetItemsList(context.Background())
	require.NoError(s.T(), err)
	assert.Equal(s.T(), 4, len(itemsShort))

	for _, itemShort := range itemsShort {
		item, err := s.Client.GetItem(context.TODO(), itemShort.Name, itemShort.Type)
		require.NoError(s.T(), err)
		assert.NotEmpty(s.T(), item.ID)
		assert.NotEmpty(s.T(), item.Hash)
		assert.NotEmpty(s.T(), item.Updated)

		if itemShort.Type != common.ItemTypeSecNote {
			assert.NotEmpty(s.T(), item.Secret)
		}

		assert.NotEmpty(s.T(), item.Notes)
		assert.NotEmpty(s.T(), item.CustomFields)
	}

	_, err = s.Client.GetItem(context.TODO(), "unexistedItem", common.ItemTypeLogin)
	assertErrWithStatusCode(s.T(), err, codes.NotFound)
}

func (s *E2ETestSuit) Test05UpdateItems() {
	originItem := TestE2ENewLoginItem()
	gotItem, err := s.Client.GetItem(context.Background(), originItem.Name, originItem.Type)
	require.NoError(s.T(), err)
	originItem.ID = gotItem.ID

	updatedLoginItem := TestE2ENewLoginItem()
	updatedLoginItem.ID = originItem.ID
	updatedLoginItem.Name = "NewLoginName"
	updatedLoginItem.Secret = api.NewSecret([]byte("newSecret"), updatedLoginItem.Type)
	updatedLoginItem.Notes = "new item notes"
	updatedLoginItem.CustomFields = api.NewCustomFields([]byte("new custom fields"))
	updatedLoginItem.URIs = api.NewURIs([]byte("new uris"))
	updatedLoginItem.Reprompt = true

	err = s.Client.SaveItem(context.Background(), updatedLoginItem)
	require.NoError(s.T(), err)

	gotItem, err = s.Client.GetItem(context.Background(), updatedLoginItem.Name, updatedLoginItem.Type)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), updatedLoginItem.Name, gotItem.Name)
	assert.Equal(s.T(), updatedLoginItem.Secret, gotItem.Secret)
	assert.Equal(s.T(), updatedLoginItem.Notes, gotItem.Notes)
	assert.Equal(s.T(), updatedLoginItem.CustomFields, gotItem.CustomFields)
	assert.Equal(s.T(), updatedLoginItem.URIs, gotItem.URIs)
	assert.Equal(s.T(), updatedLoginItem.Reprompt, gotItem.Reprompt)
	assert.NotEqual(s.T(), updatedLoginItem.Hash, gotItem.Hash)

	err = s.Client.SaveItem(context.Background(), originItem)
	require.NoError(s.T(), err)

	gotItem, err = s.Client.GetItem(context.Background(), originItem.Name, originItem.Type)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), originItem.Name, gotItem.Name)
	assert.Equal(s.T(), originItem.Secret, gotItem.Secret)
	assert.Equal(s.T(), originItem.Notes, gotItem.Notes)
	assert.Equal(s.T(), originItem.CustomFields, gotItem.CustomFields)
	assert.Equal(s.T(), originItem.URIs, gotItem.URIs)
	assert.Equal(s.T(), originItem.Reprompt, gotItem.Reprompt)
	assert.NotEqual(s.T(), originItem.Hash, gotItem.Hash)
}

func (s *E2ETestSuit) Test06DeleteItems() {
	LoginItem := TestE2ENewLoginItem()
	gotItem, err := s.Client.GetItem(context.Background(), LoginItem.Name, LoginItem.Type)
	require.NoError(s.T(), err)

	err = s.Client.DeleteItem(context.Background(), gotItem)
	require.NoError(s.T(), err)

	itemsShort, err := s.Client.GetItemsList(context.Background())
	require.NoError(s.T(), err)
	assert.Equal(s.T(), 3, len(itemsShort))

	unexistedItem := TestE2ENewLoginItem()
	unexistedItem.ID = 99999999

	err = s.Client.DeleteItem(context.Background(), unexistedItem)
	assertErrWithStatusCode(s.T(), err, codes.NotFound)
}

func assertErrWithStatusCode(t *testing.T, err error, expectedCode codes.Code) {
	t.Helper()

	if err != nil {
		assert.Error(t, err)
		return
	}

	st, ok := status.FromError(err)
	if !ok {
		t.Errorf("not found status code, want - %s, got error %v", expectedCode, err)
	}
	assert.Equal(t, expectedCode, st.Code())
}
