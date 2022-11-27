package utstring

import (
	"strings"

	"repo-scanner/internal/utils/utarray"
)

// MergeString to Merge two map string
//
// **@Params:** [ `a`: map 1; `b`: map 2 ]
func MergeString(a *map[string]string, b map[string]string) {
	for k, v := range b {
		if _, ok := b[k]; ok {
			(*a)[k] = v
		}
	}
}

// ArrContains function
//
// **@Params:** [ `a`: array; `i`: value ]
//
// **@Returns:** [ `$1`: exists flag ]
func ArrContains(a []string, i string) bool {
	for _, v := range a {
		if v == i {
			return true
		}
	}
	return false
}

// ArrUniqueString function
//
// **@Params:** [ `s`: string array ]
//
// **@Returns:** [ `$1`: unique string array ]
func ArrUniqueString(s []string) []string {
	keys := make(map[string]bool)
	var list []string
	for _, v := range s {
		if _, v2 := keys[v]; !v2 {
			keys[v] = true
			list = append(list, v)
		}
	}
	return list
}

// GeneratePattern to collecting string pattern
//
// **@Params:** [ `s`: start from; `l`: length; `a`: material ]
//
// **@Returns:** [ `$1`: string pattern ]
func GeneratePattern(s string, l int, a string) []string {
	return pGeneratePattern(false, s, l, a)
}
func pGeneratePattern(r bool, s string, l int, a string) []string {
	start := 0
	if !r && len(s) > 0 {
		x := ""
		x, s = Sub(s, 0, 1), Sub(s, 1, -1)
		start = strings.Index(a, x)
		if start < 0 {
			start = 0
		}
	}

	res := []string{}
	for i := start; i < len(a); i++ {
		tmp := []string{a[i : i+1]}
		if l > 1 {
			isReset := false
			if i > start {
				isReset = true
			}
			mrx := utarray.MatrixString(tmp, pGeneratePattern(isReset, s, l-1, a))
			res = append(res, mrx...)
		} else {
			res = append(res, tmp...)
		}
	}
	return res
}

// CleanSpit to split then trim each items
func CleanSpit(value string, sep string) []string {
	values := strings.Split(value, sep)
	for k, v := range values {
		values[k] = Trim(v)
	}
	return values
}
