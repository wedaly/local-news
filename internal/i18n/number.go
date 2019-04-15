package i18n

/*
#include <stdlib.h>
#include <stdio.h>

int formatNumber(char *s, size_t slen, int val) {
	int n = snprintf(s, slen, "%'d", val);
	s[slen - 1] = '\0';  // ensure NULL termination
	if (n >= slen) {
		return slen - 1;
	} else {
		return n;
	}
}
*/
import "C"

import "unsafe"

const MaxDigits int = 127

// FormatNumber formats the integer to a numeric string
// according to the current locale.
func FormatNumber(val int) string {
	bufsize := C.ulong(MaxDigits + 1)
	buf := (*C.char)(C.malloc(bufsize))
	n := C.formatNumber(buf, bufsize, C.int(val))
	result := C.GoStringN(buf, n)
	C.free(unsafe.Pointer(buf))
	return result
}
