package i18n

/*
#include <stdlib.h>
#include <locale.h>

int setAllLocaleCategories(const char *locale) {
	return (setlocale(LC_ALL, locale) != NULL);
}

int setAllLocaleCategoriesFromEnv() {
	return setAllLocaleCategories("");
}

*/
import "C"

import (
	"errors"
	"unsafe"
)

// GetMessageLocaleFromEnv loads the current effective locale
// for user-facing messages.
func GetMessageLocaleFromEnv() string {
	locale := C.setlocale(C.LC_MESSAGES, nil)
	return C.GoString(locale)
}

// SetLocaleFromEnv sets the current locale based on environment variables
// such as LANG and LC_ALL.  See the setlocale man page for details.
// This should be called at program startup.
func SetLocaleFromEnv() error {
	result := C.setAllLocaleCategoriesFromEnv()
	if int(result) == 0 {
		return errors.New("Could not set locale")
	}
	return nil
}

// SetLocale sets the current locale explicitly.
// This is useful for testing.
func SetLocale(locale string) error {
	localeCStr := C.CString(locale)
	result := C.setAllLocaleCategories(localeCStr)
	C.free(unsafe.Pointer(localeCStr))
	if int(result) == 0 {
		return errors.New("Could not set locale")
	}
	return nil
}
