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
