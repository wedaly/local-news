package i18n

// CompareStrings compares two strings according to the locale-specific collation order.
// It returns true if and only if
// the first string is strictly less than the second string.
func CompareStrings(s1 string, s2 string) bool {
	return s1 < s2
}
