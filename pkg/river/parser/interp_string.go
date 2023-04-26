package parser

import (
	"fmt"
	"unicode"
	"unicode/utf8"

	"github.com/grafana/agent/pkg/river/scanner"
	"github.com/grafana/agent/pkg/river/token"
)

// interpStringScanner is a specialized scanner for interpolated strings. It
// looks for sequences of "raw" text and sequences of "interpolated
// expressions."
//
// Interpolated expressions start with "${" and end with "}". A "}" inside the
// interpolated expression is permitted as long as it has a matching "{" to the
// left of it.
//
// interpStringScanner supports the same string escape sequences as normal
// strings.
type interpStringScanner struct {
	file       *token.File
	input      []byte
	err        scanner.ErrorHandler // Error reporting (may be nil)
	baseOffset int

	ch         rune // Current character
	offset     int  // Byte offset of ch
	readOffset int  // Byte offset of first character *after* ch
	numErrors  int  // Number of errors encountered during scanning.
}

// interpStringToken is a token for interpolated strings.
type interpStringToken int

const (
	interpStringTokenIllegal interpStringToken = iota
	interpStringTokenEOF

	interpStringTokenRaw  // Hello, world!
	interpStringTokenExpr // ${INTERPOLATED_EXPR}
)

const (
	bom = 0xFEFF // byte order mark, permitted as very first character
	eof = -1     // end of file
)

// newInterpStringScanner creates a new interpolated string scanner to tokenize
// the provided input config.
//
// input must be a subset of text in file containing an interpolated string.
// baseOff must be the offset of input within the original source for file.
// baseOff will be added to all returned position offsets from Scan.
//
// Calls to Scan will invoke the error handler eh when a lexical error is found
// if eh is not nil.
func newInterpStringScanner(file *token.File, input []byte, eh scanner.ErrorHandler, baseOff int) *interpStringScanner {
	s := &interpStringScanner{
		file:       file,
		input:      input,
		err:        eh,
		baseOffset: baseOff,
	}

	// Preload first character. Note that BOM are not handled since input never
	// corresponds to a physical file.
	s.next()
	return s
}

// peek gets the next byte after the current character without advancing the
// scanner. Returns 0 if the scanner is at EOF.
func (s *interpStringScanner) peek() byte {
	if s.readOffset < len(s.input) {
		return s.input[s.readOffset]
	}
	return 0
}

// next advances the scanner and reads the next Unicode character into s.ch.
// s.ch == eof indicates end of file.
func (s *interpStringScanner) next() {
	if s.readOffset >= len(s.input) {
		s.offset = len(s.input)
		if s.ch == '\n' {
			// Make sure we track final newlines at the end of the file
			s.file.AddLine(s.offset)
		}
		s.ch = eof
		return
	}

	s.offset = s.readOffset
	if s.ch == '\n' {
		s.file.AddLine(s.offset)
	}

	r, width := rune(s.input[s.readOffset]), 1
	switch {
	case r == 0:
		s.onError(s.offset, "illegal character NUL")
	case r >= utf8.RuneSelf:
		r, width = utf8.DecodeRune(s.input[s.readOffset:])
		if r == utf8.RuneError && width == 1 {
			s.onError(s.offset, "illegal UTF-8 encoding")
		} else if r == bom && s.offset > 0 {
			s.onError(s.offset, "illegal byte order mark")
		}
	}
	s.readOffset += width
	s.ch = r
}

func (s *interpStringScanner) onError(offset int, msg string) {
	if s.err != nil {
		s.err(s.file.Pos(offset), msg)
	}
	s.numErrors++
}

// Scan scans the next token and return the token's position, the token itself,
// and the token's literal string. The end of the input is indicated by
// [interpStringTokenEOF].
//
// If the returned token is interpStringTokenRaw, then lit conains the
// corresponding raw text from the interpolated string.
//
// If the returned token is interpStringTokenExpr, then lit contains the
// interpolated expression, including the surrounding "${" and "}" characters.
//
// It is possible for Scan to return multiple interpStringTokenRaw tokens in a
// row. The caller should combine these tokens as needed.
func (s *interpStringScanner) Scan() (pos token.Pos, tok interpStringToken, lit string) {
	// Start of current token.
	pos = s.file.Pos(s.baseOffset + s.offset)

	switch ch := s.ch; {
	case ch == eof:
		tok = interpStringTokenEOF
		return

	case ch == '$' && s.peek() == '{':
		s.next() // Consume $.
		s.next() // Consume {.

		if text, terminated := s.scanExpr(); terminated {
			tok = interpStringTokenExpr
			lit = text
		} else {
			tok = interpStringTokenRaw
			lit = text
		}

	default:
		tok = interpStringTokenRaw
		lit = s.scanRaw()
	}

	return
}

func (s *interpStringScanner) scanExpr() (fragment string, terminated bool) {
	// s.offset is first byte after ${; subtract 2 to get the start of $.
	off := s.offset - 2

	// curlyPairs tracks the number of pairs of {} we expect. We start at 1 to
	// count the first { from ${. We only stop scanning the expr once we've
	// consumed the last } from all the expected pairs.
	curlyPairs := 1

	for {
		ch := s.ch
		if ch == eof {
			break
		}

		s.next()

		if ch == '{' {
			curlyPairs++
		} else if ch == '}' {
			curlyPairs--
			if curlyPairs == 0 {
				terminated = true
				break
			}
		}
	}

	return string(s.input[off:s.offset]), terminated
}

func (s *interpStringScanner) scanRaw() string {
	off := s.offset

	for {
		ch := s.ch
		if ch == eof {
			break
		}

		if ch == '$' && s.peek() == '{' {
			break
		}

		s.next()
		if ch == '\\' {
			s.scanEscape()
		}
	}

	return string(s.input[off:s.offset])
}

// scanEscape parses an escape sequence. In case of a syntax error, scanEscape
// stops at the offending character without consuming it.
func (s *interpStringScanner) scanEscape() {
	off := s.offset

	var (
		n         int
		base, max uint32
	)

	switch s.ch {
	case 'a', 'b', 'f', 'n', 'r', 't', 'v', '\\', '"', '$':
		s.next()
		return
	case '0', '1', '2', '3', '4', '5', '6', '7':
		n, base, max = 3, 8, 255
	case 'x':
		s.next()
		n, base, max = 2, 16, 255
	case 'u':
		s.next()
		n, base, max = 4, 16, unicode.MaxRune
	case 'U':
		s.next()
		n, base, max = 8, 16, unicode.MaxRune
	default:
		msg := "unknown escape sequence"
		if s.ch == eof {
			msg = "escape sequence not terminated"
		}
		s.onError(off, msg)
		return
	}

	var x uint32
	for n > 0 {
		d := uint32(digitVal(s.ch))
		if d >= base {
			msg := fmt.Sprintf("illegal character %#U in escape sequence", s.ch)
			if s.ch == eof {
				msg = "escape sequence not terminated"
			}
			s.onError(off, msg)
			return
		}
		x = x*base + d
		s.next()
		n--
	}

	if x > max || x >= 0xD800 && x < 0xE000 {
		s.onError(off, "escape sequence is invalid Unicode code point")
	}
}

func digitVal(ch rune) int {
	switch {
	case ch >= '0' && ch <= '9':
		return int(ch - '0')
	case lower(ch) >= 'a' && lower(ch) <= 'f':
		return int(lower(ch) - 'a' + 10)
	}
	return 16 // Larger than any legal digit val
}

func lower(ch rune) rune { return ('a' - 'A') | ch }
