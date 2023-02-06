package db

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/artfuldog/gophkeeper/internal/common"
	"github.com/artfuldog/gophkeeper/internal/pb"
	"github.com/stretchr/testify/assert"
)

func TestDBPosgtre_CreateUser(t *testing.T) {
	newUser1 := &pb.User{
		Username: "newuser1",
		Pwdhash:  common.PtrTo("newuser1pwdhash"),
		Ekey:     []byte("somekey"),
		Email:    common.PtrTo("newemail@mail.com"),
		OtpKey:   common.PtrTo("OTPKEY"),
	}
	newUser2 := &pb.User{
		Username: "newuser2",
		Pwdhash:  common.PtrTo("newuser2pwdhash"),
		Ekey:     []byte("somekey"),
	}

	newUserMissedUsername := &pb.User{
		Pwdhash: common.PtrTo("newuser2pwdhash"),
		Ekey:    []byte("somekey"),
	}
	newUserMissedPwdHash := &pb.User{
		Username: "userMissedHash",
		Ekey:     []byte("somekey"),
	}
	newUserMissedEkey := &pb.User{
		Username: "userMissedHash",
		Pwdhash:  common.PtrTo("newuser2pwdhash"),
	}
	newUserWrongEmail := &pb.User{
		Username: "userWrongEmail",
		Pwdhash:  common.PtrTo("newuser1pwdhash"),
		Ekey:     []byte("somekey"),
		Email:    common.PtrTo("userWrongEmailmail.com"),
	}

	type args struct {
		ctx  context.Context
		user *pb.User
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		err     error
	}{
		{
			name: "Create new user",
			args: args{
				ctx:  context.Background(),
				user: newUser1,
			},
			wantErr: false,
		},
		{
			name: "Create new user #2",
			args: args{
				ctx:  context.Background(),
				user: newUser2,
			},
			wantErr: false,
		},
		{
			name: "Create new user with missed username",
			args: args{
				ctx:  context.Background(),
				user: newUserMissedUsername,
			},
			wantErr: true,
			err:     ErrConstraintViolation,
		},
		{
			name: "Create new user with missed password hash",
			args: args{
				ctx:  context.Background(),
				user: newUserMissedPwdHash,
			},
			wantErr: true,
			err:     ErrConstraintViolation,
		},
		{
			name: "Create new user with missed ecnryption key",
			args: args{
				ctx:  context.Background(),
				user: newUserMissedEkey,
			},
			wantErr: true,
			err:     ErrConstraintViolation,
		},
		{
			name: "Create new user with wrong email format",
			args: args{
				ctx:  context.Background(),
				user: newUserWrongEmail,
			},
			wantErr: true,
			err:     ErrConstraintViolation,
		},
		{
			name: "Create duplicate user",
			args: args{
				ctx:  context.Background(),
				user: newUser1,
			},
			wantErr: true,
			err:     ErrDuplicateEntry,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := testDB.CreateUser(tt.args.ctx, tt.args.user)
			if (err != nil) != tt.wantErr {
				t.Errorf("DBPosgtre.CreateUser() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && !errors.Is(tt.err, assert.AnError) {
				assert.ErrorIs(t, err, tt.err)
			}
		})
	}

	t.Run("Check creating user 1 was succesfull", func(t *testing.T) {
		newUser, _ := testDB.GetUserByName(context.Background(), newUser1.Username)

		assert.Equal(t, newUser.Email, newUser1.Email)
		assert.Equal(t, newUser.Pwdhash, newUser1.Pwdhash)
		assert.Equal(t, newUser.OtpKey, newUser1.OtpKey)

		if err := testDB.DeleteUserByName(context.Background(), newUser1.Username); err != nil {
			t.Errorf("DBPosgtre.CreateUser() - failed delete test user: %v", err)
		}
	})
	t.Run("Check creating user 2 was succesfull", func(t *testing.T) {
		newUser, _ := testDB.GetUserByName(context.Background(), newUser2.Username)

		assert.Equal(t, newUser.Email, newUser2.Email)
		assert.Equal(t, newUser.Pwdhash, newUser2.Pwdhash)
		assert.Equal(t, newUser.OtpKey, newUser2.OtpKey)

		if err := testDB.DeleteUserByName(context.Background(), newUser2.Username); err != nil {
			t.Errorf("DBPosgtre.CreateUser() - failed delete test user: %v", err)
		}
	})
}

func TestDBPosgtre_GetUserByName(t *testing.T) {
	canceledCtx, cancel := context.WithCancel(context.Background())
	cancel()

	type args struct {
		ctx      context.Context
		username string
	}
	tests := []struct {
		name    string
		args    args
		want    *pb.User
		wantErr bool
		err     error
	}{
		{
			name: "Get existing user",
			args: args{
				ctx:      context.Background(),
				username: testUser1.Username,
			},
			want:    testUser1,
			wantErr: false,
		},
		{
			name: "Get unexisting user",
			args: args{
				ctx:      context.Background(),
				username: "unexisting_user",
			},
			want:    nil,
			wantErr: true,
			err:     ErrNotFound,
		},
		{
			name: "Empty user name",
			args: args{
				ctx:      context.Background(),
				username: "",
			},
			want:    nil,
			wantErr: true,
			err:     ErrNotFound,
		},
		{
			name: "Canceled context",
			args: args{
				ctx:      canceledCtx,
				username: testUser1.Username,
			},
			want:    nil,
			wantErr: true,
			err:     ErrUndefinedError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := testDB.GetUserByName(tt.args.ctx, tt.args.username)
			if (err != nil) != tt.wantErr {
				t.Errorf("DBPosgtre.GetUserByName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && !errors.Is(tt.err, assert.AnError) {
				assert.ErrorIs(t, err, tt.err)
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DBPosgtre.GetUserByName() - \ngot = %v \nwant = %v", got, tt.want)
			}

		})
	}
}

func TestDBPosgtre_GetUserAuthData(t *testing.T) {
	canceledCtx, cancel := context.WithCancel(context.Background())
	cancel()

	type args struct {
		ctx      context.Context
		username string
	}
	tests := []struct {
		name    string
		args    args
		wantPwd Password
		wantKey OTPKey
		wantErr bool
		err     error
	}{
		{
			name: "Get existing user with hash and OTP",
			args: args{
				ctx:      context.Background(),
				username: testUser1.Username,
			},
			wantPwd: *testUser1.Pwdhash,
			wantKey: *testUser1.OtpKey,
			wantErr: false,
		},
		{
			name: "Get existing user only with Hash",
			args: args{
				ctx:      context.Background(),
				username: testUser2.Username,
			},
			wantPwd: *testUser2.Pwdhash,
			wantErr: false,
		},
		{
			name: "Get unexisting user",
			args: args{
				ctx:      context.Background(),
				username: "unexisting_user",
			},
			wantErr: true,
			err:     ErrNotFound,
		},
		{
			name: "Empty user name",
			args: args{
				ctx:      context.Background(),
				username: "",
			},
			wantErr: true,
			err:     ErrNotFound,
		},
		{
			name: "Canceled context",
			args: args{
				ctx:      canceledCtx,
				username: testUser1.Username,
			},
			wantErr: true,
			err:     ErrUndefinedError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPwd, gotKey, err := testDB.GetUserAuthData(tt.args.ctx, tt.args.username)
			if (err != nil) != tt.wantErr {
				t.Errorf("DBPosgtre.GetUserPwdHash() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && !errors.Is(tt.err, assert.AnError) {
				assert.ErrorIs(t, err, tt.err)
			}

			assert.Equal(t, gotPwd, tt.wantPwd)
			assert.Equal(t, gotKey, tt.wantKey)
		})
	}
}

func TestDBPosgtre_UpdateUser(t *testing.T) {
	newEmail := common.PtrTo("newemail@mail.com")
	newPwdHash := common.PtrTo("newupdatepwdhash")
	newOtpKey := common.PtrTo("neasda123")

	updateEmailUser := &pb.User{
		Username: testUser1.Username,
		Email:    newEmail,
	}
	updateWrongEmailUser := &pb.User{
		Username: testUser1.Username,
		Email:    common.PtrTo("newemailmail.com"),
	}

	updatePwdHashUser := &pb.User{
		Username: testUser1.Username,
		Pwdhash:  newPwdHash,
	}
	updateOTPUser := &pb.User{
		Username: testUser1.Username,
		OtpKey:   newOtpKey,
	}

	updateUnexistinglUser := &pb.User{
		Username: "unexisteduser",
		Pwdhash:  newPwdHash,
	}

	type args struct {
		ctx  context.Context
		user *pb.User
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		err     error
	}{
		{
			name: "Update existing user's email",
			args: args{
				ctx:  context.Background(),
				user: updateEmailUser,
			},
			wantErr: false,
		},
		{
			name: "Unsuccesfull Update existing user's email",
			args: args{
				ctx:  context.Background(),
				user: updateWrongEmailUser,
			},
			wantErr: true,
			err:     ErrConstraintViolation,
		},
		{
			name: "Update existing user's password hash",
			args: args{
				ctx:  context.Background(),
				user: updatePwdHashUser,
			},
			wantErr: false,
		},
		{
			name: "Update existing user's OTP secret key",
			args: args{
				ctx:  context.Background(),
				user: updateOTPUser,
			},
			wantErr: false,
		},
		{
			name: "Update unexisting user",
			args: args{
				ctx:  context.Background(),
				user: updateUnexistinglUser,
			},
			wantErr: true,
			err:     ErrNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := testDB.UpdateUser(tt.args.ctx, tt.args.user)
			if (err != nil) != tt.wantErr {
				t.Errorf("DBPosgtre.UpdateUser() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && !errors.Is(tt.err, assert.AnError) {
				assert.ErrorIs(t, err, tt.err)
			}
		})
	}

	t.Run("Check update was succesfull", func(t *testing.T) {
		updatedUser, _ := testDB.GetUserByName(context.Background(), testUser1.Username)

		assert.Equal(t, updatedUser.Email, newEmail)
		assert.Equal(t, updatedUser.Pwdhash, newPwdHash)
		assert.Equal(t, updatedUser.OtpKey, newOtpKey)
		assert.Equal(t, updatedUser.Regdate, testUser1.Regdate)
	})
}

func TestDBPosgtre_DeleteUserByName(t *testing.T) {
	newUserUsername := "newuser1fordelete"

	canceledCtx, cancel := context.WithCancel(context.Background())
	cancel()

	newUser1 := &pb.User{
		Username: newUserUsername,
		Pwdhash:  common.PtrTo("newuser1pwdhash"),
		Email:    common.PtrTo("newemailfordelete@mail.com"),
		Ekey:     []byte("somekey"),
	}

	if err := testDB.CreateUser(context.Background(), newUser1); err != nil {
		t.Errorf("DBPosgtre.UpdateUser() - failed create test user: %v", err)
	}

	type args struct {
		ctx      context.Context
		username Username
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		err     error
	}{
		{
			name: "Delete existing user with canceled context",
			args: args{
				ctx:      canceledCtx,
				username: newUserUsername,
			},
			wantErr: true,
			err:     ErrUndefinedError,
		},
		{
			name: "Delete existing user",
			args: args{
				ctx:      context.Background(),
				username: newUserUsername,
			},
			wantErr: false,
		},
		{
			name: "Delete unexisting user",
			args: args{
				ctx:      context.Background(),
				username: "unexisted_user",
			},
			wantErr: true,
			err:     ErrNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := testDB.DeleteUserByName(tt.args.ctx, tt.args.username)
			if (err != nil) != tt.wantErr {
				t.Errorf("DBPosgtre.UpdateUser() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && !errors.Is(tt.err, assert.AnError) {
				assert.ErrorIs(t, err, tt.err)
			}
		})
	}

	t.Run("Check delete was succesfull", func(t *testing.T) {
		_, err := testDB.GetUserByName(context.Background(), newUserUsername)

		if !errors.Is(err, ErrNotFound) {
			t.Errorf("DBPosgtre.DeleteUserByName() error = %v, wantErr %v", err, ErrNotFound)
		}
	})
}
