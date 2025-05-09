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

	if state.output[len(state.output)-1] == UTF8_LF {
		state.output = state.output[0 : len(state.output)-1]
	}

	return state.output, nil
}

func (state *EncodeState) marshal(source any, depth int) {
	refValSource := reflect.ValueOf(source)
	refTypeSource := reflect.TypeOf(source)

	for i := 0; i < refValSource.NumField(); i++ {
		state.encodeValue(
			depth,
			refValSource.Field(i),
			refTypeSource.Field(i).Type.Kind(),
			refTypeSource.Field(i),
		)
	}
}

func (state *EncodeState) encodeValue (depth int, refValue reflect.Value, refType reflect.Kind, structField reflect.StructField) {
	switch refType {
	case reflect.Struct:
		state.encodeStruct(depth, structField)
		state.marshal(refValue.Interface(), depth+1)
	case reflect.Int:
		state.encodeInt(depth, refValue, structField)
	case reflect.String:
		state.encodeString(depth, refValue, structField)
	case reflect.Float32:
		state.encodeFloat32(depth, refValue, structField)
	case reflect.Float64:
		state.encodeFloat64(depth, refValue, structField)
	case reflect.Bool:
		state.encodeBool(depth, refValue, structField)
	case reflect.Slice:
		state.encodeSlice(depth, refValue, structField)
	}
}

func appendDepth(data *[]byte, depth int) {
	for i := 0; i < depth; i++ {
		for i2 := 0; i2 < INDENT_SPACING; i2++ {
			*data = append(*data, UTF8_SPACE)
		}
	}
}

func (state *EncodeState) encodeStruct(depth int, structField reflect.StructField) {
	data := make([]byte, 0)

	appendDepth(&data, depth)
	data = append(data, []byte(structField.Tag.Get("yaml"))...)
	data = append(data, []byte{UTF8_COLON, UTF8_LF}...)

	state.output = append(state.output, data...)
}

func (state *EncodeState) encodeInt(depth int, value reflect.Value, structField reflect.StructField) {
	data := make([]byte, 0)

	appendDepth(&data, depth)
	if structField.Tag != "" {
		data = append(data, []byte(structField.Tag.Get("yaml"))...)
		data = append(data, []byte{UTF8_COLON, UTF8_SPACE}...)
	}
	data = append(data, []byte(strconv.FormatInt(value.Int(), 10))...)
	data = append(data, UTF8_LF)

	state.output = append(state.output, data...)
}

func (state *EncodeState) encodeString(depth int, value reflect.Value, structField reflect.StructField) {
	data := make([]byte, 0)

	appendDepth(&data, depth)
	if structField.Tag != "" {
		data = append(data, []byte(structField.Tag.Get("yaml"))...)
		data = append(data, []byte{UTF8_COLON, UTF8_SPACE}...)
	}
	data = append(data, []byte(value.String())...)
	data = append(data, UTF8_LF)

	state.output = append(state.output, data...)
}

func (state *EncodeState) encodeFloat32(depth int, value reflect.Value, structField reflect.StructField) {
	data := make([]byte, 0)

	appendDepth(&data, depth)
	if structField.Tag != "" {
		data = append(data, []byte(structField.Tag.Get("yaml"))...)
		data = append(data, []byte{UTF8_COLON, UTF8_SPACE}...)
	}
	data = append(data, []byte(strconv.FormatFloat(value.Float(), 'E', -1, 32))...)
	data = append(data, UTF8_LF)

	state.output = append(state.output, data...)
}

func (state *EncodeState) encodeFloat64(depth int, value reflect.Value, structField reflect.StructField) {
	data := make([]byte, 0)

	appendDepth(&data, depth)
	if structField.Tag != "" {
		data = append(data, []byte(structField.Tag.Get("yaml"))...)
		data = append(data, []byte{UTF8_COLON, UTF8_SPACE}...)
	}
	data = append(data, []byte(strconv.FormatFloat(value.Float(), 'E', -1, 64))...)
	data = append(data, UTF8_LF)

	state.output = append(state.output, data...)
}

func (state *EncodeState) encodeBool(depth int, value reflect.Value, structField reflect.StructField) {
	data := make([]byte, 0)

	appendDepth(&data, depth)
	if structField.Tag != "" {
		data = append(data, []byte(structField.Tag.Get("yaml"))...)
		data = append(data, []byte{UTF8_COLON, UTF8_SPACE}...)
	}
	if value.Bool() {
		data = append(data, []byte("true")...)
	} else {
		data = append(data, []byte("false")...)
	}
	data = append(data, UTF8_LF)

	state.output = append(state.output, data...)
}

func (state *EncodeState) encodeSlice(depth int, value reflect.Value, structField reflect.StructField) {
	if value.Cap() == 0 {
		return
	}
	state.encodeStruct(depth, structField)

	for i := range value.Len() {
		data := make([]byte, 0)
		appendDepth(&data, depth + 1)
		data = append(data, []byte{UTF8_DASH, UTF8_SPACE}...)

		if structField.Type.Elem().Kind() == reflect.Struct {
			for i2 := range value.Index(i).NumField() {
				if i2 == 0 {
					state.output = append(state.output, data...)
					state.encodeValue(
						0,
						value.Index(i).Field(i2),
					 	structField.Type.Elem().Field(i2).Type.Kind(),
					 	structField.Type.Elem().Field(i2),
					)
				} else {
					state.encodeValue(
						depth+2,
						value.Index(i).Field(i2),
					 	structField.Type.Elem().Field(i2).Type.Kind(),
					 	structField.Type.Elem().Field(i2),
					)
				}
			}
		} else {
			state.output = append(state.output, data...)
			state.encodeValue(
				0,
				value.Index(i),
				structField.Type.Elem().Kind(),
				reflect.StructField{ Tag: "" },
			)
		}
	}
}

