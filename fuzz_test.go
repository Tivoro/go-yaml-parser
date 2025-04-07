package main

import (
	"bytes"
	"testing"
)

type Object struct {
	Object struct {
		ObjectInt    int    `yaml:"objectInt"`
		ObjectString string `yaml:"objectString"`
	} `yaml:"object"`
	Int    int    `yaml:"int"`
	String string `yaml:"string"`
}

func FuzzUnmarshal (f *testing.F) {
	f.Add([]byte(`object:
  objectInt: 10
  objectString: pelle
int: 1
string: this is a string`))

	f.Fuzz(func (t *testing.T, data []byte) {
		var testObject Object

		err := Unmarshal(data, &testObject)
		if err != nil {
			return
		}

		encoded, err := Marshal(testObject)
		if err != nil {
			t.Fatalf("failed to marshal: %s", err)
		}

		err = Unmarshal(encoded, &testObject)
		if err != nil {
			t.Fatalf("failed to unmarshal: %s", err)
		}

		// Clean input and compare?
		if !bytes.Equal(data, encoded) {
			return
			t.Errorf("data and encoded doesn't match:\n")
			t.Errorf("data:\n%s\n", data)
			t.Errorf("encoded:\n%s\n", encoded)
			t.Errorf("data (byte):\n%v\n",data)
			t.Errorf("enc  (byte):\n%v\n", encoded)
		}
	})
}
