package utarray

import (
	"fmt"
	"reflect"
)

// Operator : Math Operator
type Operator int

const (
	// ADD : Add
	ADD Operator = 1 + iota

	// SUBTRACT : Subtract
	SUBTRACT

	// MULTIPLY : Multiply
	MULTIPLY

	// DIVIDE : Divide
	DIVIDE
)

// MatrixString to Create Metrix Array String
//
// **@Params:** [ `a`: array 1; `b`: array 2 ]
//
// **@Returns:** [ `$1`: array metrix ]
func MatrixString(a []string, b []string) []string {
	ai := []interface{}{}
	for _, v := range a {
		ai = append(ai, v)
	}

	bi := []interface{}{}
	for _, v := range b {
		bi = append(bi, v)
	}

	resi := MatrixDynamic(ai, bi, func(ax interface{}, bx interface{}) interface{} {
		return ax.(string) + bx.(string)
	})

	res := []string{}
	for _, v := range resi {
		res = append(res, v.(string))
	}

	return res
}

// MatrixInt to Create Metrix Array Int
//
// **@Params:** [ `a`: array 1; `b`: array2; `o`: operator ]
//
// **@Returns:** [ `$1`: array metrix ]
func MatrixInt(a []int, b []int, o Operator) []int {
	ai := []interface{}{}
	for _, v := range a {
		ai = append(ai, v)
	}

	bi := []interface{}{}
	for _, v := range b {
		bi = append(bi, v)
	}

	resi := MatrixDynamic(ai, bi, func(ax interface{}, bx interface{}) interface{} {
		c := 0
		switch o {
		case ADD:
			c = ax.(int) + bx.(int)

		case SUBTRACT:
			c = ax.(int) - bx.(int)

		case MULTIPLY:
			c = ax.(int) * bx.(int)

		case DIVIDE:
			c = ax.(int) / bx.(int)
		}
		return c
	})

	res := []int{}
	for _, v := range resi {
		res = append(res, v.(int))
	}

	return res
}

// MatrixInt64 to Create Metrix Array Int64
//
// **@Params:** [ `a`: array 1; `b`: array 2 ]
//
// **@Returns:** [ `$1`: array metrix ]
func MatrixInt64(a []int64, b []int64, o Operator) []int64 {
	ai := []interface{}{}
	for _, v := range a {
		ai = append(ai, v)
	}

	bi := []interface{}{}
	for _, v := range b {
		bi = append(bi, v)
	}

	resi := MatrixDynamic(ai, bi, func(ax interface{}, bx interface{}) interface{} {
		c := int64(0)
		switch o {
		case ADD:
			c = ax.(int64) + bx.(int64)

		case SUBTRACT:
			c = ax.(int64) - bx.(int64)

		case MULTIPLY:
			c = ax.(int64) * bx.(int64)

		case DIVIDE:
			c = ax.(int64) / bx.(int64)
		}
		return c
	})

	res := []int64{}
	for _, v := range resi {
		res = append(res, v.(int64))
	}

	return res
}

// MatrixDynamic to Dynamic Create Array Metrix
//
// **@Params:** [ `a`: array 1; `b`: array 2; `c`: resolver ]
//
// **@Returns:** [ `$1`: array metrix ]
func MatrixDynamic(a []interface{}, b []interface{}, f func(interface{}, interface{}) interface{}) []interface{} {
	res := []interface{}{}
	for _, v := range a {
		for _, v2 := range b {
			c := f(v, v2)
			res = append(res, c)
		}
	}
	return res
}

// IsExist is Exists in Array?
func IsExist(value interface{}, array interface{}) (exist bool) {
	exist = false
	switch reflect.TypeOf(array).Kind() {
	case reflect.Slice:
		s := reflect.ValueOf(array)

		for i := 0; i < s.Len(); i++ {
			if reflect.DeepEqual(value, s.Index(i).Interface()) {
				exist = true
				return exist
			}
		}
	}

	return exist
}

// IsExists Is Exists in Array?
func IsExists(value interface{}, array interface{}) (exists bool, index int) {
	exists, index = false, -1

	switch reflect.TypeOf(array).Kind() {
	case reflect.Slice:
		s := reflect.ValueOf(array)

		for i := 0; i < s.Len(); i++ {
			if reflect.DeepEqual(value, s.Index(i).Interface()) {
				index, exists = i, true
				return exists, index
			}
		}
	}

	return exists, index
}

// CheckAllowedLayer to check is allow? from multiple layer
func CheckAllowedLayer(value []string, layer [][]string) bool {
	for _, v := range layer {
		for _, v2 := range value {
			if IsExist(fmt.Sprintf("?%s", v2), v) {
				return false
			}
		}

		if IsExist("*", v) {
			return true
		}

		if len(v) == 1 && v[0] == "-" {
			return false
		}

		ok := false
		for _, v2 := range value {
			if IsExist(fmt.Sprintf("!%s", v2), v) {
				return true

			} else if len(v) <= 0 || IsExist(v2, v) || IsExist("@", v) {
				ok = true
				break
			}
		}

		if !ok {
			return false
		}
	}

	return true
}
