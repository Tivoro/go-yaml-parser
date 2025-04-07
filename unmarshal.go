package main

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
)

type DataState struct {
    data       []byte
    offset     int
	prevOffset int
    minDepth   int
	error      error
}

func Unmarshal (data []byte, target any) error {
    var dataState DataState
    dataState.init(data)
	dataState.unmarshal(target, 0)
	if dataState.error != nil {
		return dataState.error
	}
	return nil
}

func (dataState *DataState) init (data []byte) {
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

    if minDepth == 0 || (minDepth > depth && depth > 0) {
        minDepth = depth
    }

    dataState.data = data
    dataState.minDepth = minDepth

	line := 0
	for _, k, _ := dataState.readLine(); k != ""; _, k, _ = dataState.readLine() {
		line++
		if dataState.error != nil {
			dataState.error = fmt.Errorf("%v, line: %v", dataState.error, line)
			return
		}
	}
	dataState.offset = 0
}

func (dataState *DataState) unmarshal (target any, depth int) {
    refValTarget := reflect.ValueOf(target).Elem()
    refTypeTarget := reflect.TypeOf(target).Elem()

    for d, k, v := dataState.readLine(); true; d, k, v = dataState.readLine() {
		if dataState.error != nil {
			return
		}
        if depth > 0 && d < depth {
			dataState.offset = dataState.prevOffset
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

		if dataState.offset == len(dataState.data) {
			return
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
	dataState.prevOffset = dataState.offset

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
            if len(dataState.data) > i + 1 && dataState.data[i + 1] == UTF8_SPACE {
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

    if depth > 0 && dataState.minDepth > 0 && depth % dataState.minDepth != 0 {
    	dataState.error = errors.New("Inconsistent indentation")
		return 0, "", ""
	}
	if depth > 0 && dataState.minDepth > 0 {
		depth = depth / dataState.minDepth
	}

    return depth, key, value
}
