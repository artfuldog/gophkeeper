package api

import (
	"bytes"
	"context"
	"time"

	"github.com/artfuldog/gophkeeper/internal/client/config"
	"github.com/artfuldog/gophkeeper/internal/client/storage"
	"github.com/artfuldog/gophkeeper/internal/pb"
)

// ItemsToSync contains IDs of items, which need to be created/updated/deleted.
type ItemsToSync struct {
	Create []int64
	Update []int64
	Delete []int64
}

// NewItemsToSync creates new intance of ItemsToSync.
func NewItemsToSync() *ItemsToSync {
	return &ItemsToSync{
		Create: []int64{},
		Update: []int64{},
		Delete: []int64{},
	}
}

// SyncType defines type of force synchronization.
type SyncType uint8

const (
	// Force synchronization in background.
	SyncBackground SyncType = iota
	// Force synchronization and wait for it finishes.
	SyncWaitForComplete
)

const (
	SyncStatusOK         = "Synced"
	SyncStatusInProgress = "Syncing..."
	SyncStatusFailed     = "Sync error"
)

// ForceSyncBackground forces synchronization in background.
func (c *GRPCClient) ForceSyncBackground() {
	if c.config.GetMode() == config.ModeLocal {
		c.syncControlCh <- SyncBackground
	}
}

// ForceSyncWithWait forces synchronization and block execution until synchronization will stop.
func (c *GRPCClient) ForceSyncWithWait() {
	if c.config.GetMode() == config.ModeLocal {
		c.syncControlCh <- SyncWaitForComplete
		<-c.syncCompleteCh
	}
}

// StorageInit intializes agent's storage.
func (c *GRPCClient) StorageInit(ctx context.Context, statusCh chan<- string) (err error) {
	if c.config.GetMode() == config.ModeServer {
		return nil
	}

	c.storage, err = storage.New(storage.TypeSQLite, c.config.GetUser(), c.config.GetAppConfigDir())
	if err != nil {
		return err
	}

	c.storageStopCh = make(chan struct{})

	if err := c.storage.Connect(ctx, c.storageStopCh); err != nil {
		return err
	}

	go c.Sync(ctx, statusCh)

	c.ForceSyncBackground()

	return nil
}

// Sync maintains synchronization process.
func (c *GRPCClient) Sync(ctx context.Context, statusCh chan<- string) {
	ticker := time.NewTicker(c.config.GetSyncInterval())

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			statusCh <- SyncStatusInProgress

			if err := c.syncExec(ctx); err != nil {
				statusCh <- SyncStatusFailed
				continue
			}

			statusCh <- SyncStatusOK
		case sync := <-c.syncControlCh:
			statusCh <- SyncStatusInProgress

			if err := c.syncExec(ctx); err != nil {
				statusCh <- SyncStatusFailed

				if sync == SyncWaitForComplete {
					c.syncCompleteCh <- struct{}{}
				}

				continue
			}

			statusCh <- SyncStatusOK

			if sync == SyncWaitForComplete {
				c.syncCompleteCh <- struct{}{}
			}
		}
	}
}

// syncExec runs synchronization process.
//
// Synchronization process contains following steps:
//   - check current revision
//   - if revision is different synchronize items with server (please watch syncItems)
func (c *GRPCClient) syncExec(ctx context.Context) error {
	revEqual, srvRevision, err := c.revisionsIsEqual(ctx)
	if err != nil {
		return err
	}

	if revEqual {
		return nil
	}

	return c.syncItems(ctx, srvRevision)
}

// revisionsIsEqual compares server's and local storage's revisions.
//
// Returns true if equal, false if not. Also returns server's revision.
func (c *GRPCClient) revisionsIsEqual(ctx context.Context) (bool, []byte, error) {
	resp, err := c.usersClient.GetRevision(ctx, &pb.GetRevisionRequest{Username: c.config.GetUser()})
	if err != nil {
		return false, nil, err
	}

	storRevision, err := c.storage.GetRevision(ctx)
	if err != nil {
		return false, nil, err
	}

	if bytes.Equal(resp.Revision, storRevision) {
		return true, resp.Revision, nil
	}

	return false, resp.Revision, nil
}

// syncItems synchronizes items with server and update local storage.
func (c *GRPCClient) syncItems(ctx context.Context, revision []byte) error {
	resp, err := c.itemsClient.GetItemList(ctx, &pb.GetItemListRequest{Username: c.config.GetUser()})
	if err != nil {
		return err
	}

	storItemsList, err := c.storage.GetItemsList(ctx)
	if err != nil {
		return err
	}

	itemToSync := c.prepareItemsToSync(resp.Items, storItemsList)

	if len(itemToSync.Create) > 0 {
		createItems, err := c.GetItemsForStorage(ctx, itemToSync.Create)
		if err != nil {
			return err
		}

		err = c.storage.CreateItems(ctx, createItems)
		if err != nil {
			return err
		}
	}

	if len(itemToSync.Update) > 0 {
		updateItems, err := c.GetItemsForStorage(ctx, itemToSync.Update)
		if err != nil {
			return err
		}

		err = c.storage.UpdateItems(ctx, updateItems)
		if err != nil {
			return err
		}
	}

	if len(itemToSync.Delete) > 0 {
		err = c.storage.DeleteItems(ctx, itemToSync.Delete)
		if err != nil {
			return err
		}
	}

	err = c.storage.SaveRevision(ctx, revision)
	if err != nil {
		return err
	}

	return nil
}

// prepareItemsToSync is a helper function which prepares list of items' IDs for update,
// delete of create.
func (c *GRPCClient) prepareItemsToSync(dbItemsList []*pb.ItemShort,
	storItemsList storage.Items) *ItemsToSync {

	dbItemsMap := map[int64]*pb.ItemShort{}
	for _, item := range dbItemsList {
		dbItemsMap[item.Id] = item
	}

	res := NewItemsToSync()

	for _, item := range storItemsList {
		dbItem, ok := dbItemsMap[item.ID]
		if !ok {
			res.Delete = append(res.Delete, item.ID)
			continue
		}

		if bytes.Equal(dbItem.Hash, item.Hash) {
			delete(dbItemsMap, dbItem.Id)
			continue
		}

		res.Update = append(res.Update, item.ID)

		delete(dbItemsMap, dbItem.Id)
	}

	for _, dbItem := range dbItemsMap {
		res.Create =
			append(res.Create, dbItem.Id)
	}

	return res
}
