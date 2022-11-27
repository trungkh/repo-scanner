package utinterface

import (
	"encoding/json"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// IsNil is nil value ?
func IsNil(value interface{}) (res bool) {
	return (value == nil || (reflect.TypeOf(value).Kind() == reflect.Ptr && reflect.ValueOf(value).IsNil()))
}

// IsZero is zero value ?
func IsZero(value interface{}) (res bool) {
	res = true

	if !reflect.ValueOf(value).IsNil() {
		res = false
	}
	return
}

// ToString to casting interface to string
func ToString(value interface{}) (res string) {
	if !IsNil(value) {
		val := reflect.ValueOf(value)
		switch val.Kind() {
		case reflect.String:
			res = val.String()

		case reflect.Ptr:
			res = ToString(reflect.Indirect(val))

		default:
			switch valx := value.(type) {
			case []byte:
				res = string(valx)

			case time.Time:
				res = valx.Format(time.RFC3339Nano)

			default:
				byt, err := json.Marshal(value)
				if err == nil {
					res = string(byt)
				}
			}
		}
	}
	return
}

// Clone to cloning interface{}
func Clone(value interface{}) interface{} {
	val := reflect.ValueOf(value)
	typ := reflect.TypeOf(value)

	ptr := false
	if val.Kind() == reflect.Ptr {
		val = reflect.Indirect(val)
		typ = reflect.TypeOf(val.Interface())
		ptr = true
	}

	nval := reflect.New(typ)
	if ptr {
		nval.Elem().Set(reflect.ValueOf(val.Interface()))
	} else {
		nval = reflect.Indirect(nval)
		nval.Set(reflect.ValueOf(val.Interface()))
	}
	return nval.Interface()
}

// ToInt to casting interface to int64
func ToInt(value interface{}, def int64) int64 {
	r, err := strconv.ParseInt(ToString(value), 10, 64)
	if err != nil {
		r = def
	}
	return r
}

// ToFloat to casting interface to float64
func ToFloat(value interface{}, def float64) float64 {
	r, err := strconv.ParseFloat(ToString(value), 64)
	if err != nil {
		r = def
	}
	return r
}

// ToBool to casting interface to bool
func ToBool(value interface{}, def bool) bool {
	vx := reflect.ValueOf(value)
	switch vx.Kind() {
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int, reflect.Int64:
		switch vx.Int() {
		case 1:
			return true

		case 0:
			return false
		}

	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint, reflect.Uint64:
		switch vx.Uint() {
		case 1:
			return true

		case 0:
			return false
		}

	case reflect.Bool:
		return vx.Bool()

	default:
		switch strings.ToLower(ToString(value)) {
		case "true", "1":
			return true

		case "false", "0":
			return false
		}
	}

	return def
}
