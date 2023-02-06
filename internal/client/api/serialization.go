package api

import (
	"bytes"
	"encoding/gob"

	"gopkg.in/yaml.v3"
)

// ToBytesUnsafe is a helper function to serilaize value to byte array.
//
// This is unsafe function, which means there is no errors checks.
// If you want this please use ToBytesSafe function.
func toBytesUnsafe(val any) []byte {
	buf := bytes.Buffer{}
	enc := gob.NewEncoder(&buf)
	enc.Encode(val)
	return buf.Bytes()
}

// ToBytesSafe is a helper function to serilaize value to byte array.
//
// Unlike ToBytesUnsafe this is safe function and in case of serialization' failure returns error.
func toBytesSafe(val any) ([]byte, error) {
	buf := bytes.Buffer{}
	enc := gob.NewEncoder(&buf)

	err := enc.Encode(val)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// marshalYamlUnsafe is a helper function to marshal data to YAML.
//
// This is unsafe function, which means there is no errors checks.
// If you want this please use marshalYamlSafe function.
func marshalYamlUnsafe(data any) string {
	bytes, _ := yaml.Marshal(data)
	return string(bytes)
}

// marshalYamlSafe is a helper function to marshal data to YAML.
//
// Unlike marshalYamlSafe this is safe function and in case of marshalling failure returns error.
func marshalYamlSafe(data any) (string, error) {
	var none string

	bytes, err := yaml.Marshal(data)
	if err != nil {
		return none, err
	}

	return string(bytes), nil
}

// serializeUnsafe is a helper function to serilaize byte array to data type.
//
// For correct serialization pointer to data should be passed.
// This is unsafe function, which means there is no errors checks.
// If you want this please use ToBytesSafe function.
func serializeUnsafe(data any, b []byte) {
	dec := gob.NewDecoder(bytes.NewReader(b))
	dec.Decode(data)
}

// serializeUnsafe is a helper function to serilaize byte array to data type.
//
// For correct serialization pointer to data should be passed.
// Unlike serializeUnsafe this is safe function and in case of serialization' failure returns error.
func serializeSafe(data any, b []byte) error {
	dec := gob.NewDecoder(bytes.NewReader(b))

	err := dec.Decode(data)
	if err != nil {
		return err
	}
	return nil
}

// GobRegister registers implemnations of Secret interfaces for serizalization.
func GobRegister() {
	gob.Register(&SecretLogin{})
	gob.Register(&SecretCard{})
	gob.Register(&SecretData{})
}
