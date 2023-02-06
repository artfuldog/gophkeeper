package authorizer

import (
	"reflect"
	"testing"
	"time"

	"github.com/o1egl/paseto"
	"github.com/stretchr/testify/assert"
)

func TestNewPasetoAuthorizer(t *testing.T) {
	type args struct {
		key           string
		tokenDuration time.Duration
	}
	tests := []struct {
		name    string
		args    args
		want    *PasetoAuthorizer
		wantErr bool
	}{
		{
			name: "Create authorizer",
			args: args{
				key:           "123456789a123456789a123456789a32",
				tokenDuration: 5 * time.Minute,
			},
			want: &PasetoAuthorizer{
				paseto:        paseto.NewV2(),
				key:           []byte("123456789a123456789a123456789a32"),
				tokenDuration: 5 * time.Minute,
			},
			wantErr: false,
		},
		{
			name: "Create authorizer with wrong key",
			args: args{
				key:           "123456789a1",
				tokenDuration: 5 * time.Minute,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewPasetoAuthorizer(tt.args.key, tt.args.tokenDuration)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewPasetoAuthorizer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewPasetoAuthorizer() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPasetoAuthorizer_CreateAndVerify(t *testing.T) {
	a, err := NewPasetoAuthorizer("123456789a123456789a123456789a32", 5*time.Minute)
	assert.NoError(t, err)

	fields := AuthFields{Username: "user123"}
	token, err := a.CreateToken(fields)
	assert.NoError(t, err)

	err = a.VerifyToken(token, fields)
	assert.NoError(t, err)

	err = a.VerifyToken(token, AuthFields{Username: "wrong_user"})
	assert.Error(t, err)

	a.key = []byte("asdasdasd")
	err = a.VerifyToken(token, fields)
	assert.Error(t, err)
}
