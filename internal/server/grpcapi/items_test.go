package grpcapi

import (
	"testing"

	"github.com/artfuldog/gophkeeper/internal/common"
	"github.com/artfuldog/gophkeeper/internal/mocks/mockdb"
	"github.com/artfuldog/gophkeeper/internal/mocks/mocklogger"
	"github.com/artfuldog/gophkeeper/internal/pb"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewItemsService(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	assert.NotEmpty(t, NewItemsService(mockdb.NewMockDB(mockCtrl), mocklogger.NewMockLogger()))
	mockCtrl.Finish()
}

func TestItemsService_CreateItem(t *testing.T) {
	ts, tsErr := NewTestSuiteGRPCServer(t)
	if tsErr != nil {
		t.Errorf("failed to init test suite: %v", tsErr)
	}
	defer ts.Stop()

	t.Run("DB returns error", func(t *testing.T) {
		ts.DB.EXPECT().CreateItem(mockAny, mockAny, mockAny).Return(assert.AnError)
		req := &pb.CreateItemRequest{}
		_, err := ts.ItemsClient.CreateItem(testCtx, req)
		assert.Error(t, err)
	})

	t.Run("Successfully create", func(t *testing.T) {
		ts.DB.EXPECT().CreateItem(mockAny, mockAny, mockAny).Return(nil)
		req := &pb.CreateItemRequest{
			Item: &pb.Item{
				Name: "name",
				Type: common.ItemTypeLogin,
			},
		}
		resp, err := ts.ItemsClient.CreateItem(testCtx, req)
		require.NoError(t, err)
		assert.NotEmpty(t, resp)
	})
}

func TestItemsService_GetItem(t *testing.T) {
	ts, tsErr := NewTestSuiteGRPCServer(t)
	if tsErr != nil {
		t.Errorf("failed to init test suite: %v", tsErr)
	}
	defer ts.Stop()

	t.Run("DB returns error", func(t *testing.T) {
		ts.DB.EXPECT().GetItemByNameAndType(mockAny, mockAny, mockAny, mockAny).Return(nil, assert.AnError)
		req := &pb.GetItemRequest{}
		_, err := ts.ItemsClient.GetItem(testCtx, req)
		assert.Error(t, err)
	})

	t.Run("Successfully get", func(t *testing.T) {
		respItem := &pb.Item{
			Name: "name",
		}
		resp := &pb.GetItemResponse{
			Item: respItem,
		}
		ts.DB.EXPECT().GetItemByNameAndType(mockAny, mockAny, mockAny, mockAny).Return(respItem, nil)
		req := &pb.GetItemRequest{}
		gotResp, err := ts.ItemsClient.GetItem(testCtx, req)
		require.NoError(t, err)
		assert.Equal(t, resp.Item.Name, gotResp.Item.Name)
	})
}

func TestItemsService_GetItemList(t *testing.T) {
	ts, tsErr := NewTestSuiteGRPCServer(t)
	if tsErr != nil {
		t.Errorf("failed to init test suite: %v", tsErr)
	}
	defer ts.Stop()

	t.Run("DB returns error", func(t *testing.T) {
		ts.DB.EXPECT().GetItemList(mockAny, mockAny).Return(nil, assert.AnError)
		req := &pb.GetItemListRequest{}
		_, err := ts.ItemsClient.GetItemList(testCtx, req)
		assert.Error(t, err)
	})

	t.Run("Successfully get list", func(t *testing.T) {
		respItems := []*pb.ItemShort{
			{
				Name: "name",
			},
		}
		ts.DB.EXPECT().GetItemList(mockAny, mockAny).Return(respItems, nil)
		req := &pb.GetItemListRequest{}
		gotResp, err := ts.ItemsClient.GetItemList(testCtx, req)
		require.NoError(t, err)
		assert.NotEmpty(t, gotResp)
	})
}

func TestItemsService_GetItems(t *testing.T) {
	ts, tsErr := NewTestSuiteGRPCServer(t)
	if tsErr != nil {
		t.Errorf("failed to init test suite: %v", tsErr)
	}
	defer ts.Stop()

	t.Run("DB returns error", func(t *testing.T) {
		ts.DB.EXPECT().GetItemsByID(mockAny, mockAny, mockAny).Return(nil, assert.AnError)
		req := &pb.GetItemsRequest{}
		_, err := ts.ItemsClient.GetItems(testCtx, req)
		assert.Error(t, err)
	})

	t.Run("Successfully get items", func(t *testing.T) {
		respItems := []*pb.Item{
			{
				Name: "name",
			},
		}
		ts.DB.EXPECT().GetItemsByID(mockAny, mockAny, mockAny).Return(respItems, nil)
		req := &pb.GetItemsRequest{}
		gotResp, err := ts.ItemsClient.GetItems(testCtx, req)
		require.NoError(t, err)
		assert.NotEmpty(t, gotResp)
	})
}

func TestItemsService_GetItemHash(t *testing.T) {
	ts, tsErr := NewTestSuiteGRPCServer(t)
	if tsErr != nil {
		t.Errorf("failed to init test suite: %v", tsErr)
	}
	defer ts.Stop()

	t.Run("DB returns error", func(t *testing.T) {
		ts.DB.EXPECT().GetItemHashByID(mockAny, mockAny).Return(nil, assert.AnError)
		req := &pb.GetItemHashRequest{}
		_, err := ts.ItemsClient.GetItemHash(testCtx, req)
		assert.Error(t, err)
	})

	t.Run("Successfully get hash", func(t *testing.T) {
		ts.DB.EXPECT().GetItemHashByID(mockAny, mockAny).Return([]byte("hash"), nil)
		req := &pb.GetItemHashRequest{}
		resp, err := ts.ItemsClient.GetItemHash(testCtx, req)
		require.NoError(t, err)
		assert.Equal(t, []byte("hash"), resp.Hash)
	})
}

func TestItemsService_UpdateItem(t *testing.T) {
	ts, tsErr := NewTestSuiteGRPCServer(t)
	if tsErr != nil {
		t.Errorf("failed to init test suite: %v", tsErr)
	}
	defer ts.Stop()

	t.Run("DB returns error", func(t *testing.T) {
		ts.DB.EXPECT().UpdateItem(mockAny, mockAny, mockAny).Return(assert.AnError)
		req := &pb.UpdateItemRequest{}
		_, err := ts.ItemsClient.UpdateItem(testCtx, req)
		assert.Error(t, err)
	})

	t.Run("Successfully updated", func(t *testing.T) {
		ts.DB.EXPECT().UpdateItem(mockAny, mockAny, mockAny).Return(nil)
		req := &pb.UpdateItemRequest{
			Item: &pb.Item{
				Name: "name",
				Type: common.ItemTypeLogin,
			},
		}
		resp, err := ts.ItemsClient.UpdateItem(testCtx, req)
		require.NoError(t, err)
		assert.NotEmpty(t, resp)
	})
}

func TestItemsService_DeleteItem(t *testing.T) {
	ts, tsErr := NewTestSuiteGRPCServer(t)
	if tsErr != nil {
		t.Errorf("failed to init test suite: %v", tsErr)
	}
	defer ts.Stop()

	t.Run("DB returns error", func(t *testing.T) {
		ts.DB.EXPECT().DeleteItem(mockAny, mockAny, mockAny).Return(assert.AnError)
		req := &pb.DeleteItemRequest{}
		_, err := ts.ItemsClient.DeleteItem(testCtx, req)
		assert.Error(t, err)
	})

	t.Run("Successfully deleted", func(t *testing.T) {
		ts.DB.EXPECT().DeleteItem(mockAny, mockAny, mockAny).Return(nil)
		req := &pb.DeleteItemRequest{}
		resp, err := ts.ItemsClient.DeleteItem(testCtx, req)
		require.NoError(t, err)
		assert.NotEmpty(t, resp)
	})
}
