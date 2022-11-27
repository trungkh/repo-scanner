package utint

import (
	"reflect"
	"strconv"
	"strings"
)

// IsInteger : Is Integer ?
//
// **@Params:** [ `v`: string ]
//
// **@Returns:** [ `$1`: integer status ]
func IsInteger(v string) bool {
	if v == "" {
		return false
	}

	a := "1234567890"
	for _, v := range v {
		if !strings.Contains(a, string(v)) {
			return false
		}
	}

	return true
}

// StringToInt to casting string to integer
func StringToInt(value string, def int64) int64 {
	r, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		r = def
	}
	return r
}

// MinInt get minimal value
func MinInt(a int, b int) int {
	if a > b {
		return b
	}
	return a
}

// MaxInt get maximum value
func MaxInt(a int, b int) int {
	if a < b {
		return b
	}
	return a
}

// IsIntegerType to check is integer from type
func IsIntegerType(typ reflect.Type) bool {
	switch typ.Kind() {
	case reflect.Int:
		fallthrough
	case reflect.Int8:
		fallthrough
	case reflect.Int16:
		fallthrough
	case reflect.Int32:
		fallthrough
	case reflect.Int64:
		return true
	}

	return false
}
