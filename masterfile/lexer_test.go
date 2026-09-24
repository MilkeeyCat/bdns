package masterfile

import (
	"io"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLexer(t *testing.T) {
	tests := []struct {
		input  string
		tokens []token
		err    error
	}{
		{
			input:  `\1`,
			tokens: nil,
			err:    io.ErrUnexpectedEOF,
		},
		{
			input:  `\12`,
			tokens: nil,
			err:    io.ErrUnexpectedEOF,
		},
		{
			input:  `\123`,
			tokens: []token{{tokenTypeString, "\\123"}, {tokenTypeEOF, ""}},
			err:    nil,
		},
		{
			input:  `\256`,
			tokens: nil,
			err:    strconv.ErrRange,
		},
		{
			input:  `"not closed`,
			tokens: nil,
			err:    io.EOF,
		},
		{
			input:  `a\`,
			tokens: nil,
			err:    io.EOF,
		},
		{
			input:  `a\ b`,
			tokens: []token{{tokenTypeString, "a\\ b"}, {tokenTypeEOF, ""}},
			err:    nil,
		},
		{
			input:  `"a\"b"`,
			tokens: []token{{tokenTypeString, "a\\\"b"}, {tokenTypeEOF, ""}},
			err:    nil,
		},
		// yoinked from https://www.rfc-editor.org/info/rfc1035/#section-5.3
		{
			input: `
@   IN  SOA     VENERA      Action\.domains (
								 20     ; SERIAL
								 7200   ; REFRESH
								 600    ; RETRY
								 3600000; EXPIRE
								 60)    ; MINIMUM

		NS      A.ISI.EDU.
		NS      VENERA
		NS      VAXA
		MX      10      VENERA
		MX      20      VAXA

A       A       26.3.0.103

VENERA  A       10.1.0.52
		A       128.9.0.32

VAXA    A       10.2.0.27
		A       128.9.0.33


$INCLUDE <SUBSYS>ISI-MAILBOXES.TXT
`,
			tokens: []token{
				{tokenTypeNewline, ""},
				{tokenTypeString, "@"},
				{tokenTypeString, "IN"},
				{tokenTypeString, "SOA"},
				{tokenTypeString, "VENERA"},
				{tokenTypeString, "Action\\.domains"},
				{tokenTypeString, "20"},
				{tokenTypeString, "7200"},
				{tokenTypeString, "600"},
				{tokenTypeString, "3600000"},
				{tokenTypeString, "60"},
				{tokenTypeNewline, ""},
				{tokenTypeNewline, ""},
				{tokenTypeIndent, ""},
				{tokenTypeString, "NS"},
				{tokenTypeString, "A.ISI.EDU."},
				{tokenTypeNewline, ""},
				{tokenTypeIndent, ""},
				{tokenTypeString, "NS"},
				{tokenTypeString, "VENERA"},
				{tokenTypeNewline, ""},
				{tokenTypeIndent, ""},
				{tokenTypeString, "NS"},
				{tokenTypeString, "VAXA"},
				{tokenTypeNewline, ""},
				{tokenTypeIndent, ""},
				{tokenTypeString, "MX"},
				{tokenTypeString, "10"},
				{tokenTypeString, "VENERA"},
				{tokenTypeNewline, ""},
				{tokenTypeIndent, ""},
				{tokenTypeString, "MX"},
				{tokenTypeString, "20"},
				{tokenTypeString, "VAXA"},
				{tokenTypeNewline, ""},
				{tokenTypeNewline, ""},
				{tokenTypeString, "A"},
				{tokenTypeString, "A"},
				{tokenTypeString, "26.3.0.103"},
				{tokenTypeNewline, ""},
				{tokenTypeNewline, ""},
				{tokenTypeString, "VENERA"},
				{tokenTypeString, "A"},
				{tokenTypeString, "10.1.0.52"},
				{tokenTypeNewline, ""},
				{tokenTypeIndent, ""},
				{tokenTypeString, "A"},
				{tokenTypeString, "128.9.0.32"},
				{tokenTypeNewline, ""},
				{tokenTypeNewline, ""},
				{tokenTypeString, "VAXA"},
				{tokenTypeString, "A"},
				{tokenTypeString, "10.2.0.27"},
				{tokenTypeNewline, ""},
				{tokenTypeIndent, ""},
				{tokenTypeString, "A"},
				{tokenTypeString, "128.9.0.33"},
				{tokenTypeNewline, ""},
				{tokenTypeNewline, ""},
				{tokenTypeNewline, ""},
				{tokenTypeString, "$INCLUDE"},
				{tokenTypeString, "<SUBSYS>ISI-MAILBOXES.TXT"},
				{tokenTypeNewline, ""},
				{tokenTypeEOF, ""},
			},
			err: nil,
		},
		{
			input: `
MOE     MB      A.ISI.EDU.
LARRY   MB      A.ISI.EDU.
CURLEY  MB      A.ISI.EDU.
STOOGES MG      MOE
		MG      LARRY
		MG      CURLEY
`,
			tokens: []token{
				{tokenTypeNewline, ""},
				{tokenTypeString, "MOE"},
				{tokenTypeString, "MB"},
				{tokenTypeString, "A.ISI.EDU."},
				{tokenTypeNewline, ""},
				{tokenTypeString, "LARRY"},
				{tokenTypeString, "MB"},
				{tokenTypeString, "A.ISI.EDU."},
				{tokenTypeNewline, ""},
				{tokenTypeString, "CURLEY"},
				{tokenTypeString, "MB"},
				{tokenTypeString, "A.ISI.EDU."},
				{tokenTypeNewline, ""},
				{tokenTypeString, "STOOGES"},
				{tokenTypeString, "MG"},
				{tokenTypeString, "MOE"},
				{tokenTypeNewline, ""},
				{tokenTypeIndent, ""},
				{tokenTypeString, "MG"},
				{tokenTypeString, "LARRY"},
				{tokenTypeNewline, ""},
				{tokenTypeIndent, ""},
				{tokenTypeString, "MG"},
				{tokenTypeString, "CURLEY"},
				{tokenTypeNewline, ""},
				{tokenTypeEOF, ""},
			},
			err: nil,
		},
	}

	for _, tc := range tests {
		l := newLexer(strings.NewReader(tc.input))

		var (
			tokens []token
			err    error
		)

		for {
			var token token

			if token, err = l.next(); err != nil {
				break
			}

			tokens = append(tokens, token)

			if token.tt == tokenTypeEOF {
				break
			}
		}

		require.ErrorIs(t, err, tc.err)
		assert.Equal(t, tc.tokens, tokens)
	}
}
