package masterfile

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCanonicalizeString(t *testing.T) {
	tests := []struct {
		input  string
		result string
		panics bool
	}{
		{
			input:  "foo\\032bar",
			result: "foo bar",
			panics: false,
		},
		{
			input:  "\\032",
			result: " ",
			panics: false,
		},
		{
			input:  "a\\.b",
			result: "a.b",
			panics: false,
		},
		{
			input:  "\\104\\105\\033",
			result: "hi!",
			panics: false,
		},
		{
			input:  "\\12",
			result: "",
			panics: true,
		},
		{
			input:  "\\12t",
			result: "",
			panics: true,
		},
		{
			input:  "blah\\",
			result: "",
			panics: true,
		},
	}

	for _, tc := range tests {
		if tc.panics {
			assert.Panics(t, func() {
				canonicalizeString(tc.input)
			})
		} else {
			assert.Equal(t, tc.result, canonicalizeString(tc.input))
		}
	}
}
