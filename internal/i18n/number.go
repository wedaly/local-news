package i18n

/*
#include <stdlib.h>
#include <stdio.h>

char* formatNumber(size_t sz, int val) {
	char* s = malloc(sz);
	if (s != NULL) {
		snprintf(s, sz, "%'d", val);
		s[sz - 1] = '\0';  // ensure NULL termination
	}
	return s;
}
*/
import "C"

import "unsafe"

const MaxDigits int = 127

// FormatNumber formats the integer to a numeric string
// according to the current locale.
func FormatNumber(val int) string {
	cstr := C.formatNumber(C.ulong(MaxDigits+1), C.int(val))
	if cstr == nil {
		return ""
	}
	result := C.GoString(cstr)
	C.free(unsafe.Pointer(cstr))
	return result
}
