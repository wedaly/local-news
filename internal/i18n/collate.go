package i18n

/*
#include <stdlib.h>
#include <string.h>
*/
import "C"

import "unsafe"

// CompareStrings compares two strings according to the locale-specific collation order.
// It returns true if and only if
// the first string is strictly less than the second string.
func CompareStrings(s1 string, s2 string) bool {
	cstr1 := C.CString(s1)
	cstr2 := C.CString(s2)
	result := C.strcoll(cstr1, cstr2)
	C.free(unsafe.Pointer(cstr1))
	C.free(unsafe.Pointer(cstr2))
	return result < 0
}
