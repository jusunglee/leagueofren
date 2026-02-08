package transliteration

import (
	"strings"
	"unicode"
)

const (
	hangulBase = 0xAC00
	hangulEnd  = 0xD7A3
	jongN      = 28
	jungN      = 21
)

// Revised Romanization of Korean
var (
	choseong = []string{
		"g", "kk", "n", "d", "tt", "r", "m", "b", "pp",
		"s", "ss", "", "j", "jj", "ch", "k", "t", "p", "h",
	}
	jungseong = []string{
		"a", "ae", "ya", "yae", "eo", "e", "yeo", "ye", "o",
		"wa", "wae", "oe", "yo", "u", "wo", "we", "wi", "yu",
		"eu", "ui", "i",
	}
	jongseong = []string{
		"", "g", "kk", "gs", "n", "nj", "nh", "d", "l", "lg",
		"lm", "lb", "ls", "lt", "lp", "lh", "m", "b", "bs",
		"s", "ss", "ng", "j", "ch", "k", "t", "p", "h",
	}
)

func romanizeKorean(text string) string {
	var b strings.Builder
	for _, r := range text {
		if r >= hangulBase && r <= hangulEnd {
			code := int(r) - hangulBase
			jong := code % jongN
			jung := (code / jongN) % jungN
			cho := code / (jongN * jungN)
			b.WriteString(choseong[cho])
			b.WriteString(jungseong[jung])
			b.WriteString(jongseong[jong])
		} else if unicode.Is(unicode.Hangul, r) {
			// Jamo or compatibility jamo — skip gracefully
			b.WriteRune(r)
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}
