package masterfile

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
)

type tokenType uint8

func (tt tokenType) String() string {
	switch tt {
	case tokenTypeIndent:
		return "indentation"

	case tokenTypeNewline:
		return "newline"

	case tokenTypeString:
		return "string"

	case tokenTypeEOF:
		return "eof"

	default:
		panic(fmt.Sprintf("unexpected tokenType: %d", tt))
	}
}

const (
	tokenTypeIndent tokenType = iota
	tokenTypeNewline
	tokenTypeString
	tokenTypeEOF
)

type token struct {
	tt    tokenType
	value string
}

type lexer struct {
	rd            *bufio.Reader
	isAtLineStart bool
	isInParens    bool
}

func newLexer(rd io.Reader) lexer {
	return lexer{
		rd:            bufio.NewReader(rd),
		isAtLineStart: true,
		isInParens:    false,
	}
}

func (l *lexer) next() (token, error) {
	skipped, err := l.skipWhitespace()
	if err != nil {
		return token{}, err
	}

	if l.isAtLineStart {
		l.isAtLineStart = false

		if skipped && !l.isInParens {
			return token{tt: tokenTypeIndent}, err
		}
	}

	ch, _, err := l.rd.ReadRune()
	if err != nil && !errors.Is(err, io.EOF) {
		return token{}, err
	}

	switch ch {
	case '(', ')':
		l.isInParens = ch == '('

		return l.next()

	case '\n':
		if l.isInParens {
			return l.next()
		}

		l.isAtLineStart = true

		return token{tt: tokenTypeNewline}, nil

	case ';':
		if err := l.skipComment(); err != nil {
			return token{}, err
		}

		return l.next()

	case 0:
		return token{tt: tokenTypeEOF}, nil

	default:
		if err := l.rd.UnreadRune(); err != nil {
			return token{}, err
		}

		str, err := l.readString()
		if err != nil {
			return token{}, err
		}

		return token{tt: tokenTypeString, value: str}, nil
	}
}

func (l *lexer) skipWhitespace() (bool, error) {
	skipped := false

	for {
		ch, _, err := l.rd.ReadRune()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return false, nil
			}

			return false, err
		}

		switch ch {
		case ' ', '\t', '\r':
			skipped = true

		default:
			return skipped, l.rd.UnreadRune()
		}
	}
}

func (l *lexer) skipComment() error {
	for {
		ch, _, err := l.rd.ReadRune()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}

			return err
		}

		if ch == '\n' {
			return l.rd.UnreadRune()
		}
	}
}

func (l *lexer) readString() (string, error) {
	ch, _, err := l.rd.ReadRune()
	if err != nil {
		return "", err
	}

	quoted := ch == '"'

	if !quoted {
		if err := l.rd.UnreadRune(); err != nil {
			return "", err
		}
	}

	var buf strings.Builder

	for {
		ch, _, err := l.rd.ReadRune()
		if err != nil {
			if !quoted && errors.Is(err, io.EOF) {
				break
			}

			return "", err
		}

		if ch == '\\' {
			_, _ = buf.WriteRune('\\')

			ch, _, err := l.rd.ReadRune()
			if err != nil {
				return "", err
			}

			if ch >= '0' && ch <= '9' {
				if err := l.rd.UnreadRune(); err != nil {
					return "", err
				}

				var digits [3]byte

				if _, err := io.ReadFull(l.rd, digits[:]); err != nil {
					return "", err
				}

				if _, err := strconv.ParseUint(string(digits[:]), 10, 8); err != nil {
					return "", err
				}

				_, _ = buf.Write(digits[:])
			} else {
				_, _ = buf.WriteRune(ch)
			}

			continue
		}

		if quoted && ch == '"' {
			break
		}

		if !quoted && (ch == ' ' || ch == ';' || ch == '(' || ch == ')' || ch == '\n') {
			if err := l.rd.UnreadRune(); err != nil {
				return "", err
			}

			break
		}

		_, _ = buf.WriteRune(ch)
	}

	return buf.String(), nil
}
