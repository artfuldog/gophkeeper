package crypt

import (
	"crypto/md5"
	"crypto/sha256"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetMD5hash(t *testing.T) {
	tests := []struct {
		name string
		data string
		hash string
	}{
		{
			name: "Get MD5 hash",
			data: "some string",
			hash: "5ac749fbeec93607fc28d666be85e73a",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			want := fmt.Sprintf("%x", md5.Sum([]byte(tt.data)))
			got := fmt.Sprintf("%x", GetMD5hash(tt.data))
			if want != got {
				t.Errorf("GetMD5hash() = %v, want %v", got, want)
			}
		})
	}
}

func TestGetSHA256hash(t *testing.T) {
	tests := []struct {
		name string
		data string
		hash string
	}{
		{
			name: "Get MD5 hash",
			data: "some string",
			hash: "61d034473102d7dac305902770471fd50f4c5b26f6831a56dd90b5184b3c30fc",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			want := fmt.Sprintf("%x", sha256.Sum256([]byte(tt.data)))
			got := fmt.Sprintf("%x", GetSHA256hash(tt.data))
			if want != got {
				t.Errorf("GetSHA256hash() = %v, want %v", got, want)
			}
		})
	}
}

func TestCalculateCheckPasswordHash(t *testing.T) {
	testpwd := "CPOAsud01n340p91982()!(23"
	hash, err := CalculatePasswordHash(testpwd)
	assert.NoError(t, err)

	wrongHash, err := CalculatePasswordHash("adsapo10)!@#map;sd-1")
	assert.NoError(t, err)

	assert.Equal(t, CheckPasswordHashStr(testpwd, hash), true)
	assert.Equal(t, CheckPasswordHashStr(testpwd, wrongHash), false)
}
