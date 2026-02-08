package transliteration

import "testing"

func TestTransliterateKorean(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"페이커#KR1", "peikeo"},
		{"김치#KR1", "gimchi"},
		{"토르소#NA1", "toreuso"},
		{"꿈을꾸다#KR1", "kkumeulkkuda"},
	}
	for _, tt := range tests {
		got := Transliterate(tt.input)
		if got != tt.want {
			t.Errorf("Transliterate(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestTransliterateChinese(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"不知火舞#CN1", "buzhihuowu"},
		{"大魔王#TW1", "damowang"},
		{"人人人#NA1", "renrenren"},
	}
	for _, tt := range tests {
		got := Transliterate(tt.input)
		if got != tt.want {
			t.Errorf("Transliterate(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestTransliterateLatin(t *testing.T) {
	got := Transliterate("Faker#NA1")
	if got != "" {
		t.Errorf("Transliterate(Latin) = %q, want empty string", got)
	}
}

func TestTransliterateNoTag(t *testing.T) {
	got := Transliterate("페이커")
	if got != "peikeo" {
		t.Errorf("Transliterate(%q) = %q, want %q", "페이커", got, "peikeo")
	}
}
