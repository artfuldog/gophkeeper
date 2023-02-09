//nolint:gochecknoglobals
package db

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/artfuldog/gophkeeper/internal/common"
	"github.com/artfuldog/gophkeeper/internal/mocks/mocklogger"
	"github.com/artfuldog/gophkeeper/internal/pb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Testing helpers vars and functions

var testDB *Posgtre

var (
	testDBConnParams = Parameters{
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

	testItems = []*pb.Item{testItemLogin, testItemCard, testItemData,
		testItemNotes, testItemEmptyAdditionsNotes, testItemEmptyNotesSecrets}
)

// TestMain initializes testing database before test will start,
// clears all records and closes database's connections after all tests have finished.
func TestMain(m *testing.M) {
	var err error
	ctx := context.Background()

	testLogger := mocklogger.NewMockLogger()

	if testDB, err = newPosgtre(&testDBConnParams, testLogger); err != nil {
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

	for i, item := range testItems {
		testDB.CreateItem(ctx, testUser1.Username, item)
		newItem, _ := testDB.GetItemByNameAndType(ctx, testUser1.Username, item.Name, item.Type)
		testItems[i].Updated = newItem.Updated
		testItems[i].Id = newItem.Id
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

// Tests.

func TestNewPostgre(t *testing.T) {
	logger := mocklogger.NewMockLogger()
	t.Run("No DB address", func(t *testing.T) {
		dbParams := NewParameters("", "", "", 10000000)
		_, err := newPosgtre(dbParams, logger)
		assert.Error(t, err)
	})
	t.Run("Empty username and password", func(t *testing.T) {
		dbParams := NewParameters("localhost", "", "", 10000000)
		db, err := newPosgtre(dbParams, logger)
		require.NoError(t, err)
		assert.NotEmpty(t, db)
	})

	t.Run("Empty password", func(t *testing.T) {
		dbParams := NewParameters("localhost", "user", "", 10000000)
		db, err := newPosgtre(dbParams, logger)
		require.NoError(t, err)
		assert.NotEmpty(t, db)
	})
	t.Run("New postgres DB", func(t *testing.T) {
		dbParams := NewParameters("localhost", "user", "password", 10000000)
		db, err := newPosgtre(dbParams, logger)
		require.NoError(t, err)
		assert.NotEmpty(t, db)
	})
}

func TestPostgre_Connect(t *testing.T) {
	logger := mocklogger.NewMockLogger()
	dbParams := NewParameters("wrong dsn", "", "", 10000000)
	db, err := newPosgtre(dbParams, logger)
	require.NoError(t, err)
	assert.NotEmpty(t, db)

	err = db.Connect(context.Background())
	assert.Error(t, err)
}

func TestPostgre_ConnectAndSetupRun(t *testing.T) {
	logger := mocklogger.NewMockLogger()
	db, err := newPosgtre(&testDBConnParams, logger)
	require.NoError(t, err)
	assert.NotEmpty(t, db)

	testCtx, cancel := context.WithCancel(context.Background())

	err = db.ConnectAndSetup(testCtx)
	require.NoError(t, err)

	ch := make(chan struct{})
	go db.Run(testCtx, ch)

	cancel()
	<-ch

	db.Clear(testCtx)
}

func TestPostgre_GetMaxSecretSize(t *testing.T) {
	logger := mocklogger.NewMockLogger()
	db, err := newPosgtre(&testDBConnParams, logger)
	require.NoError(t, err)
	assert.NotEmpty(t, db)

	assert.Equal(t, testDBConnParams.maxSecretSize, db.GetMaxSecretSize())
}
