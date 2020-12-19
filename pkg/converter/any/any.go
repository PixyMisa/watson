// Package any provides between `vm.Value`s and built-in types.
package any

import (
	"fmt"
	"reflect"

	"github.com/genkami/watson/pkg/vm"
)

// FromValue converts vm.Value into one of the following type:
// * int64
// * uint64
// * string
// * bool
// * (interface{})(nil)
// * map[string]interface{} (the value of which is also one or many of these types)
func FromValue(val *vm.Value) interface{} {
	switch val.Kind {
	case vm.KInt:
		return val.Int
	case vm.KUint:
		return val.Uint
	case vm.KFloat:
		return val.Float
	case vm.KString:
		return string(val.String)
	case vm.KObject:
		obj := map[string]interface{}{}
		for k, v := range val.Object {
			obj[k] = FromValue(v)
		}
		return obj
	case vm.KArray:
		arr := make([]interface{}, 0, len(val.Array))
		for _, v := range val.Array {
			arr = append(arr, FromValue(v))
		}
		return arr
	case vm.KBool:
		return val.Bool
	case vm.KNil:
		return nil
	default:
		panic(fmt.Errorf("invalid kind: %d", val.Kind))
	}
}

func ToValue(v interface{}) *vm.Value {
	if v == nil {
		return vm.NewNilValue()
	}
	switch v := v.(type) {
	case bool:
		return vm.NewBoolValue(v)
	case int:
		return vm.NewIntValue(int64(v))
	case int8:
		return vm.NewIntValue(int64(v))
	case int16:
		return vm.NewIntValue(int64(v))
	case int32:
		return vm.NewIntValue(int64(v))
	case int64:
		return vm.NewIntValue(v)
	case uint:
		return vm.NewUintValue(uint64(v))
	case uint8:
		return vm.NewUintValue(uint64(v))
	case uint16:
		return vm.NewUintValue(uint64(v))
	case uint32:
		return vm.NewUintValue(uint64(v))
	case uint64:
		return vm.NewUintValue(uint64(v))
	case []byte:
		w := make([]byte, len(v))
		copy(w, v)
		return vm.NewStringValue(w)
	case string:
		return vm.NewStringValue([]byte(v))
	case float32:
		return vm.NewFloatValue(float64(v))
	case float64:
		return vm.NewFloatValue(v)
	}
	vv := reflect.ValueOf(v)
	return reflectValueToValue(vv)
}

func reflectValueToValue(v reflect.Value) *vm.Value {
	if isIntFamily(v) {
		return reflectIntToValue(v)
	} else if isUintFamily(v) {
		return reflectUintToValue(v)
	} else if isFloatFamily(v) {
		return reflectFloatToValue(v)
	} else if isBool(v) {
		return reflectBoolToValue(v)
	} else if isString(v) {
		return reflectStringToValue(v)
	} else if v.IsNil() {
		// Marshalers should be placed before nil so as to handle `MarshalWatson` correctly.
		return vm.NewNilValue()
		// Maps, slices, and structs should be placed after nil so as to convert nil into Nil correctly.
	} else if isBytes(v) {
		return reflectBytesToValue(v)
	} else if isMapConvertibleToValue(v) {
		return reflectMapToValue(v)
	}

	panic(fmt.Errorf("can't convert %s to *vm.Value", v.Type().String()))
}

func reflectIntToValue(v reflect.Value) *vm.Value {
	return vm.NewIntValue(v.Int())
}

func reflectUintToValue(v reflect.Value) *vm.Value {
	return vm.NewUintValue(v.Uint())
}

func reflectFloatToValue(v reflect.Value) *vm.Value {
	return vm.NewFloatValue(v.Float())
}

func reflectBoolToValue(v reflect.Value) *vm.Value {
	return vm.NewBoolValue(v.Bool())
}

func reflectStringToValue(v reflect.Value) *vm.Value {
	return vm.NewStringValue([]byte(v.String()))
}

func reflectBytesToValue(v reflect.Value) *vm.Value {
	bytes := v.Bytes()
	clone := make([]byte, len(bytes))
	copy(clone, bytes)
	return vm.NewStringValue(clone)
}

func reflectMapToValue(v reflect.Value) *vm.Value {
	obj := map[string]*vm.Value{}
	iter := v.MapRange()
	for iter.Next() {
		k := iter.Key().String()
		v := iter.Value()
		obj[k] = reflectValueToValue(v)
	}
	return vm.NewObjectValue(obj)
}

func isIntFamily(v reflect.Value) bool {
	switch v.Type().Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return true
	default:
		return false
	}
}

func isUintFamily(v reflect.Value) bool {
	switch v.Type().Kind() {
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return true
	default:
		return false
	}
}

func isFloatFamily(v reflect.Value) bool {
	switch v.Type().Kind() {
	case reflect.Float32, reflect.Float64:
		return true
	default:
		return false
	}
}

func isBool(v reflect.Value) bool {
	return v.Type().Kind() == reflect.Bool
}

func isString(v reflect.Value) bool {
	return v.Type().Kind() == reflect.String
}

func isBytes(v reflect.Value) bool {
	return v.Type().Kind() == reflect.Slice && v.Type().Elem() == reflect.TypeOf(byte(0))
}

func isMapConvertibleToValue(v reflect.Value) bool {
	t := v.Type()
	return t.Kind() == reflect.Map && t.Key().Kind() == reflect.String
}