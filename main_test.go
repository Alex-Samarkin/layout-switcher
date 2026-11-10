package main

import "testing"

func TestTransliterate(t *testing.T) {
	cases := map[string]string{
		"апрове": "fghjdt",
		"fghjdt": "апрове",
		"ghbdtn": "привет",
		"привет": "ghbdtn",
		",QETno": "бЙУЕтщ",
		"бЙУЕтщ": ",QETno",
		"ёЁ":     "`~",
	}
	for in, want := range cases {
		if got := transliterate(in); got != want {
			t.Errorf("transliterate(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestContainsCyrillic(t *testing.T) {
	cases := map[string]bool{
		"привет":    true,
		"ghbdtn":    false,
		"смесь mix": true,
		"12345":     false,
		"":          false,
	}
	for in, want := range cases {
		if got := containsCyrillic(in); got != want {
			t.Errorf("containsCyrillic(%q) = %v, want %v", in, got, want)
		}
	}
}
