package db

import (
	"time"

	"github.com/artfuldog/gophkeeper/internal/pb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// TODO find way to deal with protobuf's timestamppb

// User represents database's user record (raw from users table).
type User struct {
	ID       int       `db:"id"`
	Username string    `db:"username"`
	Email    *string   `db:"email"`
	Pwdhash  *string   `db:"pwdhash"`
	OTPKey   *string   `db:"otpkey"`
	Ekey     []byte    `db:"ekey"`
	Revision []byte    `db:"revision"`
	Updated  time.Time `db:"updated"`
	Regdate  time.Time `db:"regdate"`
}

// toPB converts User to protobuf format.
func (u User) toPB() *pb.User {
	return &pb.User{
		Username: u.Username,
		Email:    u.Email,
		Pwdhash:  u.Pwdhash,
		OtpKey:   u.OTPKey,
		Ekey:     u.Ekey,
		Revision: u.Revision,
		Updated:  timestamppb.New(u.Updated),
		Regdate:  timestamppb.New(u.Regdate),
	}
}

// ItemShort represents short message information from database.
type ItemShort struct {
	ID      int64     `db:"id"`
	Name    string    `db:"name"`
	Type    string    `db:"type"`
	Updated time.Time `db:"updated"`
	Hash    []byte    `db:"hash"`
}

// toPB converts ItemShort to protobuf format.
func (i ItemShort) toPB() *pb.ItemShort {
	return &pb.ItemShort{
		Id:      i.ID,
		Name:    i.Name,
		Type:    i.Type,
		Updated: timestamppb.New(i.Updated),
		Hash:    i.Hash,
	}
}
