package api

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/artfuldog/gophkeeper/internal/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestNewItemFromPB(t *testing.T) {
	pbItem := TestingNewPbLoginItem()
	item := NewItemFromPB(pbItem)

	assert.NotEmpty(t, item)
	assert.Equal(t, item.Name, pbItem.Name)
}

func TestItem_ToPB(t *testing.T) {
	item := TestingNewLoginItem()
	pbItem := item.ToPB()

	assert.NotEmpty(t, item)
	assert.Equal(t, item.Name, pbItem.Name)
}

func TestItem_Bytes(t *testing.T) {
	items := map[string]*Item{
		"Login":   TestingNewLoginItem(),
		"Car":     TestingNewCardItem(),
		"SecData": TestingNewSecDataItem(),
		"SecNote": TestingNewSecNoteItem()}

	for name, item := range items {
		t.Run(name+"_Safe", func(t *testing.T) {
			itemBytes := item.ToBytes()
			gotItem := NewItemFromBytes(itemBytes)

			if !reflect.DeepEqual(item, gotItem) {
				t.Errorf("Response not equal - got:  %v, want %v", gotItem, item)
			}
		})
		t.Run(name+"_UnSafe", func(t *testing.T) {
			itemBytes, err := item.ToBytesSafe()
			require.NoError(t, err)
			gotItem, err := NewItemFromBytesSafe(itemBytes)
			require.NoError(t, err)

			if !reflect.DeepEqual(item, gotItem) {
				t.Errorf("Response not equal - got:  %v, want %v", gotItem, item)
			}
		})
	}

	t.Run("Errors", func(t *testing.T) {
		item := TestingNewLoginItem()
		item.Secret = NewTestWrongSecret()

		_, err := item.ToBytesSafe()
		assert.Error(t, err)

		_, err = NewItemFromBytesSafe([]byte("yipdasd12"))
		assert.Error(t, err)
	})
}

func TestItem_MarshallUnmarshallYAML(t *testing.T) {
	items := map[string]*Item{
		"Login":   TestingNewLoginItem(),
		"Car":     TestingNewCardItem(),
		"SecData": TestingNewSecDataItem(),
		"SecNote": TestingNewSecNoteItem()}

	for name, item := range items {
		t.Run(name, func(t *testing.T) {
			gotYaml := item.ToYaml()

			gotItem := new(Item)
			err := yaml.Unmarshal([]byte(gotYaml), &gotItem)
			require.NoError(t, err)
			if !reflect.DeepEqual(gotItem, item) {
				t.Errorf("Response not equal - got:  %v, want %v", gotItem, item)
			}

			gotYamlSafe, err := item.ToYamlSafe()
			require.NoError(t, err)

			err = yaml.Unmarshal([]byte(gotYamlSafe), &gotItem)
			require.NoError(t, err)
			if !reflect.DeepEqual(gotItem, item) {
				t.Errorf("Response not equal - got:  %v, want %v", gotItem, item)
			}
		})
	}
}

func TestItem_GetSecrets(t *testing.T) {
	t.Run("Login", func(t *testing.T) {
		item := TestingNewLoginItem()

		secret := item.GetLogin()
		if !reflect.DeepEqual(item.Secret, secret) {
			t.Errorf("Response not equal - got:  %v, want %v", secret, item.Secret)
		}

		secret, err := item.GetLoginSafe()
		require.NoError(t, err)
		if !reflect.DeepEqual(item.Secret, secret) {
			t.Errorf("Response not equal - got:  %v, want %v", secret, item.Secret)
		}
	})

	t.Run("Login Errors", func(t *testing.T) {
		item := TestingNewLoginItem()
		correctSecret := item.GetLogin()
		item.Type = common.ItemTypeCard

		secret := item.GetLogin()
		assert.Nil(t, secret)

		secret, err := item.GetLoginSafe()
		require.ErrorIs(t, err, ErrWrongItemType)
		assert.Nil(t, secret)

		item.Secret = *correctSecret
		item.Type = common.ItemTypeLogin
		secret, err = item.GetLoginSafe()
		require.ErrorIs(t, err, ErrWrongItemType)
		assert.Nil(t, secret)
	})

	t.Run("Card", func(t *testing.T) {
		item := TestingNewCardItem()

		secret := item.GetCard()
		if !reflect.DeepEqual(item.Secret, secret) {
			t.Errorf("Response not equal - got:  %v, want %v", secret, item.Secret)
		}

		secret, err := item.GetCardSafe()
		require.NoError(t, err)
		if !reflect.DeepEqual(item.Secret, secret) {
			t.Errorf("Response not equal - got:  %v, want %v", secret, item.Secret)
		}
	})

	t.Run("Card Errors", func(t *testing.T) {
		item := TestingNewCardItem()
		correctSecret := item.GetCard()
		item.Type = common.ItemTypeLogin

		secret := item.GetCard()
		assert.Nil(t, secret)

		secret, err := item.GetCardSafe()
		require.ErrorIs(t, err, ErrWrongItemType)
		assert.Nil(t, secret)

		item.Secret = *correctSecret
		item.Type = common.ItemTypeCard
		secret, err = item.GetCardSafe()
		require.ErrorIs(t, err, ErrWrongItemType)
		assert.Nil(t, secret)
	})

	t.Run("SecData", func(t *testing.T) {
		item := TestingNewSecDataItem()

		secret := item.GetSecData()
		if !reflect.DeepEqual(item.Secret, secret) {
			t.Errorf("Response not equal - got:  %v, want %v", secret, item.Secret)
		}

		secret, err := item.GetSecDataSafe()
		require.NoError(t, err)
		if !reflect.DeepEqual(item.Secret, secret) {
			t.Errorf("Response not equal - got:  %v, want %v", secret, item.Secret)
		}
	})

	t.Run("SecData Errors", func(t *testing.T) {
		item := TestingNewSecDataItem()
		correctSecret := item.GetSecData()
		item.Type = common.ItemTypeCard

		secret := item.GetSecData()
		assert.Nil(t, secret)

		secret, err := item.GetSecDataSafe()
		require.ErrorIs(t, err, ErrWrongItemType)
		assert.Nil(t, secret)

		item.Secret = *correctSecret
		item.Type = common.ItemTypeSecData
		secret, err = item.GetSecDataSafe()
		require.ErrorIs(t, err, ErrWrongItemType)
		assert.Nil(t, secret)
	})
}

func TestCustomFields_Bytes(t *testing.T) {
	t.Run("Safe", func(t *testing.T) {
		cf := TestingNewCustomFields()

		gotBytes := cf.ToBytes()
		gotCf := NewCustomFields(gotBytes)

		if !reflect.DeepEqual(cf, gotCf) {
			t.Errorf("Response not equal - got:  %v, want %v", gotCf, cf)
		}
	})
	t.Run("_UnSafe", func(t *testing.T) {
		cf := TestingNewCustomFields()

		gotBytes, err := cf.ToBytesSafe()
		require.NoError(t, err)
		gotCf, err := NewCustomFieldsSafe(gotBytes)
		require.NoError(t, err)

		if !reflect.DeepEqual(cf, gotCf) {
			t.Errorf("Response not equal - got:  %v, want %v", gotCf, cf)
		}
	})

	t.Run("Errors", func(t *testing.T) {
		_, err := NewCustomFieldsSafe([]byte("yipdasd12"))
		assert.Error(t, err)
	})
}

func TestURIs_Bytes(t *testing.T) {
	t.Run("Safe", func(t *testing.T) {
		uris := TestingNewURIs()

		gotBytes := uris.ToBytes()
		gotUris := NewURIs(gotBytes)

		if !reflect.DeepEqual(uris, gotUris) {
			t.Errorf("Response not equal - got:  %v, want %v", gotUris, uris)
		}
	})
	t.Run("_UnSafe", func(t *testing.T) {
		uris := TestingNewURIs()

		gotBytes, err := uris.ToBytesSafe()
		require.NoError(t, err)
		gotUris, err := NewURIsSafe(gotBytes)
		require.NoError(t, err)

		if !reflect.DeepEqual(uris, gotUris) {
			t.Errorf("Response not equal - got:  %v, want %v", gotUris, uris)
		}
	})

	t.Run("Errors", func(t *testing.T) {
		_, err := NewURIsSafe([]byte("yipdasd12"))
		assert.Error(t, err)
	})
}

func TestURIs_String(t *testing.T) {
	assert.NotEmpty(t, fmt.Sprint(TestingNewCustomFields()))
	assert.NotEmpty(t, fmt.Sprint(TestingNewURIs()))
}
