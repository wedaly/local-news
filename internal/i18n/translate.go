package i18n

/*
#include <stdlib.h>
#include <libintl.h>
*/
import "C"

import (
	"os"
	"unsafe"
)

// MsgId identifies a user-facing string that can be translated.
type MsgId string

// InitTranslations prepares the application to load translation strings
// The `domain` uniquely identifies this application (see gettext documentation for details)
// Search paths is a list of paths containing locale information in priority order.
// If none of the search paths exist, the system default will be used instead.
func InitTranslations(domain string, searchPaths []string) {
	// Set the text domain (should equal the name of the basename of the ".mo" files)
	domainCStr := C.CString(domain)
	defer C.free(unsafe.Pointer(domainCStr))
	if C.textdomain(domainCStr) == nil {
		panic("Could not set text domain")
	}

	// Prefer "./configs/locale" to the system locale directory
	// if it exists.  This is useful for development so we can
	// test translations without installing them in /usr/share
	for _, dir := range searchPaths {
		if fileInfo, err := os.Stat(dir); err == nil && fileInfo.IsDir() {
			dirnameCStr := C.CString(dir)
			defer C.free(unsafe.Pointer(dirnameCStr))
			if C.bindtextdomain(domainCStr, dirnameCStr) == nil {
				panic("Could not bind text domain")
			}
			break
		}
	}
}

// Gettext translates a message using the currently configured locale
// If no translation is found, it returns the message ID untranslated.
func Gettext(msgId MsgId) string {
	msgIdCStr := C.CString(string(msgId))
	defer C.free(unsafe.Pointer(msgIdCStr))
	resultCStr := C.gettext(msgIdCStr)
	result := C.GoString(resultCStr)
	return result
}

// NGettext translates a message into either the singular or plural form
// using the currently configured locale.
func NGettext(singularMsgId MsgId, pluralMsgId MsgId, count int) string {
	singularMsgIdCStr := C.CString(string(singularMsgId))
	defer C.free(unsafe.Pointer(singularMsgIdCStr))

	pluralMsgIdCStr := C.CString(string(pluralMsgId))
	defer C.free(unsafe.Pointer(pluralMsgIdCStr))

	resultCStr := C.ngettext(singularMsgIdCStr, pluralMsgIdCStr, C.ulong(count))
	result := C.GoString(resultCStr)
	return result
}
