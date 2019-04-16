package i18n

import "testing"

func setTranslationLocale(t *testing.T, locale string) {
	if err := SetLocale(locale); err != nil {
		t.Fatalf("Could not set locale")
	}
	InitTranslations("localnews", []string{"../../configs/locale"})
}

func TestGettextEnglish(t *testing.T) {
	setTranslationLocale(t, "en_US.UTF-8")
	result := Gettext("All Feeds")
	if result != "All Feeds" {
		t.Errorf("Could not get English translation")
	}
}

func TestGettextPseudo(t *testing.T) {
	setTranslationLocale(t, "eo")
	result := Gettext("All Feeds")
	if result != "[Ⱥłł Fɇɇđs łøɍɇm ɨᵽsᵾm]" {
		t.Errorf("Could not get pseudo language translation: %v", result)
	}
}

func TestNGettextEnglish(t *testing.T) {
	setTranslationLocale(t, "en_US.UTF-8")
	result := NGettext("Refreshing %v feed...", "Refreshing %v feeds...", 2)
	if result != "Refreshing %v feeds..." {
		t.Errorf("Could not get English translation: %v", result)
	}
}

func TestNGettextPseudo(t *testing.T) {
	setTranslationLocale(t, "eo")
	result := NGettext("Refreshing %v feed...", "Refreshing %v feeds...", 2)
	if result != "[Ɍɇfɍɇsħɨnǥ %v fɇɇđs... łøɍɇm ɨᵽsᵾm]" {
		t.Errorf("Could not get pseudo language translation: %v", result)
	}
}
