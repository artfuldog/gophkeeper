package db

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/artfuldog/gophkeeper/internal/common"
	"github.com/artfuldog/gophkeeper/internal/mocks/mocklogger"
	"github.com/artfuldog/gophkeeper/internal/pb"
)

var testDB *DBPosgtre

var (
	testDBConnParams = DBParameters{
		address:       "localhost:5432/gophkeeper_db_tests",
		user:          "gksa",
		maxSecretSize: uint32(50 * 1024 * 1024),
	}

	testUser1 = &pb.User{
		Username: "testuser1",
		Email:    common.PtrTo("testuser1@example.com"),
		Pwdhash:  common.PtrTo("testusr1pwdhash"),
		OtpKey:   common.PtrTo("123df0123"),
		Ekey:     []byte("somekey"),
	}
	testUser2 = &pb.User{
		Username: "testuser2",
		Email:    common.PtrTo("testuser2@contoso.com"),
		Pwdhash:  common.PtrTo("testusr2swtorngpwdhash"),
		Ekey:     []byte("somekey"),
	}

	testItemLogin = &pb.Item{
		Name:     "testlogin1",
		Type:     common.ItemTypeLogin,
		Reprompt: common.PtrTo(true),
		Secrets: &pb.Secrets{
			Notes:  []byte("notes"),
			Secret: []byte("secret"),
		},
		Additions: &pb.Additions{
			Uris:         []byte("uris"),
			CustomFields: []byte("custom_fields"),
		},
	}

	testItemCard = &pb.Item{
		Name:     "testcard1",
		Type:     common.ItemTypeCard,
		Reprompt: common.PtrTo(false),
		Secrets: &pb.Secrets{
			Notes:  []byte("notes"),
			Secret: []byte("secret"),
		},
		Additions: &pb.Additions{
			CustomFields: []byte("custom_fields"),
		},
	}

	testItemData = &pb.Item{
		Name:     "testdata1",
		Type:     common.ItemTypeSecData,
		Reprompt: common.PtrTo(false),
		Secrets: &pb.Secrets{
			Notes:  []byte("notes"),
			Secret: []byte("secret"),
		},
		Additions: &pb.Additions{
			CustomFields: []byte("custom_fields"),
		},
	}

	testItemNotes = &pb.Item{
		Name:     "testnotes1",
		Type:     common.ItemTypeSecNote,
		Reprompt: common.PtrTo(false),
		Secrets: &pb.Secrets{
			Notes: []byte("notes"),
		},
		Additions: &pb.Additions{
			CustomFields: []byte("custom_fields"),
		},
	}

	testItemEmptyNotesSecrets = &pb.Item{
		Name:     "ItemEmptyNotesSecrets",
		Type:     common.ItemTypeLogin,
		Reprompt: common.PtrTo(false),
		Additions: &pb.Additions{
			CustomFields: []byte("custom_fields"),
		},
	}

	testItemEmptyAdditionsNotes = &pb.Item{
		Name:     "ItemEmptyAdditionsNotes",
		Type:     common.ItemTypeLogin,
		Reprompt: common.PtrTo(false),
		Secrets: &pb.Secrets{
			Secret: []byte("secret"),
		},
	}
)

// TestMain initializes testing database before test will start,
// clears all records and closes database's connections after all tests have finished
func TestMain(m *testing.M) {
	var err error
	ctx := context.Background()

	testLogger := mocklogger.NewMockLogger()

	if testDB, err = newDBPosgtre(&testDBConnParams, testLogger); err != nil {
		log.Printf("No testing database is available: %v\nSkikipping DB tests", err)
		os.Exit(0)
	}
	if err = testDB.Connect(ctx); err != nil {
		log.Printf("No testing database is available: %v\nSkikipping DB tests", err)
		os.Exit(0)
	}
	testDB.Clear(ctx)
	testDB.Setup(ctx)

	testUsers := []*pb.User{testUser1, testUser2}
	for i, user := range testUsers {
		testDB.CreateUser(ctx, user)
		newUser, _ := testDB.GetUserByName(ctx, user.Username)
		testUsers[i].Updated = newUser.Updated
		testUsers[i].Regdate = newUser.Regdate
	}

	testItems := []*pb.Item{testItemLogin, testItemCard, testItemData,
		testItemNotes, testItemEmptyAdditionsNotes, testItemEmptyNotesSecrets}

	for i, item := range testItems {
		testDB.CreateItem(ctx, testUser1.Username, item)
		newItem, _ := testDB.GetItemByNameAndType(ctx, testUser1.Username, item.Name, item.Type)
		testItems[i].Updated = newItem.Updated
	}

	for i, user := range testUsers {
		newUser, _ := testDB.GetUserByName(ctx, user.Username)
		testUsers[i].Revision = newUser.Revision
	}

	exitCode := m.Run()

	testDB.Clear(ctx)
	testDB.pool.Close()

	os.Exit(exitCode)
}
