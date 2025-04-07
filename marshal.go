package main

import (
	"reflect"
	"strconv"
)

type EncodeState struct {
	source any
	output []byte
	error  error
}

const INDENT_SPACING = 2

func Marshal(source any) ([]byte, error) {
	var state EncodeState = EncodeState{
		output: make([]byte, 0),
	}

	state.marshal(source, 0)
	if state.error != nil {
		return nil, state.error
	}

	if state.output[len(state.output) - 1] == UTF8_LF {
		state.output = state.output[0:len(state.output) - 1]
	}

	return state.output, nil
}

func (state *EncodeState) marshal(source any, depth int) {
	refValSource := reflect.ValueOf(source)
	refTypeSource := reflect.TypeOf(source)


	for i := 0; i < refValSource.NumField(); i++ {
		switch refTypeSource.Field(i).Type.Kind() {
			case reflect.Struct:
				state.encodeStruct(depth, refTypeSource.Field(i))
				state.marshal(refValSource.Field(i).Interface(), depth + 1)
			case reflect.Int:
				state.encodeInt(depth, refValSource.Field(i), refTypeSource.Field(i))
			case reflect.String:
				state.encodeString(depth, refValSource.Field(i), refTypeSource.Field(i))
		}
	}
}

func appendDepth (data *[]byte, depth int) {
	for i := 0; i < depth; i++ {
		for i2 := 0; i2 < INDENT_SPACING; i2++ {
			*data = append(*data, UTF8_SPACE)
		}
	}
}

func (state *EncodeState) encodeStruct (depth int, structField reflect.StructField) {
	data := make([]byte, 0)

	appendDepth(&data, depth)
	data = append(data, []byte(structField.Tag.Get("yaml"))...)
	data = append(data, []byte{UTF8_COLON, UTF8_LF}...)

	state.output = append(state.output, data...)
}

func (state *EncodeState) encodeInt (depth int, value reflect.Value, structField reflect.StructField) {
	data := make([]byte, 0)

	appendDepth(&data, depth)
	data = append(data, []byte(structField.Tag.Get("yaml"))...)
	data = append(data, []byte{UTF8_COLON, UTF8_SPACE}...)
	data = append(data, []byte(strconv.FormatInt(value.Int(), 10))...)
	data = append(data, UTF8_LF)

	state.output = append(state.output, data...)
}

func (state *EncodeState) encodeString (depth int, value reflect.Value, structField reflect.StructField) {
	data := make([]byte, 0)

	appendDepth(&data, depth)
	data = append(data, []byte(structField.Tag.Get("yaml"))...)
	data = append(data, []byte{UTF8_COLON, UTF8_SPACE}...)
	data = append(data, []byte(value.String())...)
	data = append(data, UTF8_LF)

	state.output = append(state.output, data...)
}
