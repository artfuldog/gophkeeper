package it

import (
	"github.com/artfuldog/gophkeeper/internal/client/api"
	"github.com/artfuldog/gophkeeper/internal/common"
)

//nolint:gosec
const (
	usr01Username  = "IntTestUser"
	usr01Password  = "Telekommunikationsuberwachungsverordnung"
	usr01SecretKey = "cisco-123"
)

func TestE2ENewLoginItem() *api.Item {
	return &api.Item{
		Name:     "seclogin1",
		Type:     common.ItemTypeLogin,
		Reprompt: true,
		Notes:    "May the force be with you",
		Secret: &api.SecretLogin{
			Username: "testuser",
			Password: "testpwd",
		},
		URIs: api.URIs{
			{
				URI:   "https://www.jornlande.com/",
				Match: "match",
			},
			{
				URI:   "https://avantasia.com/",
				Match: "match",
			},
		},
		CustomFields: api.CustomFields{
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

func TestE2ENewCardItem() *api.Item {
	return &api.Item{
		Name:     "seccard1",
		Type:     common.ItemTypeCard,
		Reprompt: false,
		Notes:    `Life is like a box of chocolates, you never know what you’re gonna get`,
		Secret: &api.SecretCard{
			ChName:   "Jopi Ivanov",
			Number:   "12390932124011",
			ExpMonth: 1,
			ExpYear:  30,
			Cvv:      731,
		},
		CustomFields: api.CustomFields{
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

func TestE2ENewSecNoteItem() *api.Item {
	return &api.Item{
		Name:     "secnote1",
		Type:     common.ItemTypeSecNote,
		Reprompt: false,
		Notes:    "Work hard and be nice to people",
		CustomFields: api.CustomFields{
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

func TestE2ENewSecDataItem() *api.Item {
	return &api.Item{
		Name:     "secdata1",
		Type:     common.ItemTypeSecData,
		Reprompt: false,
		Notes:    "some notes",
		Secret: &api.SecretData{
			Data: []byte("You become responsible, forever, for what you have tamed"),
		},
		CustomFields: api.CustomFields{
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
