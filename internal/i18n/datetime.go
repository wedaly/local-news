package i18n

/*
#include <stdlib.h>
#include <stdio.h>
#include <langinfo.h>
#include <time.h>
#include <nl_types.h>

char* formatDatetime(char* fmt, long unixTs) {
	time_t t = (time_t)(unixTs);
	struct tm *tm = localtime(&t);
	size_t sz = 512;
	char* s = (char *)malloc(sz);
	strftime(s, sz, fmt, tm);
	s[sz - 1] = '\0';
	return s;
}
*/
import "C"

import (
	"time"
	"unsafe"
)

var dateFmt (*C.char)
var datetimeFmt (*C.char)

func InitDateFormats() {
	dateFmt = C.nl_langinfo(C.D_FMT)
	datetimeFmt = C.nl_langinfo(C.D_T_FMT)
}

// FormatDate converts the specified date (year, month, and day)
// according to the current locale
func FormatDate(t time.Time) string {
	return formatDatetime(dateFmt, t)
}

// FormatDatetime formats the specified datetime (date, hour, and minute)
// according to the current locale.
func FormatDatetime(t time.Time) string {
	return formatDatetime(datetimeFmt, t)
}

func formatDatetime(fmt *C.char, t time.Time) string {
	cstr := C.formatDatetime(fmt, C.long(t.Unix()))
	if cstr == nil {
		return ""
	}

	result := C.GoString(cstr)
	C.free(unsafe.Pointer(cstr))
	return result
}
