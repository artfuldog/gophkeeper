package api

import (
	"errors"
	"fmt"
	"time"

	"github.com/artfuldog/gophkeeper/internal/common"
	"github.com/artfuldog/gophkeeper/internal/pb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gopkg.in/yaml.v3"
)

// Errors.
var (
	ErrWrongItemType = errors.New(`wrong item type`)
)

// Item represents secured dataitem.
type Item struct {
	ID           int64        `yaml:"id,omitempty"`
	Name         string       `yaml:"name"`
	Type         string       `yaml:"type"`
	Reprompt     bool         `yaml:"reprompt"`
	Updated      time.Time    `yaml:"updated,omitempty"`
	Hash         string       `yaml:"hash,omitempty"`
	Notes        string       `yaml:"notes,omitempty"`
	Secret       Secret       `yaml:"-"`
	URIs         URIs         `yaml:"uris,omitempty"`
	CustomFields CustomFields `yaml:"custom_fields,omitempty"`
}

// NewItemFromPB creates new Item based on protobuf format.
func NewItemFromPB(pbItem *pb.Item) *Item {
	return &Item{
		ID:           pbItem.Id,
		Name:         pbItem.Name,
		Type:         pbItem.Type,
		Reprompt:     pbItem.GetReprompt(),
		Updated:      pbItem.GetUpdated().AsTime(),
		Hash:         string(pbItem.Hash),
		Notes:        string(pbItem.Secrets.Notes),
		Secret:       NewSecret(pbItem.Secrets.Secret, pbItem.Type),
		URIs:         NewURIs(pbItem.Additions.Uris),
		CustomFields: NewCustomFields(pbItem.Additions.CustomFields),
	}
}

// toPB converts Item to protobuf format.
func (i Item) ToPB() *pb.Item {
	item := new(pb.Item)

	item.Id = i.ID
	item.Name = i.Name
	item.Type = i.Type
	item.Reprompt = &i.Reprompt
	item.Updated = timestamppb.New(i.Updated)

	secrets := new(pb.Secrets)
	secrets.Notes = []byte(i.Notes)

	if i.Secret != nil {
		secrets.Secret = i.Secret.ToBytes()
	}

	item.Secrets = secrets

	additions := new(pb.Additions)

	if item.Type == common.ItemTypeLogin {
		additions.Uris = i.URIs.ToBytes()
	}

	additions.CustomFields = i.CustomFields.ToBytes()
	item.Additions = additions

	return item
}

// NewItemFromBytes creates new Item from byte array.
//
// This is unsafe function, which means there is no errors checks.
// If you want this please use NewItemFromBytesSafe function.
//
//nolint:wsl
func NewItemFromBytes(b []byte) *Item {
	i := new(Item)
	serializeUnsafe(i, b)
	return i
}

// NewItemFromBytesSafe creates new Item from byte array.
//
// Unlike NewItemFromBytes this is safe function and in case of serialization' failure returns error.
func NewItemFromBytesSafe(b []byte) (*Item, error) {
	i := new(Item)
	if err := serializeSafe(i, b); err != nil {
		return nil, err
	}

	return i, nil
}

// ToBytes serializes Item structure to byte array.
//
// This is unsafe function, which means there is no errors checks.
// If you want this please use ToBytesSafe function.
func (i Item) ToBytes() []byte {
	return toBytesUnsafe(i)
}

// ToBytesSafe serializes Item structure to byte array.
//
// Unlike ToBytes this is safe function and in case of serialization' failure returns error.
func (i Item) ToBytesSafe() ([]byte, error) {
	return toBytesSafe(i)
}

// MarshalYAML implements yaml.Marshaler interface and used for encoding item to correct YAML format.
func (i Item) MarshalYAML() (interface{}, error) {
	type ItemAlias Item

	alias := &struct {
		ItemAlias `yaml:",inline"`
		Secret    Secret `yaml:"secret"`
	}{
		ItemAlias: ItemAlias(i),
		Secret:    i.Secret,
	}

	node := yaml.Node{}
	if err := node.Encode(alias); err != nil {
		return nil, err
	}

	return node, nil
}

// UnmarshalYAML implements yaml.Unmarshaler interface and used for decoding YAML to Item structure.
func (i *Item) UnmarshalYAML(value *yaml.Node) (err error) {
	type ItemAlias Item

	alias := &struct {
		*ItemAlias `yaml:",inline"`
		Secret     yaml.Node `yaml:"secret"`
	}{
		ItemAlias: (*ItemAlias)(i),
	}

	if err = value.Decode(alias); err != nil {
		return
	}

	if i.Type != common.ItemTypeSecNote {
		secret := NewSecretEmpty(i.Type)
		if err = alias.Secret.Decode(secret); err != nil {
			return
		}

		i.Secret = secret
	}

	return
}

// ToYaml returns Item in YAML format.
//
// This is unsafe function, which means there is no errors checks.
// If you want this please use ToYamlSafe function.
func (i Item) ToYaml() string {
	return marshalYamlUnsafe(i)
}

// ToYaml returns Item in YAML format.
//
// Unlike ToYaml this is safe function and in case of serialization' failure returns error.
func (i Item) ToYamlSafe() (string, error) {
	return marshalYamlSafe(i)
}

// GetLogin performs type assertion and return field Secret of Item as SecretLogin.
//
// This is unsafe function, which means there is no errors checks.
// If you want this please use GetLoginSafe function.
func (i Item) GetLogin() *SecretLogin {
	if i.Type != common.ItemTypeLogin || i.Secret == nil {
		return nil
	}

	return i.Secret.(*SecretLogin) //nolint:forcetypeassert //unsafe function
}

// GetLoginSafe performs type assertion and return field Secret of Item as SecretLogin.
//
// This is unsafe function, which means there is no errors checks.
// If you want this please use GetLogin function.
func (i Item) GetLoginSafe() (*SecretLogin, error) {
	if i.Type != common.ItemTypeLogin || i.Secret == nil {
		return nil, ErrWrongItemType
	}

	s, ok := i.Secret.(*SecretLogin)
	if !ok {
		return nil, ErrWrongItemType
	}

	return s, nil
}

// GetCard performs type assertion and return field Secret of Item as SecretCard.
//
// This is unsafe function, which means there is no errors checks.
// If you want this please use GetCardSafe function.
func (i Item) GetCard() *SecretCard {
	if i.Type != common.ItemTypeCard || i.Secret == nil {
		return nil
	}

	return i.Secret.(*SecretCard) //nolint:forcetypeassert //unsafe function
}

// GetCardSafe performs type assertion and return field Secret of Item as SecretCard.
//
// This is unsafe function, which means there is no errors checks.
// If you want this please use GetCard function.
func (i Item) GetCardSafe() (*SecretCard, error) {
	if i.Type != common.ItemTypeCard || i.Secret == nil {
		return nil, ErrWrongItemType
	}

	s, ok := i.Secret.(*SecretCard)
	if !ok {
		return nil, ErrWrongItemType
	}

	return s, nil
}

// GetSecData performs type assertion and return field Secret of Item as SecData.
//
// This is unsafe function, which means there is no errors checks.
// If you want this please use GetSecDataSafe function.
func (i Item) GetSecData() *SecretData {
	if i.Type != common.ItemTypeSecData || i.Secret == nil {
		return nil
	}

	return i.Secret.(*SecretData) //nolint:forcetypeassert //unsafe function
}

// GetSecDataSafe performs type assertion and return field Secret of Item as SecData.
//
// This is unsafe function, which means there is no errors checks.
// If you want this please use GetSecData function.
func (i Item) GetSecDataSafe() (*SecretData, error) {
	if i.Type != common.ItemTypeSecData || i.Secret == nil {
		return nil, ErrWrongItemType
	}

	s, ok := i.Secret.(*SecretData)
	if !ok {
		return nil, ErrWrongItemType
	}

	return s, nil
}

// Items represent slice of pointers to Items.
type Items []*Item

// CustomField represents database's custom field record.
type CustomField struct {
	Name      string `yaml:"name"`
	Type      string `yaml:"type"`
	ValueStr  string `yaml:"text"`
	ValueBool bool   `yaml:"flag"`
}

// CustomFields list of custom fields.
type CustomFields []CustomField

// ToBytes serializes CustomFields structure to byte array.
//
// This is unsafe function, which means there is no errors checks.
// If you want this please use ToBytesSafe function.
func (c *CustomFields) ToBytes() []byte {
	return toBytesUnsafe(c)
}

// ToBytesSafe serializes CustomFields structure to byte array.
//
// Unlike ToBytes this is safe function and in case of serialization' failure returns error.
func (c *CustomFields) ToBytesSafe() ([]byte, error) {
	return toBytesSafe(c)
}

// NewCustomFields serializes bytes into CustomFields structure.
//
// This is unsafe function, which means there is no errors checks.
// If you want this please use NewSecretLoginSafe function.
//
//nolint:wsl
func NewCustomFields(b []byte) CustomFields {
	var c CustomFields
	serializeUnsafe(&c, b)
	return c
}

// NewCustomFieldsSafe serializes bytes into CustomFields structure.
//
// Unlike NewCustomFields this is safe function and in case of serialization' failure returns error.
func NewCustomFieldsSafe(b []byte) (CustomFields, error) {
	var c CustomFields
	if err := serializeSafe(&c, b); err != nil {
		return nil, err
	}

	return c, nil
}

// String implements Stringer interface.
func (c CustomFields) String() string {
	output := "["
	for _, cf := range c {
		output += fmt.Sprint(cf)
	}

	return output + "]"
}

// URI represents database's URI record.
type URI struct {
	URI   string `yaml:"uri"`
	Match string `yaml:"match"`
}

// URIs represents list of database's URI record.
type URIs []URI

// NewURIs serializes bytes into URIs structure.
//
// This is unsafe function, which means there is no errors checks.
// If you want this please use NewSecretLoginSafe function.
//
//nolint:wsl
func NewURIs(b []byte) URIs {
	var u URIs
	serializeUnsafe(&u, b) //nolint:wsl
	return u
}

// NewURIsSafe serializes bytes into URIs structure.
//
// Unlike NewURIs this is safe function and in case of serialization' failure returns error.
func NewURIsSafe(b []byte) (URIs, error) {
	var u URIs
	if err := serializeSafe(&u, b); err != nil {
		return nil, err
	}

	return u, nil
}

// ToBytes serializes URI structure to byte array.
//
// This is unsafe function, which means there is no errors checks.
// If you want this please use ToBytesSafe function.
func (u *URIs) ToBytes() []byte {
	return toBytesUnsafe(u)
}

// ToBytesSafe serializes URI structure to byte array.
//
// Unlike ToBytes this is safe function and in case of serialization' failure returns error.
func (u *URIs) ToBytesSafe() ([]byte, error) {
	return toBytesSafe(u)
}

// String implements Stringer interface.
func (u URIs) String() string {
	output := "["
	for _, uri := range u {
		output += output + fmt.Sprint(uri)
	}

	return output + "]"
}
