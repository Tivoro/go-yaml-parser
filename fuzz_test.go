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
	IntReg  int     `yaml:"intReg"`
	IntOct  int     `yaml:"intOct"`
	IntHex  int     `yaml:"intHex"`
	Float32 float32 `yaml:"float32"`
	Float64 float64 `yaml:"float64"`
	True    bool    `yaml:"true"`
	False   bool    `yaml:"false"`
	StringObject struct {
		String            string `yaml:"string"`
		StringSingleQuote string `yaml:"stringSingleQuote"`
		StringDoubleQuote string `yaml:"stringDoubleQuote"`
	} `yaml:"stringObject"`
	ObjectArray []struct {
		String string `yaml:"string"`
		Int    int    `yaml:"int"`
	} `yaml:"objectArray"`
}

func FuzzUnmarshalStructured(f *testing.F) {
	f.Fuzz(func(t *testing.T, data []byte) {
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

func populateFuzzObject(o any, data []byte) ([]byte, error) {
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
			return getInt(data, i, v)
		case reflect.String:
			return getString(data, i, v)
		}
	}

	return data, nil
}

func getInt(data []byte, i int, v reflect.Value) ([]byte, error) {
	v.Field(i).SetInt(int64(data[0]))
	data = data[1:]
	return data, nil
}

func getString(data []byte, i int, v reflect.Value) ([]byte, error) {
	length := rand.Intn(len(data)-1) + 1
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
	return data, nil
}


