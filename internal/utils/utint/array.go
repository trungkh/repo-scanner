package utint

// ArrContainsInt function
//
// **@Params:** [ `a`: array; `i`: value ]
//
// **@Returns:** [ `$1`: exists flag ]
func ArrContainsInt(a []int, i int) bool {
	for _, v := range a {
		if v == i {
			return true
		}
	}
	return false
}

// ArrContainsInt8 function
//
// **@Params:** [ `a`: array; `i`: value ]
//
// **@Returns:** [ `$1`: exists flag ]
func ArrContainsInt8(a []int8, i int8) bool {
	for _, v := range a {
		if v == i {
			return true
		}
	}
	return false
}

// ArrContainsInt64 function
//
// **@Params:** [ `a`: array; `i`: value ]
//
// **@Returns:** [ `$1`: exists flag ]
func ArrContainsInt64(a []int64, i int64) bool {
	for _, v := range a {
		if v == i {
			return true
		}
	}
	return false
}
