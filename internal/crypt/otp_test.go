package crypt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateValidateTOTP(t *testing.T) {
	t.Run("Check generate and validation", func(t *testing.T) {
		username := "SomeUsername12309"
		issuer := "seriousservice"

		totpkey, err := GenerateTOTP(username, issuer, 100, 100)
		assert.NoError(t, err)
		assert.NotEmpty(t, totpkey)

		code, _ := GenerateVerificationCode(totpkey.Secret)

		assert.Equal(t, ValidateTOTP(code, totpkey.Secret), true)
		assert.Equal(t, ValidateTOTP("123901", totpkey.Secret), false)
	})
	t.Run("Check errors", func(t *testing.T) {
		username := "SomeUsername12309"
		issuer := "seriousservice"

		_, errIssuer := GenerateTOTP(username, "", 100, 100)
		assert.Error(t, errIssuer)

		_, errImage := GenerateTOTP(username, issuer, 47, 47)
		assert.Error(t, errImage)

		_, errGenVerCode := GenerateVerificationCode("asd123")
		assert.Error(t, errGenVerCode)
	})
}

func TestPrintQRCodeToTerminal(t *testing.T) {
	PrintQRCodeToTerminal("http://1.com", QRRecoveryLow)
	PrintQRCodeToTerminal("http://1.com", QRRecoveryMid)
	PrintQRCodeToTerminal("http://1.com", QRRecoveryHigh)
	PrintQRCodeToTerminal("http://1.com", QRRecoveryHighest)
}
