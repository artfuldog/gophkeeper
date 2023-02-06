package api

import (
	"os"
	"testing"
	"time"

	"github.com/artfuldog/gophkeeper/internal/common"
	"github.com/artfuldog/gophkeeper/internal/pb"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestMain(m *testing.M) {
	GobRegister()

	exitCode := m.Run()

	os.Exit(exitCode)
}

type TestWrongSecret struct {
}

var _ Secret = (*TestWrongSecret)(nil)

func NewTestWrongSecret() *SecretLogin {
	return nil
}
func NewTestWrongSecretSafe() (*SecretLogin, error) {
	return nil, assert.AnError
}
func (s TestWrongSecret) ToBytes() []byte {
	return nil
}
func (s TestWrongSecret) ToBytesSafe() ([]byte, error) {
	return nil, assert.AnError
}
func (s TestWrongSecret) ToYaml() string {
	return ""
}
func (s TestWrongSecret) ToYamlSafe() (string, error) {
	return "", assert.AnError
}
func (s TestWrongSecret) String() string {
	return "test secret"
}

func TestingNewLoginItem() *Item {
	return &Item{
		Name:     "seclogin1",
		Type:     common.ItemTypeLogin,
		Reprompt: true,
		Notes:    "May the force be with you",
		Secret: &SecretLogin{
			Username: "testuser",
			Password: "testpwd",
		},
		URIs: URIs{
			{
				URI:   "https://www.jornlande.com/",
				Match: "match",
			},
			{
				URI:   "https://avantasia.com/",
				Match: "match",
			},
		},
		CustomFields: CustomFields{
			{
				Name:     "loginCF1",
				Type:     common.CfTypeText,
				ValueStr: "val login CF1",
			},
			{
				Name:     "loginCF2",
				Type:     common.CfTypeHidden,
				ValueStr: "val login CF2",
			},
			{
				Name:      "loginCF3",
				Type:      common.CfTypeBool,
				ValueBool: true,
			},
		},
	}
}

func TestingNewPbLoginItem() *pb.Item {
	return &pb.Item{
		Name:     "seclogin1",
		Type:     common.ItemTypeLogin,
		Reprompt: common.PtrTo(true),
		Updated:  timestamppb.New(time.Now()),
		Hash:     []byte("some hash"),
		Secrets: &pb.Secrets{
			Notes:  []byte("notes"),
			Secret: []byte("secret"),
		},
		Additions: &pb.Additions{
			Uris:         []byte("uris"),
			CustomFields: []byte("custom_fields"),
		},
	}
}

func TestingNewCardItem() *Item {
	return &Item{
		Name:     "seccard1",
		Type:     common.ItemTypeCard,
		Reprompt: false,
		Notes:    `Life is like a box of chocolates, you never know what you’re gonna get`,
		Secret: &SecretCard{
			ChName:   "Jopi Ivanov",
			Number:   "12390932124011",
			ExpMonth: 1,
			ExpYear:  30,
			Cvv:      731,
		},
		CustomFields: CustomFields{
			{
				Name:     "cardCF1",
				Type:     common.CfTypeText,
				ValueStr: "val card CF1",
			},
			{
				Name:     "cardCF2",
				Type:     common.CfTypeHidden,
				ValueStr: "val card CF2",
			},
			{
				Name:      "cardCF3",
				Type:      common.CfTypeBool,
				ValueBool: true,
			},
		},
	}
}

func TestingNewSecNoteItem() *Item {
	return &Item{
		Name:     "secnote1",
		Type:     common.ItemTypeSecNote,
		Reprompt: false,
		Notes:    "Work hard and be nice to people",
		CustomFields: CustomFields{
			{
				Name:     "noteCF1",
				Type:     common.CfTypeText,
				ValueStr: "val note CF1",
			},
			{
				Name:     "noteCF2",
				Type:     common.CfTypeHidden,
				ValueStr: "val note CF2",
			},
			{
				Name:      "noteCF3",
				Type:      common.CfTypeBool,
				ValueBool: true,
			},
		},
	}
}

func TestingNewPbSecNoteItem() *pb.Item {
	return &pb.Item{
		Name:     "secnote1",
		Type:     common.ItemTypeSecNote,
		Reprompt: common.PtrTo(true),
		Updated:  timestamppb.New(time.Now()),
		Hash:     []byte("some hash"),
		Secrets: &pb.Secrets{
			Notes: []byte("Work hard and be nice to people"),
		},
	}
}

func TestingNewSecDataItem() *Item {
	return &Item{
		Name:     "secnote1",
		Type:     common.ItemTypeSecData,
		Reprompt: false,
		Notes:    "some notes",
		Secret: &SecretData{
			Data: []byte("You become responsible, forever, for what you have tamed"),
		},
		CustomFields: CustomFields{
			{
				Name:     "dataCF1",
				Type:     common.CfTypeText,
				ValueStr: "val data CF1",
			},
			{
				Name:     "noteCF2",
				Type:     common.CfTypeHidden,
				ValueStr: "val data CF2",
			},
			{
				Name:      "dataCF3",
				Type:      common.CfTypeBool,
				ValueBool: true,
			},
		},
	}
}

func TestingNewCustomFields() CustomFields {
	return []CustomField{
		{
			Name:     "custom field test",
			Type:     common.CfTypeText,
			ValueStr: "Simplicity is the ultimate sophistication",
		},
		{
			Name:     "custom field hidden",
			Type:     common.CfTypeHidden,
			ValueStr: "Chop your own wood and it will warm you twice.",
		},
		{
			Name:      "custom field bool",
			Type:      common.CfTypeBool,
			ValueBool: true,
		},
	}
}

func TestingNewURIs() URIs {
	return []URI{
		{
			URI:   "https://sadtrombone.com/",
			Match: "",
		},
		{
			URI:   "https://akinator.com",
			Match: "",
		},
	}
}
