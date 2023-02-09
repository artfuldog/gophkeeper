package db

import (
	"reflect"
	"testing"
	"time"

	"github.com/artfuldog/gophkeeper/internal/common"
	"github.com/artfuldog/gophkeeper/internal/pb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestUser_toPB(t *testing.T) {
	pbUser := &pb.User{
		Username: "testuser",
		Email:    common.PtrTo("email@com.com"),
		Pwdhash:  common.PtrTo("pwdhash"),
		OtpKey:   common.PtrTo("SDJAWMAASIDQ<:DFNCZASD"),
		Ekey:     []byte("ekey"),
		Revision: []byte("revision"),
		Updated:  timestamppb.New(time.Now()),
		Regdate:  timestamppb.New(time.Now()),
	}

	user := User{
		ID:       10,
		Username: pbUser.Username,
		Email:    pbUser.Email,
		Pwdhash:  pbUser.Pwdhash,
		OTPKey:   pbUser.OtpKey,
		Ekey:     pbUser.Ekey,
		Revision: pbUser.Revision,
		Updated:  pbUser.Updated.AsTime(),
		Regdate:  pbUser.Regdate.AsTime(),
	}

	gotPbUser := user.toPB()

	if !reflect.DeepEqual(pbUser, gotPbUser) {
		t.Errorf("Response not equal - got:  %v, want %v", gotPbUser, pbUser)
	}
}

func TestItemShort_toPB(t *testing.T) {
	pbItem := &pb.ItemShort{
		Name:    "itemname",
		Type:    common.ItemTypeLogin,
		Updated: timestamppb.New(time.Now()),
		Hash:    []byte("i.Hash"),
	}

	itemShort := ItemShort{
		Name:    pbItem.Name,
		Type:    pbItem.Type,
		Updated: pbItem.Updated.AsTime(),
		Hash:    pbItem.Hash,
	}

	gotPbItem := itemShort.toPB()

	if !reflect.DeepEqual(pbItem, gotPbItem) {
		t.Errorf("Response not equal - got:  %v, want %v", gotPbItem, pbItem)
	}
}
