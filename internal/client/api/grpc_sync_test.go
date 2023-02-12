package api

import (
	"context"
	"os"
	"testing"

	"github.com/artfuldog/gophkeeper/internal/client/config"
	"github.com/artfuldog/gophkeeper/internal/client/storage"
	"github.com/artfuldog/gophkeeper/internal/pb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestGRPCClient_StorageInit(t *testing.T) {
	ts := NewTestSuiteGRPClient(t)
	defer ts.Stop()

	t.Run("Server mode", func(t *testing.T) {
		assert.NoError(t, ts.Client.StorageInit(testGRPCctx, nil))
	})

	ts.Client.config.SetMode(config.ModeLocal)

	t.Run("Connect to storage failed", func(t *testing.T) {
		ts.Client.config.SetUser("<>&|:::")
		assert.Error(t, ts.Client.StorageInit(testGRPCctx, nil))
	})

	t.Run("Successful init", func(t *testing.T) {
		ts.Client.config.SetUser("testgrpcsyncuser")
		ts.Client.config.CreateAppDir()
		defer os.RemoveAll(ts.Client.config.GetAppConfigDir())

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		require.NoError(t, ts.Client.StorageInit(ctx, nil))
	})
}

func TestGRPCClient_syncExec(t *testing.T) {
	ts := NewTestSuiteGRPClient(t)
	defer ts.Stop()

	t.Run("Get revision error", func(t *testing.T) {
		ts.UsersClient.EXPECT().GetRevision(testGRPCctx, mockAnyVal).Return(nil, assert.AnError)
		assert.Error(t, ts.Client.syncExec(testGRPCctx))
	})

	t.Run("Equal revisions", func(t *testing.T) {
		srvResp := &pb.GetRevisionResponse{Revision: []byte("revision1")}
		ts.UsersClient.EXPECT().GetRevision(testGRPCctx, mockAnyVal).Return(srvResp, nil)
		ts.Storage.EXPECT().GetRevision(testGRPCctx).Return([]byte("revision1"), nil)
		assert.NoError(t, ts.Client.syncExec(testGRPCctx))
	})

	t.Run("Unequal revisions", func(t *testing.T) {
		srvResp := &pb.GetRevisionResponse{Revision: []byte("revision1")}
		ts.UsersClient.EXPECT().GetRevision(testGRPCctx, mockAnyVal).Return(srvResp, nil)
		ts.Storage.EXPECT().GetRevision(testGRPCctx).Return([]byte("revision2"), nil)
		ts.ItemsClient.EXPECT().GetItemList(testGRPCctx, mockAnyVal).Return(nil, assert.AnError)
		assert.Error(t, ts.Client.syncExec(testGRPCctx))
	})
}

func TestGRPCClient_revisionsIsEqual(t *testing.T) {
	ts := NewTestSuiteGRPClient(t)
	defer ts.Stop()

	t.Run("Get revision from server error", func(t *testing.T) {
		ts.UsersClient.EXPECT().GetRevision(testGRPCctx, mockAnyVal).Return(nil, assert.AnError)

		_, _, err := ts.Client.revisionsIsEqual(testGRPCctx)
		assert.Error(t, err)
	})

	t.Run("Get revision from local storage error", func(t *testing.T) {
		ts.UsersClient.EXPECT().GetRevision(testGRPCctx, mockAnyVal).Return(nil, nil)
		ts.Storage.EXPECT().GetRevision(testGRPCctx).Return(nil, assert.AnError)

		_, _, err := ts.Client.revisionsIsEqual(testGRPCctx)
		assert.Error(t, err)
	})

	t.Run("Same revision", func(t *testing.T) {
		srvResp := &pb.GetRevisionResponse{Revision: []byte("revision1")}

		ts.UsersClient.EXPECT().GetRevision(testGRPCctx, mockAnyVal).Return(srvResp, nil)
		ts.Storage.EXPECT().GetRevision(testGRPCctx).Return([]byte("revision1"), nil)

		equal, revision, err := ts.Client.revisionsIsEqual(testGRPCctx)
		require.NoError(t, err)
		assert.Equal(t, true, equal)
		assert.Equal(t, []byte("revision1"), revision)
	})

	t.Run("Different revisions", func(t *testing.T) {
		srvResp := &pb.GetRevisionResponse{Revision: []byte("srvrevision")}

		ts.UsersClient.EXPECT().GetRevision(testGRPCctx, mockAnyVal).Return(srvResp, nil)
		ts.Storage.EXPECT().GetRevision(testGRPCctx).Return([]byte("localrevision"), nil)

		equal, revision, err := ts.Client.revisionsIsEqual(testGRPCctx)
		require.NoError(t, err)
		assert.Equal(t, false, equal)
		assert.Equal(t, []byte("srvrevision"), revision)
	})
}

func TestGRPCClient_syncItems(t *testing.T) {
	ts := NewTestSuiteGRPClient(t)
	defer ts.Stop()

	ts.Client.config.SetUser("testgrpcsyncuser")
	ts.Client.config.CreateAppDir()
	defer os.RemoveAll(ts.Client.config.GetAppConfigDir())

	t.Run("Get item list from server error", func(t *testing.T) {
		ts.ItemsClient.EXPECT().GetItemList(testGRPCctx, mockAnyVal).Return(nil, assert.AnError)

		assert.Error(t, ts.Client.syncItems(testGRPCctx, nil))
	})

	t.Run("Get item list from local storage error", func(t *testing.T) {
		ts.ItemsClient.EXPECT().GetItemList(testGRPCctx, mockAnyVal).Return(nil, nil)
		ts.Storage.EXPECT().GetItemsList(testGRPCctx).Return(nil, assert.AnError)

		assert.Error(t, ts.Client.syncItems(testGRPCctx, nil))
	})

	t.Run("Save revision error", func(t *testing.T) {
		srvResp := &pb.GetItemListResponse{
			Items: []*pb.ItemShort{},
		}
		storResp := storage.Items{}

		ts.ItemsClient.EXPECT().GetItemList(testGRPCctx, mockAnyVal).Return(srvResp, nil)
		ts.Storage.EXPECT().GetItemsList(testGRPCctx).Return(storResp, nil)
		ts.Storage.EXPECT().SaveRevision(testGRPCctx, mockAnyVal).Return(assert.AnError)

		assert.Error(t, ts.Client.syncItems(testGRPCctx, nil))
	})

	t.Run("Empty sync list", func(t *testing.T) {
		srvResp := &pb.GetItemListResponse{
			Items: []*pb.ItemShort{},
		}
		storResp := storage.Items{}

		ts.ItemsClient.EXPECT().GetItemList(testGRPCctx, mockAnyVal).Return(srvResp, nil)
		ts.Storage.EXPECT().GetItemsList(testGRPCctx).Return(storResp, nil)
		ts.Storage.EXPECT().SaveRevision(testGRPCctx, mockAnyVal).Return(nil)

		require.NoError(t, ts.Client.syncItems(testGRPCctx, nil))
	})

	t.Run("Create new items", func(t *testing.T) {
		srvRespList := &pb.GetItemListResponse{
			Items: []*pb.ItemShort{
				{
					Id:      500,
					Name:    "Item",
					Type:    "l",
					Hash:    []byte("hash"),
					Updated: timestamppb.Now(),
				},
			},
		}

		srvRespItem := &pb.GetItemsResponse{Items: []*pb.Item{}}

		storResp := storage.Items{}

		ts.ItemsClient.EXPECT().GetItemList(testGRPCctx, mockAnyVal).Return(srvRespList, nil)
		ts.Storage.EXPECT().GetItemsList(testGRPCctx).Return(storResp, nil)
		ts.ItemsClient.EXPECT().GetItems(testGRPCctx, mockAnyVal).Return(nil, assert.AnError)

		assert.Error(t, ts.Client.syncItems(testGRPCctx, nil))

		ts.ItemsClient.EXPECT().GetItemList(testGRPCctx, mockAnyVal).Return(srvRespList, nil)
		ts.Storage.EXPECT().GetItemsList(testGRPCctx).Return(storResp, nil)
		ts.ItemsClient.EXPECT().GetItems(testGRPCctx, mockAnyVal).Return(srvRespItem, nil)
		ts.Storage.EXPECT().CreateItems(testGRPCctx, mockAnyVal).Return(assert.AnError)

		assert.Error(t, ts.Client.syncItems(testGRPCctx, nil))

		ts.ItemsClient.EXPECT().GetItemList(testGRPCctx, mockAnyVal).Return(srvRespList, nil)
		ts.Storage.EXPECT().GetItemsList(testGRPCctx).Return(storResp, nil)
		ts.ItemsClient.EXPECT().GetItems(testGRPCctx, mockAnyVal).Return(srvRespItem, nil)
		ts.Storage.EXPECT().CreateItems(testGRPCctx, mockAnyVal).Return(nil)
		ts.Storage.EXPECT().SaveRevision(testGRPCctx, mockAnyVal).Return(nil)

		require.NoError(t, ts.Client.syncItems(testGRPCctx, nil))
	})

	t.Run("Update items", func(t *testing.T) {
		srvRespList := &pb.GetItemListResponse{
			Items: []*pb.ItemShort{
				{
					Id:      500,
					Name:    "Item",
					Type:    "l",
					Hash:    []byte("hash"),
					Updated: timestamppb.Now(),
				},
			},
		}

		srvRespItem := &pb.GetItemsResponse{Items: []*pb.Item{}}

		storResp := storage.Items{
			{
				ID:   500,
				Name: "Item new name",
				Type: "l",
				Hash: []byte("new hash"),
			},
		}

		ts.ItemsClient.EXPECT().GetItemList(testGRPCctx, mockAnyVal).Return(srvRespList, nil)
		ts.Storage.EXPECT().GetItemsList(testGRPCctx).Return(storResp, nil)
		ts.ItemsClient.EXPECT().GetItems(testGRPCctx, mockAnyVal).Return(nil, assert.AnError)

		assert.Error(t, ts.Client.syncItems(testGRPCctx, nil))

		ts.ItemsClient.EXPECT().GetItemList(testGRPCctx, mockAnyVal).Return(srvRespList, nil)
		ts.Storage.EXPECT().GetItemsList(testGRPCctx).Return(storResp, nil)
		ts.ItemsClient.EXPECT().GetItems(testGRPCctx, mockAnyVal).Return(srvRespItem, nil)
		ts.Storage.EXPECT().UpdateItems(testGRPCctx, mockAnyVal).Return(assert.AnError)

		assert.Error(t, ts.Client.syncItems(testGRPCctx, nil))

		ts.ItemsClient.EXPECT().GetItemList(testGRPCctx, mockAnyVal).Return(srvRespList, nil)
		ts.Storage.EXPECT().GetItemsList(testGRPCctx).Return(storResp, nil)
		ts.ItemsClient.EXPECT().GetItems(testGRPCctx, mockAnyVal).Return(srvRespItem, nil)
		ts.Storage.EXPECT().UpdateItems(testGRPCctx, mockAnyVal).Return(nil)
		ts.Storage.EXPECT().SaveRevision(testGRPCctx, mockAnyVal).Return(nil)

		require.NoError(t, ts.Client.syncItems(testGRPCctx, nil))
	})

	t.Run("Delete items", func(t *testing.T) {
		srvRespList := &pb.GetItemListResponse{
			Items: []*pb.ItemShort{},
		}

		storResp := storage.Items{
			{
				ID:   500,
				Name: "Item new name",
				Type: "l",
				Hash: []byte("new hash"),
			},
		}

		ts.ItemsClient.EXPECT().GetItemList(testGRPCctx, mockAnyVal).Return(srvRespList, nil)
		ts.Storage.EXPECT().GetItemsList(testGRPCctx).Return(storResp, nil)
		ts.Storage.EXPECT().DeleteItems(testGRPCctx, mockAnyVal).Return(assert.AnError)

		assert.Error(t, ts.Client.syncItems(testGRPCctx, nil))

		ts.ItemsClient.EXPECT().GetItemList(testGRPCctx, mockAnyVal).Return(srvRespList, nil)
		ts.Storage.EXPECT().GetItemsList(testGRPCctx).Return(storResp, nil)
		ts.Storage.EXPECT().DeleteItems(testGRPCctx, mockAnyVal).Return(nil)
		ts.Storage.EXPECT().SaveRevision(testGRPCctx, mockAnyVal).Return(nil)

		require.NoError(t, ts.Client.syncItems(testGRPCctx, nil))
	})
}

func TestGRPCClient_PrepareItemsToSync(t *testing.T) {
	ts := NewTestSuiteGRPClient(t)
	defer ts.Stop()

	dbItems := []*pb.ItemShort{
		{
			Id:      201,
			Name:    "UpdatedItem",
			Type:    "l",
			Hash:    []byte("updated hash"),
			Updated: timestamppb.Now(),
		},
		{
			Id:      100,
			Name:    "NewItem",
			Type:    "l",
			Hash:    []byte("new item hash"),
			Updated: timestamppb.Now(),
		},
		{
			Id:      401,
			Name:    "UnchangeItem",
			Type:    "l",
			Hash:    []byte("item hash"),
			Updated: timestamppb.Now(),
		},
		{
			Id:      202,
			Name:    "NewNameUpdatedItem2",
			Type:    "l",
			Hash:    []byte("updated hash"),
			Updated: timestamppb.Now(),
		},
	}
	storeItems := []*storage.Item{
		{
			ID:   301,
			Name: "DeleteItem",
			Type: "l",
			Hash: []byte("item hash"),
		},
		{
			ID:   201,
			Name: "UpdatedItem",
			Type: "l",
			Hash: []byte("old hash"),
		},
		{
			ID:   202,
			Name: "UpdatedItem2",
			Type: "l",
			Hash: []byte("old hash"),
		},
		{
			ID:   302,
			Name: "DeleteItem2",
			Type: "c",
			Hash: []byte("item hash"),
		},
		{
			ID:   401,
			Name: "UnchangeItem",
			Type: "l",
			Hash: []byte("item hash"),
		},
	}

	items := ts.Client.prepareItemsToSync(dbItems, storeItems)

	assert.Equal(t, items.Create, []int64{100})
	assert.Equal(t, items.Update, []int64{201, 202})
	assert.Equal(t, items.Delete, []int64{301, 302})
}
