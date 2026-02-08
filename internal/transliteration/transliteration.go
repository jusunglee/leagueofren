package transliteration

import (
	"strings"
	"unicode"
)

// Transliterate converts a username (gameName#tag) to its romanized form.
// Only the gameName part is transliterated; the #tag is stripped.
// Returns empty string for Latin-only names.
func Transliterate(username string) string {
	gameName := username
	if idx := strings.IndexByte(username, '#'); idx >= 0 {
		gameName = username[:idx]
	}

	script := detectScript(gameName)
	switch script {
	case "korean":
		return romanizeKorean(gameName)
	case "chinese":
		return romanizeChinese(gameName)
	default:
		return ""
	}
}

func detectScript(text string) string {
	for _, r := range text {
		if unicode.Is(unicode.Hangul, r) {
			return "korean"
		}
	}
	for _, r := range text {
		if unicode.Is(unicode.Han, r) {
			return "chinese"
		}
	}
	return "latin"
}
