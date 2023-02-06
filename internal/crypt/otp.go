package crypt

import (
	"bytes"
	"fmt"
	"image/png"
	"time"

	"github.com/pquerna/otp/totp"
	"github.com/skip2/go-qrcode"
)

// Error detection/recovery capacity.
// According to RecoveryLevel in "github.com/skip2/go-qrcode"
const (
	QRRecoveryLow = iota
	QRRecoveryMid
	QRRecoveryHigh
	QRRecoveryHighest
)

// TOTPKey TOTP's key secret, url and QR Code (as []byte) for the user.
type TOTPKey struct {
	Secret string
	Url    string
	QRCode []byte
}

// GenerateTOTP generate TOTP key.
func GenerateTOTP(username string, issuer string, qrWidth int, qrHeight int) (*TOTPKey, error) {
	opts := totp.GenerateOpts{
		Issuer:      issuer,
		AccountName: username,
	}
	key, err := totp.Generate(opts)
	if err != nil {
		return nil, err
	}

	var imageBuffer bytes.Buffer
	image, err := key.Image(qrWidth, qrHeight)
	if err != nil {
		return nil, err
	}

	if err := png.Encode(&imageBuffer, image); err != nil {
		return nil, err
	}

	return &TOTPKey{
		Secret: key.Secret(),
		Url:    key.URL(),
		QRCode: imageBuffer.Bytes(),
	}, nil
}

// ValidateTOTP validates TOTP using the current time.
func ValidateTOTP(verificationCode string, secret string) bool {
	return totp.Validate(verificationCode, secret)
}

// GenerateVerificationCode generation user verification code.
func GenerateVerificationCode(secret string) (string, error) {
	return totp.GenerateCode(secret, time.Now())
}

// PrintQRCodeToTerminal prints QR code to terminal.
func PrintQRCodeToTerminal(url string, level int) {
	var qrcodeLevel qrcode.RecoveryLevel
	switch level {
	case QRRecoveryLow:
		qrcodeLevel = qrcode.Low
	case QRRecoveryHigh:
		qrcodeLevel = qrcode.High
	case QRRecoveryHighest:
		qrcodeLevel = qrcode.Highest
	default:
		qrcodeLevel = qrcode.Medium
	}

	qr, _ := qrcode.New(url, qrcodeLevel)

	blackBox := "\033[48;5;0m  \033[0m"
	whiteBox := "\033[48;5;7m  \033[0m"

	for ir, row := range qr.Bitmap() {
		lr := len(row)
		if ir == 0 || ir == 1 || ir == 2 ||
			ir == lr-1 || ir == lr-2 || ir == lr-3 {
			continue
		}
		for ic, col := range row {
			lc := len(qr.Bitmap())
			if ic == 0 || ic == 1 || ic == 2 ||
				ic == lc-1 || ic == lc-2 || ic == lc-3 {
				continue
			}
			if col {
				fmt.Print(blackBox)
			} else {
				fmt.Print(whiteBox)
			}
		}
		fmt.Println()
	}
}
