package main

import (
	"bytes"
	"errors"
	"math/rand"
	"reflect"
	"testing"
	"unicode/utf8"
)

type Object struct {
	Object struct {
		ObjectInt    int    `yaml:"objectInt"`
		ObjectString string `yaml:"objectString"`
	} `yaml:"object"`
	Int    int    `yaml:"int"`
	String string `yaml:"string"`
}

func FuzzUnmarshalStructured (f *testing.F) {
	f.Fuzz(func (t *testing.T, data []byte) {
		var fuzzObject Object

		data, err := populateFuzzObject(&fuzzObject, data)
		if err != nil {
			t.Skip()
		}

		yaml, err := Marshal(fuzzObject)
		if err != nil {
			t.Errorf("Failed to marshal fuzzObject: data: %s", data)
		}

		var newObject Object
		err = Unmarshal(yaml, &newObject)
		if err != nil {
			t.Errorf("Failed to marshal fuzzYaml: err: %v, yaml: %s", err, yaml)
		}

		newYaml, err := Marshal(newObject)
		if err != nil {
			t.Errorf("Failed to marshal newYaml: err: %v", err)
		}

		if !bytes.Equal(yaml, newYaml) {	
 			t.Errorf("yaml and newYaml doesn't match:\n")
 			t.Errorf("yaml:\n%s\n", string(yaml))
 			t.Errorf("newYaml:\n%s\n", string(newYaml))
			t.Errorf("object: %v\n", newObject)
			t.Errorf("validstring: %v\n", utf8.ValidString(string(yaml)))
			t.Errorf("validstring: %v\n", utf8.ValidString(string(newYaml)))
 			t.Errorf("yaml     (byte):\n%v\n", yaml)
 			t.Errorf("newYaml  (byte):\n%v\n", newYaml)
		}
	})
}

func populateFuzzObject (o any, data []byte) ([]byte, error) {
	v := reflect.ValueOf(o).Elem()
	t := reflect.TypeOf(o).Elem()

	for i := 0; i < v.NumField(); i++ {
		if len(data) < 2 {
			return data, errors.New("Not enough data")
		}

		switch t.Field(i).Type.Kind() {
			case reflect.Struct:
				populateFuzzObject(v.Field(i).Addr().Interface(), data)
			case reflect.Int:
				v.Field(i).SetInt(int64(data[0]))
				data = data[1:]
			case reflect.String:
				length := rand.Intn(len(data) - 1) + 1
				if length > len(data) {
					return data, errors.New("Not enough data")
				}

				inputBytes := data[0:length]
				validBytes := make([]byte, 0)
				for len(inputBytes) > 0 {
					rune, size := utf8.DecodeRune(inputBytes)
					if rune == utf8.RuneError && size == 1 {
						inputBytes = inputBytes[1:]
					} else {
						validBytes = append(validBytes, inputBytes[:size]...)
						inputBytes = inputBytes[size:]
					}
				}

				inputBytes = validBytes
				validBytes = make([]byte, 0)
				for _, byte := range inputBytes {
					if byte == UTF8_LF {
						continue
					}
					validBytes = append(validBytes, byte)
				}

				v.Field(i).SetString(string(validBytes))
				data = data[length:]
		}
	}

	return data, nil
}

// func OldFuzzUnmarshal (f *testing.F) {
// 	f.Add([]byte(`object:
//   objectInt: 10
//   objectString: pelle
// int: 1
// string: this is a string`))
//
// 	f.Fuzz(func (t *testing.T, data []byte) {
// 		t.Log("here??")
// 		var testObject Object
//
// 		err := Unmarshal(data, &testObject)
// 		if err != nil {
// 			return
// 		}
//
// 		encoded, err := Marshal(testObject)
// 		if err != nil {
// 			t.Fatalf("failed to marshal: %s", err)
// 		}
//
// 		err = Unmarshal(encoded, &testObject)
// 		if err != nil {
// 			t.Fatalf("failed to unmarshal: %s", err)
// 		}
//
// 		// Clean input and compare?
// 		if !bytes.Equal(data, encoded) {
// 			return
// 			t.Errorf("data and encoded doesn't match:\n")
// 			t.Errorf("data:\n%s\n", data)
// 			t.Errorf("encoded:\n%s\n", encoded)
// 			t.Errorf("data (byte):\n%v\n",data)
// 			t.Errorf("enc  (byte):\n%v\n", encoded)
// 		}
// 	})
// }
