package i18n

import (
	"testing"
	"time"
)

func TestFormatDate(t *testing.T) {
	SetLocale("C")
	InitDateFormats()
	d := time.Date(2020, 1, 2, 3, 4, 5, 0, time.Local)
	result := FormatDate(d)
	expected := "01/02/20"
	if result != expected {
		t.Errorf("Wrong date, expected %v but got %v", expected, result)
	}
}

func TestFormatDateLocalized(t *testing.T) {
	SetLocale("de_DE")
	InitDateFormats()
	d := time.Date(2020, 1, 2, 3, 4, 5, 0, time.Local)
	result := FormatDate(d)
	expected := "02.01.2020"
	if result != expected {
		t.Errorf("Wrong date, expected %v but got %v", expected, result)
	}
}

func TestFormatDatetime(t *testing.T) {
	SetLocale("C")
	InitDateFormats()
	d := time.Date(2020, 1, 2, 3, 4, 5, 0, time.Local)
	result := FormatDatetime(d)
	expected := "Thu Jan  2 03:04:05 2020"
	if result != expected {
		t.Errorf("Wrong date, expected %v but got %v", expected, result)
	}
}

func TestFormatDatetimeLocalized(t *testing.T) {
	SetLocale("de_DE")
	InitDateFormats()
	d := time.Date(2020, 1, 2, 3, 4, 5, 0, time.Local)
	result := FormatDatetime(d)
	expected := "Do 02 Jan 2020 03:04:05 PST"
	if result != expected {
		t.Errorf("Wrong date, expected %v but got %v", expected, result)
	}
}
