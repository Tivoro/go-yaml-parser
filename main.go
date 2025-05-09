package main

import (
	"fmt"
)

type TestStruct struct {
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

var data = []byte(`
intReg: 10
intOct: 0o12
intHex: 0x0A
float32: 1230.15
float64: 1.23015e+3
true: true
false: false
stringObject:
  string: this is a string
  stringSingleQuote: 'this is a single quote string'
  stringDoubleQuote: "this is a double quote string"
objectArray:
  - string: this is a string
    int: 10
  - string: this is another string
    int: 20
`)
var unmarshalledData TestStruct

func main() {
	err := Unmarshal(data, &unmarshalledData)
	if err != nil {
		fmt.Println("Err:", err)
	}
	fmt.Println(unmarshalledData)
	data, err = Marshal(unmarshalledData)
	if err != nil {
		fmt.Println("Err:", err)
	}
	fmt.Println(string(data))
}
