package crypt

import (
	"bytes"
	"crypto/rand"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncryptDecryptAES(t *testing.T) {
	key := make([]byte, AESKeyLength)
	rand.Read(key)

	messageE := []byte("")
	messageS := []byte("!asd 123 55")
	messageM := make([]byte, 1111)
	rand.Read(messageM)
	messageL := make([]byte, 5123456)
	rand.Read(messageL)

	tests := []struct {
		name    string
		key     []byte
		message []byte
	}{
		{
			name:    "Encrypt/decrypt message",
			key:     key,
			message: messageE,
		},
		{
			name:    "Encrypt/decrypt message",
			key:     key,
			message: messageS,
		},
		{
			name:    "Encrypt/decrypt message",
			key:     key,
			message: messageM,
		},
		{
			name:    "Encrypt/decrypt message",
			key:     key,
			message: messageL,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encr, err := EncryptAES(tt.key, tt.message)
			if err != nil {
				t.Errorf("EncryptDecryptAES() unexpected error = %v", err)
			}

			decr, err := DecryptAES(tt.key, encr)
			if err != nil {
				t.Errorf("EncryptDecryptAES() unexpected error = %v", err)
			}

			if !bytes.Equal(tt.message, decr) {
				t.Error("EncryptDecryptAES - Original and decrypted messages are not equal")
			}
		})
	}
}

func TestEncryptDecryptAESwithAD(t *testing.T) {
	messageM := make([]byte, 1111)
	rand.Read(messageM)

	tests := []struct {
		name    string
		key     []byte
		message []byte
	}{
		{
			name:    "Enrcypt/decrypt message with AD",
			key:     []byte("Passw@rdAsAkEY&8"),
			message: messageM,
		},
		{
			name:    "Encrypt/decrypt message with AD",
			key:     []byte("Passw@rdAsAkEY&8asd-11 a asd 123 asd 12 dasd 12 ----asd ;aksd;asdasdSADawwas;l"),
			message: messageM,
		},
		{
			name:    "Encrypt/decrypt message with AD",
			key:     []byte(""),
			message: messageM,
		},
		{
			name:    "Encrypt/decrypt message with AD",
			key:     nil,
			message: messageM,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encr, err := EncryptAESwithAD(tt.key, tt.message)
			if err != nil {
				t.Errorf("EncryptDecryptAESwithAD() unexpected error = %v", err)
			}

			decr, err := DecryptAESwithAD(tt.key, encr)
			if err != nil {
				t.Errorf("EncryptDecryptAESwithAD() unexpected error = %v", err)
			}

			if !bytes.Equal(tt.message, decr) {
				t.Error("EncryptDecryptAESwithAD - Original and decrypted messages are not equal")
			}
		})
	}
}

func TestEncryptAES(t *testing.T) {
	message := []byte("sample message")

	type args struct {
		key     []byte
		message []byte
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name: "Wrong key length",
			args: args{
				key:     []byte("123456789abcd"),
				message: message,
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := EncryptAES(tt.args.key, tt.args.message)
			if (err != nil) != tt.wantErr {
				t.Errorf("EncryptAES() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("EncryptAES() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDecryptAES(t *testing.T) {
	key := make([]byte, AESKeyLength)
	rand.Read(key)
	message, _ := EncryptAES(key, []byte("sample message"))

	type args struct {
		key       []byte
		encrypted []byte
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name: "Wrong key length",
			args: args{
				key:       []byte("123456789abcd"),
				encrypted: message,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "To short encrypted message",
			args: args{
				key:       key,
				encrypted: []byte("iuqweo123"),
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DecryptAES(tt.args.key, tt.args.encrypted)
			if (err != nil) != tt.wantErr {
				t.Errorf("DecryptAES() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DecryptAES() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDecryptAESwithAD(t *testing.T) {
	type args struct {
		key       []byte
		encrypted []byte
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name: "To short encrypted message",
			args: args{
				key:       []byte("123456789abcd"),
				encrypted: []byte("iuqweo123"),
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DecryptAESwithAD(tt.args.key, tt.args.encrypted)
			if (err != nil) != tt.wantErr {
				t.Errorf("DecryptAESwithAD() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DecryptAESwithAD() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGenerateRandomKey32(t *testing.T) {
	key := GenerateRandomKey32()
	assert.NotEmpty(t, key)
	assert.Equal(t, 32, len(key))
}
