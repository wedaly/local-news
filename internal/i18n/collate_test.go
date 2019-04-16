package i18n

import (
	"reflect"
	"sort"
	"testing"
)

func TestCollation(t *testing.T) {
	SetLocale("C")
	items := []string{
		"inconclusive",
		"van",
		"educated",
		"imagine",
		"elite",
		"eye",
		"rob",
		"wise",
		"uninterested",
		"envious",
		"super",
		"impossible",
	}

	sort.SliceStable(items, func(i, j int) bool {
		return CompareStrings(items[i], items[j])
	})

	if !sort.StringsAreSorted(items) {
		t.Errorf("Strings are not sorted")
	}
}

func TestCollationLocalizedEnglish(t *testing.T) {
	if err := SetLocale("en_US.UTF-8"); err != nil {
		t.Fatalf("Could not set locale")
	}

	items := []string{
		"ñandú",
		"ñú",
		"coté",
		"número",
		"côté",
		"cote",
		"Namibia",
		"côte",
	}

	sort.SliceStable(items, func(i, j int) bool {
		return CompareStrings(items[i], items[j])
	})

	// Matches output of `LC_ALL=en_US.UTF-8 sort input.txt`
	expected := []string{
		"cote",
		"coté",
		"côte",
		"côté",
		"Namibia",
		"ñandú",
		"ñú",
		"número",
	}

	if !reflect.DeepEqual(expected, items) {
		t.Errorf("Strings are not sorted, expected %v but got %v",
			expected, items)
	}
}

func TestCollationLocalizedSpanish(t *testing.T) {
	if err := SetLocale("es_ES.UTF-8"); err != nil {
		t.Fatalf("Could not set locale")
	}

	items := []string{
		"ñandú",
		"ñú",
		"coté",
		"número",
		"côté",
		"cote",
		"Namibia",
		"côte",
	}

	sort.SliceStable(items, func(i, j int) bool {
		return CompareStrings(items[i], items[j])
	})

	// Matches output of `LC_ALL=es_ES.UTF-8 sort input.txt`
	expected := []string{
		"cote",
		"coté",
		"côte",
		"côté",
		"Namibia",
		"número",
		"ñandú",
		"ñú",
	}

	if !reflect.DeepEqual(expected, items) {
		t.Errorf("Strings are not sorted, expected %v but got %v",
			expected, items)
	}
}
