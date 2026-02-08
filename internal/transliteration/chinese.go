package transliteration

import (
	"strings"
	"unicode"

	"github.com/mozillazg/go-pinyin"
)

var pinyinArgs pinyin.Args

func init() {
	pinyinArgs = pinyin.NewArgs()
	pinyinArgs.Style = pinyin.Normal // no tone marks
}

func romanizeChinese(text string) string {
	var b strings.Builder
	for _, r := range text {
		if unicode.Is(unicode.Han, r) {
			py := pinyin.SinglePinyin(r, pinyinArgs)
			if len(py) > 0 {
				b.WriteString(py[0])
			} else {
				b.WriteRune(r)
			}
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}
