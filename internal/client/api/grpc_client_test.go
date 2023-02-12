//nolint:gochecknoglobals
package api

import (
	"context"
	"reflect"
	"sync"
	"testing"

	"github.com/artfuldog/gophkeeper/internal/client/config"
	"github.com/artfuldog/gophkeeper/internal/client/storage"
	"github.com/artfuldog/gophkeeper/internal/crypt"
	"github.com/artfuldog/gophkeeper/internal/mocks/mockgrpc"
	"github.com/artfuldog/gophkeeper/internal/mocks/mocklogger"
	"github.com/artfuldog/gophkeeper/internal/mocks/mockstorage"
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
	testGRPCctx       = context.Background()
	testGRPCencKey    = []byte("123456789a123456789b123456789abc")
)

type TestSuiteGRPClient struct {
	MockCtrl    *gomock.Controller
	Client      *GRPCClient
	UsersClient *mockgrpc.MockUsersClient
	ItemsClient *mockgrpc.MockItemsClient
	Storage     *mockstorage.MockS
}

func NewTestSuiteGRPClient(t gomock.TestReporter) *TestSuiteGRPClient {
	mockCtrl := gomock.NewController(t)

	testLogger := mocklogger.NewMockLogger()

	testConfig, _ := config.NewConfiger(nil)
	testConfig.SetSecretKey(testGRPCSecretKey)

	testUsersService := mockgrpc.NewMockUsersClient(mockCtrl)
	testItemsService := mockgrpc.NewMockItemsClient(mockCtrl)
	testStorage := mockstorage.NewMockS(mockCtrl)

	testClient := NewGRPCClient(testConfig, testLogger)
	testClient.itemsClient = testItemsService
	testClient.usersClient = testUsersService
	testClient.storage = testStorage

	return &TestSuiteGRPClient{
		MockCtrl:    mockCtrl,
		Client:      testClient,
		UsersClient: testUsersService,
		ItemsClient: testItemsService,
		Storage:     testStorage,
	}
}

func (s *TestSuiteGRPClient) Stop() {
	s.MockCtrl.Finish()
}

// Tests

func TestGRPCClient_Connect(t *testing.T) {
	ts := NewTestSuiteGRPClient(t)
	defer ts.Stop()
	ctrlCh := make(chan struct{})

	ts.Client.config.SetCACert("/wrong/file/path/to/ca")
	t.Run("Creds error", func(t *testing.T) {
		assert.Error(t, ts.Client.Connect(testGRPCctx, ctrlCh))
	})

	ts.Client.config.SetCACert("")
	ts.Client.config.SetServer("")
	t.Run("Connection error", func(t *testing.T) {
		assert.Error(t, ts.Client.Connect(testGRPCctx, ctrlCh))
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ts.Client.config.SetServer("127.0.0.1:3200")
	t.Run("Connection error", func(t *testing.T) {
		require.NoError(t, ts.Client.Connect(ctx, ctrlCh))
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
		ts.UsersClient.EXPECT().UserLogin(testGRPCctx, mockAnyVal).Return(nil, assert.AnError)
		assert.Error(t, ts.Client.UserLogin(testGRPCctx, "", "", ""))
	})

	t.Run("Second factor", func(t *testing.T) {
		resp := &pb.UserLoginResponse{
			SecondFactor: true,
		}
		ts.UsersClient.EXPECT().UserLogin(testGRPCctx, mockAnyVal).Return(resp, nil)
		assert.ErrorIs(t, ts.Client.UserLogin(testGRPCctx, "", "", ""), ErrSecondFactorRequired)
	})

	t.Run("Encryption error", func(t *testing.T) {
		resp := &pb.UserLoginResponse{
			SecondFactor: false,
			Ekey:         []byte("asdasd"),
		}

		ts.UsersClient.EXPECT().UserLogin(testGRPCctx, mockAnyVal).Return(resp, nil)
		assert.ErrorIs(t, ts.Client.UserLogin(testGRPCctx, "", "", ""), ErrEKeyDecryptionFailed)
	})

	t.Run("User logged in", func(t *testing.T) {
		encKey, err := crypt.EncryptAESwithAD([]byte(testGRPCSecretKey), []byte("enckey"))
		require.NoError(t, err)
		resp := &pb.UserLoginResponse{
			SecondFactor: false,
			Ekey:         encKey,
			Token:        "123",
			ServerLimits: &pb.ServerLimits{
				MaxSecretSize: 1024,
			},
		}

		ts.UsersClient.EXPECT().UserLogin(testGRPCctx, mockAnyVal).Return(resp, nil)
		assert.NoError(t, ts.Client.UserLogin(testGRPCctx, "", "", ""))
	})
}

func TestGRPCClient_UserRegister(t *testing.T) {
	ts := NewTestSuiteGRPClient(t)
	defer ts.Stop()

	t.Run("Encryption failed", func(t *testing.T) {
		ts.UsersClient.EXPECT().CreateUser(testGRPCctx, mockAnyVal).Return(nil, assert.AnError)
		user := &NewUser{}
		_, err := ts.Client.UserRegister(testGRPCctx, user)
		assert.Error(t, err)
	})

	t.Run("Two-factor enabled with nil response", func(t *testing.T) {
		ts.UsersClient.EXPECT().CreateUser(testGRPCctx, mockAnyVal).Return(nil, nil)
		user := &NewUser{
			Email:           "someone@example.com",
			TwoFactorEnable: true,
		}
		_, err := ts.Client.UserRegister(testGRPCctx, user)
		assert.ErrorIs(t, err, ErrMissedServerResponse)
	})

	t.Run("Two-factor enabled", func(t *testing.T) {
		resp := &pb.CreateUserResponse{
			Totpkey: &pb.TOTPKey{
				Secret: "MNOAWDw",
				Qrcode: []byte("qrcode"),
			},
		}
		ts.UsersClient.EXPECT().CreateUser(testGRPCctx, mockAnyVal).Return(resp, nil)
		user := &NewUser{
			Email:           "someone@example.com",
			TwoFactorEnable: true,
		}
		totpkey, err := ts.Client.UserRegister(testGRPCctx, user)
		require.NoError(t, err)
		assert.NotEmpty(t, totpkey)
	})

	t.Run("Two-factor disabled", func(t *testing.T) {
		resp := &pb.CreateUserResponse{}
		ts.UsersClient.EXPECT().CreateUser(testGRPCctx, mockAnyVal).Return(resp, nil)
		user := &NewUser{
			Email:           "someone@example.com",
			TwoFactorEnable: false,
		}
		totpkey, err := ts.Client.UserRegister(testGRPCctx, user)
		require.NoError(t, err)
		assert.Empty(t, totpkey)
	})
}

func TestGRPCClient_GetItemsList(t *testing.T) {
	ts := NewTestSuiteGRPClient(t)
	defer ts.Stop()

	t.Run("Server response error", func(t *testing.T) {
		ts.ItemsClient.EXPECT().GetItemList(testGRPCctx, mockAnyVal).Return(nil, assert.AnError)
		_, err := ts.Client.GetItemsList(testGRPCctx)
		assert.Error(t, err)
	})

	t.Run("Server response OK", func(t *testing.T) {
		resp := &pb.GetItemListResponse{
			Items: []*pb.ItemShort{{Name: "123"}},
		}
		ts.ItemsClient.EXPECT().GetItemList(testGRPCctx, mockAnyVal).Return(resp, nil)
		items, err := ts.Client.GetItemsList(testGRPCctx)
		require.NoError(t, err)
		assert.NotEmpty(t, items)
	})

	ts.Client.config.SetMode(config.ModeLocal)

	t.Run("Local storage error", func(t *testing.T) {
		ts.Storage.EXPECT().GetItemsList(testGRPCctx).Return(nil, assert.AnError)
		_, err := ts.Client.GetItemsList(testGRPCctx)
		assert.Error(t, err)
	})
	t.Run("Local storage OK", func(t *testing.T) {
		resp := storage.Items{{Name: "123"}}
		ts.Storage.EXPECT().GetItemsList(testGRPCctx).Return(resp, nil)
		items, err := ts.Client.GetItemsList(testGRPCctx)
		require.NoError(t, err)
		assert.NotEmpty(t, items)
	})
}

func TestGRPCClient_GetItem(t *testing.T) {
	ts := NewTestSuiteGRPClient(t)
	defer ts.Stop()

	itemName := "itemname"
	itemType := "itemType"

	t.Run("Server response error", func(t *testing.T) {
		ts.ItemsClient.EXPECT().GetItem(testGRPCctx, mockAnyVal).Return(nil, assert.AnError)
		_, err := ts.Client.GetItem(testGRPCctx, itemName, itemType)
		assert.Error(t, err)
	})

	t.Run("Decryption error", func(t *testing.T) {
		resp := &pb.GetItemResponse{}
		ts.ItemsClient.EXPECT().GetItem(testGRPCctx, mockAnyVal).Return(resp, nil)
		_, err := ts.Client.GetItem(testGRPCctx, itemName, itemType)
		assert.Error(t, err)
	})

	t.Run("Successfully get item from server", func(t *testing.T) {
		ts.Client.encKey = testGRPCencKey
		pbItem := TestingNewPbLoginItem()
		ts.Client.EncryptPbItem(pbItem)

		resp := &pb.GetItemResponse{
			Item: pbItem,
		}
		ts.ItemsClient.EXPECT().GetItem(testGRPCctx, mockAnyVal).Return(resp, nil)
		item, err := ts.Client.GetItem(testGRPCctx, itemName, itemType)
		require.NoError(t, err)
		assert.NotEmpty(t, item)
	})

	ts.Client.config.SetMode(config.ModeLocal)

	t.Run("Local storage error", func(t *testing.T) {
		ts.Storage.EXPECT().GetItem(testGRPCctx, mockAnyVal, mockAnyVal).Return(nil, assert.AnError)
		_, err := ts.Client.GetItem(testGRPCctx, itemName, itemType)
		assert.Error(t, err)
	})

	t.Run("Serialization error", func(t *testing.T) {
		storItem := []byte("random bytes")
		ts.Storage.EXPECT().GetItem(testGRPCctx, mockAnyVal, mockAnyVal).Return(storItem, nil)
		_, err := ts.Client.GetItem(testGRPCctx, itemName, itemType)
		assert.Error(t, err)
	})

	t.Run("Successfully get item from local storage", func(t *testing.T) {
		pbItem := TestingNewPbLoginItem()
		ts.Client.EncryptPbItem(pbItem)
		storItem := toBytesUnsafe(pbItem)

		ts.Storage.EXPECT().GetItem(testGRPCctx, mockAnyVal, mockAnyVal).Return(storItem, nil)
		item, err := ts.Client.GetItem(testGRPCctx, itemName, itemType)
		require.NoError(t, err)
		assert.NotEmpty(t, item)
	})
}

func TestGRPCClient_GetItemsForStorage(t *testing.T) {
	ts := NewTestSuiteGRPClient(t)
	defer ts.Stop()

	t.Run("Server response error", func(t *testing.T) {
		ts.ItemsClient.EXPECT().GetItems(testGRPCctx, mockAnyVal).Return(nil, assert.AnError)
		_, err := ts.Client.GetItemsForStorage(testGRPCctx, []int64{1, 2, 3})
		assert.Error(t, err)
	})

	t.Run("Successfully get item", func(t *testing.T) {
		ts.Client.encKey = testGRPCencKey
		pbItem := TestingNewPbLoginItem()
		ts.Client.EncryptPbItem(pbItem)

		resp := &pb.GetItemsResponse{
			Items: []*pb.Item{pbItem},
		}
		ts.ItemsClient.EXPECT().GetItems(testGRPCctx, mockAnyVal).Return(resp, nil)
		items, err := ts.Client.GetItemsForStorage(testGRPCctx, []int64{1, 2, 3})
		require.NoError(t, err)
		assert.Equal(t, 1, len(items))
	})
}

func TestGRPCClient_SaveItem(t *testing.T) {
	ts := NewTestSuiteGRPClient(t)
	defer ts.Stop()
	ts.Client.MaxSecretSize = 1024 * 1024

	t.Run("Encryption error", func(t *testing.T) {
		item := TestingNewLoginItem()
		assert.Error(t, ts.Client.SaveItem(testGRPCctx, item))
	})

	ts.Client.encKey = testGRPCencKey

	t.Run("Secret too big", func(t *testing.T) {
		item := TestingNewSecDataItem()
		ts.Client.MaxSecretSize = 100
		assert.ErrorIs(t, ts.Client.SaveItem(testGRPCctx, item), ErrSecretTooBig)
		ts.Client.MaxSecretSize = 50 * 1024 * 1024
	})

	t.Run("Create item error", func(t *testing.T) {
		item := TestingNewLoginItem()

		ts.ItemsClient.EXPECT().CreateItem(testGRPCctx, mockAnyVal).Return(nil, assert.AnError)
		assert.Error(t, ts.Client.SaveItem(testGRPCctx, item))
	})

	t.Run("Create item", func(t *testing.T) {
		item := TestingNewLoginItem()
		resp := &pb.CreateItemResponse{}

		ts.ItemsClient.EXPECT().CreateItem(testGRPCctx, mockAnyVal).Return(resp, nil)
		require.NoError(t, ts.Client.SaveItem(testGRPCctx, item))
	})

	t.Run("Update item error", func(t *testing.T) {
		item := TestingNewLoginItem()
		item.ID = 100

		ts.ItemsClient.EXPECT().UpdateItem(testGRPCctx, mockAnyVal).Return(nil, assert.AnError)
		assert.Error(t, ts.Client.SaveItem(testGRPCctx, item))
	})

	t.Run("Update item", func(t *testing.T) {
		item := TestingNewLoginItem()
		item.ID = 100
		resp := &pb.UpdateItemResponse{}

		ts.ItemsClient.EXPECT().UpdateItem(testGRPCctx, mockAnyVal).Return(resp, nil)
		require.NoError(t, ts.Client.SaveItem(testGRPCctx, item))
	})

	ts.Client.config.SetMode(config.ModeLocal)

	t.Run("Update item - failed to get revision", func(t *testing.T) {
		item := TestingNewLoginItem()
		item.ID = 100

		ts.UsersClient.EXPECT().GetRevision(testGRPCctx, mockAnyVal).Return(nil, assert.AnError)
		assert.Error(t, ts.Client.SaveItem(testGRPCctx, item))
	})

	t.Run("Update item - get item error", func(t *testing.T) {
		item := TestingNewLoginItem()
		item.ID = 100

		revResp := &pb.GetRevisionResponse{Revision: []byte("revision1")}

		ts.UsersClient.EXPECT().GetRevision(testGRPCctx, mockAnyVal).Return(revResp, nil)
		ts.Storage.EXPECT().GetRevision(testGRPCctx).Return([]byte("revision2"), nil)
		ts.ItemsClient.EXPECT().GetItemHash(mockAnyVal, mockAnyVal).Return(nil, assert.AnError)
		assert.ErrorIs(t, ErrOutOfSync, ts.Client.SaveItem(testGRPCctx, item))
	})

	t.Run("Update item - remote changes", func(t *testing.T) {
		item := TestingNewLoginItem()
		item.ID = 100

		revResp := &pb.GetRevisionResponse{Revision: []byte("revision1")}
		hashResp := &pb.GetItemHashResponse{Hash: []byte("wrhash")}

		ts.UsersClient.EXPECT().GetRevision(testGRPCctx, mockAnyVal).Return(revResp, nil)
		ts.Storage.EXPECT().GetRevision(testGRPCctx).Return([]byte("revision2"), nil)
		ts.ItemsClient.EXPECT().GetItemHash(mockAnyVal, mockAnyVal).Return(hashResp, nil)
		assert.ErrorIs(t, ErrOutOfSync, ts.Client.SaveItem(testGRPCctx, item))
	})

	t.Run("Update item with local storage", func(t *testing.T) {
		item := TestingNewLoginItem()
		item.ID = 100

		revResp := &pb.GetRevisionResponse{Revision: []byte("revision1")}
		ts.UsersClient.EXPECT().GetRevision(testGRPCctx, mockAnyVal).Return(revResp, nil)
		ts.Storage.EXPECT().GetRevision(testGRPCctx).Return([]byte("revision1"), nil)

		resp := &pb.UpdateItemResponse{}
		ts.ItemsClient.EXPECT().UpdateItem(testGRPCctx, mockAnyVal).Return(resp, nil)

		go func() {
			syncMode := <-ts.Client.syncControlCh
			assert.Equal(t, SyncWaitForComplete, syncMode)

			ts.Client.syncCompleteCh <- struct{}{}
		}()

		require.NoError(t, ts.Client.SaveItem(testGRPCctx, item))
	})
}

func TestGRPCClient_DeleteItem(t *testing.T) {
	ts := NewTestSuiteGRPClient(t)
	defer ts.Stop()
	ts.Client.config.SetUser("username")

	t.Run("Delete item error", func(t *testing.T) {
		item := TestingNewLoginItem()

		ts.ItemsClient.EXPECT().DeleteItem(testGRPCctx, mockAnyVal).Return(nil, assert.AnError)
		assert.Error(t, ts.Client.DeleteItem(testGRPCctx, item))
	})

	t.Run("Delete item", func(t *testing.T) {
		item := TestingNewLoginItem()
		resp := &pb.DeleteItemResponse{}

		ts.ItemsClient.EXPECT().DeleteItem(testGRPCctx, mockAnyVal).Return(resp, nil)
		require.NoError(t, ts.Client.DeleteItem(testGRPCctx, item))
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

func TestGRPCClient_encrypt_decryptPbItem(t *testing.T) {
	ts := NewTestSuiteGRPClient(t)
	defer ts.Stop()

	t.Run("Secret encryption/decryption error", func(t *testing.T) {
		pbItem := TestingNewPbLoginItem()
		err := ts.Client.EncryptPbItem(pbItem)
		assert.Error(t, err)
		err = ts.Client.DecryptPbItem(pbItem)
		assert.Error(t, err)

		wantItem := TestingNewPbLoginItem()
		pbItem.Updated = wantItem.Updated
		if !reflect.DeepEqual(pbItem, wantItem) {
			t.Errorf("Encrypt/decrypt err: got %v, want %v", pbItem, wantItem)
		}
	})

	t.Run("Notes encryption/decryption error", func(t *testing.T) {
		pbItem := TestingNewPbSecNoteItem()
		err := ts.Client.EncryptPbItem(pbItem)
		assert.Error(t, err)
		err = ts.Client.DecryptPbItem(pbItem)
		assert.Error(t, err)

		wantItem := TestingNewPbSecNoteItem()
		pbItem.Updated = wantItem.Updated
		if !reflect.DeepEqual(pbItem, wantItem) {
			t.Errorf("Encrypt/decrypt err: got %v, want %v", pbItem, wantItem)
		}
	})

	t.Run("URIs encryption/decryption error", func(t *testing.T) {
		pbItem := TestingNewPbLoginItem()
		pbItem.Secrets.Secret = nil
		pbItem.Secrets.Notes = nil

		err := ts.Client.EncryptPbItem(pbItem)
		assert.Error(t, err)
		err = ts.Client.DecryptPbItem(pbItem)
		assert.Error(t, err)

		wantItem := TestingNewPbLoginItem()
		wantItem.Secrets.Secret = nil
		wantItem.Secrets.Notes = nil
		pbItem.Updated = wantItem.Updated
		if !reflect.DeepEqual(pbItem, wantItem) {
			t.Errorf("Encrypt/decrypt err: got %v, want %v", pbItem, wantItem)
		}
	})

	t.Run("Custom field encryption error", func(t *testing.T) {
		pbItem := TestingNewPbLoginItem()
		pbItem.Secrets.Secret = nil
		pbItem.Secrets.Notes = nil
		pbItem.Additions.Uris = nil

		err := ts.Client.EncryptPbItem(pbItem)
		assert.Error(t, err)
		err = ts.Client.DecryptPbItem(pbItem)
		assert.Error(t, err)

		wantItem := TestingNewPbLoginItem()
		wantItem.Secrets.Secret = nil
		wantItem.Secrets.Notes = nil
		wantItem.Additions.Uris = nil
		pbItem.Updated = wantItem.Updated
		if !reflect.DeepEqual(pbItem, wantItem) {
			t.Errorf("Encrypt/decrypt err: got %v, want %v", pbItem, wantItem)
		}
	})
}

func BenchmarkGRPCClient_EncryptDecrypt(b *testing.B) {
	ts := NewTestSuiteGRPClient(b)
	defer ts.Stop()
	ts.Client.encKey = testGRPCencKey

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
