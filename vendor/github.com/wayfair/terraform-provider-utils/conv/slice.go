// Package conv contains functions to gracefully handle conversions of
// types, such as asserting interface values to concrete types.
package conv

// InterfaceSliceToIntSlice converts a slice of interfaces to a slice of
// int.  If a value in the input slice cannot be asserted as an int, then
// the corresponding value in the output slice will be 0.
func InterfaceSliceToIntSlice(iArr []interface{}) []int {
	numInterfaces := len(iArr)
	if numInterfaces == 0 {
		return []int{}
	}
	intArr := make([]int, numInterfaces)
	for idx, val := range iArr {
		var ok bool
		if intArr[idx], ok = val.(int); !ok {
			intArr[idx] = 0
		}
	}
	return intArr
}

// InterfaceSliceToStringSlice converts a slice of interfaces to a slice of
// strings.  If a value in the input slice cannot be asserted as a string, then
// the corresponding value in the output slice will be the empty string.
func InterfaceSliceToStringSlice(iArr []interface{}) []string {
	iArrLen := len(iArr)
	strArr := make([]string, iArrLen)
	if iArrLen == 0 {
		return strArr
	}
	var ok bool
	for idx, val := range iArr {
		if strArr[idx], ok = val.(string); !ok {
			strArr[idx] = ""
		}
	}
	return strArr
}
