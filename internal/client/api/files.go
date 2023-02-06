package api

import (
	"os"

	"gopkg.in/yaml.v3"
)

// UnmarshallYamlFromFile decodes data from YAML file to object.
func UnmarshallYamlFromFile(obj any, filename string) error {
	file, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(file, obj)
	if err != nil {
		return err
	}

	return nil
}
