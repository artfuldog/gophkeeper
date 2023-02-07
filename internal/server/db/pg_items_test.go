package db

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/artfuldog/gophkeeper/internal/pb"
	"github.com/stretchr/testify/assert"
)

func TestDBPosgtre_CreateItem(t *testing.T) {
	canceledCtx, cancel := context.WithCancel(context.Background())
	cancel()

	itemUsername := testUser2.Username

	type args struct {
		ctx      context.Context
		username string
		item     *pb.Item
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		err     error
	}{
		{
			name: "Create new login item",
			args: args{
				ctx:      context.Background(),
				username: itemUsername,
				item:     testItemLogin,
			},
			wantErr: false,
		},
		{
			name: "Create new card item",
			args: args{
				ctx:      context.Background(),
				username: itemUsername,
				item:     testItemCard,
			},
			wantErr: false,
		},
		{
			name: "Create new note item",
			args: args{
				ctx:      context.Background(),
				username: itemUsername,
				item:     testItemNotes,
			},
			wantErr: false,
		},
		{
			name: "Create new data item",
			args: args{
				ctx:      context.Background(),
				username: itemUsername,
				item:     testItemData,
			},
			wantErr: false,
		},
		{
			name: "Create new item without notes and secrets",
			args: args{
				ctx:      context.Background(),
				username: itemUsername,
				item:     testItemEmptyNotesSecrets,
			},
			wantErr: false,
		},
		{
			name: "Create new item without notes and additions",
			args: args{
				ctx:      context.Background(),
				username: itemUsername,
				item:     testItemEmptyAdditionsNotes,
			},
			wantErr: false,
		},
		{
			name: "Missed username",
			args: args{
				ctx:  context.Background(),
				item: testItemLogin,
			},
			wantErr: true,
			err:     ErrNotFound,
		},
		{
			name: "Duplicate entry",
			args: args{
				ctx:      context.Background(),
				username: testUser1.Username,
				item:     testItemLogin,
			},
			wantErr: true,
			err:     ErrDuplicateEntry,
		},
		{
			name: "Canceled context",
			args: args{
				ctx:      canceledCtx,
				username: itemUsername,
				item:     testItemLogin,
			},
			wantErr: true,
			err:     ErrTransactionFailed,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := testDB.CreateItem(tt.args.ctx, tt.args.username, tt.args.item)
			if (err != nil) != tt.wantErr {
				t.Errorf("DBPosgtre.CreateUser() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && !errors.Is(tt.err, assert.AnError) {
				assert.ErrorIs(t, err, tt.err)
			}
		})
	}

	for _, item := range testItems {
		t.Run(fmt.Sprintf("Check creating %s was succesfull", item.Name), func(t *testing.T) {
			newItem, _ := testDB.GetItemByNameAndType(context.Background(),
				testUser2.Username, item.Name, item.Type)

			if !reflect.DeepEqual(newItem.Secrets, item.Secrets) {
				t.Errorf("Responce not equal - got:  %v, want %v", newItem.Secrets, item.Secrets)
			}

			if !reflect.DeepEqual(newItem.Additions, item.Additions) {
				t.Errorf("Responce not equal - got:  %v, want %v", newItem.Additions, item.Additions)
			}

			if err := testDB.DeleteItem(context.Background(), testUser2.Username, newItem.Id); err != nil {
				t.Errorf("DBPosgtre.CreateUser() - failed delete test item: %v", err)
			}
		})
	}
}

func TestDBPosgtre_GetItemByNameAndType(t *testing.T) {
	canceledCtx, cancel := context.WithCancel(context.Background())
	cancel()
	itemUsername := testUser1.Username

	type args struct {
		ctx      context.Context
		username Username
		itemName string
		itemType string
	}
	tests := []struct {
		name    string
		args    args
		want    *pb.Item
		wantErr bool
		err     error
	}{
		{
			name: "Get login item",
			args: args{
				ctx:      context.Background(),
				username: itemUsername,
				itemName: testItemLogin.Name,
				itemType: testItemLogin.Type,
			},
			want:    testItemLogin,
			wantErr: false,
		},
		{
			name: "Get card item",
			args: args{
				ctx:      context.Background(),
				username: itemUsername,
				itemName: testItemCard.Name,
				itemType: testItemCard.Type,
			},
			want:    testItemCard,
			wantErr: false,
		},
		{
			name: "Get notes item",
			args: args{
				ctx:      context.Background(),
				username: itemUsername,
				itemName: testItemNotes.Name,
				itemType: testItemNotes.Type,
			},
			want:    testItemNotes,
			wantErr: false,
		},
		{
			name: "Get data item",
			args: args{
				ctx:      context.Background(),
				username: itemUsername,
				itemName: testItemData.Name,
				itemType: testItemData.Type,
			},
			want:    testItemData,
			wantErr: false,
		},
		{
			name: "Get item without notes and secrets",
			args: args{
				ctx:      context.Background(),
				username: itemUsername,
				itemName: testItemEmptyNotesSecrets.Name,
				itemType: testItemEmptyNotesSecrets.Type,
			},
			want:    testItemEmptyNotesSecrets,
			wantErr: false,
		},
		{
			name: "Get item without notes and additions",
			args: args{
				ctx:      context.Background(),
				username: itemUsername,
				itemName: testItemEmptyAdditionsNotes.Name,
				itemType: testItemEmptyAdditionsNotes.Type,
			},
			want:    testItemEmptyAdditionsNotes,
			wantErr: false,
		},
		{
			name: "Get item without notes and additions",
			args: args{
				ctx:      context.Background(),
				username: itemUsername,
				itemName: testItemEmptyAdditionsNotes.Name,
				itemType: testItemEmptyAdditionsNotes.Type,
			},
			want:    testItemEmptyAdditionsNotes,
			wantErr: false,
		},
		{
			name: "Get item for unexisted user",
			args: args{
				ctx:      context.Background(),
				username: "wrong user",
				itemName: testItemLogin.Name,
				itemType: testItemLogin.Type,
			},
			wantErr: true,
			err:     ErrNotFound,
		},
		{
			name: "Get unexisted item",
			args: args{
				ctx:      context.Background(),
				username: itemUsername,
				itemName: "wrong item name",
				itemType: testItemLogin.Type,
			},
			wantErr: true,
			err:     ErrNotFound,
		},
		{
			name: "Canceled context",
			args: args{
				ctx:      canceledCtx,
				username: itemUsername,
				itemName: testItemLogin.Name,
				itemType: testItemLogin.Type,
			},
			wantErr: true,
			err:     ErrTransactionFailed,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := testDB.GetItemByNameAndType(tt.args.ctx, tt.args.username, tt.args.itemName, tt.args.itemType)
			if (err != nil) != tt.wantErr {
				t.Errorf("DBPosgtre.GetItemByNameAndType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && !errors.Is(tt.err, assert.AnError) {
				assert.ErrorIs(t, err, tt.err)
			}

			if tt.wantErr {
				return
			}

			if !reflect.DeepEqual(got.Secrets, tt.want.Secrets) {
				t.Errorf("DBPosgtre.GetItemByNameAndType() = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got.Additions, tt.want.Additions) {
				t.Errorf("DBPosgtre.GetItemByNameAndType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDBPosgtre_GetItemList(t *testing.T) {
	canceledCtx, cancel := context.WithCancel(context.Background())
	cancel()

	type args struct {
		ctx      context.Context
		username Username
	}
	tests := []struct {
		name    string
		args    args
		wantLen int
		wantErr bool
		err     error
	}{
		{
			name: "Get item list for existing user",
			args: args{
				ctx:      context.Background(),
				username: testUser1.Username,
			},
			wantLen: len(testItems),
			wantErr: false,
		},
		{
			name: "Get item list for existing user #2",
			args: args{
				ctx:      context.Background(),
				username: testUser2.Username,
			},
			wantLen: 0,
			wantErr: false,
		},
		{
			name: "Get item list for unexisting user",
			args: args{
				ctx:      context.Background(),
				username: testUser2.Username,
			},
			wantLen: 0,
			wantErr: false,
		},
		{
			name: "Canceled context",
			args: args{
				ctx:      canceledCtx,
				username: testUser2.Username,
			},
			wantErr: true,
			err:     ErrTransactionFailed,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := testDB.GetItemList(tt.args.ctx, tt.args.username)
			if (err != nil) != tt.wantErr {
				t.Errorf("DBPosgtre.GetItemList() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && !errors.Is(tt.err, assert.AnError) {
				assert.ErrorIs(t, err, tt.err)
			}

			if tt.wantErr {
				return
			}

			assert.Equal(t, tt.wantLen, len(got))
		})
	}
}

func TestDBPosgtre_UpdateItem(t *testing.T) {
	err := testDB.CreateItem(context.Background(), testUser2.Username, testItemEmptyNotesSecrets)
	if err != nil {
		t.Errorf("Failed to create test item: %v", err)
	}
	itemUsername := testUser2.Username

	newItem, err := testDB.GetItemByNameAndType(context.Background(), itemUsername,
		testItemEmptyNotesSecrets.Name, testItemEmptyNotesSecrets.Type)
	if err != nil {
		t.Errorf("Failed to get test item: %v", err)
	}

	newItemWithSecret := &pb.Item{
		Id:        newItem.Id,
		Name:      newItem.Name,
		Type:      newItem.Type,
		Reprompt:  newItem.Reprompt,
		Additions: newItem.Additions,
		Secrets: &pb.Secrets{
			Secret: []byte("secret"),
		},
	}

	newItemWithNotes := &pb.Item{
		Id:        newItem.Id,
		Name:      newItem.Name,
		Type:      newItem.Type,
		Reprompt:  newItem.Reprompt,
		Additions: newItem.Additions,
	}
	newItemWithNotes.Secrets = newItemWithSecret.Secrets
	newItemWithNotes.Secrets.Notes = []byte("new notes")

	unexistedItem := &pb.Item{
		Id:        999999,
		Name:      newItem.Name,
		Type:      newItem.Type,
		Reprompt:  newItem.Reprompt,
		Additions: newItem.Additions,
	}

	type args struct {
		ctx      context.Context
		username string
		item     *pb.Item
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		err     error
	}{
		{
			name: "Update item with same data",
			args: args{
				ctx:      context.Background(),
				username: itemUsername,
				item:     newItem,
			},
			wantErr: false,
			err:     ErrOperationFailed,
		},
		{
			name: "Update item's secret",
			args: args{
				ctx:      context.Background(),
				username: itemUsername,
				item:     newItemWithSecret,
			},
			wantErr: false,
		},
		{
			name: "Update item's notes'",
			args: args{
				ctx:      context.Background(),
				username: itemUsername,
				item:     newItemWithNotes,
			},
			wantErr: false,
		},
		{
			name: "Update unexisting item",
			args: args{
				ctx:      context.Background(),
				username: itemUsername,
				item:     unexistedItem,
			},
			wantErr: true,
			err:     ErrOperationFailed,
		},
		{
			name: "Update unexisting user's item",
			args: args{
				ctx:      context.Background(),
				username: "wrong username",
				item:     newItem,
			},
			wantErr: true,
			err:     ErrOperationFailed,
		},
		{
			name: "Missed username",
			args: args{
				ctx:  context.Background(),
				item: newItem,
			},
			wantErr: true,
			err:     ErrNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := testDB.UpdateItem(tt.args.ctx, tt.args.username, tt.args.item)
			if (err != nil) != tt.wantErr {
				t.Errorf("DBPosgtre.UpdateItem() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr && !errors.Is(tt.err, assert.AnError) {
				assert.ErrorIs(t, err, tt.err)
			}
		})
	}

	err = testDB.DeleteItem(context.Background(), itemUsername, newItem.Id)
	if err != nil {
		t.Errorf("Failed to delete test item: %v", err)
	}
}

func TestDBPosgtre_DeleteItem(t *testing.T) {
	err := testDB.CreateItem(context.Background(), testUser2.Username, testItemEmptyNotesSecrets)
	if err != nil {
		t.Errorf("Failed to create test item: %v", err)
	}
	itemUsername := testUser2.Username

	newItem, err := testDB.GetItemByNameAndType(context.Background(), itemUsername,
		testItemEmptyNotesSecrets.Name, testItemEmptyNotesSecrets.Type)
	if err != nil {
		t.Errorf("Failed to get test item: %v", err)
	}

	type args struct {
		ctx      context.Context
		username Username
		itemID   int64
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		err     error
	}{
		{
			name: "Wrong username",
			args: args{
				ctx:      context.Background(),
				username: "wrong username",
				itemID:   newItem.Id,
			},
			wantErr: true,
			err:     ErrOperationFailed,
		},
		{
			name: "Missed username",
			args: args{
				ctx:    context.Background(),
				itemID: newItem.Id,
			},
			wantErr: true,
			err:     ErrNotFound,
		},
		{
			name: "Delete item",
			args: args{
				ctx:      context.Background(),
				username: itemUsername,
				itemID:   newItem.Id,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := testDB.DeleteItem(tt.args.ctx, tt.args.username, tt.args.itemID)
			if (err != nil) != tt.wantErr {
				t.Errorf("DBPosgtre.DeleteItem() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && !errors.Is(tt.err, assert.AnError) {
				assert.ErrorIs(t, err, tt.err)
			}
		})
	}
}
