package main

import (
	"fmt"
	"reflect"
	"strconv"
)

type TestStruct struct {
    Name string `yaml:"name"`
    Value int `yaml:"value"`
    Address struct {
        Street string `yaml:"street"`
    } `yaml:"address"`
}

type DataState struct {
    data     []byte
    offset   int
    minDepth int
}

const (
    UTF8_LF    byte = 10
    UTF8_SPACE byte = 32
    UTF8_COLON byte = 58
)

var data = []byte(`name: pelle
value: 10
address:
  street: pellegatan 1`)
var unmarshalledData TestStruct

func main() {
    Unmarshal(data, &unmarshalledData)
    fmt.Println(unmarshalledData)
}

func Unmarshal (data []byte, target any) {
    var dataState DataState
    dataState.init(data)
    dataState.unmarshal(target, 0)
}

func (dataState *DataState) init (data []byte) {
    // normalize
    var (
        depth      int  = 0
        minDepth   int  = 0
        depthFound bool = false
    )

    for i := 0; i < len(data); i++ {
        if data[i] == UTF8_LF {
            if depth > 0 && (minDepth == 0 || minDepth > depth) {
                minDepth = depth
            }
            depthFound = false
            depth = 0
        } else if !depthFound && data[i] == UTF8_SPACE {
            depth++
        } else if data[i] != UTF8_SPACE {
            depthFound = true
        }
    }

    if minDepth == 0 || minDepth > depth {
        minDepth = depth
    }

    dataState.data = data
    dataState.minDepth = minDepth
}

func (dataState *DataState) unmarshal (target any, depth int) {
    refValTarget := reflect.ValueOf(target).Elem()
    refTypeTarget := reflect.TypeOf(target).Elem()

    for d, k, v := dataState.readLine(); k != ""; d, k, v = dataState.readLine() {
        if depth > 0 && d < depth {
            fmt.Println("return")
            return
        }
        if depth != d {
            continue
        }

        for i := 0; i < refValTarget.NumField(); i++ {
            if k != refTypeTarget.Field(i).Tag.Get("yaml") {
                continue;
            }

            switch refTypeTarget.Field(i).Type.Kind() {
                case reflect.Struct:
                    dataState.unmarshal(refValTarget.Field(i).Addr().Interface(), depth + 1)
                case reflect.Int:
                    value, err := strconv.ParseInt(v, 10, 0)
                    if err != nil {
                        fmt.Printf("Unable to parse int: k: %v, v: %v, value: %v\n", k, v, value)
                        continue
                    }
                    refValTarget.Field(i).SetInt(value)
                case reflect.String:
                    refValTarget.Field(i).SetString(v)
                default:
                    fmt.Println("unhandled type", refTypeTarget.Field(i).Type.Kind(), refTypeTarget.Field(i).Tag.Get("yaml"))
            }
        }
    } 
}

func (dataState *DataState) readLine () (int, string, string) {
    var (
        depth      int    = 0
        key        string
        value      string
        depthFound bool   = false
        keyFound   bool   = false
    )

    for i := dataState.offset; i < len(dataState.data); i++ {
        dataState.offset++

        if !depthFound && dataState.data[i] == UTF8_SPACE {
            depth++
            continue
        } else if !depthFound {
            depthFound = true
        }

        if !keyFound && dataState.data[i] != UTF8_COLON {
            key += string(dataState.data[i])
            continue
        } else if dataState.data[i] == UTF8_COLON {
            keyFound = true
            if dataState.data[i + 1] == UTF8_SPACE {
                i++
                dataState.offset++
                continue
            }
        }

        if dataState.data[i] != UTF8_LF {
            value += string(dataState.data[i])
        } else if dataState.data[i] == UTF8_LF {
            break
        }
    }

    if depth % dataState.minDepth != 0 {
        fmt.Println("err! mindepth")
    }
    depth = depth / dataState.minDepth

    return depth, key, value
}

func valueIsZero (target any) {
    var value = reflect.ValueOf(target)
    fmt.Println(value)
    fmt.Println(value.Elem().Kind())
    field := value.Elem().FieldByName("Name")
    fmt.Println(reflect.ValueOf(field).IsZero())
}

