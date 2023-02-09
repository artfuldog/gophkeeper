//nolint:goconst
package api

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/artfuldog/gophkeeper/internal/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSecretEmpty(t *testing.T) {
	tests := []struct {
		name     string
		itemType string
		want     Secret
	}{
		{
			name:     "Login",
			itemType: common.ItemTypeLogin,
			want:     new(SecretLogin),
		},
		{
			name:     "Card",
			itemType: common.ItemTypeCard,
			want:     new(SecretCard),
		},
		{
			name:     "SecData",
			itemType: common.ItemTypeSecData,
			want:     new(SecretData),
		},
		{
			name:     "SecNote",
			itemType: common.ItemTypeSecNote,
			want:     nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewSecretEmpty(tt.itemType); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewSecretEmpty() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewSecret(t *testing.T) {
	login := TestingNewLoginItem().Secret
	card := TestingNewCardItem().Secret
	secdata := TestingNewSecDataItem().Secret

	type args struct {
		b        []byte
		itemType string
	}
	tests := []struct {
		name string
		args args
		want Secret
	}{
		{
			name: "Login",
			args: args{
				b:        login.ToBytes(),
				itemType: common.ItemTypeLogin,
			},
			want: login,
		},
		{
			name: "Card",
			args: args{
				b:        card.ToBytes(),
				itemType: common.ItemTypeCard,
			},
			want: card,
		},
		{
			name: "SecData",
			args: args{
				b:        secdata.ToBytes(),
				itemType: common.ItemTypeSecData,
			},
			want: secdata,
		},
		{
			name: "SecNote",
			args: args{
				b:        []byte(""),
				itemType: common.ItemTypeSecNote,
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewSecret(tt.args.b, tt.args.itemType)

			if tt.want != nil {
				assert.NotEmpty(t, got)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewSecret() = %v, want %v", got, tt.want)
			}
		})
	}
}

//nolint:cyclop
func TestSecrets_Bytes(t *testing.T) {
	secrets := map[string]Secret{
		"Login":   TestingNewLoginItem().Secret,
		"Card":    TestingNewCardItem().Secret,
		"SecData": TestingNewSecDataItem().Secret,
	}

	for name, secret := range secrets {
		t.Run(name+"_Safe", func(t *testing.T) {
			gotBytes := secret.ToBytes()

			var gotSecret Secret
			switch name {
			case "Login":
				gotSecret = NewSecretLogin(gotBytes)
			case "Card":
				gotSecret = NewSecretCard(gotBytes)
			case "SecData":
				gotSecret = NewSecretData(gotBytes)
			default:
				t.Errorf("unexprected test name")
			}

			require.NotEmpty(t, gotSecret)
			if !reflect.DeepEqual(secret, gotSecret) {
				t.Errorf("Response not equal - got:  %v, want %v", gotSecret, secret)
			}
		})
		t.Run(name+"_UnSafe", func(t *testing.T) {
			gotBytes, err := secret.ToBytesSafe()
			require.NoError(t, err)

			var gotSecret Secret
			switch name {
			case "Login":
				gotSecret, err = NewSecretLoginSafe(gotBytes)
			case "Card":
				gotSecret, err = NewSecretCardSafe(gotBytes)
			case "SecData":
				gotSecret, err = NewSecretDataSafe(gotBytes)
			default:
				t.Errorf("unexprected test name")
			}
			require.NoError(t, err)

			require.NotEmpty(t, gotSecret)
			if !reflect.DeepEqual(secret, gotSecret) {
				t.Errorf("Response not equal - got:  %v, want %v", gotSecret, secret)
			}
		})

		t.Run("Errors", func(t *testing.T) {
			gotBytes := []byte("isad asado asd1")

			var err error
			switch name {
			case "Login":
				_, err = NewSecretLoginSafe(gotBytes)
			case "Card":
				_, err = NewSecretCardSafe(gotBytes)
			case "SecData":
				_, err = NewSecretDataSafe(gotBytes)
			default:
				t.Errorf("unexprected test name")
			}
			require.Error(t, err)
		})
	}
}

func TestSecrets_Print(t *testing.T) {
	assert.NotEmpty(t, fmt.Sprint(TestingNewLoginItem().Secret))
	assert.NotEmpty(t, fmt.Sprint(TestingNewCardItem().Secret))
	assert.NotEmpty(t, fmt.Sprint(TestingNewSecDataItem().Secret))
}
