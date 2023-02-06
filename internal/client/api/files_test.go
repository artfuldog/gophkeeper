package api

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMarshallYamlFromFile(t *testing.T) {
	type NestedObj struct {
		Textfield string
		Intfield  int
		Boolfield bool
	}

	type Obj struct {
		Textfield string
		Intfield  int
		Boolfield bool
		Nested    NestedObj
		List      []string
		Map       map[string]int
	}

	wantObj := Obj{
		Textfield: "text1",
		Intfield:  100500,
		Boolfield: true,
		Nested: NestedObj{
			Textfield: "text2",
			Intfield:  200500,
			Boolfield: true,
		},
		List: []string{"listfield1", "listfield2", "listfield3"},
		Map: map[string]int{
			"mapkey1": 200,
			"mapkey2": 300,
		},
	}

	obj := new(Obj)

	assert.Error(t, UnmarshallYamlFromFile(obj, "/wrong/file/path.yaml"))
	assert.Error(t, UnmarshallYamlFromFile(obj, "./test_data/invalid_unmarshallfile.yaml"))

	assert.NoError(t, UnmarshallYamlFromFile(obj, "./test_data/unmarshallfile.yaml"))
	if !reflect.DeepEqual(wantObj, *obj) {
		t.Errorf("Responce not equal - got:  %v, want %v", obj, wantObj)
	}
}
