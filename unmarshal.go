package main

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"reflect"
)

type DataState struct {
	data       []byte
	offset     int
	prevOffset int
	minDepth   int
	error      error
	sliceIndex int
}

func Unmarshal(data []byte, target any) error {
	var dataState DataState
	dataState.init(data)
	dataState.unmarshal(target, 0, UnmarshalOptions{})
	if dataState.error != nil {
		return dataState.error
	}
	return nil
}

func (dataState *DataState) init(data []byte) {
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

	dataState.data = bytes.Trim(data, "\n")
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
	dataState.sliceIndex = 0
}

type UnmarshalOptions struct {
	parseSlice       bool
	parseSliceStruct bool
}

func (dataState *DataState) unmarshal(target any, depth int, options UnmarshalOptions) {
	elemVal := reflect.ValueOf(target).Elem()
	elemType := reflect.TypeOf(target).Elem()
	sliceStartFound := false

	for d, k, v := dataState.readLine(); true; d, k, v = dataState.readLine() {
		// fmt.Printf("d: '%v', k: '%v', v: '%v'\n", d, k, v)
		if options.parseSlice {
			if !strings.HasPrefix(k, "- ") {
				dataState.offset = dataState.prevOffset
				break
			}
			v = strings.Trim(k, "- ")
			dataState.setValue(v, depth, elemVal.Index(dataState.sliceIndex), elemType.Elem())
			dataState.sliceIndex++
			continue
		}
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

		if options.parseSliceStruct {
			if sliceStartFound && strings.HasPrefix(k, "- ") {
				sliceStartFound = false
			}
			if !sliceStartFound && strings.HasPrefix(k, "- ") {
				elemVal = reflect.ValueOf(target).Elem().Index(dataState.sliceIndex)
				elemType = elemVal.Type()
				sliceStartFound = true
				dataState.sliceIndex++
			}
			k = strings.Trim(k, "- ")
		}

		for i := 0; i < elemVal.NumField(); i++ {
			fieldVal := elemVal.Field(i)
			fieldType := elemType.Field(i)
			if k != fieldType.Tag.Get("yaml") {
				continue
			}

			dataState.setValue(v, depth, fieldVal, fieldType.Type)
		}

		if dataState.offset == len(dataState.data) {
			return
		}
	}
}

func (dataState *DataState) setValue (v string, depth int, fieldVal reflect.Value, fieldType reflect.Type) {
	// fmt.Printf("v: '%v', depth: '%v', fieldVal: '%v', fieldType: '%v'\n", v, depth, fieldVal, fieldType)
	switch fieldType.Kind() {
	case reflect.Struct:
		dataState.unmarshal(fieldVal.Addr().Interface(), depth+1, UnmarshalOptions{})
	case reflect.String:
		setString(v, fieldVal)
	case reflect.Bool:
		setBool(v, fieldVal)
	case reflect.Int:
		setInt(v, fieldVal)
	case reflect.Float32:
		setFloat32(v, fieldVal)
	case reflect.Float64:
		setFloat64(v, fieldVal)
	case reflect.Slice:
		// Dynamically extend?
		fieldVal.Set(reflect.MakeSlice(fieldType, 8, 8))
		if fieldType.Elem().Kind() == reflect.Struct {
			dataState.unmarshal(fieldVal.Addr().Interface(), depth+2, UnmarshalOptions{ parseSliceStruct: true })
		} else {
			if len(v) > 0 && v[0] == UTF8_LEFT_SQUARE_BRACKET && v[len(v)-1] == UTF8_RIGHT_SQUARE_BRACKET {
				v = strings.Trim(v, "[]")
				tempDataState := DataState{
					data: make([]byte, 0),
					offset: 0, 
				}
				for _, v2 := range strings.Split(v, ",") {
					tempDataState.data = append(tempDataState.data, []byte("- " + strings.Trim(v2, " ") + "\n")...)
				}
				tempDataState.unmarshal(fieldVal.Addr().Interface(), depth, UnmarshalOptions{ parseSlice: true })
				dataState.sliceIndex = tempDataState.sliceIndex
			} else {
				dataState.unmarshal(fieldVal.Addr().Interface(), depth+2, UnmarshalOptions{ parseSlice: true })
			}
		}
		fieldVal.SetLen(dataState.sliceIndex)
		dataState.sliceIndex = 0
	default:
		fmt.Println("unhandled type", fieldType.Kind())
	}
} 

func (dataState *DataState) readLine() (int, string, string) {
	var (
		depth      int    = 0
		key        []byte = make([]byte, 0)
		value      []byte = make([]byte, 0)
		depthFound bool   = false
		keyFound   bool   = false
	)
	dataState.prevOffset = dataState.offset

	for i := dataState.offset; i < len(dataState.data); i++ {
		dataState.offset++
		if dataState.data[i] == UTF8_LF {
			break
		}

		if !depthFound && dataState.data[i] == UTF8_SPACE {
			depth++
			continue
		} else if !depthFound {
			depthFound = true
		}

		if !keyFound && dataState.data[i] != UTF8_COLON {
			key = append(key, dataState.data[i])
			continue
		} else if !keyFound && dataState.data[i] == UTF8_COLON {
			keyFound = true
			if len(dataState.data) > i+1 && dataState.data[i+1] == UTF8_SPACE {
				i++
				dataState.offset++
			}
			continue
		}

		if dataState.data[i] != UTF8_LF {
			value = append(value, dataState.data[i])
		} else if dataState.data[i] == UTF8_LF {
			break
		}
	}

	if depth > 0 && dataState.minDepth > 0 && depth%dataState.minDepth != 0 {
		dataState.error = errors.New("Inconsistent indentation")
		return 0, "", ""
	}
	if depth > 0 && dataState.minDepth > 0 {
		depth = depth / dataState.minDepth
	}

	if len(key) != 0 && key[0] == UTF8_DASH {
		depth++
	}

	return depth, string(key), string(bytes.Trim(value, "\"'"))
}

func setString (v string, field reflect.Value) { 
	field.SetString(v)
}

func setBool (v string, field reflect.Value) {
	if v == "true" || v == "1" {
		field.SetBool(true)
	} else if v == "false" || v == "0" {
		field.SetBool(false)
	} else {
		fmt.Printf("Unable to parse bool: v: %v\n", v)
	}
}

func setInt (v string, field reflect.Value) {
	value, err := strconv.ParseInt(v, 0, 0)
	if err != nil {
		fmt.Printf("Unable to parse int: v: %v\n", v)
	}
	field.SetInt(value)
}

func setFloat32 (v string, field reflect.Value) {
	value, err := strconv.ParseFloat(v, 32)
	if err != nil {
		fmt.Printf("Unable to parse float32: v: %v\n", v)
	}
	field.SetFloat(value)
}

func setFloat64 (v string, field reflect.Value) {
	value, err := strconv.ParseFloat(v, 64)
	if err != nil {
		fmt.Printf("Unable to parse float64: v: %v\n", v)
	}
	field.SetFloat(value)
}

