package grpcapi

import (
	"context"
	"fmt"

	"github.com/artfuldog/gophkeeper/internal/common"
	"github.com/artfuldog/gophkeeper/internal/logger"
	"github.com/artfuldog/gophkeeper/internal/pb"
	"github.com/artfuldog/gophkeeper/internal/server/db"
)

// GRPCService implements all GRPC-method for handling request and stores service options.
// Used for registering with GRPC-server.
type ItemsService struct {
	pb.UnimplementedItemsServer
	db     db.DB
	logger logger.L
}

// NewGRPCService a constructor for GRPCService.
func NewItemsService(db db.DB, l logger.L) *ItemsService {
	return &ItemsService{
		db:     db,
		logger: l,
	}
}

// CreateItem creates new item.
func (s *ItemsService) CreateItem(ctx context.Context, req *pb.CreateItemRequest) (*pb.CreateItemResponse, error) {
	componentName := "ItemsService:CreateItem"
	resp := new(pb.CreateItemResponse)

	if err := s.db.CreateItem(ctx, req.Username, req.Item); err != nil {
		s.logger.Warn(err, "db error", componentName)
		return nil, wrapErrorToClient(err)
	}

	resp.Info = fmt.Sprintf("successfully create item type %s '%s'",
		common.ItemTypeText(req.Item.Type), req.Item.Name)

	return resp, nil
}

// GetItem returns item's information.
func (s *ItemsService) GetItem(ctx context.Context, req *pb.GetItemRequest) (*pb.GetItemResponse, error) {
	componentName := "ItemsService:GetItem"
	resp := new(pb.GetItemResponse)

	var err error

	resp.Item, err = s.db.GetItemByNameAndType(ctx, req.Username, req.ItemName, req.ItemType)
	if err != nil {
		s.logger.Warn(err, "db error", componentName)
		return nil, wrapErrorToClient(err)
	}

	return resp, nil
}

// GetItemList returns list with items' short representation.
func (s *ItemsService) GetItemList(ctx context.Context, req *pb.GetItemListRequest) (*pb.GetItemListResponse, error) {
	componentName := "ItemsService:GetItemList"
	resp := new(pb.GetItemListResponse)

	var err error

	resp.Items, err = s.db.GetItemList(ctx, req.Username)
	if err != nil {
		s.logger.Warn(err, "db error", componentName)
		return nil, wrapErrorToClient(err)
	}

	return resp, nil
}

// GetItem returns item's information.
func (s *ItemsService) GetItems(ctx context.Context, req *pb.GetItemsRequest) (*pb.GetItemsResponse, error) {
	componentName := "ItemsService:GetItems"
	resp := new(pb.GetItemsResponse)

	var err error

	resp.Items, err = s.db.GetItemsByID(ctx, req.Username, req.Ids)
	if err != nil {
		s.logger.Warn(err, "db error", componentName)
		return nil, wrapErrorToClient(err)
	}

	return resp, nil
}

// GetItem returns item's information.
func (s *ItemsService) GetItemHash(ctx context.Context, req *pb.GetItemHashRequest) (*pb.GetItemHashResponse, error) {
	componentName := "ItemsService:GetItemHash"
	resp := new(pb.GetItemHashResponse)

	var err error

	resp.Hash, err = s.db.GetItemHashByID(ctx, req.Id)
	if err != nil {
		s.logger.Warn(err, "db error", componentName)
		return nil, wrapErrorToClient(err)
	}

	return resp, nil
}

// UpdateItem updates existing item.
func (s *ItemsService) UpdateItem(ctx context.Context, req *pb.UpdateItemRequest) (*pb.UpdateItemResponse, error) {
	componentName := "ItemsService:UpdateItem"
	resp := new(pb.UpdateItemResponse)

	if err := s.db.UpdateItem(ctx, req.Username, req.Item); err != nil {
		s.logger.Warn(err, "db error", componentName)
		return nil, wrapErrorToClient(err)
	}

	resp.Info = fmt.Sprintf("successfully update item type %s '%s'",
		common.ItemTypeText(req.Item.Type), req.Item.Name)

	return resp, nil
}

// DeleteItem deletes item.
func (s *ItemsService) DeleteItem(ctx context.Context, req *pb.DeleteItemRequest) (*pb.DeleteItemResponse, error) {
	componentName := "ItemsService:DeleteItem"
	resp := new(pb.DeleteItemResponse)

	if err := s.db.DeleteItem(ctx, req.Username, req.Id); err != nil {
		s.logger.Warn(err, "db error", componentName)
		return nil, wrapErrorToClient(err)
	}

	resp.Info = "item deleted"

	return resp, nil
}
