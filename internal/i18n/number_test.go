package i18n

import (
	"strconv"
	"testing"
)

func TestFormatNumber(t *testing.T) {
	if err := SetLocale("C"); err != nil {
		t.Fatalf("Could not set locale")
	}

	// Each value has one more digit than the previous val
	// If MaxDigits is sufficiently large, val will wrap around due to
	// integer overflow, which is okay for this test.
	for d := 0; d < MaxDigits; d++ {
		val := d * 10
		result := FormatNumber(val)
		if result != strconv.Itoa(val) {
			t.Errorf("Incorrect value, expected %v but got %v", val, result)
		}
	}
}

func TestFormatNumberWithLocale(t *testing.T) {
	if err := SetLocale("de_DE.UTF-8"); err != nil {
		t.Fatalf("Could not set locale")
	}
	val := 123456789
	result := FormatNumber(val)
	expected := "123.456.789"
	if result != expected {
		t.Errorf(
			"Incorrect format for German locale, expected %v but got %v",
			expected, result)
	}
}
