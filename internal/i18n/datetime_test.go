package i18n

import (
	"testing"
	"time"
)

func initDateLocale(t *testing.T, locale string) {
	if err := SetLocale(locale); err != nil {
		t.Fatalf("Could not set locale to %v", locale)
	}
	InitDateFormats()
}

func TestFormatDate(t *testing.T) {
	initDateLocale(t, "C")
	d := time.Date(2020, 1, 2, 3, 4, 5, 0, time.Local)
	result := FormatDate(d)
	expected := "01/02/20"
	if result != expected {
		t.Errorf("Wrong date, expected %v but got %v", expected, result)
	}
}

func TestFormatDateLocalized(t *testing.T) {
	initDateLocale(t, "de_DE.UTF-8")
	d := time.Date(2020, 1, 2, 3, 4, 5, 0, time.Local)
	result := FormatDate(d)
	expected := "02.01.2020"
	if result != expected {
		t.Errorf("Wrong date, expected %v but got %v", expected, result)
	}
}

func TestFormatDatetime(t *testing.T) {
	initDateLocale(t, "C")
	d := time.Date(2020, 1, 2, 3, 4, 5, 0, time.Local)
	result := FormatDatetime(d)
	expected := "Thu Jan  2 03:04:05 2020"
	if result != expected {
		t.Errorf("Wrong date, expected %v but got %v", expected, result)
	}
}

func TestFormatDatetimeLocalized(t *testing.T) {
	initDateLocale(t, "de_DE.UTF-8")
	d := time.Date(2020, 1, 2, 3, 4, 5, 0, time.Local)
	result := FormatDatetime(d)
	expected := "Do 02 Jan 2020 03:04:05 PST"
	if result != expected {
		t.Errorf("Wrong date, expected %v but got %v", expected, result)
	}
}
