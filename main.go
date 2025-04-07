package main

import (
	"fmt"
)

type TestStruct struct {
    Name string `yaml:"name"`
    Value int `yaml:"value"`
    Address struct {
        Street string `yaml:"street"`
    } `yaml:"address"`
}
type TestStruct2 struct {
	Object struct {
		ObjectInt    int    `yaml:"objectInt"`
		ObjectString string `yaml:"objectString"`
	} `yaml:"object"`
	Int    int    `yaml:"int"`
	String string `yaml:"string"`
}

var data = []byte(`name: pelle
value: 10
address:
  street: pellegatan 20`)
var data2 = []byte(`object2:
  notUsed: 1
object:
  objectInt: 10
  objectString: pelle
int: 11
string: this is a string`)
var unmarshalledData TestStruct
var unmarshalledData2 TestStruct2

func main() {
	//err := Unmarshal(data, &unmarshalledData)
	//if err != nil {
	//	fmt.Println("Err:", err)
	//}
    //fmt.Println(unmarshalledData)
	//data, err := Marshal(unmarshalledData)
	//if err != nil {
	//	fmt.Println("Err:", err)
	//}
	//fmt.Println(string(data))

	err := Unmarshal(data2, &unmarshalledData2)
	if err != nil {
		fmt.Println("Err:", err)
	}
	fmt.Println(unmarshalledData2)
	data, err = Marshal(unmarshalledData2)
	if err != nil {
		fmt.Println("Err:", err)
	}
	fmt.Println(string(data))
}

