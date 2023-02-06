package api

import (
	"fmt"

	c "github.com/artfuldog/gophkeeper/internal/common"
)

// Secret is general secret item inteface.
type Secret interface {
	// Serializes structure to byte array unsafe (without errors checks).
	ToBytes() []byte
	// Serializes structure to byte array safe.
	ToBytesSafe() ([]byte, error)
	// Returns structure in YAML format unsafe (without errors checks).
	String() string
}

// NewSecretEmpty is a fabric method for create empty Secret with provided type.
func NewSecretEmpty(itemType string) Secret {
	switch itemType {
	case c.ItemTypeLogin:
		return new(SecretLogin)
	case c.ItemTypeCard:
		return new(SecretCard)
	case c.ItemTypeSecData:
		return new(SecretData)
	default:
		return nil
	}
}

// NewSecret is a fabric method for create Secret with provided type and data.
func NewSecret(b []byte, itemType string) Secret {
	switch itemType {
	case c.ItemTypeLogin:
		return NewSecretLogin(b)
	case c.ItemTypeCard:
		return NewSecretCard(b)
	case c.ItemTypeSecData:
		return NewSecretData(b)
	default:
		return nil
	}
}

// SecretLogin represents secret for login item's type.
type SecretLogin struct {
	Username string `yaml:"username,omitempty"`
	Password string `yaml:"password,omitempty"`
	Authkey  string `yaml:"authkey,omitempty"`
}

var _ Secret = (*SecretLogin)(nil)

// NewSecretLogin serializes bytes into SecretLogin structure.
//
// This is unsafe function, which means there is no errors checks.
// If you want this please use NewSecretLoginSafe function.
func NewSecretLogin(b []byte) *SecretLogin {
	s := new(SecretLogin)
	serializeUnsafe(s, b)
	return s
}

// NewSecretLoginSafe serializes bytes into SecretLogin structure.
//
// Unlike NewSecretLogin this is safe function and in case of serialization' failure returns error.
func NewSecretLoginSafe(b []byte) (*SecretLogin, error) {
	s := new(SecretLogin)
	if err := serializeSafe(s, b); err != nil {
		return nil, err
	}
	return s, nil
}

// ToBytes serializes SecretLogin structure to byte array.
//
// This is unsafe function, which means there is no errors checks.
// If you want this please use ToBytesSafe function.
func (s SecretLogin) ToBytes() []byte {
	return toBytesUnsafe(s)
}

// ToBytesSafe serializes SecretLogin structure to byte array.
//
// Unlike ToBytes this is safe function and in case of serialization' failure returns error.
func (s SecretLogin) ToBytesSafe() ([]byte, error) {
	return toBytesSafe(s)
}

// String is used for printing text representation of SecretLogin.
func (s SecretLogin) String() string {
	return fmt.Sprintf("username: %s | password: %s | authkey: %s",
		s.Username, c.MaskLeft(s.Password, 2), c.MaskAll(8))
}

// SecretCard represents secret for login item's type.
type SecretCard struct {
	ChName   string `yaml:"cardholder,omitempty"`
	Number   string `yaml:"number,omitempty"`
	ExpMonth uint8  `yaml:"exp_month,omitempty"`
	ExpYear  uint8  `yaml:"exp_year,omitempty"`
	Cvv      uint16 `yaml:"cvv,omitempty"`
}

var _ Secret = (*SecretCard)(nil)

// NewSecretCard serializes bytes into SecretLogin structure.
//
// This is unsafe function, which means there is no errors checks.
// If you want this please use NewSecretLoginSafe function.
func NewSecretCard(b []byte) *SecretCard {
	s := new(SecretCard)
	serializeUnsafe(s, b)
	return s
}

// NewSecretCardSafe serializes bytes into SecretLogin structure.
//
// Unlike NewSecretCard this is safe function and in case of serialization' failure returns error.
func NewSecretCardSafe(b []byte) (*SecretCard, error) {
	s := new(SecretCard)
	if err := serializeSafe(s, b); err != nil {
		return nil, err
	}
	return s, nil
}

// ToBytes serializes SecretLogin structure to byte array.
//
// This is unsafe function, which means there is no errors checks.
// If you want this please use ToBytesSafe function.
func (s SecretCard) ToBytes() []byte {
	return toBytesUnsafe(s)
}

// ToBytesSafe serializes SecretCard structure to byte array.
//
// Unlike ToBytes this is safe function and in case of serialization' failure returns error.
func (s SecretCard) ToBytesSafe() ([]byte, error) {
	return toBytesSafe(s)
}

// String is used for printing text representation of SecretCard.
func (s SecretCard) String() string {
	return fmt.Sprintf("cardholder: %s | number: %s | exp_month: %s | exp_year: %s | CVV: %s",
		s.ChName, c.MaskLeft(s.Number, 4), c.MaskAll(2), c.MaskAll(2), c.MaskAll(3))
}

// SecretData represents secret for login item's type.
type SecretData struct {
	Data []byte `yaml:"data,omitempty"`
}

var _ Secret = (*SecretData)(nil)

// NewSecretData serializes bytes into SecretLogin structure.
//
// This is unsafe function, which means there is no errors checks.
// If you want this please use NewSecretLoginSafe function.
func NewSecretData(b []byte) *SecretData {
	s := new(SecretData)
	serializeUnsafe(s, b)
	return s
}

// NewSecretDataSafe serializes bytes into SecretLogin structure.
//
// Unlike NewSecretData this is safe function and in case of serialization' failure returns error.
func NewSecretDataSafe(b []byte) (*SecretData, error) {
	s := new(SecretData)
	if err := serializeSafe(s, b); err != nil {
		return nil, err
	}
	return s, nil
}

// ToBytes serializes SecretLogin structure to byte array.
//
// This is unsafe function, which means there is no errors checks.
// If you want this please use ToBytesSafe function.
func (s SecretData) ToBytes() []byte {
	return toBytesUnsafe(s)
}

// ToBytesSafe serializes SecretData structure to byte array.
//
// Unlike ToBytes this is safe function and in case of serialization' failure returns error.
func (s SecretData) ToBytesSafe() ([]byte, error) {
	return toBytesSafe(s)
}

// String is used for printing text representation of SecretData.
func (s SecretData) String() string {
	return "binary data"
}
