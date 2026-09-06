package domain

import "strings"

type Label []byte

type Domain []Label

func (d Domain) String() string {
	if len(d) == 1 && len(d[0]) == 0 {
		return "."
	}

	var buf strings.Builder

	for i, label := range d {
		_, _ = buf.WriteString(string(label))

		if len(d)-1 > i {
			buf.WriteByte('.')
		}
	}

	return buf.String()
}
