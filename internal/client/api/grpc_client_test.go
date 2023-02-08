package api

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"testing"

	"github.com/artfuldog/gophkeeper/internal/client/config"
	"github.com/artfuldog/gophkeeper/internal/crypt"
	"github.com/artfuldog/gophkeeper/internal/mocks/mockgrpc"
	"github.com/artfuldog/gophkeeper/internal/mocks/mocklogger"
	"github.com/artfuldog/gophkeeper/internal/pb"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

// Helpers

var (
	mockAnyVal        = gomock.Any()
	testGRPCSecretKey = "secretkey"
	testGRPC_Ctx      = context.Background()
	testGRP_EncKey    = []byte("123456789a123456789b123456789abc")
)

type TestSuiteGRPClient struct {
	MockCtrl    *gomock.Controller
	Client      *GRPCClient
	UsersClient *mockgrpc.MockUsersClient
	ItemsClient *mockgrpc.MockItemsClient
}

func NewTestSuiteGRPClient(t gomock.TestReporter) *TestSuiteGRPClient {
	mockCtrl := gomock.NewController(t)

	testLogger := mocklogger.NewMockLogger()

	testConfig, _ := config.NewConfiger(nil)
	testConfig.SetSecretKey(testGRPCSecretKey)

	testUsersService := mockgrpc.NewMockUsersClient(mockCtrl)
	testItemsService := mockgrpc.NewMockItemsClient(mockCtrl)

	testClient := NewGRPCClient(testConfig, testLogger)
	//testClient.encKey = testGRP_EncKey
	testClient.itemsClient = testItemsService
	testClient.usersClient = testUsersService

	return &TestSuiteGRPClient{
		MockCtrl:    mockCtrl,
		Client:      testClient,
		UsersClient: testUsersService,
		ItemsClient: testItemsService,
	}
}

func (s *TestSuiteGRPClient) Stop() {
	s.MockCtrl.Finish()
}

// Tests

func TestGRPCClient_Connect(t *testing.T) {
	ts := NewTestSuiteGRPClient(t)
	defer ts.Stop()

	ts.Client.config.SetCACert("/wrong/file/path/to/ca")
	t.Run("Creds error", func(t *testing.T) {
		assert.Error(t, ts.Client.Connect(testGRPC_Ctx))
	})

	ts.Client.config.SetCACert("")
	ts.Client.config.SetServer("")
	t.Run("Connection error", func(t *testing.T) {
		assert.Error(t, ts.Client.Connect(testGRPC_Ctx))
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ts.Client.config.SetServer("127.0.0.1:3200")
	t.Run("Connection error", func(t *testing.T) {
		require.NoError(t, ts.Client.Connect(ctx))
	})
}

func TestGRPCClient_getCredentials(t *testing.T) {
	ts := NewTestSuiteGRPClient(t)
	defer ts.Stop()

	ts.Client.config.SetTLSDisable(true)
	t.Run("TLS disable", func(t *testing.T) {
		cred, err := ts.Client.getCredentials()
		require.NoError(t, err)
		assert.Equal(t, cred, insecure.NewCredentials())
	})
}

func TestGRPCClient_UserLogin(t *testing.T) {
	ts := NewTestSuiteGRPClient(t)
	defer ts.Stop()

	t.Run("Login Error", func(t *testing.T) {
		ts.UsersClient.EXPECT().UserLogin(testGRPC_Ctx, mockAnyVal).Return(nil, assert.AnError)
		assert.Error(t, ts.Client.UserLogin(testGRPC_Ctx, "", "", ""))
	})

	t.Run("Second factor", func(t *testing.T) {
		resp := &pb.UserLoginResponce{
			SecondFactor: true,
		}
		ts.UsersClient.EXPECT().UserLogin(testGRPC_Ctx, mockAnyVal).Return(resp, nil)
		assert.ErrorIs(t, ts.Client.UserLogin(testGRPC_Ctx, "", "", ""), ErrSecondFactorRequired)
	})

	t.Run("Encryption error", func(t *testing.T) {
		resp := &pb.UserLoginResponce{
			SecondFactor: false,
			Ekey:         []byte("asdasd"),
		}

		ts.UsersClient.EXPECT().UserLogin(testGRPC_Ctx, mockAnyVal).Return(resp, nil)
		assert.ErrorIs(t, ts.Client.UserLogin(testGRPC_Ctx, "", "", ""), ErrEKeyDecryptionFailed)
	})

	t.Run("User logged in", func(t *testing.T) {
		encKey, err := crypt.EncryptAESwithAD([]byte(testGRPCSecretKey), []byte("enckey"))
		require.NoError(t, err)
		resp := &pb.UserLoginResponce{
			SecondFactor: false,
			Ekey:         encKey,
			Token:        "123",
			ServerLimits: &pb.ServerLimits{
				MaxSecretSize: 1024,
			},
		}

		ts.UsersClient.EXPECT().UserLogin(testGRPC_Ctx, mockAnyVal).Return(resp, nil)
		assert.NoError(t, ts.Client.UserLogin(testGRPC_Ctx, "", "", ""))
	})
}

func TestGRPCClient_UserRegister(t *testing.T) {
	ts := NewTestSuiteGRPClient(t)
	defer ts.Stop()

	t.Run("Encryption failed", func(t *testing.T) {
		ts.UsersClient.EXPECT().CreateUser(testGRPC_Ctx, mockAnyVal).Return(nil, assert.AnError)
		user := &NewUser{}
		_, err := ts.Client.UserRegister(testGRPC_Ctx, user)
		assert.Error(t, err)
	})

	t.Run("Two-factor enabled with nil responce", func(t *testing.T) {
		ts.UsersClient.EXPECT().CreateUser(testGRPC_Ctx, mockAnyVal).Return(nil, nil)
		user := &NewUser{
			Email:           "someone@example.com",
			TwoFactorEnable: true,
		}
		_, err := ts.Client.UserRegister(testGRPC_Ctx, user)
		assert.ErrorIs(t, err, ErrMissedServerResponce)
	})

	t.Run("Two-factor enabled", func(t *testing.T) {
		resp := &pb.CreateUserResponce{
			Totpkey: &pb.TOTPKey{
				Secret: "MNOAWDw",
				Qrcode: []byte("qrcode"),
			},
		}
		ts.UsersClient.EXPECT().CreateUser(testGRPC_Ctx, mockAnyVal).Return(resp, nil)
		user := &NewUser{
			Email:           "someone@example.com",
			TwoFactorEnable: true,
		}
		totpkey, err := ts.Client.UserRegister(testGRPC_Ctx, user)
		require.NoError(t, err)
		assert.NotEmpty(t, totpkey)
	})

	t.Run("Two-factor disabled", func(t *testing.T) {
		resp := &pb.CreateUserResponce{}
		ts.UsersClient.EXPECT().CreateUser(testGRPC_Ctx, mockAnyVal).Return(resp, nil)
		user := &NewUser{
			Email:           "someone@example.com",
			TwoFactorEnable: false,
		}
		totpkey, err := ts.Client.UserRegister(testGRPC_Ctx, user)
		require.NoError(t, err)
		assert.Empty(t, totpkey)
	})
}

func TestGRPCClient_GetItemsList(t *testing.T) {
	ts := NewTestSuiteGRPClient(t)
	defer ts.Stop()

	t.Run("Server responce error", func(t *testing.T) {
		ts.ItemsClient.EXPECT().GetItemList(testGRPC_Ctx, mockAnyVal).Return(nil, assert.AnError)
		_, err := ts.Client.GetItemsList(testGRPC_Ctx)
		assert.Error(t, err)
	})

	t.Run("Server responce PK", func(t *testing.T) {
		resp := &pb.GetItemListResponce{
			Items: []*pb.ItemShort{{Name: "123"}},
		}
		ts.ItemsClient.EXPECT().GetItemList(testGRPC_Ctx, mockAnyVal).Return(resp, nil)
		items, err := ts.Client.GetItemsList(testGRPC_Ctx)
		require.NoError(t, err)
		assert.NotEmpty(t, items)
	})
}

func TestGRPCClient_GetItem(t *testing.T) {
	ts := NewTestSuiteGRPClient(t)
	defer ts.Stop()

	itemName := "itemname"
	itemType := "itemType"

	t.Run("Server responce error", func(t *testing.T) {
		ts.ItemsClient.EXPECT().GetItem(testGRPC_Ctx, mockAnyVal).Return(nil, assert.AnError)
		_, err := ts.Client.GetItem(testGRPC_Ctx, itemName, itemType)
		assert.Error(t, err)
	})

	t.Run("Decryption error", func(t *testing.T) {
		resp := &pb.GetItemResponce{}
		ts.ItemsClient.EXPECT().GetItem(testGRPC_Ctx, mockAnyVal).Return(resp, nil)
		_, err := ts.Client.GetItem(testGRPC_Ctx, itemName, itemType)
		assert.Error(t, err)
	})

	t.Run("Succesfully get item", func(t *testing.T) {
		ts.Client.encKey = testGRP_EncKey
		pbItem := TestingNewPbLoginItem()
		ts.Client.EncryptPbItem(pbItem)

		resp := &pb.GetItemResponce{
			Item: pbItem,
		}
		ts.ItemsClient.EXPECT().GetItem(testGRPC_Ctx, mockAnyVal).Return(resp, nil)
		item, err := ts.Client.GetItem(testGRPC_Ctx, itemName, itemType)
		fmt.Println(*item)
		require.NoError(t, err)
		assert.NotEmpty(t, item)
	})

}

func TestGRPCClient_SaveItem(t *testing.T) {
	ts := NewTestSuiteGRPClient(t)
	defer ts.Stop()
	ts.Client.MaxSecretSize = 1024 * 1024

	t.Run("Encryption error", func(t *testing.T) {
		item := TestingNewLoginItem()
		assert.Error(t, ts.Client.SaveItem(testGRPC_Ctx, item))
	})

	ts.Client.encKey = testGRP_EncKey

	t.Run("Secret too big", func(t *testing.T) {
		item := TestingNewSecDataItem()
		ts.Client.MaxSecretSize = 100
		assert.ErrorIs(t, ts.Client.SaveItem(testGRPC_Ctx, item), ErrSecretTooBig)
		ts.Client.MaxSecretSize = 50 * 1024 * 1024
	})

	t.Run("Create item error", func(t *testing.T) {
		item := TestingNewLoginItem()

		ts.ItemsClient.EXPECT().CreateItem(testGRPC_Ctx, mockAnyVal).Return(nil, assert.AnError)
		assert.Error(t, ts.Client.SaveItem(testGRPC_Ctx, item))
	})

	t.Run("Create item", func(t *testing.T) {
		item := TestingNewLoginItem()
		resp := &pb.CreateItemResponce{}

		ts.ItemsClient.EXPECT().CreateItem(testGRPC_Ctx, mockAnyVal).Return(resp, nil)
		require.NoError(t, ts.Client.SaveItem(testGRPC_Ctx, item))
	})

	t.Run("Update item error", func(t *testing.T) {
		item := TestingNewLoginItem()
		item.Id = 100

		ts.ItemsClient.EXPECT().UpdateItem(testGRPC_Ctx, mockAnyVal).Return(nil, assert.AnError)
		assert.Error(t, ts.Client.SaveItem(testGRPC_Ctx, item))
	})

	t.Run("Update item", func(t *testing.T) {
		item := TestingNewLoginItem()
		item.Id = 100
		resp := &pb.UpdateItemResponce{}

		ts.ItemsClient.EXPECT().UpdateItem(testGRPC_Ctx, mockAnyVal).Return(resp, nil)
		require.NoError(t, ts.Client.SaveItem(testGRPC_Ctx, item))
	})
}

func TestGRPCClient_DeleteItem(t *testing.T) {
	ts := NewTestSuiteGRPClient(t)
	defer ts.Stop()
	ts.Client.config.SetUser("username")

	t.Run("Delete item error", func(t *testing.T) {
		item := TestingNewLoginItem()

		ts.ItemsClient.EXPECT().DeleteItem(testGRPC_Ctx, mockAnyVal).Return(nil, assert.AnError)
		assert.Error(t, ts.Client.DeleteItem(testGRPC_Ctx, item))
	})

	t.Run("Delete item", func(t *testing.T) {
		item := TestingNewLoginItem()
		resp := &pb.DeleteItemResponce{}

		ts.ItemsClient.EXPECT().DeleteItem(testGRPC_Ctx, mockAnyVal).Return(resp, nil)
		require.NoError(t, ts.Client.DeleteItem(testGRPC_Ctx, item))
	})
}

func TestGRPCClient_wrapError(t *testing.T) {
	ts := NewTestSuiteGRPClient(t)
	defer ts.Stop()

	t.Run("Not status error", func(t *testing.T) {
		assert.ErrorIs(t, ts.Client.wrapError(assert.AnError), assert.AnError)
	})
	t.Run("Status error", func(t *testing.T) {
		err := status.Error(codes.NotFound, "")
		assert.ErrorIs(t, ts.Client.wrapError(err), err)
	})
	t.Run("Permission denied", func(t *testing.T) {
		err := status.Error(codes.PermissionDenied, "")
		assert.ErrorIs(t, ts.Client.wrapError(err), ErrSessionExpired)
	})
}

func TestGRPCClient_encryptPbItem(t *testing.T) {
	ts := NewTestSuiteGRPClient(t)
	defer ts.Stop()

	t.Run("Secret encryption error", func(t *testing.T) {
		pbItem := TestingNewPbLoginItem()
		assert.Error(t, ts.Client.EncryptPbItem(pbItem))
	})

	t.Run("Notes encryption error", func(t *testing.T) {
		pbItem := TestingNewPbSecNoteItem()
		assert.Error(t, ts.Client.EncryptPbItem(pbItem))
	})

	t.Run("URIs encryption error", func(t *testing.T) {
		pbItem := TestingNewPbLoginItem()
		pbItem.Secrets.Secret = nil
		pbItem.Secrets.Notes = nil
		assert.Error(t, ts.Client.EncryptPbItem(pbItem))
	})

	t.Run("Custom field encryption error", func(t *testing.T) {
		pbItem := TestingNewPbLoginItem()
		pbItem.Secrets.Secret = nil
		pbItem.Secrets.Notes = nil
		pbItem.Additions.Uris = nil
		assert.Error(t, ts.Client.EncryptPbItem(pbItem))
	})
}

func TestGRPCClient_decryptPbItem(t *testing.T) {
	ts := NewTestSuiteGRPClient(t)
	defer ts.Stop()

	t.Run("Secret decryption error", func(t *testing.T) {
		pbItem := TestingNewPbLoginItem()
		assert.Error(t, ts.Client.DecryptPbItem(pbItem))
	})

	t.Run("Notes decryption error", func(t *testing.T) {
		pbItem := TestingNewPbSecNoteItem()
		assert.Error(t, ts.Client.DecryptPbItem(pbItem))
	})

	t.Run("URIs decryption error", func(t *testing.T) {
		pbItem := TestingNewPbLoginItem()
		pbItem.Secrets.Secret = nil
		pbItem.Secrets.Notes = nil
		assert.Error(t, ts.Client.DecryptPbItem(pbItem))
	})

	t.Run("Custom field decryption error", func(t *testing.T) {
		pbItem := TestingNewPbLoginItem()
		pbItem.Secrets.Secret = nil
		pbItem.Secrets.Notes = nil
		pbItem.Additions.Uris = nil
		assert.Error(t, ts.Client.DecryptPbItem(pbItem))
	})
}

func BenchmarkGRPCClient_EncryptDecrypt(b *testing.B) {
	ts := NewTestSuiteGRPClient(b)
	defer ts.Stop()
	ts.Client.encKey = testGRP_EncKey

	var wg sync.WaitGroup
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		wg.Add(1)
		go func() {
			item := TestingNewLoginItem()
			pbItem := item.ToPB()

			b.StartTimer()
			err := ts.Client.EncryptPbItem(pbItem)
			assert.NoError(b, err)
			err = ts.Client.DecryptPbItem(pbItem)
			assert.NoError(b, err)
			b.StopTimer()

			gotItem := NewItemFromPB(pbItem)
			if !reflect.DeepEqual(gotItem, item) {
				b.Errorf("GRPCClient.EncryptDecrypt() = %v, want %v", gotItem, item)
			}
			wg.Done()
		}()
	}

	wg.Wait()
}
