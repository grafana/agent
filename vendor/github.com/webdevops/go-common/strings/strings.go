package strings

import (
	"unicode"
	"unicode/utf8"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func Title(s string) string {
	return cases.Title(language.English, cases.NoLower).String(s)
}

func UppercaseFirst(s string) string {
	if len(s) > 0 {
		r, size := utf8.DecodeRuneInString(s)
		if r != utf8.RuneError || size > 1 {
			firstChar := unicode.ToUpper(r)
			if firstChar != r {
				s = string(firstChar) + s[size:]
			}
		}
	}
	return s
}
